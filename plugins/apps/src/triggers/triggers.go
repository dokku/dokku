package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/apps"
	"github.com/dokku/dokku/plugins/common"
)

// main entrypoint to all triggers
func main() {
	parts := strings.Split(os.Args[0], "/")
	trigger := parts[len(parts)-1]
	flag.Parse()

	var err error
	switch trigger {
	case "app-create":
		appName := flag.Arg(0)
		err = apps.TriggerAppCreate(appName)
	case "app-destroy":
		appName := flag.Arg(0)
		err = apps.TriggerAppDestroy(appName)
	case "app-exists":
		appName := flag.Arg(0)
		err = apps.TriggerAppExists(appName)
	case "app-maybe-create":
		appName := flag.Arg(0)
		err = apps.TriggerAppMaybeCreate(appName)
	case "deploy-source-set":
		appName := flag.Arg(0)
		sourceType := flag.Arg(1)
		sourceMetadata := flag.Arg(2)
		err = apps.TriggerDeploySourceSet(appName, sourceType, sourceMetadata)
	case "install":
		err = apps.TriggerInstall()
	case "post-app-clone-setup":
		oldAppName := flag.Arg(0)
		newAppName := flag.Arg(1)
		err = apps.TriggerPostAppCloneSetup(oldAppName, newAppName)
	case "post-app-rename-setup":
		oldAppName := flag.Arg(0)
		newAppName := flag.Arg(1)
		err = apps.TriggerPostAppRenameSetup(oldAppName, newAppName)
	case "post-delete":
		appName := flag.Arg(0)
		err = apps.TriggerPostDelete(appName)
	case "report":
		appName := flag.Arg(0)
		err = apps.ReportSingleApp(appName, "", "")
	default:
		err = fmt.Errorf("Invalid plugin trigger call: %s", trigger)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
