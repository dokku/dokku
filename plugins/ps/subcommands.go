package ps

import (
	"errors"
	"fmt"
	"strings"

	dockeroptions "github.com/dokku/dokku/plugins/docker-options"

	"github.com/dokku/dokku/plugins/common"
)

// CommandInspect displays a sanitized version of docker inspect for an app
func CommandInspect(appName string) error {
	if appName == "" {
		return errors.New("Please specify an app to run the command on")
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	scheduler := common.GetAppScheduler(appName)
	return common.PlugnTrigger("scheduler-inspect", []string{scheduler, appName}...)
}

// CommandRebuild rebuilds an app from source
func CommandRebuild(appName string, allApps bool, parallelCount int) error {
	if allApps {
		return RunCommandAgainstAllApps(Rebuild, "rebuild", parallelCount)
	}

	if appName == "" {
		return errors.New("Please specify an app to run the command on")
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	return Rebuild(appName)
}

// CommandReport displays a ps report for one or more apps
func CommandReport(appName string, infoFlag string) error {
	if strings.HasPrefix(appName, "--") {
		infoFlag = appName
		appName = ""
	}

	if len(appName) == 0 {
		apps, err := common.DokkuApps()
		if err != nil {
			return err
		}
		for _, appName := range apps {
			if err := ReportSingleApp(appName, infoFlag); err != nil {
				return err
			}
		}
		return nil
	}

	return ReportSingleApp(appName, infoFlag)
}

// CommandRestart restarts an app
func CommandRestart(appName string, allApps bool, parallelCount int) error {
	if allApps {
		return RunCommandAgainstAllApps(Restart, "restart", parallelCount)
	}

	if appName == "" {
		return errors.New("Please specify an app to run the command on")
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	return Restart(appName)
}

// CommandRestore starts previously running apps e.g. after reboot
// TODO: implement me
func CommandRestore(appName string) error {
	return nil
}

// CommandRetire ensures old containers are retired
// TODO: implement me
func CommandRetire() error {
	return nil
}

// CommandScale gets or sets how many instances of a given process to run
func CommandScale(appName string, skipDeploy bool, processTuples []string) error {
	if appName == "" {
		return errors.New("Please specify an app to run the command on")
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	procfilePath := getProcfilePath(appName)
	if !common.FileExists(procfilePath) {
		image := common.GetAppImageRepo(appName)
		common.SuppressOutput(func() error {
			extractProcfile(appName, image, procfilePath)
			return nil
		})
	}

	if !hasScaleFile(appName) {
		err := common.SuppressOutput(func() error {
			return updateScalefile(appName, make(map[string]int))
		})
		if err != nil {
			return err
		}
	} else if common.FileExists(procfilePath) {
		if err := updateScalefile(appName, make(map[string]int)); err != nil {
			return err
		}
	}

	if len(processTuples) == 0 {
		return scaleReport(appName)
	}

	return scaleSet(appName, skipDeploy, processTuples)
}

// CommandSet sets or clears a ps property for an app
func CommandSet(appName string, property string, value string) error {
	if property == "restart-policy" {
		if !isValidRestartPolicy(value) {
			return errors.New("Invalid restart-policy specified")
		}

		common.LogInfo2Quiet(fmt.Sprintf("Setting %s to %s", property, value))
		return dockeroptions.SetDockerOptionForPhases(appName, []string{"deploy"}, "restart", value)
	}

	common.CommandPropertySet("ps", appName, property, value, DefaultProperties)
	return nil
}

// CommandStart starts an app
func CommandStart(appName string, allApps bool, parallelCount int) error {
	if allApps {
		return RunCommandAgainstAllApps(Start, "start", parallelCount)
	}

	if appName == "" {
		return errors.New("Please specify an app to run the command on")
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	return Start(appName)
}

// CommandStop stops an app
func CommandStop(appName string, allApps bool, parallelCount int) error {
	if allApps {
		return RunCommandAgainstAllApps(Stop, "stop", parallelCount)
	}

	if appName == "" {
		return errors.New("Please specify an app to run the command on")
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	return Stop(appName)
}
