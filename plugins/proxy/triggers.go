package proxy

import (
	"fmt"

	"github.com/dokku/dokku/plugins/common"
)

// TriggerInstall runs the install step for the proxy plugin
func TriggerInstall() error {
	if err := common.PropertySetup("proxy"); err != nil {
		return fmt.Errorf("Unable to install the proxy plugin: %s", err.Error())
	}

	if err := common.MigrateConfigToProperties("proxy", []common.MigrateConfigEntry{
		{
			ConfigVar:       "DOKKU_APP_PROXY_TYPE",
			GlobalConfigVar: "DOKKU_PROXY_TYPE",
			Property:        "type",
		},
		{
			ConfigVar: "DOKKU_DISABLE_PROXY",
			Property:  "disabled",
			Transform: func(value string) string {
				if value != "" {
					return "true"
				}
				return value
			},
		},
		{
			ConfigVar:       "DOKKU_PROXY_PORT",
			GlobalConfigVar: "DOKKU_PROXY_PORT",
			Property:        "proxy-port",
		},
		{
			ConfigVar:       "DOKKU_PROXY_SSL_PORT",
			GlobalConfigVar: "DOKKU_PROXY_SSL_PORT",
			Property:        "proxy-ssl-port",
		},
	}); err != nil {
		return err
	}

	return nil
}

// TriggerProxyIsEnabled prints true or false depending on whether the proxy is enabled
func TriggerProxyIsEnabled(appName string) error {
	if IsAppProxyEnabled(appName) {
		fmt.Println("true")
	} else {
		fmt.Println("false")
	}

	return nil
}

// TriggerProxyType prints out the current proxy type, defaulting to nginx
func TriggerProxyType(appName string) error {
	proxyType := getComputedProxyType(appName)
	fmt.Println(proxyType)

	return nil
}

// TriggerProxyRoutesList prints one "process|port|path|strip" line per route,
// sorted by descending path length. Backends use this to render path-based
// routes in their config or labels.
func TriggerProxyRoutesList(appName string) error {
	routes, err := GetRoutes(appName)
	if err != nil {
		return err
	}
	for _, r := range routes {
		strip := "0"
		if r.StripPrefix {
			strip = "1"
		}
		fmt.Printf("%s|%d|%s|%s\n", r.Process, r.Port, r.Path, strip)
	}
	return nil
}

// TriggerPostAppCloneSetup creates new proxy files
func TriggerPostAppCloneSetup(oldAppName string, newAppName string) error {
	return common.PropertyClone("proxy", oldAppName, newAppName)
}

// TriggerPostAppRenameSetup renames proxy files
func TriggerPostAppRenameSetup(oldAppName string, newAppName string) error {
	if err := common.PropertyClone("proxy", oldAppName, newAppName); err != nil {
		return err
	}

	return common.PropertyDestroy("proxy", oldAppName)
}

// TriggerPostDelete destroys the proxy property for a given app container
func TriggerPostDelete(appName string) error {
	return common.PropertyDestroy("proxy", appName)
}
