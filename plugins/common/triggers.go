package common

import (
	"fmt"
	"os"
)

// TriggerAppList outputs each app name to stdout on a newline
func TriggerAppList() error {
	apps, _ := DokkuApps()
	for _, app := range apps {
		Log(app)
	}

	return nil
}

// TriggerCorePostDeploy associates the container with a specified network
func TriggerCorePostDeploy(appName string) error {
	quiet := os.Getenv("DOKKU_QUIET_OUTPUT")
	os.Setenv("DOKKU_QUIET_OUTPUT", "1")
	CommandPropertySet("common", appName, "deployed", "true", DefaultProperties, GlobalProperties)
	os.Setenv("DOKKU_QUIET_OUTPUT", quiet)
	return nil
}

// TriggerInstall runs the install step for the common plugin
func TriggerInstall() error {
	if err := PropertySetup("common"); err != nil {
		return fmt.Errorf("Unable to install the common plugin: %s", err.Error())
	}

	apps, err := DokkuApps()
	if err != nil {
		return nil
	}

	// migrate all is-deployed values from trigger to property
	for _, appName := range apps {
		IsDeployed(appName)
	}

	return nil
}

// TriggerPostAppCloneSetup copies common files
func TriggerPostAppCloneSetup(oldAppName string, newAppName string) error {
	if err := PropertyClone("common", oldAppName, newAppName); err != nil {
		return err
	}

	if err := PropertyDelete("common", oldAppName, "deployed"); err != nil {
		return err
	}

	return nil
}

// TriggerPostAppRenameSetup renames common files
func TriggerPostAppRenameSetup(oldAppName string, newAppName string) error {
	if err := PropertyClone("common", oldAppName, newAppName); err != nil {
		return err
	}

	if err := PropertyDestroy("common", oldAppName); err != nil {
		return err
	}

	return nil
}

// TriggerPostDelete destroys the common property for a given app container
func TriggerPostDelete(appName string) error {
	return PropertyDestroy("common", appName)
}
