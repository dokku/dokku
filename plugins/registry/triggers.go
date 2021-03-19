package registry

import (
	"fmt"

	"github.com/dokku/dokku/plugins/common"
)

// TriggerDeployedAppImageRepo outputs the associated image repo to stdout
func TriggerDeployedAppImageRepo(appName string) error {
	imageRepo := common.PropertyGet("registry", appName, "image-repo")
	if imageRepo == "" {
		imageRepo = common.GetAppImageRepo(appName)
	}

	fmt.Println(imageRepo)
	return nil
}

// TriggerDeployedAppImageTag outputs the associated image tag to stdout
func TriggerDeployedAppImageTag(appName string) error {
	tagVersion := common.PropertyGet("registry", appName, "tag-version")
	if tagVersion == "" {
		tagVersion = "1"
	}

	fmt.Println(tagVersion)
	return nil
}

// TriggerDeployedAppRepository outputs the associated registry repository to stdout
func TriggerDeployedAppRepository(appName string) error {
	fmt.Println(getRegistryServerForApp(appName))
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
