package ps

import (
	"fmt"

	"github.com/dokku/dokku/plugins/common"
)

const RUN_IN_SERIAL = -1

type ParallelCommand func(string) error

var (
	// DefaultProperties is a map of all valid ps properties with corresponding default property values
	DefaultProperties = map[string]string{
		"restart-policy": "on-failure:10",
	}
)

func RunCommandAgainstAllApps(command ParallelCommand, commandName string, runInSerial bool, parallelCount int) error {
	if runInSerial && parallelCount != RUN_IN_SERIAL {
		common.LogWarn("Ignoring --parallel value and running in serial mode")
	}

	if runInSerial {
		return RunCommandAgainstAllAppsSerially(command, commandName)
	}

	return RunCommandAgainstAllAppsInParallel(command, commandName, parallelCount)
}

// TODO: implement me
func RunCommandAgainstAllAppsInParallel(command ParallelCommand, commandName string, parallelCount int) error {
	return nil
}

func RunCommandAgainstAllAppsSerially(command ParallelCommand, commandName string) error {
	apps, err := common.DokkuApps()
	if err != nil {
		common.LogWarn(err.Error())
		return nil
	}

	errorCount := 0
	for _, appName := range apps {
		if err = command(appName); err != nil {
			errorCount += 1
		}
	}

	if errorCount > 0 {
		return fmt.Errorf("%s command returned %d errors", commandName, errorCount)
	}

	return nil
}

// Rebuild rebuilds app from base image
func Rebuild(appName string) error {
	return common.PlugnTrigger("receive-app", []string{appName}...)
}

// ReportSingleApp is an internal function that displays the ps report for one or more apps
// TODO: implement me
func ReportSingleApp(appName, infoFlag string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}


	policy, _ := getRestartPolicy(appName)
	if policy == "" {
		policy = DefaultProperties["restart-policy"]
	}

	infoFlags := map[string]string{
		"--ps-restart-policy": policy,
	}

	trimPrefix := false
	uppercaseFirstCharacter := true
	return common.ReportSingleApp("ps", appName, infoFlag, infoFlags, trimPrefix, uppercaseFirstCharacter)
}

// Restart restarts the app
func Restart(appName string) error {
	if !common.IsDeployed(appName) {
		common.LogWarn(fmt.Sprintf("App %s has not been deployed", appName))
		return nil
	}

	return common.PlugnTrigger("release-and-deploy", []string{appName}...)
}

// Start starts the app
func Start(appName string) error {
	imageTag, _ := common.GetRunningImageTag(appName)

	if !common.IsDeployed(appName) {
		common.LogWarn(fmt.Sprintf("App %s has not been deployed", appName))
		return nil
	}

	if err := common.PlugnTrigger("pre-start", []string{appName}...); err != nil {
		return fmt.Errorf("Failure in pre-start hook: %s", err)
	}

	runningState := getRunningState(appName)

	if runningState == "mixed" {
		common.LogWarn("App is running in mixed mode, releasing")
	}

	if runningState != "true" {
		if err := common.PlugnTrigger("release-and-deploy", []string{appName, imageTag}...); err != nil {
			return err
		}
	} else {
		common.LogWarn(fmt.Sprintf("App %s already running", appName))
	}

	if err := common.PlugnTrigger("proxy-build-config", []string{appName}...); err != nil {
		return err
	}

	return nil
}

// Stop stops the app
func Stop(appName string) error {
	if !common.IsDeployed(appName) {
		common.LogWarn(fmt.Sprintf("App %s has not been deployed", appName))
		return nil
	}

	common.LogQuiet(fmt.Sprintf("Stopping %s", appName))
	scheduler := common.GetAppScheduler(appName)

	if err := common.PlugnTrigger("scheduler-stop", []string{scheduler, appName}...); err != nil {
		return err
	}

	if err := common.PlugnTrigger("post-stop", []string{appName}...); err != nil {
		return err
	}

	return nil
}
