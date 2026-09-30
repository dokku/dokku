package cron

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"text/template"

	"github.com/dokku/dokku/plugins/common"
	"golang.org/x/sync/errgroup"

	base36 "github.com/multiformats/go-base36"
)

//go:embed templates/cron.tmpl
var cronTemplate string

// usesHostCron reports whether the given scheduler writes its cron tasks to the
// host crontab (as opposed to managing its own cron backend). An empty scheduler
// or an unimplemented scheduler-uses-host-cron trigger is treated as false.
func usesHostCron(scheduler string) bool {
	if scheduler == "" {
		return false
	}

	results, _ := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "scheduler-uses-host-cron",
		Args:    []string{scheduler},
	})
	return results.StdoutContents() == "true"
}

// hostCronSchedulers returns a map keyed by every distinct scheduler seen across
// the given apps plus the global scheduler, with a boolean value indicating
// whether that scheduler uses the host crontab. Deduplicating up front avoids
// redundant scheduler-uses-host-cron dispatches and any shared-cache race across
// concurrent task collection.
func hostCronSchedulers(appSchedulers []string) map[string]bool {
	schedulers := map[string]bool{}
	for _, scheduler := range append(appSchedulers, common.GetGlobalScheduler()) {
		if scheduler == "" {
			continue
		}
		if _, ok := schedulers[scheduler]; ok {
			continue
		}
		schedulers[scheduler] = usesHostCron(scheduler)
	}
	return schedulers
}

// CronEntryFormat is passed to the cron-entries trigger to signal that each
// entry may be printed as a json object on its own line. Implementations that
// do not know it keep printing $SCHEDULE;$COMMAND[;$LOGFILE] lines, which are
// still read.
const CronEntryFormat = "json"

// injectedCronEntry is a task printed as a json line by the cron-entries trigger
type injectedCronEntry struct {
	// Schedule is the cron schedule
	Schedule string `json:"schedule"`

	// Command is the command to run
	Command string `json:"command"`

	// LogFile is the log file the command's output is appended to
	LogFile string `json:"log-file"`

	// Mailto is the MAILTO value cron uses for the task instead of the global one
	Mailto string `json:"mailto"`
}

// parseInjectedCronEntry parses a single line printed by the cron-entries
// trigger, either a json object or in the form $SCHEDULE;$COMMAND[;$LOGFILE].
// The fields the task id is generated from are returned alongside the task.
func parseInjectedCronEntry(line string) (injectedCronEntry, []string, bool) {
	if strings.HasPrefix(strings.TrimSpace(line), "{") {
		entry := injectedCronEntry{}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return injectedCronEntry{}, nil, false
		}
		if entry.Schedule == "" || entry.Command == "" {
			return injectedCronEntry{}, nil, false
		}

		// json can encode a newline, which would add a line to the crontab
		fields := []string{entry.Schedule, entry.Command, entry.LogFile, entry.Mailto}
		for _, field := range fields {
			if strings.ContainsAny(field, "\r\n") {
				return injectedCronEntry{}, nil, false
			}
		}

		fields = []string{entry.Schedule, entry.Command}
		if entry.LogFile != "" {
			fields = append(fields, entry.LogFile)
		}
		if entry.Mailto != "" {
			fields = append(fields, entry.Mailto)
		}
		return entry, fields, true
	}

	parts := strings.Split(line, ";")
	if len(parts) != 2 && len(parts) != 3 {
		return injectedCronEntry{}, nil, false
	}

	entry := injectedCronEntry{
		Schedule: parts[0],
		Command:  parts[1],
	}
	if len(parts) == 3 {
		entry.LogFile = parts[2]
	}
	return entry, parts, true
}

// parseInjectedCronEntries parses the newline delimited output of the
// cron-entries trigger, returning the tasks and any lines that are not valid
// entries
func parseInjectedCronEntries(output string) ([]CronTask, []string) {
	tasks := []CronTask{}
	invalid := []string{}
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}

		entry, fields, ok := parseInjectedCronEntry(line)
		if !ok {
			invalid = append(invalid, line)
			continue
		}

		tasks = append(tasks, CronTask{
			ID:          base36.EncodeToStringLc([]byte(strings.Join(fields, ";;;"))),
			Schedule:    entry.Schedule,
			Command:     entry.Command,
			AltCommand:  entry.Command,
			LogFile:     entry.LogFile,
			Mailto:      entry.Mailto,
			Maintenance: false,
		})
	}
	return tasks, invalid
}

// injectedCronTasks parses the tasks injected via the cron-entries trigger for a
// given scheduler
func injectedCronTasks(scheduler string) ([]CronTask, error) {
	response, _ := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "cron-entries",
		Args:    []string{scheduler, CronEntryFormat},
	})

	tasks, invalid := parseInjectedCronEntries(response.StdoutContents())
	if len(invalid) > 0 {
		return []CronTask{}, fmt.Errorf("Invalid injected cron task: %v", invalid[0])
	}
	return tasks, nil
}

// generateCronTasks returns all cron tasks that should be written to the host
// crontab: the app.json cron tasks for every app whose scheduler uses the host
// crontab, plus any tasks injected via the cron-entries trigger for each such
// scheduler. Tasks in maintenance are omitted.
func generateCronTasks() ([]CronTask, error) {
	apps, _ := common.UnfilteredDokkuApps()

	appSchedulers := make([]string, len(apps))
	sg := new(errgroup.Group)
	for i, appName := range apps {
		i := i
		appName := appName
		sg.Go(func() error {
			appSchedulers[i] = common.GetAppScheduler(appName)
			return nil
		})
	}
	if err := sg.Wait(); err != nil {
		return []CronTask{}, err
	}

	hostCron := hostCronSchedulers(appSchedulers)

	g := new(errgroup.Group)
	results := make(chan []CronTask, len(apps)+len(hostCron))

	for i, appName := range apps {
		if !hostCron[appSchedulers[i]] {
			continue
		}
		appName := appName
		g.Go(func() error {
			c, err := FetchCronTasks(FetchCronTasksInput{AppName: appName})
			if err != nil {
				results <- []CronTask{}
				common.LogWarn(err.Error())
				return nil
			}

			results <- c
			return nil
		})
	}

	for scheduler, isHostCron := range hostCron {
		if !isHostCron {
			continue
		}
		scheduler := scheduler
		g.Go(func() error {
			tasks, err := injectedCronTasks(scheduler)
			if err != nil {
				results <- []CronTask{}
				return err
			}

			results <- tasks
			return nil
		})
	}

	err := g.Wait()
	close(results)

	tasks := []CronTask{}
	if err != nil {
		return tasks, err
	}

	for result := range results {
		for _, task := range result {
			if !task.Maintenance {
				tasks = append(tasks, task)
			}
		}
	}

	return tasks, nil
}

// writeCronTab regenerates the dokku user crontab from every host-cron app. It
// is always a full regeneration, so there is no clobbering when multiple
// schedulers use the host crontab.
func writeCronTab() error {
	tasks, err := generateCronTasks()
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		return deleteCrontab()
	}

	mailfrom := common.PropertyGetDefault("cron", "--global", "mailfrom", DefaultProperties["mailfrom"])
	mailto := common.PropertyGetDefault("cron", "--global", "mailto", DefaultProperties["mailto"])
	data := cronTemplateData(tasks, mailfrom, mailto)

	t, err := getCronTemplate()
	if err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("dokku-%s-%s", common.MustGetEnv("DOKKU_PID"), "WriteCronTab"))
	if err != nil {
		return fmt.Errorf("Cannot create temporary schedule file: %v", err)
	}

	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	if err := t.Execute(tmpFile, data); err != nil {
		return fmt.Errorf("Unable to template out schedule file: %v", err)
	}

	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command: "crontab",
		Args:    []string{"-u", "dokku", tmpFile.Name()},
	})
	if err != nil {
		return fmt.Errorf("Unable to update schedule file: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Unable to update schedule file: %s", result.StderrContents())
	}

	common.LogInfo1("Updated schedule file")

	return nil
}

// cronMailtoGroup is the set of tasks written to the crontab under a MAILTO of
// their own
type cronMailtoGroup struct {
	// Mailto is the MAILTO value the tasks are written under
	Mailto string

	// Tasks are the tasks written under the MAILTO value
	Tasks []CronTask
}

// cronTemplateData builds the data the cron template is executed with. Tasks
// that set their own MAILTO are grouped by it and written after every other
// task, so the global MAILTO applies to all of the tasks before them.
func cronTemplateData(tasks []CronTask, mailfrom string, mailto string) map[string]interface{} {
	globalTasks := []CronTask{}
	groups := map[string][]CronTask{}
	for _, task := range tasks {
		if task.Mailto == "" {
			globalTasks = append(globalTasks, task)
			continue
		}
		groups[task.Mailto] = append(groups[task.Mailto], task)
	}

	mailtoGroups := []cronMailtoGroup{}
	for _, groupMailto := range slices.Sorted(maps.Keys(groups)) {
		mailtoGroups = append(mailtoGroups, cronMailtoGroup{
			Mailto: groupMailto,
			Tasks:  groups[groupMailto],
		})
	}

	return map[string]interface{}{
		"Tasks":        globalTasks,
		"MailtoGroups": mailtoGroups,
		"Mailfrom":     mailfrom,
		"Mailto":       mailto,
	}
}

// deleteCrontab removes the dokku user crontab
func deleteCrontab() error {
	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command: "crontab",
		Args:    []string{"-l", "-u", "dokku"},
	})
	if err != nil || result.ExitCode != 0 {
		return nil
	}

	result, err = common.CallExecCommand(common.ExecCommandInput{
		Command: "crontab",
		Args:    []string{"-r", "-u", "dokku"},
	})
	if err != nil {
		return fmt.Errorf("Unable to remove schedule file: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Unable to remove schedule file: %s", result.StderrContents())
	}

	common.LogInfo1("Removed")
	return nil
}

// getCronTemplate parses the embedded cron template
func getCronTemplate() (*template.Template, error) {
	t := template.New("cron")
	s := strings.TrimSpace(cronTemplate)
	return t.Parse(s)
}
