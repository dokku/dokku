package cron

import (
	"errors"
	"fmt"

	appjson "github.com/dokku/dokku/plugins/app-json"
	"github.com/dokku/dokku/plugins/common"
)

// TriggerCronGetProperty writes the cron key to stdout for a given app container
func TriggerCronGetProperty(appName string, key string) error {
	validProperties := map[string]bool{
		"mailfrom":    true,
		"mailto":      true,
		"maintenance": true,
	}
	if !validProperties[key] {
		return errors.New("Invalid cron property specified")
	}

	value := common.PropertyGetDefault("cron", appName, key, DefaultProperties[key])
	fmt.Println(value)
	return nil
}

// TriggerAppJSONIsValid validates the cron entries for a given app
func TriggerAppJSONIsValid(appName string, appJSONPath string) error {
	if !common.FileExists(appJSONPath) {
		return nil
	}

	appJSON, err := appjson.ReadAppJSON(appJSONPath)
	if err != nil {
		return err
	}

	_, err = FetchCronEntries(FetchCronEntriesInput{AppName: appName, AppJSON: &appJSON})
	if err != nil {
		return err
	}

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
	scheduler := common.GetAppScheduler(appName)
	_, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "scheduler-cron-write",
		Args:        []string{scheduler, appName},
		StreamStdio: true,
	})
	if err != nil {
		return err
	}

	if err := common.PropertyDestroy("cron", appName); err != nil {
		return err
	}

	return nil
}

// TriggerSchedulerStop stops the scheduler for a given app container
func TriggerSchedulerStop(scheduler string, appName string, removeContainers string) error {
	if scheduler != "docker-local" {
		return nil
	}

	return nil
}
