package cron

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	appjson "github.com/dokku/dokku/plugins/app-json"
	"github.com/dokku/dokku/plugins/common"

	"github.com/multiformats/go-base36"
	cronparser "github.com/robfig/cron/v3"
)

var (
	// DefaultProperties is a map of all valid network properties with corresponding default property values
	DefaultProperties = map[string]string{
		"mailto": "",
	}

	// GlobalProperties is a map of all valid global network properties
	GlobalProperties = map[string]bool{
		"mailto": true,
	}
)

type TemplateCommand struct {
	ID         string `json:"id"`
	App        string `json:"app"`
	Command    string `json:"command"`
	Schedule   string `json:"schedule"`
	AltCommand string `json:"-"`
	LogFile    string `json:"-"`
}

func (t TemplateCommand) CronCommand() string {
	if t.AltCommand != "" {
		if t.LogFile != "" {
			return fmt.Sprintf("%s &>> %s", t.AltCommand, t.LogFile)
		}
		return t.AltCommand
	}

	return fmt.Sprintf("dokku run --cron-id %s %s %s", t.ID, t.App, t.Command)
}

func FetchCronEntries(appName string) ([]TemplateCommand, error) {
	commands := []TemplateCommand{}
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

		commands = append(commands, TemplateCommand{
			App:      appName,
			Command:  c.Command,
			Schedule: c.Schedule,
			ID:       GenerateCommandID(appName, c),
		})
	}

	return commands, nil
}

// GenerateCommandID creates a unique ID for a given app/command/schedule combination
func GenerateCommandID(appName string, c appjson.CronCommand) string {
	return base36.EncodeToStringLc([]byte(appName + "===" + c.Command + "===" + c.Schedule))
}
