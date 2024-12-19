package cron

import (
	"fmt"

	appjson "github.com/dokku/dokku/plugins/app-json"

	"github.com/multiformats/go-base36"
	cronparser "github.com/robfig/cron/v3"
)

var (
	// DefaultProperties is a map of all valid cron properties with corresponding default property values
	DefaultProperties = map[string]string{
		"mailfrom": "",
		"mailto":   "",
	}

	// GlobalProperties is a map of all valid global cron properties
	GlobalProperties = map[string]bool{
		"mailfrom": true,
		"mailto":   true,
	}
)

// TemplateCommand is a struct that represents a cron command
type TemplateCommand struct {
	// ID is a unique identifier for the cron command
	ID string `json:"id"`

	// App is the app the cron command belongs to
	App string `json:"app"`

	// Command is the command to run
	Command string `json:"command"`

	// Schedule is the cron schedule
	Schedule string `json:"schedule"`

	// AltCommand is an alternate command to run
	AltCommand string `json:"-"`

	// LogFile is the log file to write to
	LogFile string `json:"-"`
}

// CronCommand returns the command to run for a given cron command
func (t TemplateCommand) CronCommand() string {
	if t.AltCommand != "" {
		if t.LogFile != "" {
			return fmt.Sprintf("%s &>> %s", t.AltCommand, t.LogFile)
		}
		return t.AltCommand
	}

	return fmt.Sprintf("dokku run --cron-id %s %s %s", t.ID, t.App, t.Command)
}

// FetchCronEntries returns a list of cron commands for a given app
func FetchCronEntries(appName string) ([]TemplateCommand, error) {
	commands := []TemplateCommand{}
	appJSON, err := appjson.GetAppJSON(appName)
	if err != nil {
		return commands, fmt.Errorf("Unable to fetch app.json for app %s: %s", appName, err.Error())
	}

	if appJSON.Cron == nil {
		return commands, nil
	}

	for _, c := range appJSON.Cron {
		parser := cronparser.NewParser(cronparser.Minute | cronparser.Hour | cronparser.Dom | cronparser.Month | cronparser.Dow | cronparser.Descriptor)
		_, err := parser.Parse(c.Schedule)
		if err != nil {
			return commands, fmt.Errorf("Invalid cron schedule %s: %s", c.Schedule, err.Error())
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
