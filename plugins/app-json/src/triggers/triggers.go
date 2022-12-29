package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	appjson "github.com/dokku/dokku/plugins/app-json"
	"github.com/dokku/dokku/plugins/common"
)

// main entrypoint to all triggers
func main() {
	parts := strings.Split(os.Args[0], "/")
	trigger := parts[len(parts)-1]
	flag.Parse()

	var err error
	switch trigger {
	case "app-json-process-deploy-parallelism":
		appName := flag.Arg(0)
		processType := flag.Arg(1)
		err = appjson.TriggerAppJSONProcessDeployParallelism(appName, processType)
	case "core-post-deploy":
		appName := flag.Arg(0)
		err = appjson.TriggerCorePostDeploy(appName)
	case "core-post-extract":
		appName := flag.Arg(0)
		sourceWorkDir := flag.Arg(1)
		err = appjson.TriggerCorePostExtract(appName, sourceWorkDir)
	case "install":
		err = appjson.TriggerInstall()
	case "post-app-clone-setup":
		oldAppName := flag.Arg(0)
		newAppName := flag.Arg(1)
		err = appjson.TriggerPostAppCloneSetup(oldAppName, newAppName)
	case "post-app-rename":
		oldAppName := flag.Arg(0)
		newAppName := flag.Arg(1)
		err = appjson.TriggerPostAppRename(oldAppName, newAppName)
	case "post-app-rename-setup":
		oldAppName := flag.Arg(0)
		newAppName := flag.Arg(1)
		err = appjson.TriggerPostAppRenameSetup(oldAppName, newAppName)
	case "post-create":
		appName := flag.Arg(0)
		err = appjson.TriggerPostCreate(appName)
	case "post-delete":
		appName := flag.Arg(0)
		err = appjson.TriggerPostDelete(appName)
	case "post-deploy":
		appName := flag.Arg(0)
		imageTag := flag.Arg(3)
		err = appjson.TriggerPostDeploy(appName, imageTag)
	case "pre-deploy":
		appName := flag.Arg(0)
		imageTag := flag.Arg(1)
		err = appjson.TriggerPreDeploy(appName, imageTag)
	case "report":
		appName := flag.Arg(0)
		err = appjson.ReportSingleApp(appName, "", "")
	default:
		err = fmt.Errorf("Invalid plugin trigger call: %s", trigger)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
