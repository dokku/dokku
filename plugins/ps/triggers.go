package ps

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	dockeroptions "github.com/dokku/dokku/plugins/docker-options"
)

// TriggerAppRestart restarts an app
func TriggerAppRestart(appName string) error {
	return Restart(appName)
}

// TriggerCorePostDeploy sets a property to
// allow the app to be restored on boot
func TriggerCorePostDeploy(appName string) error {
	err := common.CorePostDeploy(common.CorePostDeployInput{
		AppName:     appName,
		Destination: common.GetAppDataDirectory("ps", appName),
		PluginName:  "ps",
		ExtractedPaths: []common.CorePostDeployPath{
			{Path: "Procfile", IsDirectory: false},
		},
	})
	if err != nil {
		return err
	}

	if err := common.PropertyDelete("ps", appName, "scale.old"); err != nil {
		return err
	}

	return common.PropertyWrite("ps", appName, "restore", "true")
}

// TriggerCorePostExtract ensures that the main Procfile is the one specified by procfile-path
func TriggerCorePostExtract(appName string, sourceWorkDir string) error {
	destination := common.GetAppDataDirectory("ps", appName)
	procfilePath := strings.Trim(reportComputedProcfilePath(appName), "/")
	if procfilePath == "" {
		procfilePath = "Procfile"
	}

	validator := func(appName string, path string) error {
		if !common.FileExists(path) {
			return nil
		}

		result, err := common.CallExecCommand(common.ExecCommandInput{
			Command: "procfile-util",
			Args:    []string{"check", "-P", path},
		})
		if err != nil {
			return err
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("Invalid Procfile: %s", result.StderrContents())
		}
		return nil
	}

	return common.CorePostExtract(common.CorePostExtractInput{
		AppName:       appName,
		Destination:   destination,
		PluginName:    "ps",
		SourceWorkDir: sourceWorkDir,
		ToExtract: []common.CorePostExtractToExtract{
			{
				Path:        procfilePath,
				IsDirectory: false,
				Name:        "Procfile",
				Destination: "Procfile",
				Validator:   validator,
			},
		},
	})
}

// TriggerInstall initializes app restart policies
func TriggerInstall() error {
	if err := common.PropertySetup("ps"); err != nil {
		return fmt.Errorf("Unable to install the ps plugin: %s", err.Error())
	}

	if err := common.SetupAppData("ps"); err != nil {
		return err
	}

	apps, err := common.UnfilteredDokkuApps()
	if err != nil && !errors.Is(err, common.NoAppsExist) {
		return nil
	}

	for _, appName := range apps {
		if err := common.CreateAppDataDirectory("ps", appName); err != nil {
			return err
		}
	}

	for _, appName := range apps {
		policies, err := getRestartPolicy(appName)
		if err != nil {
			return err
		}

		if len(policies) != 0 {
			continue
		}

		if err := dockeroptions.AddDockerOptionToPhases(appName, []string{"deploy"}, "--restart=on-failure:10"); err != nil {
			common.LogWarn(err.Error())
		}
	}

	for _, appName := range apps {
		dokkuScaleFile := filepath.Join(common.AppRoot(appName), "DOKKU_SCALE")
		if common.FileExists(dokkuScaleFile) {
			os.Remove(dokkuScaleFile)
		}

		dokkuScaleExtracted := filepath.Join(common.AppRoot(appName), "DOKKU_SCALE.extracted")
		if common.FileExists(dokkuScaleExtracted) {
			os.Remove(dokkuScaleExtracted)
		}
	}

	if err := common.MigrateConfigToProperties("ps", []common.MigrateConfigEntry{
		{
			ConfigVar: "DOKKU_DOCKER_STOP_TIMEOUT",
			Property:  "stop-timeout-seconds",
		},
		{
			ConfigVar: "DOKKU_APP_RESTORE",
			Property:  "restore",
			Transform: func(value string) string {
				if value == "0" {
					return "false"
				}
				return "true"
			},
		},
		{
			ConfigVar:       "DOKKU_SKIP_DEPLOY",
			GlobalConfigVar: "DOKKU_SKIP_DEPLOY",
			Property:        "skip-deploy",
		},
		{
			ConfigVar: "DOKKU_START_CMD",
			Property:  "start-cmd",
		},
		{
			ConfigVar: "DOKKU_DOCKERFILE_START_CMD",
			Property:  "dockerfile-start-cmd",
		},
	}); err != nil {
		return err
	}

	return nil
}

// TriggerPostAppClone rebuilds the new app
func TriggerPostAppClone(oldAppName string, newAppName string) error {
	if os.Getenv("SKIP_REBUILD") == "true" {
		return nil
	}

	return Rebuild(newAppName)
}

// TriggerPostAppCloneSetup creates new ps files
func TriggerPostAppCloneSetup(oldAppName string, newAppName string) error {
	err := common.PropertyClone("ps", oldAppName, newAppName)
	if err != nil {
		return err
	}

	return common.CloneAppData("ps", oldAppName, newAppName)
}

// TriggerPostAppRename rebuilds the renamed app
func TriggerPostAppRename(oldAppName string, newAppName string) error {
	if err := common.MigrateAppDataDirectory("ps", oldAppName, newAppName); err != nil {
		return err
	}

	if os.Getenv("SKIP_REBUILD") == "true" {
		return nil
	}

	return Rebuild(newAppName)
}

// TriggerPostAppRenameSetup renames ps files
func TriggerPostAppRenameSetup(oldAppName string, newAppName string) error {
	if err := common.PropertyClone("ps", oldAppName, newAppName); err != nil {
		return err
	}

	if err := common.PropertyDestroy("ps", oldAppName); err != nil {
		return err
	}

	return common.CloneAppData("ps", oldAppName, newAppName)
}

// TriggerPostCreate ensures apps have a default restart policy
// and scale value for web
func TriggerPostCreate(appName string) error {
	if err := dockeroptions.AddDockerOptionToPhases(appName, []string{"deploy"}, "--restart=on-failure:10"); err != nil {
		return err
	}

	if err := common.CreateAppDataDirectory("ps", appName); err != nil {
		return err
	}

	formations := FormationSlice{
		&Formation{
			ProcessType: "web",
			Quantity:    1,
		},
	}
	return updateScale(appName, false, formations)
}

// TriggerPostDelete destroys the ps properties for a given app container
func TriggerPostDelete(appName string) error {
	dataErr := common.RemoveAppDataDirectory("ps", appName)
	propertyErr := common.PropertyDestroy("ps", appName)

	if dataErr != nil {
		return dataErr
	}

	return propertyErr
}

// TriggerPostStop sets the restore property to false
func TriggerPostStop(appName string) error {
	return common.PropertyWrite("ps", appName, "restore", "false")
}

// TriggerPostReleaseBuilder ensures an app has an up to date scale parameters
func TriggerPostReleaseBuilder(builderType string, appName string, image string) error {
	if err := updateScale(appName, false, FormationSlice{}); err != nil {
		common.LogDebug(fmt.Sprintf("Error generating scale file: %s", err.Error()))
		return err
	}

	return nil
}

// TriggerProcfileExists checks if a procfile exists
func TriggerProcfileExists(appName string) error {
	if hasProcfile(appName) {
		return nil
	}

	return errors.New("Procfile does not exist")
}

// TriggerProcfileGetCommand fetches a command from the procfile
func TriggerProcfileGetCommand(appName string, processType string, port int) error {
	if !hasProcfile(appName) {
		return nil
	}

	command, err := getProcfileCommand(appName, getProcessSpecificProcfilePath(appName), processType, port)
	if err != nil {
		return err
	}

	if command != "" {
		fmt.Printf("%s\n", command)
	}

	return nil
}

// TriggerPsCanScale sets whether or not a user can scale an app with ps:scale
func TriggerPsCanScale(appName string, canScale bool) error {
	return common.PropertyWrite("ps", appName, "can-scale", strconv.FormatBool(canScale))
}

// TriggerPsCurrentScale prints out the current scale contents (process-type=quantity) delimited by newlines
func TriggerPsCurrentScale(appName string) error {
	formations, err := getFormations(appName)
	if err != nil {
		return err
	}

	lines := []string{}
	for _, formation := range formations {
		lines = append(lines, fmt.Sprintf("%s=%d", formation.ProcessType, formation.Quantity))
	}

	fmt.Print(strings.Join(lines, "\n"))

	return nil
}

// TriggerPsSetScale configures the scale parameters for a given app
func TriggerPsSetScale(appName string, skipDeploy bool, clearExisting bool, processTuples []string) error {
	return scaleSet(scaleSetInput{
		appName:           appName,
		skipDeploy:        skipDeploy,
		clearExisting:     clearExisting,
		processTuples:     processTuples,
		deployOnlyChanged: false,
	})
}

func TriggerPsGetProperty(appName string, property string) error {
	computedValueMap := map[string]common.ReportFunc{
		"restore":              reportRestore,
		"skip-deploy":          reportComputedSkipDeploy,
		"stop-timeout-seconds": reportComputedStopTimeoutSeconds,
	}

	fn, ok := computedValueMap[property]
	if !ok {
		return fmt.Errorf("Invalid ps property specified: %v", property)
	}

	fmt.Println(fn(appName))
	return nil
}

func reportComputedSkipDeploy(appName string) string {
	value := common.PropertyGet("ps", appName, "skip-deploy")
	if value == "" {
		value = common.PropertyGet("ps", "--global", "skip-deploy")
	}
	return value
}
