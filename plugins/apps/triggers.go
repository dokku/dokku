package apps

import (
	"fmt"
	"os"
	"time"

	"github.com/dokku/dokku/plugins/common"
)

// TriggerAppCreate is a trigger to create an app
func TriggerAppCreate(appName string) error {
	return createApp(appName)
}

// TriggerAppDestroy is a trigger to destroy an app
func TriggerAppDestroy(appName string) error {
	return destroyApp(appName)
}

// TriggerAppExists is a trigger to check if an app exists
func TriggerAppExists(appName string) error {
	return appExists(appName)
}

// TriggerAppMaybeCreate is a trigger to allow gated app creation
func TriggerAppMaybeCreate(appName string) error {
	return maybeCreateApp(appName)
}

// TriggerDeploySourceSet sets the current deploy source
func TriggerDeploySourceSet(appName string, sourceType string, sourceMetadata string) error {
	if err := common.PropertyWrite("apps", appName, "deploy-source", sourceType); err != nil {
		return err
	}

	return common.PropertyWrite("apps", appName, "deploy-source-metadata", sourceMetadata)
}

// TriggerInstall runs the install step for the apps plugin
func TriggerInstall() error {
	if err := common.PropertySetup("apps"); err != nil {
		return fmt.Errorf("Unable to install the apps plugin: %s", err.Error())
	}

	if err := common.SetupAppData("apps"); err != nil {
		return err
	}

	apps, err := common.UnfilteredDokkuApps()
	if err != nil {
		return nil
	}

	// migrate all created-at values from app mod-time to property
	for _, appName := range apps {
		if common.PropertyExists("apps", appName, "created-at") {
			continue
		}

		fi, err := os.Stat(common.AppRoot(appName))
		if err != nil {
			// if we can't get the time, just write the current one out as a stub
			if err := common.PropertyWrite("apps", appName, "created-at", fmt.Sprintf("%d", time.Now().Unix())); err != nil {
				return err
			}

			continue
		}

		if err := common.PropertyWrite("apps", appName, "created-at", fmt.Sprintf("%d", fi.ModTime().Unix())); err != nil {
			return err
		}
	}

	return nil
}

// TriggerPostAppCloneSetup creates new apps files
func TriggerPostAppCloneSetup(oldAppName string, newAppName string) error {
	err := common.PropertyClone("apps", oldAppName, newAppName)
	if err != nil {
		return err
	}

	return common.CloneAppData("apps", oldAppName, newAppName)
}

// TriggerPostAppRename removes the old app data
func TriggerPostAppRename(oldAppName string, newAppName string) error {
	return common.MigrateAppDataDirectory("apps", oldAppName, newAppName)
}

// TriggerPostAppRenameSetup renames apps files
func TriggerPostAppRenameSetup(oldAppName string, newAppName string) error {
	if err := common.PropertyClone("apps", oldAppName, newAppName); err != nil {
		return err
	}

	if err := common.PropertyDestroy("apps", oldAppName); err != nil {
		return err
	}

	return common.CloneAppData("apps", oldAppName, newAppName)
}

// TriggerPostCreate ensures apps have the correct data directory structure
func TriggerPostCreate(appName string) error {
	return common.CreateAppDataDirectory("apps", appName)
}

// TriggerPostDelete destroys the apps data for a given app container
func TriggerPostDelete(appName string) error {
	dataErr := common.RemoveAppDataDirectory("apps", appName)
	propertyErr := common.PropertyDestroy("apps", appName)

	if dataErr != nil {
		return dataErr
	}

	return propertyErr
}
