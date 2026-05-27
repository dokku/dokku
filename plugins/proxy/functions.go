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

func getAppProxyPort(appName string) string {
	return common.PropertyGet("proxy", appName, "proxy-port")
}

func getGlobalProxyPort() string {
	return common.PropertyGet("proxy", "--global", "proxy-port")
}

func getComputedProxyPort(appName string) string {
	value := getAppProxyPort(appName)
	if value == "" {
		value = getGlobalProxyPort()
	}

	return value
}

func getAppProxySSLPort(appName string) string {
	return common.PropertyGet("proxy", appName, "proxy-ssl-port")
}

func getGlobalProxySSLPort() string {
	return common.PropertyGet("proxy", "--global", "proxy-ssl-port")
}

func getComputedProxySSLPort(appName string) string {
	value := getAppProxySSLPort(appName)
	if value == "" {
		value = getGlobalProxySSLPort()
	}

	return value
}

func getAppDisabled(appName string) string {
	return common.PropertyGet("proxy", appName, "disabled")
}

func getComputedDisabled(appName string) string {
	value := getAppDisabled(appName)
	if value == "" {
		value = "false"
	}

	return value
}
