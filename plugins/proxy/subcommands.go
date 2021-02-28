package proxy

import (
	"errors"
	"strings"

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

// CommandPorts is a cmd wrapper to list proxy port mappings for an app
func CommandPorts(appName string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	return listAppProxyPorts(appName)
}

// CommandPortsAdd adds proxy port mappings to an app
func CommandPortsAdd(appName string, portMaps []string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if len(portMaps) == 0 {
		return errors.New("No port mapping specified")
	}

	proxyPortMap, err := parseProxyPortMapString(strings.Join(portMaps, " "))
	if err != nil {
		return err
	}

	if err := addProxyPorts(appName, proxyPortMap); err != nil {
		return err
	}

	return common.PlugnTrigger("post-proxy-ports-update", []string{appName, "add"}...)
}

// CommandPortsClear clears all proxy port mappings for an app
func CommandPortsClear(appName string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	keys := []string{"DOKKU_PROXY_PORT_MAP"}
	if err := config.UnsetMany(appName, keys, false); err != nil {
		return err
	}

	return common.PlugnTrigger("post-proxy-ports-update", []string{appName, "clear"}...)
}

// CommandPortsRemove removes specific proxy port mappings from an app
func CommandPortsRemove(appName string, portMaps []string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if len(portMaps) == 0 {
		return errors.New("No port mapping specified")
	}

	proxyPortMap, err := parseProxyPortMapString(strings.Join(portMaps, " "))
	if err != nil {
		return err
	}

	if err := removeProxyPorts(appName, proxyPortMap); err != nil {
		return err
	}

	return common.PlugnTrigger("post-proxy-ports-update", []string{appName, "remove"}...)
}

// CommandPortsSet sets proxy port mappings for an app
func CommandPortsSet(appName string, portMaps []string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if len(portMaps) == 0 {
		return errors.New("No port mapping specified")
	}

	proxyPortMap, err := parseProxyPortMapString(strings.Join(portMaps, " "))
	if err != nil {
		return err
	}

	if err := setProxyPorts(appName, proxyPortMap); err != nil {
		return err
	}

	return common.PlugnTrigger("post-proxy-ports-update", []string{appName, "set"}...)
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
