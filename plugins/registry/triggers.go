package registry

import (
	"fmt"

	"github.com/dokku/dokku/plugins/common"
)

// TriggerDeployedAppRepository outputs the associated registry repository to stdout
func TriggerDeployedAppRepository(appName string) error {
	// TODO
	return nil
}

// TriggerInstall runs the install step for the registry plugin
func TriggerInstall() error {
	if err := common.PropertySetup("registry"); err != nil {
		return fmt.Errorf("Unable to install the registry plugin: %s", err.Error())
	}

	return nil
}

// TriggerPostDelete destroys the registry property for a given app container
func TriggerPostDelete(appName string) error {
	return common.PropertyDestroy("registry", appName)
}

// TriggerPostReleaseBuilder pushes the image to the remote registry
func TriggerPostReleaseBuilder(appName string) error {
	if !isPushEnabled(appName) {
		return nil
	}

	common.LogInfo1("Pushing image to registry")
	imageTag, err := incrementTagVersion(appName)
	if err != nil {
		return err
	}

	return pushToRegistry(appName, imageTag)
}
