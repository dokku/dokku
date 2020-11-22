package logs

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/dokku/dokku/plugins/common"
)

// CommandDefault displays recent log output
func CommandDefault(appName string, num int64, process string, tail, quiet bool) error {
	if appName == "" {
		return errors.New("Please specify an app to run the command on")
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if !common.IsDeployed(appName) {
		return fmt.Errorf("App %s has not been deployed", appName)
	}

	s := common.GetAppScheduler(appName)
	t := strconv.FormatBool(tail)
	q := strconv.FormatBool(quiet)
	n := strconv.FormatInt(num, 10)
	if err := common.PlugnTrigger("scheduler-logs", []string{s, appName, process, t, q, n}...); err != nil {
		return err
	}
	return nil
}

// CommandFailed shows the last failed deploy logs
func CommandFailed(appName string, allApps bool) error {
	if allApps {
		return common.RunCommandAgainstAllAppsSerially(GetFailedLogs, "logs:failed")
	}

	if appName == "" {
		common.LogWarn("Restore specified without app, assuming --all")
		return common.RunCommandAgainstAllAppsSerially(GetFailedLogs, "logs:failed")
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	return GetFailedLogs(appName)
}
