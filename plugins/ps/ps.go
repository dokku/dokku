package ps

import (
	"fmt"

	"github.com/dokku/dokku/plugins/common"
)

// RunInSerial is the default value for whether to run a command in parallel or not
// and defaults to -1 (false)
const RunInSerial = 0

var (
	// DefaultProperties is a map of all valid ps properties with corresponding default property values
	DefaultProperties = map[string]string{
		"restart-policy": "on-failure:10",
		"procfile-path":  "",
	}

	// GlobalProperties is a map of all valid global ps properties
	GlobalProperties = map[string]bool{
		"procfile-path": true,
	}
)

// RetireLockFailed wraps error to distinguish between a normal error
// and an error where the retire lock could not be fetched
type RetireLockFailed struct {
	Err *error
}

// ExitCode returns an exit code to use in case this error bubbles
// up into an os.Exit() call
func (err *RetireLockFailed) ExitCode() int {
	return 137
}

// Error returns a standard non-existent app error
func (err *RetireLockFailed) Error() string {
	if err.Err != nil {
		e := *err.Err
		return fmt.Sprintf("Failed to acquire ps:retire lock: %s", e.Error())
	}

	return fmt.Sprintf("Failed to acquire ps:retire lock")
}

// Formation contains scaling information for a given process type
type Formation struct {
	ProcessType string `json:"process_type"`
	Quantity    int    `json:"quantity"`
}

// FormationSlice contains a slice of Formations that can be sorted
type FormationSlice []*Formation

func (d FormationSlice) Len() int {
	return len(d)
}

func (d FormationSlice) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d FormationSlice) Less(i, j int) bool {
	return d[i].ProcessType < d[j].ProcessType
}

// Rebuild rebuilds app from base image
func Rebuild(appName string) error {
	_, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "receive-app",
		Args:        []string{appName},
		StreamStdio: true,
	})
	return err
}

// Restart restarts the app
func Restart(appName string) error {
	if !common.IsDeployed(appName) {
		common.LogWarn(fmt.Sprintf("App %s has not been deployed", appName))
		return nil
	}

	imageTag, err := common.GetRunningImageTag(appName, "")
	if err != nil {
		return err
	}

	if imageTag == "" {
		common.LogWarn("No deployed-image-tag property saved, falling back to full release-and-deploy")
		_, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
			Trigger:     "release-and-deploy",
			Args:        []string{appName},
			StreamStdio: true,
		})
		return err
	}

	_, err = common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "deploy",
		Args:        []string{appName, imageTag},
		StreamStdio: true,
	})
	return err
}

// RestartProcess restarts a process type within an app
func RestartProcess(appName string, processName string) error {
	if !common.IsDeployed(appName) {
		common.LogWarn(fmt.Sprintf("App %s has not been deployed", appName))
		return nil
	}

	imageTag, err := common.GetRunningImageTag(appName, "")
	if err != nil {
		return err
	}

	if imageTag == "" {
		common.LogWarn("No deployed-image-tag property saved, falling back to full release-and-deploy")
		_, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
			Trigger:     "release-and-deploy",
			Args:        []string{appName},
			StreamStdio: true,
		})
		return err
	}

	_, err = common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "deploy",
		Args:        []string{appName, imageTag, processName},
		StreamStdio: true,
	})
	return err
}

// Restore ensures an app that should be running is running on boot
func Restore(appName string) error {
	scheduler := common.GetAppScheduler(appName)
	_, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "scheduler-pre-restore",
		Args:        []string{scheduler, appName},
		StreamStdio: true,
	})
	if err != nil {
		return fmt.Errorf("Error running scheduler-pre-restore: %s", err)
	}

	common.LogInfo1("Clearing potentially invalid proxy configuration")
	_, err = common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "proxy-clear-config",
		Args:        []string{appName},
		StreamStdio: true,
	})
	if err != nil {
		common.LogWarn(fmt.Sprintf("Error clearing proxy config: %s", err))
	}

	if !common.IsDeployed(appName) {
		common.LogWarn(fmt.Sprintf("App %s has not been deployed", appName))
		return nil
	}

	results, _ := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "config-get",
		Args:    []string{appName, "DOKKU_APP_RESTORE"},
	})
	restore := results.StdoutContents()
	if restore == "0" {
		common.LogWarn(fmt.Sprintf("Skipping ps:restore for %s as DOKKU_APP_RESTORE=%s", appName, restore))
		return nil
	}

	common.LogInfo1("Starting app")
	return Start(appName)
}

// Start starts the app
func Start(appName string) error {
	imageTag, _ := common.GetRunningImageTag(appName, "")

	if !common.IsDeployed(appName) {
		common.LogWarn(fmt.Sprintf("App %s has not been deployed", appName))
		return nil
	}

	_, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "pre-start",
		Args:        []string{appName},
		StreamStdio: true,
	})
	if err != nil {
		return fmt.Errorf("Failure in pre-start hook: %s", err)
	}

	runningState := getRunningState(appName)

	if runningState == "mixed" {
		common.LogWarn("App is running in mixed mode, releasing")
	} else if runningState == "false" {
		common.LogWarn("App has been detected as not running, releasing")
	}

	if runningState != "true" {
		_, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
			Trigger:     "release-and-deploy",
			Args:        []string{appName, imageTag},
			StreamStdio: true,
		})
		if err != nil {
			return err
		}
	} else {
		common.LogWarn(fmt.Sprintf("App %s already running", appName))
	}

	_, err = common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "proxy-build-config",
		Args:        []string{appName},
		StreamStdio: true,
	})
	return err
}

// Stop stops the app
func Stop(appName string) error {
	if !common.IsDeployed(appName) {
		common.LogWarn(fmt.Sprintf("App %s has not been deployed", appName))
		return nil
	}

	common.LogInfo1Quiet(fmt.Sprintf("Stopping %s", appName))
	scheduler := common.GetAppScheduler(appName)

	_, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "scheduler-stop",
		Args:        []string{scheduler, appName},
		StreamStdio: true,
	})
	if err != nil {
		return err
	}

	_, err = common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "post-stop",
		Args:        []string{appName},
		StreamStdio: true,
	})
	return err
}
