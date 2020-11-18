package ps

import (
	"errors"
	"fmt"
	"strings"

	dockeroptions "github.com/dokku/dokku/plugins/docker-options"

	"github.com/dokku/dokku/plugins/common"
)

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

func CommandRebuild(appName string, allApps bool, runInSerial bool, parallelCount int) error {
	if allApps {
		return RunCommandAgainstAllApps(Rebuild, "rebuild", runInSerial, parallelCount)
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

func CommandRestart(appName string, allApps bool, runInSerial bool, parallelCount int) error {
	if allApps {
		return RunCommandAgainstAllApps(Restart, "restart", runInSerial, parallelCount)
	}

	if appName == "" {
		return errors.New("Please specify an app to run the command on")
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	return Restart(appName)
}

// TODO: implement me
func CommandRestore(appName string) error {
	return nil
}

// TODO: implement me
func CommandRetire() error {
	return nil
}

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
			return generateScalefile(appName)
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

func CommandStart(appName string, allApps bool, runInSerial bool, parallelCount int) error {
	if allApps {
		return RunCommandAgainstAllApps(Start, "start", runInSerial, parallelCount)
	}

	if appName == "" {
		return errors.New("Please specify an app to run the command on")
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	return Start(appName)
}

func CommandStop(appName string, allApps bool, runInSerial bool, parallelCount int) error {
	if allApps {
		return RunCommandAgainstAllApps(Stop, "stop", runInSerial, parallelCount)
	}

	if appName == "" {
		return errors.New("Please specify an app to run the command on")
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	return Stop(appName)
}
