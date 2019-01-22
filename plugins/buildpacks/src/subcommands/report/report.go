package main

import (
	"flag"
	"strings"

	"github.com/dokku/dokku/plugins/buildpacks"
	"github.com/dokku/dokku/plugins/common"
)

// displays a buildpacks report for one or more apps
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
			buildpacks.ReportSingleApp(appName, infoFlag)
		}
		return
	}

	buildpacks.ReportSingleApp(appName, infoFlag)
}
