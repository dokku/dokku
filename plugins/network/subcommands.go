package network

import (
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// CommandRebuildall rebuilds network settings for all apps
func CommandRebuildall() {
	apps, err := common.DokkuApps()
	if err != nil {
		common.LogFail(err.Error())
	}
	for _, appName := range apps {
		BuildConfig(appName)
	}
}

// CommandReport displays a network report for one or more apps
func CommandReport(appName string, infoFlag string) {
	if strings.HasPrefix(appName, "--") {
		infoFlag = appName
		appName = ""
	}

	if len(appName) == 0 {
		apps, err := common.DokkuApps()
		if err != nil {
			return
		}
		for _, appName := range apps {
			ReportSingleApp(appName, infoFlag)
		}
		return
	}

	ReportSingleApp(appName, infoFlag)
}

// CommandSet set or clear a network property for an app
func CommandSet(appName string, property string, value string) {
	if property == "bind-all-interfaces" && value == "" {
		value = "false"
	}

	common.CommandPropertySet("network", appName, property, value, DefaultProperties)
}
