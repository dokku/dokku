package ps

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	dockeroptions "github.com/dokku/dokku/plugins/docker-options"

	"github.com/dokku/dokku/plugins/common"
	"github.com/ryanuber/columnize"
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

func CommandScale(appName string, processTuples []string) error {
	if appName == "" {
		return errors.New("Please specify an app to run the command on")
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	procfilePath := getProcfilePath(appName)
	scalefilePath := getScalefilePath(appName)
	if !common.FileExists(procfilePath) {
		image := common.GetAppImageRepo(appName)
		common.SuppressOutput(func() error {
			extractProcfile(appName, image, procfilePath)
			return nil
		})
	}

	if !common.FileExists(scalefilePath) {
		err := common.SuppressOutput(func() error {
			return generateScalefile(appName, scalefilePath)
		})
		if err != nil {
			return err
		}
	} else if common.FileExists(procfilePath) {
		if err := updateScalefile(appName, []string{}); err != nil {
			return err
		}
	}

	if len(processTuples) == 0 {
		lines, err := common.FileToSlice(scalefilePath)
		if err != nil {
			return err
		}

		common.LogInfo1Quiet(fmt.Sprintf("Scaling for %s", appName))
		config := columnize.DefaultConfig()
		config.Delim = "="
		config.Glue = ": "
		config.Prefix = "    "
		config.Empty = ""

		content := []string{}
		if os.Getenv("DOKKU_QUIET_OUTPUT") == "" {
			content = append(content, "proctype=qty","--------=---")
		}

		sort.Strings(lines)
		for _, line := range lines {
			content = append(content, line)
		}

		for _, line := range content {
			s := strings.Split(line, "=")
			common.Log(fmt.Sprintf("%s %s", common.RightPad(fmt.Sprintf("%s:", s[0]), 5, " "), s[1]))
		}
	} else {
		if !canScaleApp(appName) {
			return fmt.Errorf("App %s contains DOKKU_SCALE file and cannot be manually scaled", appName)
		}

		common.LogInfo1(fmt.Sprintf("Scaling %s processes: %s", appName, strings.Join(processTuples, " ")))
		if err := updateScalefile(appName, processTuples); err != nil {
			return err
		}

		if !common.IsDeployed(appName) {
			return nil
		}

		imageTag, err := common.GetRunningImageTag(appName)
		if err != nil {
			return err
		}
		return common.PlugnTrigger("release-and-deploy", []string{appName, imageTag}...)
	}

	return nil
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
