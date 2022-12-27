package proxy

import (
	"github.com/dokku/dokku/plugins/config"
)

func getAppProxyType(appName string) string {
	return config.GetWithDefault(appName, "DOKKU_APP_PROXY_TYPE", "nginx")
}
