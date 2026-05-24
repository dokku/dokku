package proxy

import (
	"github.com/dokku/dokku/plugins/common"
)

func getAppProxyType(appName string) string {
	return common.PropertyGet("proxy", appName, "type")
}

func getComputedProxyType(appName string) string {
	proxyType := getAppProxyType(appName)
	if proxyType == "" {
		proxyType = getGlobalProxyType()
	}
	if proxyType == "" {
		proxyType = "nginx"
	}

	return proxyType
}

func getGlobalProxyType() string {
	return common.PropertyGet("proxy", "--global", "type")
}
