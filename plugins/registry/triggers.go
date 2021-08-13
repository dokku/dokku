package registry

import (
	"errors"
	"fmt"
	"strings"

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
	if !isPushEnabled(appName) {
		return nil
	}

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
func TriggerPostReleaseBuilder(appName string, image string) error {
	imageID, _ := common.DockerInspect(image, "{{ .Id }}")
	imageRepo := common.GetAppImageRepo(appName)
	computedImageRepo := reportComputedImageRepo(appName)
	newImage := strings.Replace(image, imageRepo+":", computedImageRepo+":", 1)

	parts := strings.Split(newImage, ":")
	imageTag := parts[len(parts)-1]
	if err := common.PlugnTrigger("pre-deploy", []string{appName, imageTag}...); err != nil {
		return err
	}

	if computedImageRepo != imageRepo {
		if !dockerTag(imageID, newImage) {
			// TODO: better error
			return errors.New("Unable to tag image")
		}
	}

	if !isPushEnabled(appName) {
		return nil
	}

	common.LogInfo1("Pushing image to registry")
	newTag, err := incrementTagVersion(appName)
	if err != nil {
		return err
	}

	return pushToRegistry(appName, newTag, imageID, computedImageRepo)
}
