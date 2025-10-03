package cron

import (
	"fmt"
	"strings"

	appjson "github.com/dokku/dokku/plugins/app-json"
	"github.com/dokku/dokku/plugins/common"

	"github.com/multiformats/go-base36"
	cronparser "github.com/robfig/cron/v3"
)

var (
	// DefaultProperties is a map of all valid cron properties with corresponding default property values
	DefaultProperties = map[string]string{
		"mailfrom":    "",
		"mailto":      "",
		"maintenance": "false",
	}

	// GlobalProperties is a map of all valid global cron properties
	GlobalProperties = map[string]bool{
		"mailfrom":    true,
		"mailto":      true,
		"maintenance": true,
	}
)

// CronEntry is a struct that represents a cron command
type CronEntry struct {
	// ID is a unique identifier for the cron command
	ID string `json:"id"`

	// App is the app the cron command belongs to
	App string `json:"app,omitempty"`

	// Command is the command to run
	Command string `json:"command"`

	// Global is whether the cron command is global
	Global bool `json:"global,omitempty"`

	// Schedule is the cron schedule
	Schedule string `json:"schedule"`

	// ConcurrencyPolicy is the concurrency policy for the cron command
	ConcurrencyPolicy string `json:"concurrency_policy"`

	// AltCommand is an alternate command to run
	AltCommand string `json:"-"`

	// LogFile is the log file to write to
	LogFile string `json:"-"`

	// Maintenance is whether the cron command is in maintenance mode
	Maintenance bool `json:"maintenance"`
}

// DokkuRunCommand returns the command to run for a given cron command
func (t CronEntry) DokkuRunCommand() string {
	if t.AltCommand != "" {
		if t.LogFile != "" {
			return fmt.Sprintf("%s &>> %s", t.AltCommand, t.LogFile)
		}
		return t.AltCommand
	}

	return fmt.Sprintf("dokku run --concurrency-policy %s --cron-id %s %s %s", t.ConcurrencyPolicy, t.ID, t.App, t.Command)
}

// FetchCronEntriesInput is the input for the FetchCronEntries function
type FetchCronEntriesInput struct {
	AppName       string
	AppJSON       *appjson.AppJSON
	WarnToFailure bool
}

// FetchCronEntries returns a list of cron commands for a given app
func FetchCronEntries(input FetchCronEntriesInput) ([]CronEntry, error) {
	appName := input.AppName
	commands := []CronEntry{}
	isMaintenance := reportComputedMaintenance(appName) == "true"

	if input.AppJSON == nil {
		appJSON, err := appjson.GetAppJSON(appName)
		if err != nil {
			return commands, fmt.Errorf("Unable to fetch app.json for app %s: %s", appName, err.Error())
		}

		input.AppJSON = &appJSON
	}

	if input.AppJSON.Cron == nil {
		return commands, nil
	}

	for i, c := range input.AppJSON.Cron {
		if c.Command == "" {
			if input.WarnToFailure {
				return commands, fmt.Errorf("Missing cron command for app %s (index %d)", appName, i)
			}

			common.LogWarn(fmt.Sprintf("Missing cron command for app %s (index %d)", appName, i))
			continue
		}

		if c.Schedule == "" {
			if input.WarnToFailure {
				return commands, fmt.Errorf("Missing cron schedule for app %s (index %d)", appName, i)
			}

			common.LogWarn(fmt.Sprintf("Missing cron schedule for app %s (index %d)", appName, i))
			continue
		}

		parser := cronparser.NewParser(cronparser.Minute | cronparser.Hour | cronparser.Dom | cronparser.Month | cronparser.Dow | cronparser.Descriptor)
		_, err := parser.Parse(c.Schedule)
		if err != nil {
			return commands, fmt.Errorf("Invalid cron schedule for app %s (schedule %s): %s", appName, c.Schedule, err.Error())
		}

		if c.ConcurrencyPolicy == "" {
			c.ConcurrencyPolicy = "allow"
		}
		if c.ConcurrencyPolicy != "allow" && c.ConcurrencyPolicy != "forbid" && c.ConcurrencyPolicy != "replace" {
			return commands, fmt.Errorf("Invalid cron concurrency policy for app %s (schedule %s): %s", appName, c.Schedule, c.ConcurrencyPolicy)
		}

		commands = append(commands, CronEntry{
			App:               appName,
			Command:           c.Command,
			Schedule:          c.Schedule,
			ID:                GenerateCommandID(appName, c),
			Maintenance:       isMaintenance,
			ConcurrencyPolicy: c.ConcurrencyPolicy,
		})
	}

	return commands, nil
}

// FetchGlobalCronEntries returns a list of global cron commands
// This function should only be used for the cron:list --global command
// and not internally by the cron plugin
func FetchGlobalCronEntries() ([]CronEntry, error) {
	commands := []CronEntry{}
	response, _ := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "cron-entries",
		Args:    []string{"docker-local"},
	})
	for _, line := range strings.Split(response.StdoutContents(), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Split(line, ";")
		if len(parts) != 2 && len(parts) != 3 {
			common.LogWarn(fmt.Sprintf("Invalid injected cron task: %v", line))
			continue
		}

		id := base36.EncodeToStringLc([]byte(strings.Join(parts, ";;;")))
		command := CronEntry{
			ID:          id,
			Schedule:    parts[0],
			Command:     parts[1],
			AltCommand:  parts[1],
			Maintenance: false,
			Global:      true,
		}
		if len(parts) == 3 {
			command.LogFile = parts[2]
		}
		commands = append(commands, command)
	}
	return commands, nil
}

// GenerateCommandID creates a unique ID for a given app/command/schedule combination
func GenerateCommandID(appName string, c appjson.CronEntry) string {
	return base36.EncodeToStringLc([]byte(appName + "===" + c.Command + "===" + c.Schedule))
}
