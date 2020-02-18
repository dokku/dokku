package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/resource"
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
		processType := flag.Arg(3)
		err = resource.TriggerDockerArgsProcessDeploy(appName, processType)
	case "install":
		err = resource.TriggerInstall()
	case "post-app-clone-setup":
		oldAppName := flag.Arg(0)
		newAppName := flag.Arg(1)
		err := resource.TriggerPostAppCloneSetup(oldAppName, newAppName)
	case "post-app-rename-setup":
		oldAppName := flag.Arg(0)
		newAppName := flag.Arg(1)
		err := resource.TriggerPostAppRenameSetup(oldAppName, newAppName)
	case "post-delete":
		appName := flag.Arg(0)
		err = resource.TriggerPostDelete(appName)
	case "report":
		appName := flag.Arg(0)
		resource.ReportSingleApp(appName, "")
	case "resource-get-property":
		appName := flag.Arg(0)
		processType := flag.Arg(1)
		resourceType := flag.Arg(2)
		key := flag.Arg(3)
		err = resource.TriggerResourceGetProperty(appName, processType, resourceType, key)
	default:
		common.LogFail(fmt.Sprintf("Invalid plugin trigger call: %s", trigger))
	}

	if err != nil {
		common.LogFail(err.Error())
	}
}
