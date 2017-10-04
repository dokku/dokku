package main

import (
	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/network"
)

// rebuilds network settings for all apps
func main() {
	apps, err := common.DokkuApps()
	if err != nil {
		common.LogFail(err.Error())
	}
	for _, appName := range apps {
		network.BuildConfig(appName)
	}
}
