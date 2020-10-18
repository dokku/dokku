package logs

import (
	"errors"
	"fmt"
	"github.com/dokku/dokku/plugins/common"
	"strconv"
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
func CommandFailed(appName string) error {
	if appName == "" {
		apps, err := common.DokkuApps()
		if err != nil {
			return err
		}
		for _, appName := range apps {
			_ = failedSingle(appName)
		}
		return nil
	}
	return failedSingle(appName)
}

func failedSingle(appName string) error {
	common.LogInfo2Quiet(fmt.Sprintf("%s failed deploy logs", appName))
	s := common.GetAppScheduler(appName)
	if err := common.PlugnTrigger("scheduler-logs-failed", s, appName); err != nil {
		return err
	}
	return nil
}
