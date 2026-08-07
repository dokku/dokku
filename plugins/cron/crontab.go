package cron

import (
	_ "embed"
	"fmt"
	"os"
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

// injectedCronTasks parses the tasks injected via the cron-entries trigger for a
// given scheduler. Each entry is newline delimited in the form
// $SCHEDULE;$COMMAND[;$LOGFILE].
func injectedCronTasks(scheduler string) ([]CronTask, error) {
	tasks := []CronTask{}
	response, _ := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "cron-entries",
		Args:    []string{scheduler},
	})
	for _, line := range strings.Split(response.StdoutContents(), "\n") {
		if strings.TrimSpace(line) == "" {
			return []CronTask{}, nil
		}

		parts := strings.Split(line, ";")
		if len(parts) != 2 && len(parts) != 3 {
			return []CronTask{}, fmt.Errorf("Invalid injected cron task: %v", line)
		}

		id := base36.EncodeToStringLc([]byte(strings.Join(parts, ";;;")))
		task := CronTask{
			ID:          id,
			Schedule:    parts[0],
			AltCommand:  parts[1],
			Maintenance: false,
		}
		if len(parts) == 3 {
			task.LogFile = parts[2]
		}
		tasks = append(tasks, task)
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

	data := map[string]interface{}{
		"Tasks":    tasks,
		"Mailfrom": mailfrom,
		"Mailto":   mailto,
	}

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
