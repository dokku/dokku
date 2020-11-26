package logs

import (
	"fmt"

	"github.com/dokku/dokku/plugins/common"
)

// GetFailedLogs outputs failed deploy logs for a given app
func GetFailedLogs(appName string) error {
	common.LogInfo2Quiet(fmt.Sprintf("%s failed deploy logs", appName))
	s := common.GetAppScheduler(appName)
	if err := common.PlugnTrigger("scheduler-logs-failed", s, appName); err != nil {
		return err
	}
	return nil
}
