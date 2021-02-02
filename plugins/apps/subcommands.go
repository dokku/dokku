package apps

import (
	"errors"
	"fmt"
	"os"

	"github.com/dokku/dokku/plugins/common"
)

// CommandClone clones an app
func CommandClone(oldAppName string, newAppName string, skipDeploy bool, ignoreExisting bool) error {
	if oldAppName == "" {
		return errors.New("Please specify an app to run the command on")
	}

	if newAppName == "" {
		return errors.New("Please specify an new app name")
	}

	if err := common.VerifyAppName(oldAppName); err != nil {
		return err
	}

	if err := common.IsValidAppName(newAppName); err != nil {
		return err
	}

	if err := appExists(newAppName); err == nil {
		if ignoreExisting {
			common.LogWarn("Name is already taken")
			return nil
		}

		return errors.New("Name is already taken")
	}

	common.LogInfo1Quiet(fmt.Sprintf("Cloning %s to %s", oldAppName, newAppName))
	if err := createApp(newAppName); err != nil {
		return err
	}

	if err := common.PlugnTrigger("post-app-clone-setup", []string{oldAppName, newAppName}...); err != nil {
		return err
	}

	if skipDeploy {
		os.Setenv("SKIP_REBUILD", "true")
	}

	if err := common.PlugnTrigger("post-app-clone", []string{oldAppName, newAppName}...); err != nil {
		return err
	}

	return nil
}

// CommandCreate creates app via command line
func CommandCreate(appName string) error {
	if err := common.IsValidAppName(appName); err != nil {
		return err
	}

	return createApp(appName)
}

// CommandDestroy destroys an app
func CommandDestroy(appName string, force bool) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if force {
		os.Setenv("DOKKU_APPS_FORCE_DELETE", "1")
	}

	return destroyApp(appName)
}

// CommandExists checks if an app exists
func CommandExists(appName string) error {
	return appExists(appName)
}

// CommandList lists all apps
func CommandList() error {
	common.LogInfo2Quiet("My Apps")
	apps, err := common.DokkuApps()
	if err != nil {
		common.LogWarn(err.Error())
		return nil
	}

	for _, appName := range apps {
		common.Log(appName)
	}

	return nil
}

// CommandLock locks an app for deployment
func CommandLock(appName string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	lockfilePath := fmt.Sprintf("%v/.deploy.lock", common.AppRoot(appName))
	if _, err := os.Create(lockfilePath); err != nil {
		return errors.New("Unable to create deploy lock")
	}

	common.LogInfo1("Deploy lock created")
	return nil
}

// CommandLocked checks if an app is locked for deployment
func CommandLocked(appName string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if appIsLocked(appName) {
		common.LogQuiet("Deploy lock exists")
		return nil

	}
	return errors.New("Deploy lock does not exist")
}

// CommandRename renames an app
func CommandRename(oldAppName string, newAppName string, skipDeploy bool) error {
	if oldAppName == "" {
		return errors.New("Please specify an app to run the command on")
	}

	if newAppName == "" {
		return errors.New("Please specify an new app name")
	}

	if err := common.VerifyAppName(oldAppName); err != nil {
		return err
	}

	if err := common.IsValidAppName(newAppName); err != nil {
		return err
	}

	if err := appExists(newAppName); err == nil {
		return errors.New("Name is already taken")
	}

	common.LogInfo1Quiet(fmt.Sprintf("Renaming %s to %s", oldAppName, newAppName))
	if err := createApp(newAppName); err != nil {
		return err
	}

	if err := common.PlugnTrigger("post-app-rename-setup", []string{oldAppName, newAppName}...); err != nil {
		return err
	}

	os.Setenv("DOKKU_APPS_FORCE_DELETE", "1")
	if err := destroyApp(oldAppName); err != nil {
		return err
	}

	if skipDeploy {
		os.Setenv("SKIP_REBUILD", "true")
	}

	if err := common.PlugnTrigger("post-app-rename", []string{oldAppName, newAppName}...); err != nil {
		return err
	}

	return nil
}

// CommandReport displays an app report for one or more apps
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

// CommandUnlock unlocks an app for deployment
func CommandUnlock(appName string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	lockfilePath := fmt.Sprintf("%v/.deploy.lock", common.AppRoot(appName))
	if _, err := os.Stat(lockfilePath); !os.IsNotExist(err) {
		common.LogWarn("A deploy may be in progress.")
		common.LogWarn("Removing the app lock will not stop in progress deploys.")
	}

	if err := os.Remove(lockfilePath); err != nil {
		return errors.New("Unable to remove deploy lock")
	}

	common.LogInfo1("Deploy lock removed")
	return nil
}
