package proxy

import (
	"github.com/dokku/dokku/plugins/config"
)

func getAppProxyType(appName string) string {
	return config.GetWithDefault(appName, "DOKKU_APP_PROXY_TYPE", "")
}

func getComputedProxyType(appName string) string {
	proxyType := getGlobalProxyType()
	if proxyType == "" {
		proxyType = getAppProxyType(appName)
	}

	return proxyType
}

func getGlobalProxyType() string {
	return config.GetWithDefault("--global", "DOKKU_PROXY_TYPE", "nginx")
}
