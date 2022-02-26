package builder

import (
	"github.com/dokku/dokku/plugins/common"
)

// CommandReport displays a builder report for one or more apps
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

// CommandSet set or clear a builder property for an app
func CommandSet(appName string, property string, value string) error {
	common.CommandPropertySet("builder", appName, property, value, DefaultProperties, GlobalProperties)
	return nil
}
