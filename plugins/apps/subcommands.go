package apps

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// CommandClone clones an app
func CommandClone(args []string) error {
	appName, err := getAppName(args)
	if err != nil {
		return err
	}

	fmt.Println(appName)
	return nil
}

// CommandCreate creates app via command line
func CommandCreate(args []string) error {
	appName, err := getAppName(args)
	if err != nil {
		return err
	}

	return createApp(appName)
}

// CommandDestroy destroys an app
func CommandDestroy(args []string) error {
	appName, err := getAppName(args)
	if err != nil {
		return err
	}

	if appName == "tls" {
		return errors.New("Unable to destroy tls directory")
	}

	if len(args) >= 2 {
		force := args[1]
		if force == "force" {
			os.Setenv("DOKKU_APPS_FORCE_DELETE", "1")
		}
	}

	return destroyApp(appName)
}

// CommandExists checks if an app exists
func CommandExists(args []string) error {
	appName, err := getAppName(args)
	if err != nil {
		return err
	}

	return appExists(appName)
}

// CommandList lists all apps
func CommandList(args []string) error {
	common.LogInfo2Quiet("My Apps")
	apps, err := common.DokkuApps()
	if err != nil {
		return err
	}

	for _, appName := range apps {
		common.Log(appName)
	}

	return nil
}

// CommandLock locks an app for deployment
func CommandLock(args []string) error {
	appName, err := getAppName(args)
	if err != nil {
		return err
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	lockfilePath := fmt.Sprintf("%v/.deploy.lock", common.AppRoot(appName))
	if _, err = os.Create(lockfilePath); err != nil {
		return errors.New("Unable to create deploy lock")
	}

	common.LogInfo1("Deploy lock created")
	return nil
}

// CommandLocked checks if an app is locked for deployment
func CommandLocked(args []string) error {
	appName, err := getAppName(args)
	if err != nil {
		return err
	}

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
func CommandRename(args []string) error {
	appName, err := getAppName(args)
	if err != nil {
		return err
	}

	fmt.Println(appName)
	return nil
}

// CommandReport displays an app report for one or more apps
func CommandReport(args []string) error {
	appName, err := getAppName(args)
	if err != nil {
		return err
	}

	infoFlag := ""
	if len(args) > 1 {
		infoFlag = args[1]
	}

	if strings.HasPrefix(appName, "--") {
		infoFlag = appName
		appName = ""
	}

	if len(appName) == 0 {
		apps, err := common.DokkuApps()
		if err != nil {
			return err
		}
		for _, appName := range apps {
			if err := ReportSingleApp(appName, infoFlag); err != nil {
				return err
			}
		}
		return nil
	}

	return ReportSingleApp(appName, infoFlag)
}

// CommandUnlock unlocks an app for deployment
func CommandUnlock(args []string) error {
	appName, err := getAppName(args)
	if err != nil {
		return err
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	lockfilePath := fmt.Sprintf("%v/.deploy.lock", common.AppRoot(appName))
	_, err = os.Stat(lockfilePath)
	if !os.IsNotExist(err) {
		common.LogWarn("A deploy may be in progress.")
		common.LogWarn("Removing the app lock will not stop in progress deploys.")
	}

	err = os.Remove(lockfilePath)
	if err == nil {
		common.LogInfo1("Deploy lock removed")
		return nil
	}

	return errors.New("Unable to remove deploy lock")
}
