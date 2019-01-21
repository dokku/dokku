package main

import (
	"fmt"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/proxy"
)

// runs the install step for the network plugin
func main() {
	if err := common.PropertySetup("network"); err != nil {
		common.LogFail(fmt.Sprintf("Unable to install the network plugin: %s", err.Error()))
	}

	apps, err := common.DokkuApps()
	if err != nil {
		return
	}
	for _, appName := range apps {
		if common.PropertyExists("network", appName, "bind-all-interfaces") {
			continue
		}
		if proxy.IsAppProxyEnabled(appName) {
			common.LogVerboseQuiet("Setting %s network property 'bind-all-interfaces' to false")
			if err := common.PropertyWrite("network", appName, "bind-all-interfaces", "false"); err != nil {
				common.LogWarn(err.Error())
			}
		} else {
			common.LogVerboseQuiet("Setting %s network property 'bind-all-interfaces' to true")
			if err := common.PropertyWrite("network", appName, "bind-all-interfaces", "true"); err != nil {
				common.LogWarn(err.Error())
			}
		}
	}
}
