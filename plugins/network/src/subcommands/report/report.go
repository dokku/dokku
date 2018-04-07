package main

import (
	"flag"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/network"
)

// displays a network report for one or more apps
func main() {
	flag.Parse()
	appName := flag.Arg(1)
	infoFlag := flag.Arg(2)

	if strings.HasPrefix(appName, "--") {
		infoFlag = appName
		appName = ""
	}

	if len(appName) == 0 {
		apps, err := common.DokkuApps()
		if err != nil {
			return
		}
		for _, appName := range apps {
			network.ReportSingleApp(appName, infoFlag)
		}
		return
	}

	network.ReportSingleApp(appName, infoFlag)
}
