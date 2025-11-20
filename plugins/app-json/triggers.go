package appjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// TriggerAppJSONProcessDeployParallelism returns the max number of processes to deploy in parallel
func TriggerAppJSONProcessDeployParallelism(appName string, processType string) error {
	appJSON, err := GetAppJSON(appName)
	if err != nil {
		return err
	}

	parallelism := 1
	for procType, formation := range appJSON.Formation {
		if procType != processType {
			continue
		}

		if formation.MaxParallel == nil {
			continue
		}

		if *formation.MaxParallel > 0 {
			parallelism = *formation.MaxParallel
		}
	}

	fmt.Println(parallelism)
	return nil
}

// TriggerAppJSONGetContent outputs the contents of the app-json file, if any
func TriggerAppJSONGetContent(appName string) error {
	if !hasAppJSON(appName) {
		fmt.Print("{}")
		return nil
	}

	appJSON, err := ReadAppJSON(getProcessSpecificAppJSONPath(appName))
	if err != nil {
		return err
	}

	content, err := json.Marshal(appJSON)
	if err != nil {
		return err
	}

	fmt.Print(string(content))
	return nil
}

// TriggerCorePostDeploy moves the extracted app.json to the app data directory
// allowing the app to be restored on boot
func TriggerCorePostDeploy(appName string) error {
	return common.CorePostDeploy(common.CorePostDeployInput{
		AppName:     appName,
		Destination: common.GetAppDataDirectory("app-json", appName),
		PluginName:  "app-json",
		ExtractedPaths: []common.CorePostDeployPath{
			{Path: "app.json", IsDirectory: false},
		},
	})
}

// TriggerCorePostExtract ensures that the main app.json is the one specified by app-json-path
func TriggerCorePostExtract(appName string, sourceWorkDir string) error {
	destination := common.GetAppDataDirectory("app-json", appName)
	appJSONPath := strings.Trim(reportComputedAppjsonpath(appName), "/")
	if appJSONPath == "" {
		appJSONPath = "app.json"
	}

	validator := func(appName string, path string) error {
		if !common.FileExists(path) {
			return nil
		}

		result, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
			Trigger:      "app-json-is-valid",
			Args:         []string{appName, path},
			StreamStdout: true,
			StreamStderr: true,
		})

		if err != nil {
			if result.StderrContents() != "" {
				return errors.New(result.StderrContents())
			}

			return err
		}
		return nil
	}

	results, _ := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "builder-get-property",
		Args:    []string{appName, "build-dir"},
	})
	buildDir := results.StdoutContents()
	return common.CorePostExtract(common.CorePostExtractInput{
		AppName:       appName,
		BuildDir:      buildDir,
		Destination:   destination,
		PluginName:    "app-json",
		SourceWorkDir: sourceWorkDir,
		ToExtract: []common.CorePostExtractToExtract{
			{
				Path:        appJSONPath,
				IsDirectory: false,
				Name:        "app.json",
				Destination: "app.json",
				Validator:   validator,
			},
		},
	})
}

// TriggerInstall initializes app-json directory structures
func TriggerInstall() error {
	if err := common.PropertySetup("app-json"); err != nil {
		return fmt.Errorf("Unable to install the app-json plugin: %s", err.Error())
	}

	if err := common.SetupAppData("app-json"); err != nil {
		return err
	}

	return nil
}

// TriggerPostAppCloneSetup creates new app-json files
func TriggerPostAppCloneSetup(oldAppName string, newAppName string) error {
	err := common.PropertyClone("app-json", oldAppName, newAppName)
	if err != nil {
		return err
	}

	return common.CloneAppData("app-json", oldAppName, newAppName)
}

// TriggerPostAppRename removes the old app data
func TriggerPostAppRename(oldAppName string, newAppName string) error {
	return common.MigrateAppDataDirectory("app-json", oldAppName, newAppName)
}

// TriggerPostAppRenameSetup renames app-json files
func TriggerPostAppRenameSetup(oldAppName string, newAppName string) error {
	if err := common.PropertyClone("app-json", oldAppName, newAppName); err != nil {
		return err
	}

	if err := common.PropertyDestroy("app-json", oldAppName); err != nil {
		return err
	}

	return common.CloneAppData("app-json", oldAppName, newAppName)
}

// TriggerPostCreate ensures apps have the correct data directory structure
func TriggerPostCreate(appName string) error {
	return common.CreateAppDataDirectory("app-json", appName)
}

// TriggerPostDelete destroys the app-json data for a given app container
func TriggerPostDelete(appName string) error {
	dataErr := common.RemoveAppDataDirectory("app-json", appName)
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

	return executeScript(appName, image, imageTag, "postdeploy")
}

func TriggerPreReleaseBuilder(builderType string, appName string, image string) error {
	parts := strings.Split(image, ":")
	imageTag := parts[len(parts)-1]
	return executeScript(appName, image, imageTag, "predeploy")
}

// TriggerPostReleaseBuilder is a trigger to execute predeploy and release deployment tasks
func TriggerPostReleaseBuilder(builderType string, appName string, image string) error {
	parts := strings.Split(image, ":")
	imageTag := parts[len(parts)-1]
	if err := executeScript(appName, image, imageTag, "release"); err != nil {
		return err
	}

	if err := setScale(appName); err != nil {
		return err
	}

	if common.PropertyGet("common", appName, "deployed") == "true" {
		return nil
	}

	// Ensure that a failed postdeploy does not trigger twice
	if common.PropertyGet("app-json", appName, "heroku.postdeploy") == "executed" {
		return nil
	}

	if err := common.PropertyWrite("app-json", appName, "heroku.postdeploy", "executed"); err != nil {
		return err
	}

	return executeScript(appName, image, imageTag, "heroku.postdeploy")
}
