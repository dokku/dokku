package common

import (
	"fmt"
	"os"
)

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

// TriggerPostDelete destroys the common property for a given app container
func TriggerPostDelete(appName string) error {
	return PropertyDestroy("common", appName)
}
