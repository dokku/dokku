package main

import (
	"flag"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/resource"
)

// displays a resource report for one or more apps
func main() {
	flag.Parse()
	appName := flag.Arg(1)
	infoFlag := flag.Arg(2)

	if strings.HasPrefix(appName, "--") {
		common.LogFail("The resource:report command does not support flags without an app name")
	}

	if len(appName) == 0 {
		apps, err := common.DokkuApps()
		if err != nil {
			return
		}
		for _, appName := range apps {
			resource.ReportSingleApp(appName, infoFlag)
		}
		return
	}

	resource.ReportSingleApp(appName, infoFlag)
}
