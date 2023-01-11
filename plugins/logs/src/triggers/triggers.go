package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/logs"
)

// main entrypoint to all triggers
func main() {
	parts := strings.Split(os.Args[0], "/")
	trigger := parts[len(parts)-1]
	flag.Parse()

	var err error
	switch trigger {
	case "docker-args-process-deploy":
		appName := flag.Arg(0)
		err = logs.TriggerDockerArgsProcessDeploy(appName)
	case "install":
		err = logs.TriggerInstall()
	case "logs-get-property":
		appName := flag.Arg(0)
		property := flag.Arg(1)
		err = logs.TriggerLogsGetProperty(appName, property)
	case "post-app-clone-setup":
		oldAppName := flag.Arg(0)
		newAppName := flag.Arg(1)
		err = logs.TriggerPostAppCloneSetup(oldAppName, newAppName)
	case "post-app-rename":
		oldAppName := flag.Arg(0)
		newAppName := flag.Arg(1)
		err = logs.TriggerPostAppRename(oldAppName, newAppName)
	case "post-app-rename-setup":
		oldAppName := flag.Arg(0)
		newAppName := flag.Arg(1)
		err = logs.TriggerPostAppRenameSetup(oldAppName, newAppName)
	case "post-create":
		appName := flag.Arg(0)
		err = logs.TriggerPostCreate(appName)
	case "post-delete":
		appName := flag.Arg(0)
		err = logs.TriggerPostDelete(appName)
	case "report":
		appName := flag.Arg(0)
		err = logs.ReportSingleApp(appName, "", "")
	default:
		err = fmt.Errorf("Invalid plugin trigger call: %s", trigger)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
