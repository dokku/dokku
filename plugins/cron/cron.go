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

// CronTask is a struct that represents a cron task
type CronTask struct {
	// ID is a unique identifier for the cron task
	ID string `json:"id"`

	// App is the app the cron task belongs to
	App string `json:"app,omitempty"`

	// Command is the command to run
	Command string `json:"command"`

	// Global is whether the cron task is global
	Global bool `json:"global,omitempty"`

	// Schedule is the cron schedule
	Schedule string `json:"schedule"`

	// AltCommand is an alternate command to run
	AltCommand string `json:"-"`

	// LogFile is the log file to write to
	LogFile string `json:"-"`

	// Maintenance is whether the cron task is in maintenance mode
	Maintenance bool `json:"maintenance"`
}

// DokkuRunCommand returns the dokku run command to execute for a given cron task
func (t CronTask) DokkuRunCommand() string {
	if t.AltCommand != "" {
		if t.LogFile != "" {
			return fmt.Sprintf("%s &>> %s", t.AltCommand, t.LogFile)
		}
		return t.AltCommand
	}

	return fmt.Sprintf("dokku run --cron-id %s %s %s", t.ID, t.App, t.Command)
}

// FetchCronTasksInput is the input for the FetchCronTasks function
type FetchCronTasksInput struct {
	AppName       string
	AppJSON       *appjson.AppJSON
	WarnToFailure bool
}

// FetchCronTasks returns a list of cron tasks for a given app
func FetchCronTasks(input FetchCronTasksInput) ([]CronTask, error) {
	appName := input.AppName
	tasks := []CronTask{}
	isMaintenance := reportComputedMaintenance(appName) == "true"

	if input.AppJSON == nil && input.AppName == "" {
		return commands, fmt.Errorf("Missing app name or app.json")
	}

	if input.AppJSON == nil {
		appJSON, err := appjson.GetAppJSON(appName)
		if err != nil {
			return tasks, fmt.Errorf("Unable to fetch app.json for app %s: %s", appName, err.Error())
		}

		input.AppJSON = &appJSON
	}

	if input.AppJSON.Cron == nil {
		return tasks, nil
	}

	for i, c := range input.AppJSON.Cron {
		if c.Command == "" {
			if input.WarnToFailure {
				return tasks, fmt.Errorf("Missing cron task command for app %s (index %d)", appName, i)
			}

			common.LogWarn(fmt.Sprintf("Missing cron task command for app %s (index %d)", appName, i))
			continue
		}

		if c.Schedule == "" {
			if input.WarnToFailure {
				return tasks, fmt.Errorf("Missing cron schedule for app %s (index %d)", appName, i)
			}

			common.LogWarn(fmt.Sprintf("Missing cron schedule for app %s (index %d)", appName, i))
			continue
		}

		parser := cronparser.NewParser(cronparser.Minute | cronparser.Hour | cronparser.Dom | cronparser.Month | cronparser.Dow | cronparser.Descriptor)
		_, err := parser.Parse(c.Schedule)
		if err != nil {
			return tasks, fmt.Errorf("Invalid cron schedule for app %s (schedule %s): %s", appName, c.Schedule, err.Error())
		}

		tasks = append(tasks, CronTask{
			App:         appName,
			Command:     c.Command,
			Schedule:    c.Schedule,
			ID:          GenerateCommandID(appName, c),
			Maintenance: isMaintenance,
		})
	}

	return tasks, nil
}

// FetchGlobalCronTasks returns a list of global cron tasks
// This function should only be used for the cron:list --global command
// and not internally by the cron plugin
func FetchGlobalCronTasks() ([]CronTask, error) {
	tasks := []CronTask{}
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
		task := CronTask{
			ID:          id,
			Schedule:    parts[0],
			Command:     parts[1],
			AltCommand:  parts[1],
			Maintenance: false,
			Global:      true,
		}
		if len(parts) == 3 {
			task.LogFile = parts[2]
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// GenerateCommandID creates a unique ID for a given app/command/schedule combination
func GenerateCommandID(appName string, c appjson.CronTask) string {
	return base36.EncodeToStringLc([]byte(appName + "===" + c.Command + "===" + c.Schedule))
}
