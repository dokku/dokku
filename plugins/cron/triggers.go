package cron

import (
	"errors"
	"fmt"

	"github.com/dokku/dokku/plugins/common"
)

// TriggerCronGetProperty writes the cron key to stdout for a given app container
func TriggerCronGetProperty(appName string, key string) error {
	if key != "mailto" {
		return errors.New("Invalid cron property specified")
	}

	fmt.Println(common.PropertyGet("cron", appName, key))
	return nil
}

// TriggerInstall runs the install step for the cron plugin
func TriggerInstall() error {
	if err := common.PropertySetup("cron"); err != nil {
		return fmt.Errorf("Unable to install the cron plugin: %s", err.Error())
	}

	return nil
}

// TriggerPostAppCloneSetup creates new cron files
func TriggerPostAppCloneSetup(oldAppName string, newAppName string) error {
	err := common.PropertyClone("cron", oldAppName, newAppName)
	if err != nil {
		return err
	}

	return nil
}

// TriggerPostAppRenameSetup renames cron files
func TriggerPostAppRenameSetup(oldAppName string, newAppName string) error {
	if err := common.PropertyClone("cron", oldAppName, newAppName); err != nil {
		return err
	}

	if err := common.PropertyDestroy("cron", oldAppName); err != nil {
		return err
	}

	return nil
}

// TriggerPostDelete destroys the cron property for a given app container
func TriggerPostDelete(appName string) error {
	return common.PropertyDestroy("cron", appName)
}
