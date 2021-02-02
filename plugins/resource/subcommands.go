package resource

import (
	"github.com/dokku/dokku/plugins/common"
)

// CommandLimit implements resource:limit
func CommandLimit(appName string, processType string, r Resource) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	return setResourceType(appName, processType, r, "limit")
}

// CommandLimitClear implements resource:limit-clear
func CommandLimitClear(appName string, processType string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	clearByResourceType(appName, processType, "limit")
	return nil
}

// CommandReport displays a resource report for one or more apps
func CommandReport(appName string, format string, infoFlag string) error {
	if len(appName) == 0 {
		apps, err := common.DokkuApps()
		if err != nil {
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

// CommandReserve implements resource:reserve
func CommandReserve(appName string, processType string, r Resource) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	return setResourceType(appName, processType, r, "reserve")
}

// CommandReserveClear implements resource:reserve-clear
func CommandReserveClear(appName string, processType string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	clearByResourceType(appName, processType, "reserve")
	return nil
}
