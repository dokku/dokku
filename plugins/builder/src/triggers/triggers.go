package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/builder"
	"github.com/dokku/dokku/plugins/common"
)

// main entrypoint to all triggers
func main() {
	parts := strings.Split(os.Args[0], "/")
	trigger := parts[len(parts)-1]
	flag.Parse()

	var err error
	switch trigger {
	case "builder-detect":
		appName := flag.Arg(0)
		err = builder.TriggerBuilderDetect(appName)
	case "builder-get-property":
		appName := flag.Arg(0)
		property := flag.Arg(1)
		err = builder.TriggerBuilderGetProperty(appName, property)
	case "builder-image-is-cnb":
		appName := flag.Arg(0)
		image := flag.Arg(1)
		err = builder.TriggerBuilderImageIsCNB(appName, image)
	case "builder-image-is-herokuish":
		appName := flag.Arg(0)
		image := flag.Arg(1)
		err = builder.TriggerBuilderImageIsHerokuish(appName, image)
	case "core-post-extract":
		appName := flag.Arg(0)
		sourceWorkDir := flag.Arg(1)
		err = builder.TriggerCorePostExtract(appName, sourceWorkDir)
	case "install":
		err = builder.TriggerInstall()
	case "post-app-clone-setup":
		oldAppName := flag.Arg(0)
		newAppName := flag.Arg(1)
		err = builder.TriggerPostAppCloneSetup(oldAppName, newAppName)
	case "post-app-rename-setup":
		oldAppName := flag.Arg(0)
		newAppName := flag.Arg(1)
		err = builder.TriggerPostAppRenameSetup(oldAppName, newAppName)
	case "post-delete":
		appName := flag.Arg(0)
		err = builder.TriggerPostDelete(appName)
	case "report":
		appName := flag.Arg(0)
		err = builder.ReportSingleApp(appName, "", "")
	default:
		err = fmt.Errorf("Invalid plugin trigger call: %s", trigger)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
