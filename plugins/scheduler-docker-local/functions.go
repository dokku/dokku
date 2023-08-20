package schedulerdockerlocal

import (
	"fmt"
	"io/ioutil"
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
	command := common.NewShellCmd("sudo /usr/bin/crontab -l -u dokku")
	command.ShowOutput = false
	if !command.Execute() {
		return nil
	}

	command = common.NewShellCmd("sudo /usr/bin/crontab -r -u dokku")
	command.ShowOutput = false
	out, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Unable to remove schedule file: %v", string(out))
	}

	common.LogInfo1("Removed")
	return nil
}

func generateCronEntries() ([]cron.TemplateCommand, error) {
	apps, _ := common.UnfilteredDokkuApps()

	g := new(errgroup.Group)
	results := make(chan []cron.TemplateCommand, len(apps)+1)
	for _, appName := range apps {
		appName := appName
		g.Go(func() error {
			scheduler := common.GetAppScheduler(appName)
			if scheduler != "docker-local" {
				results <- []cron.TemplateCommand{}
				return nil
			}

			c, err := cron.FetchCronEntries(appName)
			if err != nil {
				results <- []cron.TemplateCommand{}
				return err
			}

			results <- c
			return nil
		})
	}

	g.Go(func() error {
		commands := []cron.TemplateCommand{}
		b, _ := common.PlugnTriggerOutput("cron-entries", "docker-local")
		for _, line := range strings.Split(strings.TrimSpace(string(b[:])), "\n") {
			if strings.TrimSpace(line) == "" {
				results <- []cron.TemplateCommand{}
				return nil
			}

			parts := strings.Split(line, ";")
			if len(parts) != 2 && len(parts) != 3 {
				results <- []cron.TemplateCommand{}
				return fmt.Errorf("Invalid injected cron task: %v", line)
			}

			id := base36.EncodeToStringLc([]byte(strings.Join(parts, ";;;")))
			command := cron.TemplateCommand{
				ID:         id,
				Schedule:   parts[0],
				AltCommand: parts[1],
			}
			if len(parts) == 3 {
				command.LogFile = parts[2]
			}
			commands = append(commands, command)
		}
		results <- commands
		return nil
	})

	err := g.Wait()
	close(results)

	commands := []cron.TemplateCommand{}
	if err != nil {
		return commands, err
	}

	for result := range results {
		c := result
		if len(c) > 0 {
			commands = append(commands, c...)
		}
	}

	return commands, nil
}

func writeCronEntries() error {
	commands, err := generateCronEntries()
	if err != nil {
		return err
	}

	if len(commands) == 0 {
		return deleteCrontab()
	}

	mailto, _ := common.PlugnTriggerOutputAsString("cron-get-property", []string{"--global", "mailto"}...)

	data := map[string]interface{}{
		"Commands": commands,
		"Mailto":   mailto,
	}

	t, err := getCronTemplate()
	if err != nil {
		return err
	}

	tmpFile, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("dokku-%s-%s", common.MustGetEnv("DOKKU_PID"), "WriteCronEntries"))
	if err != nil {
		return fmt.Errorf("Cannot create temporary schedule file: %v", err)
	}

	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	if err := t.Execute(tmpFile, data); err != nil {
		return fmt.Errorf("Unable to template out schedule file: %v", err)
	}

	command := common.NewShellCmd(fmt.Sprintf("sudo /usr/bin/crontab -u dokku %s", tmpFile.Name()))
	command.ShowOutput = false
	out, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Unable to update schedule file: %s", string(out))
	}

	common.LogInfo1("Updated schedule file")

	return nil
}

func getCronTemplate() (*template.Template, error) {
	t := template.New("cron")

	templatePath := filepath.Join(common.MustGetEnv("PLUGIN_ENABLED_PATH"), "cron", "templates", "cron.tmpl")
	b, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return t, fmt.Errorf("Cannot read template file: %v", err)
	}

	s := strings.TrimSpace(string(b))
	return t.Parse(s)
}
