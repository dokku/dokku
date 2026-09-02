package ps

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/gofrs/flock"
)

// CommandInspect displays a sanitized version of docker inspect for an app
func CommandInspect(appName string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	scheduler := common.GetAppScheduler(appName)
	_, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "scheduler-inspect",
		Args:        []string{scheduler, appName},
		StreamStdio: true,
	})
	return err
}

// CommandRebuild rebuilds an app from source
func CommandRebuild(appName string, allApps bool, parallelCount int) error {
	if allApps {
		return common.RunCommandAgainstAllApps(Rebuild, "rebuild", parallelCount)
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	return Rebuild(appName)
}

// CommandReport displays a ps report for one or more apps
func CommandReport(appName string, format string, infoFlag string) error {
	if len(appName) == 0 {
		apps, err := common.DokkuApps()
		if err != nil {
			if errors.Is(err, common.NoAppsExist) {
				common.LogWarn(err.Error())
				return nil
			}
			return err
		}
		for _, appName := range apps {
			if err := ReportSingleApp(appName, format, infoFlag); err != nil {
				return err
			}
		}
		return nil
	}

	return ReportSingleApp(appName, format, infoFlag)
}

// CommandRestart restarts an app
func CommandRestart(appName string, processName string, allApps bool, parallelCount int) error {
	if allApps {
		if processName != "" {
			return errors.New("Unable to restart all apps when specifying a process name")
		}
		return common.RunCommandAgainstAllApps(Restart, "restart", parallelCount)
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if processName != "" {
		return RestartProcess(appName, processName)
	}

	return Restart(appName)
}

// CommandRestore starts previously running apps e.g. after reboot
func CommandRestore(appName string, allApps bool, parallelCount int) error {
	_, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "pre-restore",
		StreamStdio: true,
	})
	if err != nil {
		return fmt.Errorf("Error running pre-restore: %s", err)
	}

	if allApps {
		if err := restorePrep(); err != nil {
			return err
		}

		return common.RunCommandAgainstAllApps(Restore, "restore", parallelCount)
	}

	if appName == "" {
		common.LogWarn("Restore specified without app, assuming --all")

		if err := restorePrep(); err != nil {
			return err
		}
		return common.RunCommandAgainstAllApps(Restore, "restore", parallelCount)
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	return Restore(appName)
}

// CommandRetire ensures old containers are retired
func CommandRetire(appName string) error {
	lockFile := filepath.Join(common.GetDataDirectory("ps"), "retire")
	scheduler := ""
	if appName == "" {
		scheduler = common.GetGlobalScheduler()
	} else {
		scheduler = common.GetAppScheduler(appName)
	}

	fileLock := flock.New(lockFile)
	locked, err := fileLock.TryLock()
	if err != nil {
		return &RetireLockFailed{&err}
	}
	defer fileLock.Unlock()

	if !locked {
		return &RetireLockFailed{}
	}

	common.LogInfo1("Retiring old containers and images")
	_, err = common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "scheduler-retire",
		Args:        []string{scheduler, appName},
		StreamStdio: true,
	})
	if err != nil {
		return fmt.Errorf("Error retiring containers: %w", err)
	}

	common.LogInfo1("Retiring expired run and cron containers")
	_, err = common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "scheduler-run-retire",
		StreamStdio: true,
	})
	if err != nil {
		return fmt.Errorf("Error retiring expired run containers: %w", err)
	}

	return err
}

// CommandScaleInput is the input for the CommandScale function
type CommandScaleInput struct {
	// AppName is the name of the app to scale
	AppName string

	// Clear is a flag to reset the formation to the default scale
	Clear bool

	// Format is the format to display the formation in
	Format string

	// ProcessTuples is a list of process tuples to scale
	ProcessTuples []string

	// Replace is a flag to replace the formation with the specified process tuples
	Replace bool

	// SkipDeploy is a flag to skip the deploy phase
	SkipDeploy bool
}

// CommandScale gets or sets how many instances of a given process to run
func CommandScale(input CommandScaleInput) error {
	if err := common.VerifyAppName(input.AppName); err != nil {
		return err
	}

	if input.Clear && input.Replace {
		return errors.New("The --clear and --replace flags cannot be specified together")
	}

	if input.Clear && len(input.ProcessTuples) > 0 {
		return errors.New("Process types cannot be specified when using the --clear flag, use --replace to set the entire formation")
	}

	if input.Replace && len(input.ProcessTuples) == 0 {
		return errors.New("Must specify at least one process type when using the --replace flag, use --clear to reset the formation")
	}

	if !input.Clear && len(input.ProcessTuples) == 0 {
		return scaleReport(input.AppName, input.Format)
	}

	if !canScaleApp(input.AppName) {
		return fmt.Errorf("App %s contains an app.json file with a formations key and cannot be manually scaled", input.AppName)
	}

	processTuples := input.ProcessTuples
	if input.Clear {
		var err error
		processTuples, err = defaultProcessTuples(input.AppName)
		if err != nil {
			return err
		}

		common.LogInfo1(fmt.Sprintf("Resetting %s processes to the default scale", input.AppName))
	} else {
		common.LogInfo1(fmt.Sprintf("Scaling %s processes: %s", input.AppName, strings.Join(processTuples, " ")))
	}

	return scaleSet(scaleSetInput{
		appName:           input.AppName,
		skipDeploy:        input.SkipDeploy,
		clearExisting:     input.Clear || input.Replace,
		processTuples:     processTuples,
		deployOnlyChanged: true,
	})
}

// CommandSet sets or clears a ps property for an app
func CommandSet(appName string, property string, value string) error {
	if property == "restart-policy" && value != "" && !isValidRestartPolicy(value) {
		return errors.New("Invalid restart-policy specified")
	}

	common.CommandPropertySet("ps", appName, property, value, DefaultProperties, GlobalProperties)
	return nil
}

// CommandStart starts an app
func CommandStart(appName string, allApps bool, parallelCount int) error {
	if allApps {
		return common.RunCommandAgainstAllApps(Start, "start", parallelCount)
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	return Start(appName)
}

// CommandStop stops an app
func CommandStop(appName string, allApps bool, parallelCount int) error {
	if allApps {
		return common.RunCommandAgainstAllApps(Stop, "stop", parallelCount)
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	return Stop(appName)
}
