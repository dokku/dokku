package apps

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/dokku/dokku/plugins/common"
)

// checks if an app exists
func appExists(appName string) error {
	return common.VerifyAppName(appName)
}

// checks if an app is locked
func appIsLocked(appName string) bool {
	lockPath := getLockPath(appName)
	_, err := os.Stat(lockPath)
	return !os.IsNotExist(err)
}

// verifies app name and creates an app
func createApp(appName string) error {
	if err := common.IsValidAppName(appName); err != nil {
		return err
	}

	if err := appExists(appName); err == nil {
		return errors.New("Name is already taken")
	}

	common.LogInfo1Quiet(fmt.Sprintf("Creating %s...", appName))
	os.MkdirAll(common.AppRoot(appName), 0755)

	if err := common.PropertyWrite("apps", appName, "created-at", fmt.Sprintf("%d", time.Now().Unix())); err != nil {
		return err
	}

	_, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "post-create",
		Args:        []string{appName},
		StreamStdio: true,
	})
	return err
}

// destroys an app
func destroyApp(appName string) error {
	if os.Getenv("DOKKU_APPS_FORCE_DELETE") != "1" {
		if err := common.AskForDestructiveConfirmation(appName, "app"); err != nil {
			return err
		}
	}

	common.LogInfo1(fmt.Sprintf("Destroying %s (including all add-ons)", appName))

	imageTag, _ := common.GetRunningImageTag(appName, "")
	_, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "pre-delete",
		Args:        []string{appName, imageTag},
		StreamStdio: true,
	})
	if err != nil {
		return err
	}

	scheduler := common.GetAppScheduler(appName)
	removeContainers := "true"
	_, err = common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "scheduler-stop",
		Args:        []string{scheduler, appName, removeContainers},
		StreamStdio: true,
	})
	if err != nil {
		return err
	}

	_, err = common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "scheduler-post-delete",
		Args:        []string{scheduler, appName, imageTag},
		StreamStdio: true,
	})
	if err != nil {
		return err
	}

	_, err = common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "post-delete",
		Args:        []string{appName, imageTag},
		StreamStdio: true,
	})
	if err != nil {
		return err
	}

	forceCleanup := true
	common.DockerCleanup(appName, forceCleanup)

	common.LogInfo1("Retiring old containers and images")
	_, err = common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "scheduler-retire",
		Args:        []string{scheduler, appName},
		StreamStdio: true,
	})
	if err != nil {
		return err
	}

	// remove contents for apps that are symlinks to other folders
	if err := os.RemoveAll(fmt.Sprintf("%v/", common.AppRoot(appName))); err != nil {
		common.LogWarn(err.Error())
	}

	// then remove the folder and/or the symlink
	if err := os.RemoveAll(common.AppRoot(appName)); err != nil {
		common.LogWarn(err.Error())
	}

	return nil
}

// returns the lock path
func getLockPath(appName string) string {
	return fmt.Sprintf("%v/.deploy.lock", common.GetAppDataDirectory("apps", appName))
}

// creates an app if allowed
func maybeCreateApp(appName string) error {
	if err := appExists(appName); err == nil {
		return nil
	}

	results, _ := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "config-get-global",
		Args:    []string{"DOKKU_DISABLE_APP_AUTOCREATION"},
	})
	disableAutocreate := results.StdoutContents()
	if disableAutocreate == "true" {
		common.LogWarn("App auto-creation disabled.")
		return fmt.Errorf("Re-enable app auto-creation or create an app with 'dokku apps:create %s'", appName)
	}

	return common.SuppressOutput(func() error {
		return createApp(appName)
	})
}
