package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/buildpacks"
	"github.com/dokku/dokku/plugins/common"
)

// main entrypoint to all triggers
func main() {
	parts := strings.Split(os.Args[0], "/")
	trigger := parts[len(parts)-1]
	flag.Parse()

	var err error
	switch trigger {
	case "install":
		buildpacks.TriggerInstall()
	case "post-app-clone-setup":
		oldAppName := flag.Arg(0)
		newAppName := flag.Arg(1)
		err = buildpacks.TriggerPostAppCloneSetup(oldAppName, newAppName)
	case "post-app-rename-setup":
		oldAppName := flag.Arg(0)
		newAppName := flag.Arg(1)
		err = buildpacks.TriggerPostAppRenameSetup(oldAppName, newAppName)
	case "post-delete":
		appName := flag.Arg(0)
		buildpacks.TriggerPostDelete(appName)
	case "post-extract":
		appName := flag.Arg(0)
		sourceWorkDir := flag.Arg(1)
		buildpacks.TriggerPostExtract(appName, sourceWorkDir)
	case "report":
		appName := flag.Arg(0)
		buildpacks.ReportSingleApp(appName, "")
	default:
		common.LogFail(fmt.Sprintf("Invalid plugin trigger call: %s", trigger))
	}

	if err != nil {
		common.LogFail(err.Error())
	}
}
