package builder

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// TriggerBuilderDetect outputs a manually selected builder for the app
func TriggerBuilderDetect(appName string) error {
	if builder := common.PropertyGet("builder", appName, "selected"); builder != "" {
		fmt.Println(builder)
		return nil
	}

	if builder := common.PropertyGet("builder", "--global", "selected"); builder != "" {
		fmt.Println(builder)
		return nil
	}

	return nil
}

// TriggerBuilderGetProperty writes the builder key to stdout for a given app container
func TriggerBuilderGetProperty(appName string, key string) error {
	if _, ok := DefaultProperties[key]; !ok {
		return errors.New("Invalid builder property specified")
	}

	fmt.Println(common.PropertyGet("builder", appName, key))
	return nil
}

// TriggerBuilderSetProperty writes the builder key to stdout for a given app container
func TriggerBuilderSetProperty(appName string, key string, value string) error {
	if _, ok := DefaultProperties[key]; !ok {
		return errors.New("Invalid builder property specified")
	}

	return common.PropertyWrite("builder", appName, key, value)
}

// TriggerBuilderImageIsCNB prints true if an image is cnb based, false otherwise
func TriggerBuilderImageIsCNB(appName string, image string) error {
	if common.IsImageCnbBased(image) {
		fmt.Println("true")
	} else {
		fmt.Println("false")
	}

	return nil
}

// TriggerBuilderImageIsHerokuish prints true if an image is herokuish based, false otherwise
func TriggerBuilderImageIsHerokuish(appName string, image string) error {
	if common.IsImageHerokuishBased(image, appName) {
		fmt.Println("true")
	} else {
		fmt.Println("false")
	}

	return nil
}

// TriggerCorePostExtract moves a configured build-dir to be in the app root dir
func TriggerCorePostExtract(appName string, sourceWorkDir string) error {
	buildDir := strings.Trim(reportComputedBuildDir(appName), "/")
	if buildDir == "" {
		return nil
	}

	newSourceWorkDir := filepath.Join(sourceWorkDir, buildDir)
	if !common.DirectoryExists(newSourceWorkDir) {
		return fmt.Errorf("Specified build-dir not found in sourcecode working directory: %v", buildDir)
	}

	tmpWorkDir, err := os.MkdirTemp(os.TempDir(), fmt.Sprintf("dokku-%s-%s", common.MustGetEnv("DOKKU_PID"), "CorePostExtract"))
	if err != nil {
		return fmt.Errorf("Unable to create temporary working directory: %v", err.Error())
	}

	if err := removeAllContents(tmpWorkDir); err != nil {
		return fmt.Errorf("Unable to clear out temporary working directory for rewrite: %v", err.Error())
	}

	if err := common.Copy(newSourceWorkDir, tmpWorkDir); err != nil {
		return fmt.Errorf("Unable to move build-dir to temporary working directory: %v", err.Error())
	}

	if err := removeAllContents(sourceWorkDir); err != nil {
		return fmt.Errorf("Unable to clear out sourcecode working directory for rewrite: %v", err.Error())
	}

	if err := common.Copy(tmpWorkDir, sourceWorkDir); err != nil {
		return fmt.Errorf("Unable to move build-dir to sourcecode working directory: %v", err.Error())
	}

	return nil
}

// TriggerInstall runs the install step for the builder plugin
func TriggerInstall() error {
	if err := common.PropertySetup("builder"); err != nil {
		return fmt.Errorf("Unable to install the builder plugin: %s", err.Error())
	}

	apps, err := common.UnfilteredDokkuApps()
	if err != nil {
		return nil
	}

	for _, appName := range apps {
		if common.PropertyExists("builder", appName, "detected") {
			continue
		}

		results, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
			Trigger: "config-get",
			Args:    []string{appName, "DOKKU_APP_TYPE"},
		})
		if err != nil {
			return err
		}

		if results.StdoutContents() != "" {
			common.PropertyWrite("builder", appName, "detected", results.StdoutContents())
		}
	}

	_, err = common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "install-builder-prune",
	})

	return err
}

// TriggerPostAppCloneSetup creates new builder files
func TriggerPostAppCloneSetup(oldAppName string, newAppName string) error {
	err := common.PropertyClone("builder", oldAppName, newAppName)
	if err != nil {
		return err
	}

	return nil
}

// TriggerPostAppRenameSetup renames builder files
func TriggerPostAppRenameSetup(oldAppName string, newAppName string) error {
	if err := common.PropertyClone("builder", oldAppName, newAppName); err != nil {
		return err
	}

	if err := common.PropertyDestroy("builder", oldAppName); err != nil {
		return err
	}

	return nil
}

// TriggerPostDelete destroys the builder property for a given app container
func TriggerPostDelete(appName string) error {
	if err := common.PropertyDestroy("builder", appName); err != nil {
		return err
	}

	imagesByAppLabel, err := common.DockerFilterImages([]string{
		fmt.Sprintf("label=com.dokku.app-name=%s", appName),
	})
	if err != nil {
		common.LogWarn(err.Error())
	}

	imageRepo := common.GetAppImageRepo(appName)
	imagesByRepo, err := listImagesByImageRepo(imageRepo)
	if err != nil {
		common.LogWarn(err.Error())
	}

	images := append(imagesByAppLabel, imagesByRepo...)
	common.RemoveImages(images)

	return nil
}

// TriggerPostReleaseBuilder deletes unused build images
func TriggerPostReleaseBuilder(builderType string, appName string) error {
	images, _ := common.DockerFilterImages([]string{
		"label=com.dokku.image-stage=build",
		fmt.Sprintf("label=com.dokku.app-name=%s", appName),
	})

	if err := common.RemoveImages(images); err != nil {
		common.LogWarn(err.Error())
	}

	return nil
}
