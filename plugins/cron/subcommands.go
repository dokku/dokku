package cron

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"

	"github.com/ryanuber/columnize"
	"mvdan.cc/sh/v3/shell"
)

// CommandList lists all scheduled cron tasks for a given app
func CommandList(appName string, format string) error {
	if format == "" {
		format = "stdout"
	}

	if format != "stdout" && format != "json" {
		return fmt.Errorf("Invalid format specified, supported formats: json, stdout")
	}

	var entries []TemplateCommand
	if appName == "--global" {
		var err error
		entries, err = FetchGlobalCronEntries()
		if err != nil {
			return err
		}
	} else {
		var err error
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
		entries, err = FetchCronEntries(FetchCronEntriesInput{AppName: appName})
		if err != nil {
			return err
		}
	}

	if format == "stdout" {
		output := []string{"ID | Schedule | Maintenance | Command"}
		for _, entry := range entries {
			maintenance := "false"
			if entry.Maintenance {
				if entry.TaskInMaintenance {
					maintenance = "true (task)"
				} else if entry.AppInMaintenance {
					maintenance = "true (app)"
				}
			}
			output = append(output, fmt.Sprintf("%s | %s | %s | %s", entry.ID, entry.Schedule, maintenance, entry.Command))
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
			if errors.Is(err, common.NoAppsExist) {
				common.LogWarn(err.Error())
				return nil
			}
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

// CommandResume resumes a cron command
func CommandResume(appName string, cronID string) error {
	return CommandSet(appName, fmt.Sprintf("%s%s", MaintenancePropertyPrefix, cronID), "")
}

// CommandRun executes a cron command on the fly
func CommandRun(appName string, cronID string, detached bool) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	entries, err := FetchCronEntries(FetchCronEntriesInput{AppName: appName})
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
	args := append([]string{scheduler, appName, "0", "--"}, fields...)
	_, err = common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "scheduler-run",
		Args:        args,
		StreamStdio: true,
	})
	return err
}

// CommandSet set or clear a cron property for an app
func CommandSet(appName string, property string, value string) error {
	if err := validateSetValue(appName, property, value); err != nil {
		return err
	}

	validProperties := DefaultProperties
	globalProperties := GlobalProperties
	if strings.HasPrefix(property, MaintenancePropertyPrefix) {
		if appName == "--global" {
			return fmt.Errorf("Task maintenance properties cannot be set globally")
		}

		maintenanceParts := strings.SplitN(property, ".", 2)
		if len(maintenanceParts) != 3 {
			return fmt.Errorf("Invalid task maintenance property, expected format: %s$ID: %s", MaintenancePropertyPrefix, property)
		}

		if maintenanceParts[1] == "" {
			return fmt.Errorf("Invalid task maintenance property, missing ID")
		}

		entries, err := FetchCronEntries(FetchCronEntriesInput{AppName: appName})
		if err != nil {
			return err
		}

		maintenanceID := maintenanceParts[1]
		for _, entry := range entries {
			if entry.ID == maintenanceID {
				validProperties[property] = ""
				globalProperties[property] = false
				break
			}
		}

		if _, ok := validProperties[property]; !ok {
			return fmt.Errorf("Invalid task maintenance property, no matching task ID found: %s", property)
		}
	}

	common.CommandPropertySet("cron", appName, property, value, validProperties, globalProperties)
	scheduler := common.GetAppScheduler(appName)
	_, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "scheduler-cron-write",
		Args:        []string{scheduler, appName},
		StreamStdio: true,
	})
	return err
}

// CommandSuspend suspends a cron command
func CommandSuspend(appName string, cronID string) error {
	return CommandSet(appName, fmt.Sprintf("%s%s", MaintenancePropertyPrefix, cronID), "true")
}
