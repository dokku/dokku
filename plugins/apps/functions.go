package apps

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/dokku/dokku/plugins/common"
)

// checks if an app exists
func appExists(appName string) error {
	return common.VerifyAppName(appName)
}

// checks if an app directory already exists, regardless of app visibility
func appNameTaken(appName string) bool {
	return common.DirectoryExists(common.AppRoot(appName))
}

// checks if an app is locked via apps:lock or by an in-progress deploy
func appIsLocked(appName string) bool {
	return appIsManuallyLocked(appName) || deployLockHeld(appName)
}

// verifies app name and creates an app
func createApp(appName string) error {
	if err := common.IsValidAppName(appName); err != nil {
		return err
	}

	if appNameTaken(appName) {
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

// returns the path of the lock created by apps:lock
func getManualLockPath(appName string) string {
	return fmt.Sprintf("%v/.manual-deploy.lock", common.GetAppDataDirectory("apps", appName))
}

// checks if an app has been locked via apps:lock
func appIsManuallyLocked(appName string) bool {
	_, err := os.Stat(getManualLockPath(appName))
	return err == nil
}

// checks if a deploy currently holds the flock on the deploy lock file
//
// a shared lock is attempted and immediately released. Deploys taking the
// lock exclusively wait briefly before failing, so this check does not make
// a concurrent deploy fail.
func deployLockHeld(appName string) bool {
	lockFile, err := os.Open(getLockPath(appName))
	if err != nil {
		return false
	}
	defer lockFile.Close()

	fd := int(lockFile.Fd())
	if err := syscall.Flock(fd, syscall.LOCK_SH|syscall.LOCK_NB); err != nil {
		return errors.Is(err, syscall.EWOULDBLOCK)
	}

	_ = syscall.Flock(fd, syscall.LOCK_UN)
	return false
}

// creates an app if allowed
func maybeCreateApp(appName string) error {
	if err := appExists(appName); err == nil {
		return nil
	}

	disableAutocreate := common.PropertyGetDefault("apps", "--global", "disable-autocreation", "")
	if disableAutocreate == "true" {
		common.LogWarn("App auto-creation disabled.")
		return fmt.Errorf("Re-enable app auto-creation or create an app with 'dokku apps:create %s'", appName)
	}

	return common.SuppressOutput(func() error {
		return createApp(appName)
	})
}
