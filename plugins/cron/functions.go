package cron

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	appjson "github.com/dokku/dokku/plugins/app-json"
	"github.com/dokku/dokku/plugins/common"

	cronparser "github.com/robfig/cron/v3"
)

type templateCommand struct {
	ID       string
	App      string
	Command  string
	Schedule string
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

func writeCronEntries() error {
	apps, err := common.DokkuApps()
	if err != nil {
		return nil
	}

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

	if len(commands) == 0 {
		if !common.FileExists("/var/spool/cron/crontabs/dokku") {
			return nil
		}

		command := common.NewShellCmd("crontab -r -u dokku")
		command.ShowOutput = false
		out, err := command.CombinedOutput()
		if err != nil {
			return fmt.Errorf("Unable to remove schedule file: %v", out)
		}
	}

	data := map[string]interface{}{
		"Commands": commands,
	}

	t, err := getCronTemplate()
	if err != nil {
		return err
	}

	tmpFile, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("dokku-%s-%s", common.MustGetEnv("DOKKU_PID"), "CopyFromImage"))
	if err != nil {
		return fmt.Errorf("Cannot create temporary schedule file: %v", err)
	}

	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	if err := t.Execute(tmpFile, data); err != nil {
		return fmt.Errorf("Unable to template out schedule file: %v", err)
	}

	command := common.NewShellCmd(fmt.Sprintf("crontab -u dokku %s", tmpFile.Name()))
	command.ShowOutput = false
	out, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Unable to update schedule file: %v", out)
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
	return base64.StdEncoding.EncodeToString([]byte(appName + "===" + c.Command + "===" + c.Schedule))
}
