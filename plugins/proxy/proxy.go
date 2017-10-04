package proxy

import (
	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
)

// IsAppProxyEnabled returns true if proxy is enabled; otherwise return false
func IsAppProxyEnabled(appName string) bool {
	err := common.VerifyAppName(appName)
	if err != nil {
		common.LogFail(err.Error())
	}

	proxyEnabled := true
	disableProxy := config.GetWithDefault(appName, "DOKKU_DISABLE_PROXY", "")
	if disableProxy != "" {
		proxyEnabled = false
	}
	return proxyEnabled
}
