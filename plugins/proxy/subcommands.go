package proxy

import (
	"errors"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
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

// CommandSet sets a proxy for an app
func CommandSet(appName string, proxyType string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if len(proxyType) < 2 {
		return errors.New("Please specify a proxy type")
	}

	entries := map[string]string{
		"DOKKU_APP_PROXY_TYPE": proxyType,
	}
	return config.SetMany(appName, entries, false)
}
