package cron

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dokku/dokku/plugins/common"

	"github.com/ryanuber/columnize"
	"mvdan.cc/sh/v3/shell"
)

// CommandList lists all scheduled cron tasks for a given app
func CommandList(appName string, format string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if format == "" {
		format = "stdout"
	}

	if format != "stdout" && format != "json" {
		return fmt.Errorf("Invalid format specified, supported formats: json, stdout")
	}

	entries, err := FetchCronEntries(appName)
	if err != nil {
		return err
	}

	if format == "stdout" {
		output := []string{"ID | Schedule | Command"}
		for _, entry := range entries {
			output = append(output, fmt.Sprintf("%s | %s | %s", entry.ID, entry.Schedule, entry.Command))
		}

		result := columnize.SimpleFormat(output)
		fmt.Println(result)
		return nil
	}

	out, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	common.Log(string(out))

	return nil
}

// CommandReport displays a cron report for one or more apps
func CommandReport(appName string, format string, infoFlag string) error {
	if len(appName) == 0 {
		apps, err := common.DokkuApps()
		if err != nil {
			return err
		}
		for _, appName := range apps {
			if err := ReportSingleApp(appName, format, infoFlag); err != nil {
				return err
			}
		}
		return nil
	}

	return ReportSingleApp(appName, format, infoFlag)
}

// CommandRun executes a cron command on the fly
func CommandRun(appName string, cronID string, detached bool) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	entries, err := FetchCronEntries(appName)
	if err != nil {
		return err
	}

	if cronID == "" {
		return fmt.Errorf("Please specify a Cron ID from the output of 'dokku cron:list %s'", appName)
	}

	command := ""
	for _, entry := range entries {
		if entry.ID == cronID {
			command = entry.Command
		}
	}

	if command == "" {
		return fmt.Errorf("No matching Cron ID found. Please specify a Cron ID from the output of 'dokku cron:list %s'", appName)
	}

	fields, err := shell.Fields(command, func(name string) string {
		return ""
	})
	if err != nil {
		return fmt.Errorf("Could not parse command: %s", err)
	}

	if detached {
		os.Setenv("DOKKU_DETACH_CONTAINER", "1")
		os.Setenv("DOKKU_DISABLE_TTY", "true")
	}

	os.Setenv("DOKKU_CRON_ID", cronID)
	os.Setenv("DOKKU_RM_CONTAINER", "1")
	scheduler := common.GetAppScheduler(appName)
	args := append([]string{scheduler, appName, "0", ""}, fields...)
	return common.PlugnTrigger("scheduler-run", args...)
}

// CommandSet set or clear a cron property for an app
func CommandSet(appName string, property string, value string) error {
	if err := validateSetValue(appName, property, value); err != nil {
		return err
	}

	common.CommandPropertySet("cron", appName, property, value, DefaultProperties, GlobalProperties)
	return common.PlugnTrigger("cron-write")
}
