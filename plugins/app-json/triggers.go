package appjson

import (
	"os"
	"path/filepath"

	"github.com/dokku/dokku/plugins/common"
)

// TriggerInstall initializes app restart policies
func TriggerInstall() error {
	directory := filepath.Join(common.MustGetEnv("DOKKU_LIB_ROOT"), "data", "app-json")
	if err := os.MkdirAll(directory, 0755); err != nil {
		return err
	}

	if err := common.SetPermissions(directory, 0755); err != nil {
		return err
	}

	return nil
}

// TriggerPostDelete destroys the app-json data for a given app container
func TriggerPostDelete(appName string) error {
	directory := filepath.Join(common.MustGetEnv("DOKKU_LIB_ROOT"), "data", "app-json", appName)
	dataErr := os.RemoveAll(directory)
	propertyErr := common.PropertyDestroy("app-json", appName)

	if dataErr != nil {
		return dataErr
	}

	return propertyErr
}

// TriggerPostDeploy is a trigger to execute the postdeploy deployment task
func TriggerPostDeploy(appName string, imageTag string) error {
	image, err := common.GetDeployingAppImageName(appName, imageTag, "")
	if err != nil {
		return err
	}

	if err := executeScript(appName, image, imageTag, "postdeploy"); err != nil {
		return err
	}
	return nil
}

// TriggerPreDeploy is a trigger to execute predeploy and release deployment tasks
func TriggerPreDeploy(appName string, imageTag string) error {
	image, err := common.GetDeployingAppImageName(appName, imageTag, "")
	if err != nil {
		return err
	}

	if err := refreshAppJSON(appName, image); err != nil {
		return err
	}

	if err := executeScript(appName, image, imageTag, "predeploy"); err != nil {
		return err
	}

	if err := executeScript(appName, image, imageTag, "release"); err != nil {
		return err
	}
	return nil
}
