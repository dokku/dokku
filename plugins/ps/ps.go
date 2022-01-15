package ps

import (
	"fmt"
	"strings"

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
	return common.PlugnTrigger("receive-app", []string{appName}...)
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
		return common.PlugnTrigger("release-and-deploy", []string{appName}...)
	}

	return common.PlugnTrigger("deploy", []string{appName, imageTag}...)
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
		return common.PlugnTrigger("release-and-deploy", []string{appName}...)
	}

	return common.PlugnTrigger("deploy", []string{appName, imageTag, processName}...)
}

// Restore ensures an app that should be running is running on boot
func Restore(appName string) error {
	scheduler := common.GetAppScheduler(appName)
	if err := common.PlugnTrigger("pre-restore", []string{scheduler, appName}...); err != nil {
		return fmt.Errorf("Error running pre-restore: %s", err)
	}

	common.LogInfo1("Clearing potentially invalid proxy configuration")
	if err := common.PlugnTrigger("proxy-clear-config", []string{appName}...); err != nil {
		return fmt.Errorf("Error clearing proxy config: %s", err)
	}

	if !common.IsDeployed(appName) {
		common.LogWarn(fmt.Sprintf("App %s has not been deployed", appName))
		return nil
	}

	b, _ := common.PlugnTriggerOutput("config-get", []string{appName, "DOKKU_APP_RESTORE"}...)
	restore := strings.TrimSpace(string(b[:]))
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

	if err := common.PlugnTrigger("pre-start", []string{appName}...); err != nil {
		return fmt.Errorf("Failure in pre-start hook: %s", err)
	}

	runningState := getRunningState(appName)

	if runningState == "mixed" {
		common.LogWarn("App is running in mixed mode, releasing")
	} else if runningState == "false" {
		common.LogWarn("App has been detected as not running, releasing")
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
