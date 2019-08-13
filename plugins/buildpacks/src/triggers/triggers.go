package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dokku/dokku/plugins/buildpacks"
	"github.com/dokku/dokku/plugins/common"
)

// handles all plugin triggers
func main() {
	flag.Parse()

	triggerName := os.Args[0]
	if triggerName == "install" {
		// runs the install step for the buildpacks plugin
		if err := common.PropertySetup("buildpacks"); err != nil {
			common.LogFail(fmt.Sprintf("Unable to install the buildpacks plugin: %s", err.Error()))
		}
	} else if triggerName == "post-delete" {
		// destroys the buildpacks property for a given app container
		appName := flag.Arg(0)

		err := common.PropertyDestroy("buildpacks", appName)
		if err != nil {
			common.LogFail(err.Error())
		}
	} else if triggerName == "post-extract" {
		// writes a .buildpacks file into the app
		appName := flag.Arg(0)
		tmpWorkDir := flag.Arg(1)
		buildpacks.PostExtract(appName, tmpWorkDir)
	} else if triggerName == "report" {
		// displays a buildpacks report for one or more apps
		appName := flag.Arg(0)
		buildpacks.ReportSingleApp(appName, "")
	}
}
