package cron

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	appjson "github.com/dokku/dokku/plugins/app-json"
	"github.com/dokku/dokku/plugins/common"

	base36 "github.com/multiformats/go-base36"
	cronparser "github.com/robfig/cron/v3"
)

type templateCommand struct {
	ID         string
	App        string
	Command    string
	Schedule   string
	AltCommand string
	LogFile    string
}

func (t templateCommand) CronCommand() string {
	if t.AltCommand != "" {
		if t.LogFile != "" {
			return fmt.Sprintf("%s &>> %s", t.AltCommand, t.LogFile)
		}
		return t.AltCommand
	}

	return fmt.Sprintf("dokku run --cron-id %s %s %s", t.ID, t.App, t.Command)
}

func fetchCronEntries(appName string) ([]templateCommand, error) {
	commands := []templateCommand{}
	scheduler := common.GetAppScheduler(appName)
	if scheduler != "docker-local" {
		return commands, nil
	}

	appjsonPath := appjson.GetAppjsonPath(appName)
	if !common.FileExists(appjsonPath) {
		return commands, nil
	}

	b, err := ioutil.ReadFile(appjsonPath)
	if err != nil {
		return commands, fmt.Errorf("Cannot read app.json file for %s: %v", appName, err)
	}

	if strings.TrimSpace(string(b)) == "" {
		return commands, nil
	}

	var appJSON appjson.AppJSON
	if err = json.Unmarshal(b, &appJSON); err != nil {
		return commands, fmt.Errorf("Cannot parse app.json for %s: %v", appName, err)
	}

	for _, c := range appJSON.Cron {
		parser := cronparser.NewParser(cronparser.Minute | cronparser.Hour | cronparser.Dom | cronparser.Month | cronparser.Dow | cronparser.Descriptor)
		_, err := parser.Parse(c.Schedule)
		if err != nil {
			return commands, err
		}

		commands = append(commands, templateCommand{
			App:      appName,
			Command:  c.Command,
			Schedule: c.Schedule,
			ID:       generateCommandID(appName, c),
		})
	}

	return commands, nil
}

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

func writeCronEntries() error {
	apps, _ := common.UnfilteredDokkuApps()
	commands := []templateCommand{}
	for _, appName := range apps {
		scheduler := common.GetAppScheduler(appName)
		if scheduler != "docker-local" {
			continue
		}

		c, err := fetchCronEntries(appName)
		if err != nil {
			return err
		}

		commands = append(commands, c...)
	}

	b, _ := common.PlugnTriggerOutput("cron-entries", "docker-local")
	for _, line := range strings.Split(strings.TrimSpace(string(b[:])), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Split(line, ";")
		if len(parts) != 2 && len(parts) != 3 {
			return fmt.Errorf("Invalid injected cron task: %v", line)
		}

		id := base36.EncodeToStringLc([]byte(strings.Join(parts, ";;;")))
		command := templateCommand{
			ID:         id,
			Schedule:   parts[0],
			AltCommand: parts[1],
		}
		if len(parts) == 3 {
			command.LogFile = parts[2]
		}
		commands = append(commands, command)
	}

	if len(commands) == 0 {
		return deleteCrontab()
	}

	data := map[string]interface{}{
		"Commands": commands,
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

func generateCommandID(appName string, c appjson.CronCommand) string {
	return base36.EncodeToStringLc([]byte(appName + "===" + c.Command + "===" + c.Schedule))
}
