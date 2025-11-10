package schedulerdockerlocal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/cron"
	"golang.org/x/sync/errgroup"

	base36 "github.com/multiformats/go-base36"
)

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

func generateCronTasks() ([]cron.CronTask, error) {
	apps, _ := common.UnfilteredDokkuApps()

	g := new(errgroup.Group)
	results := make(chan []cron.CronTask, len(apps)+1)
	for _, appName := range apps {
		appName := appName
		g.Go(func() error {
			scheduler := common.GetAppScheduler(appName)
			if scheduler != "docker-local" {
				results <- []cron.CronTask{}
				return nil
			}

			c, err := cron.FetchCronTasks(cron.FetchCronTasksInput{AppName: appName})
			if err != nil {
				results <- []cron.CronTask{}
				common.LogWarn(err.Error())
				return nil
			}

			results <- c
			return nil
		})
	}

	g.Go(func() error {
		tasks := []cron.CronTask{}
		response, _ := common.CallPlugnTrigger(common.PlugnTriggerInput{
			Trigger: "cron-entries",
			Args:    []string{"docker-local"},
		})
		for _, line := range strings.Split(response.StdoutContents(), "\n") {
			if strings.TrimSpace(line) == "" {
				results <- []cron.CronTask{}
				return nil
			}

			parts := strings.Split(line, ";")
			if len(parts) != 2 && len(parts) != 3 {
				results <- []cron.CronTask{}
				return fmt.Errorf("Invalid injected cron task: %v", line)
			}

			id := base36.EncodeToStringLc([]byte(strings.Join(parts, ";;;")))
			task := cron.CronTask{
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
		results <- tasks
		return nil
	})

	err := g.Wait()
	close(results)

	tasks := []cron.CronTask{}
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

func writeCronTab(scheduler string) error {
	// allow empty scheduler, which means all apps (used by letsencrypt)
	if scheduler != "docker-local" && scheduler != "" {
		return nil
	}

	tasks, err := generateCronTasks()
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		return deleteCrontab()
	}

	resultfromResults, _ := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "cron-get-property",
		Args:    []string{"--global", "mailfrom"},
	})
	mailfrom := resultfromResults.StdoutContents()

	mailtoResults, _ := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "cron-get-property",
		Args:    []string{"--global", "mailto"},
	})
	mailto := mailtoResults.StdoutContents()

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

func getCronTemplate() (*template.Template, error) {
	t := template.New("cron")

	templatePath := filepath.Join(common.MustGetEnv("PLUGIN_ENABLED_PATH"), "cron", "templates", "cron.tmpl")
	b, err := os.ReadFile(templatePath)
	if err != nil {
		return t, fmt.Errorf("Cannot read template file: %v", err)
	}

	s := strings.TrimSpace(string(b))
	return t.Parse(s)
}
