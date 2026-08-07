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

// TriggerAppJSONIsValid validates the cron tasks for a given app
func TriggerAppJSONIsValid(appName string, appJSONPath string) error {
	if !common.FileExists(appJSONPath) {
		return nil
	}

	appJSON, err := appjson.ReadAppJSON(appJSONPath)
	if err != nil {
		return err
	}

	_, err = FetchCronTasks(FetchCronTasksInput{
		AppName:       appName,
		AppJSON:       &appJSON,
		WarnToFailure: true,
	})
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

// TriggerPostDeploy regenerates the cron schedule for a given app after it is
// deployed. Dispatching scheduler-cron-write lets the cron plugin regenerate the
// host crontab for host-cron schedulers while self-managed schedulers update
// their own backends.
func TriggerPostDeploy(appName string) error {
	scheduler := common.GetAppScheduler(appName)
	_, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "scheduler-cron-write",
		Args:        []string{scheduler, appName},
		StreamStdio: true,
	})
	return err
}

// TriggerSchedulerCronWrite regenerates the host crontab when no scheduler is
// given (the letsencrypt "all apps" path) or when the given scheduler uses the
// host crontab. Self-managed schedulers implement their own scheduler-cron-write
// trigger and are no-ops here.
func TriggerSchedulerCronWrite(scheduler string, appName string) error {
	if scheduler != "" && !usesHostCron(scheduler) {
		return nil
	}

	return writeCronTab()
}

// TriggerSchedulerStop regenerates the host crontab for a host-cron app after
// its processes are stopped
func TriggerSchedulerStop(scheduler string, appName string, removeContainers string) error {
	if !usesHostCron(scheduler) {
		return nil
	}

	_, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "scheduler-cron-write",
		Args:        []string{scheduler, appName},
		StreamStdio: true,
	})
	return err
}
