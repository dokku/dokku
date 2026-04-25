package scheduler

import (
	"fmt"

	"github.com/dokku/dokku/plugins/common"
)

// TriggerSchedulerDetect outputs a manually selected scheduler for the app
func TriggerSchedulerDetect(appName string) error {
	if appName != "--global" {
		if scheduler := common.PropertyGet("scheduler", appName, "selected"); scheduler != "" {
			fmt.Println(scheduler)
			return nil
		}
	}

	if scheduler := common.PropertyGet("scheduler", "--global", "selected"); scheduler != "" {
		fmt.Println(scheduler)
		return nil
	}

	fmt.Println("docker-local")
	return nil
}

// TriggerInstall runs the install step for the scheduler plugin
func TriggerInstall() error {
	if err := common.PropertySetup("scheduler"); err != nil {
		return fmt.Errorf("Unable to install the scheduler plugin: %s", err.Error())
	}

	if err := common.MigrateConfigToProperties("scheduler", []common.MigrateConfigEntry{
		{
			ConfigVar:       "DOKKU_SCHEDULER",
			GlobalConfigVar: "DOKKU_SCHEDULER",
			Property:        "selected",
		},
		{
			ConfigVar:       "DOKKU_APP_SHELL",
			GlobalConfigVar: "DOKKU_APP_SHELL",
			Property:        "shell",
		},
	}); err != nil {
		return err
	}

	return nil
}

// TriggerPostAppCloneSetup creates new scheduler files
func TriggerPostAppCloneSetup(oldAppName string, newAppName string) error {
	err := common.PropertyClone("scheduler", oldAppName, newAppName)
	if err != nil {
		return err
	}

	return nil
}

// TriggerPostAppRenameSetup renames scheduler files
func TriggerPostAppRenameSetup(oldAppName string, newAppName string) error {
	if err := common.PropertyClone("scheduler", oldAppName, newAppName); err != nil {
		return err
	}

	if err := common.PropertyDestroy("scheduler", oldAppName); err != nil {
		return err
	}

	return nil
}

// TriggerPostDelete destroys the scheduler property for a given app container
func TriggerPostDelete(appName string) error {
	return common.PropertyDestroy("scheduler", appName)
}
