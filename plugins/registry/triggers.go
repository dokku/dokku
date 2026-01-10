package registry

import (
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// TriggerDeployedAppImageRepo outputs the associated image repo to stdout
func TriggerDeployedAppImageRepo(appName string) error {
	imageRepo := strings.TrimSpace(reportImageRepo(appName))
	if imageRepo == "" {
		var err error
		imageRepo, err = getImageRepoFromTemplate(appName)
		if err != nil {
			return fmt.Errorf("Unable to determine image repo from template: %w", err)
		}
	}

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
	tagVersion = strings.TrimSpace(tagVersion)
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

// TriggerPostCreate creates the registry config directory for a new app
func TriggerPostCreate(appName string) error {
	return common.PropertySetupApp("registry", appName)
}

// TriggerPostAppCloneSetup creates new registry files
func TriggerPostAppCloneSetup(oldAppName string, newAppName string) error {
	err := common.PropertyClone("registry", oldAppName, newAppName)
	if err != nil {
		return err
	}

	// Clone docker config.json if it exists
	oldConfigPath := GetAppRegistryConfigPath(oldAppName)
	if common.FileExists(oldConfigPath) {
		newConfigDir := GetAppRegistryConfigDir(newAppName)
		if err := os.MkdirAll(newConfigDir, 0700); err != nil {
			return fmt.Errorf("Unable to create registry config directory: %w", err)
		}
		if err := common.Copy(oldConfigPath, GetAppRegistryConfigPath(newAppName)); err != nil {
			return fmt.Errorf("Unable to clone registry config: %w", err)
		}
	}

	return nil
}

// TriggerPostAppRenameSetup renames registry files
func TriggerPostAppRenameSetup(oldAppName string, newAppName string) error {
	if err := common.PropertyClone("registry", oldAppName, newAppName); err != nil {
		return err
	}

	// Move docker config.json if it exists
	oldConfigPath := GetAppRegistryConfigPath(oldAppName)
	if common.FileExists(oldConfigPath) {
		newConfigDir := GetAppRegistryConfigDir(newAppName)
		if err := os.MkdirAll(newConfigDir, 0700); err != nil {
			return fmt.Errorf("Unable to create registry config directory: %w", err)
		}
		if err := os.Rename(oldConfigPath, GetAppRegistryConfigPath(newAppName)); err != nil {
			return fmt.Errorf("Unable to rename registry config: %w", err)
		}
	}

	if err := common.PropertyDestroy("registry", oldAppName); err != nil {
		return err
	}

	return nil
}

// TriggerPostDelete destroys the registry property for a given app container
func TriggerPostDelete(appName string) error {
	return common.PropertyDestroy("registry", appName)
}

// TriggerPostReleaseBuilder pushes the image to the remote registry
func TriggerPostReleaseBuilder(appName string, image string) error {
	parts := strings.Split(image, ":")
	imageTag := parts[len(parts)-1]
	if common.PlugnTriggerExists("pre-deploy") {
		common.LogWarn("Deprecated: please upgrade plugin to use 'pre-release-builder' plugin trigger instead of pre-deploy")
		_, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
			Trigger:     "pre-deploy",
			Args:        []string{appName, imageTag},
			StreamStdio: true,
		})
		if err != nil {
			return err
		}
	}

	imageID, _ := common.DockerInspect(image, "{{ .Id }}")
	imageRepo := common.GetAppImageRepo(appName)
	computedImageRepo := reportComputedImageRepo(appName)
	newImage := strings.Replace(image, imageRepo+":", computedImageRepo+":", 1)

	if computedImageRepo != imageRepo {
		if err := dockerTag(imageID, newImage); err != nil {
			return fmt.Errorf("unable to tag image %s as %s: %w", imageID, newImage, err)
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
