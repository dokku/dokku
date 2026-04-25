package proxy

import (
	"errors"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// CommandBuildConfig rebuilds config for a given app
func CommandBuildConfig(appName string, allApps bool, parallelCount int) error {
	if allApps {
		return common.RunCommandAgainstAllApps(BuildConfig, "build-config", parallelCount)
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	return BuildConfig(appName)
}

// CommandClearConfig clears config for a given app
func CommandClearConfig(appName string, allApps bool) error {
	if allApps {
		return ClearConfig("--all")
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	return ClearConfig(appName)
}

// CommandDisable disables the proxy for app via command line
func CommandDisable(appName string, allApps bool, parallelCount int) error {
	if allApps {
		return common.RunCommandAgainstAllApps(Disable, "disable", parallelCount)
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	return Disable(appName)
}

// CommandEnable enables the proxy for app via command line
func CommandEnable(appName string, allApps bool, parallelCount int) error {
	if allApps {
		return common.RunCommandAgainstAllApps(Enable, "enable", parallelCount)
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	return Enable(appName)
}

// CommandReport displays a proxy report for one or more apps
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

// CommandSet set or clear a proxy property for an app
func CommandSet(appName string, property string, value string) error {
	// backward compatibility: `proxy:set <app> <proxy-type>` is treated as
	// `proxy:set <app> type <proxy-type>`
	if _, ok := DefaultProperties[property]; !ok && value == "" {
		if strings.Contains(property, ":") {
			common.LogWarn("Detected potential port mapping instead of proxy type")
			return errors.New("Consider using ports:set command or specifying a valid proxy")
		}
		value = property
		property = "type"
	}

	common.CommandPropertySet("proxy", appName, property, value, DefaultProperties, GlobalProperties)
	return nil
}
