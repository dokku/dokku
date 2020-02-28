package proxy

import (
	"errors"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
)

// CommandDisable disables the proxy for app via command line
func CommandDisable(args []string) error {
	appName, err := getAppName(args)
	if err != nil {
		return err
	}

	if !IsAppProxyEnabled(appName) {
		common.LogInfo1("Proxy is already disable for app")
		return nil
	}

	common.LogInfo1("Disabling proxy for app")
	restart := true
	if len(args) >= 2 && args[1] == "--no-restart" {
		restart = false
	}

	entries := map[string]string{
		"DOKKU_DISABLE_PROXY": "1",
	}

	if err := config.SetMany(appName, entries, restart); err != nil {
		return err
	}

	return common.PlugnTrigger("proxy-disable", []string{appName}...)
}

// CommandEnable enables the proxy for app via command line
func CommandEnable(args []string) error {
	appName, err := getAppName(args)
	if err != nil {
		return err
	}

	if IsAppProxyEnabled(appName) {
		common.LogInfo1("Proxy is already enabled for app")
		return nil
	}

	common.LogInfo1("Enabling proxy for app")
	restart := true
	if len(args) >= 2 && args[1] == "--no-restart" {
		restart = false
	}

	keys := []string{"DOKKU_DISABLE_PROXY"}
	if err := config.UnsetMany(appName, keys, restart); err != nil {
		return err
	}

	return common.PlugnTrigger("proxy-enable", []string{appName}...)
}

// CommandPorts is a cmd wrapper to list proxy port mappings for an app
func CommandPorts(args []string) error {
	appName, err := getAppName(args)
	if err != nil {
		return err
	}

	return listAppProxyPorts(appName)
}

// CommandPortsAdd adds proxy port mappings to an app
func CommandPortsAdd(args []string) error {
	appName, err := getAppName(args)
	if err != nil {
		return err
	}

	if len(args) == 1 {
		return errors.New("No port mapping specified")
	}

	stringPortMap := strings.Join(args[1:], " ")
	proxyPortMap, err := parseProxyPortMapString(stringPortMap)
	if err != nil {
		return err
	}

	if err := addProxyPorts(appName, proxyPortMap); err != nil {
		return err
	}

	return common.PlugnTrigger("post-proxy-ports-update", []string{appName, "add"}...)
}

// CommandPortsClear clears all proxy port mappings for an app
func CommandPortsClear(args []string) error {
	appName, err := getAppName(args)
	if err != nil {
		return err
	}

	keys := []string{"DOKKU_PROXY_PORT_MAP"}
	err = config.UnsetMany(appName, keys, false)
	if err != nil {
		return err
	}

	return common.PlugnTrigger("post-proxy-ports-update", []string{appName, "clear"}...)
}

// CommandPortsRemove removes specific proxy port mappings from an app
func CommandPortsRemove(args []string) error {
	appName, err := getAppName(args)
	if err != nil {
		return err
	}

	if len(args) == 1 {
		return errors.New("No port mapping specified")
	}

	stringPortMap := strings.Join(args[1:], " ")
	proxyPortMap, err := parseProxyPortMapString(stringPortMap)
	if err != nil {
		return err
	}

	if err := removeProxyPorts(appName, proxyPortMap); err != nil {
		return err
	}

	return common.PlugnTrigger("post-proxy-ports-update", []string{appName, "remove"}...)
}

// CommandPortsSet sets proxy port mappings for an app
func CommandPortsSet(args []string) error {
	appName, err := getAppName(args)
	if err != nil {
		return err
	}

	if len(args) == 1 {
		return errors.New("No port mapping specified")
	}

	stringPortMap := strings.Join(args[1:], " ")
	proxyPortMap, err := parseProxyPortMapString(stringPortMap)
	if err != nil {
		return err
	}

	if err := setProxyPorts(appName, proxyPortMap); err != nil {
		return err
	}

	return common.PlugnTrigger("post-proxy-ports-update", []string{appName, "set"}...)
}

// CommandReport displays a proxy report for one or more apps
func CommandReport(args []string) error {
	appName, err := getAppName(args)
	if err != nil {
		return err
	}

	infoFlag := ""
	if len(args) > 1 {
		infoFlag = args[1]
	}

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
func CommandSet(args []string) error {
	appName, err := getAppName(args)
	if err != nil {
		return err
	}

	if len(appName) < 2 {
		return errors.New("Please specify a proxy type")
	}

	proxyType := args[1]
	entries := map[string]string{
		"DOKKU_APP_PROXY_TYPE": proxyType,
	}
	return config.SetMany(appName, entries, false)
}
