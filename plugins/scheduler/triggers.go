package scheduler

import (
	"fmt"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
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

	apps, err := common.DokkuApps()
	if err != nil {
		return nil
	}

	globalScheduler := config.GetWithDefault("--scheduler", "DOKKU_SCHEDULER", "")
	if globalScheduler != "" {
		common.LogVerboseQuiet(fmt.Sprintf("Setting scheduler property 'selected' to %v", globalScheduler))
		if err := common.PropertyWrite("scheduler", "--global", "selected", globalScheduler); err != nil {
			return err
		}

		if err := config.UnsetMany("--global", []string{"DOKKU_SCHEDULER"}, false); err != nil {
			common.LogWarn(err.Error())
		}
	}

	for _, appName := range apps {
		scheduler := config.GetWithDefault(appName, "DOKKU_SCHEDULER", "")
		if scheduler == "" {
			continue
		}

		common.LogVerboseQuiet(fmt.Sprintf("Setting scheduler property 'selected' to %v", scheduler))
		if err := common.PropertyWrite("scheduler", appName, "selected", scheduler); err != nil {
			return err
		}

		if err := config.UnsetMany(appName, []string{"DOKKU_SCHEDULER"}, false); err != nil {
			common.LogWarn(err.Error())
		}
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
