package proxy

import (
	"errors"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
)

// CommandDisable disables the proxy for app via command line
func CommandDisable(appName string, skipRestart bool) error {
	if appName == "" {
		return errors.New("Please specify an app to run the command on")
	}

	if !IsAppProxyEnabled(appName) {
		common.LogInfo1("Proxy is already disable for app")
		return nil
	}

	common.LogInfo1("Disabling proxy for app")
	entries := map[string]string{
		"DOKKU_DISABLE_PROXY": "1",
	}

	if err := config.SetMany(appName, entries, false); err != nil {
		return err
	}

	return common.PlugnTrigger("proxy-disable", []string{appName}...)
}

// CommandEnable enables the proxy for app via command line
func CommandEnable(appName string) error {
	if appName == "" {
		return errors.New("Please specify an app to run the command on")
	}

	if IsAppProxyEnabled(appName) {
		common.LogInfo1("Proxy is already enabled for app")
		return nil
	}

	common.LogInfo1("Enabling proxy for app")
	keys := []string{"DOKKU_DISABLE_PROXY"}
	if err := config.UnsetMany(appName, keys, false); err != nil {
		return err
	}

	return common.PlugnTrigger("proxy-enable", []string{appName}...)
}

// CommandPorts is a cmd wrapper to list proxy port mappings for an app
func CommandPorts(appName string) error {
	if appName == "" {
		return errors.New("Please specify an app to run the command on")
	}

	return listAppProxyPorts(appName)
}

// CommandPortsAdd adds proxy port mappings to an app
func CommandPortsAdd(appName string, portMaps []string) error {
	if appName == "" {
		return errors.New("Please specify an app to run the command on")
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
	if appName == "" {
		return errors.New("Please specify an app to run the command on")
	}

	keys := []string{"DOKKU_PROXY_PORT_MAP"}
	if err := config.UnsetMany(appName, keys, false); err != nil {
		return err
	}

	return common.PlugnTrigger("post-proxy-ports-update", []string{appName, "clear"}...)
}

// CommandPortsRemove removes specific proxy port mappings from an app
func CommandPortsRemove(appName string, portMaps []string) error {
	if appName == "" {
		return errors.New("Please specify an app to run the command on")
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
	if appName == "" {
		return errors.New("Please specify an app to run the command on")
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
func CommandReport(appName string, infoFlag string) error {
	if strings.HasPrefix(appName, "--") {
		infoFlag = appName
		appName = ""
	}

	if len(appName) == 0 {
		apps, err := common.DokkuApps()
		if err != nil {
			return err
		}
		for _, appName := range apps {
			if err := ReportSingleApp(appName, infoFlag); err != nil {
				return err
			}
		}
		return nil
	}

	return ReportSingleApp(appName, infoFlag)
}

// CommandSet sets a proxy for an app
func CommandSet(appName string, proxyType string) error {
	if appName == "" {
		return errors.New("Please specify an app to run the command on")
	}

	if len(proxyType) < 2 {
		return errors.New("Please specify a proxy type")
	}

	entries := map[string]string{
		"DOKKU_APP_PROXY_TYPE": proxyType,
	}
	return config.SetMany(appName, entries, false)
}
