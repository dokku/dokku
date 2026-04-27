package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	dockeroptions "github.com/dokku/dokku/plugins/docker-options"
)

func main() {
	parts := strings.Split(os.Args[0], "/")
	trigger := parts[len(parts)-1]
	flag.Parse()

	var err error
	switch trigger {
	case "install":
		err = dockeroptions.TriggerInstall()
	case "docker-args-build":
		appName := flag.Arg(0)
		imageSourceType := flag.Arg(1)
		err = dockeroptions.TriggerDockerArgs("build", appName, imageSourceType, "")
	case "docker-args-deploy":
		appName := flag.Arg(0)
		imageSourceType := flag.Arg(1)
		err = dockeroptions.TriggerDockerArgs("deploy", appName, imageSourceType, "")
	case "docker-args-run":
		appName := flag.Arg(0)
		imageSourceType := flag.Arg(1)
		err = dockeroptions.TriggerDockerArgs("run", appName, imageSourceType, "")
	case "docker-args-process-deploy":
		appName := flag.Arg(0)
		imageSourceType := flag.Arg(1)
		processType := flag.Arg(3)
		err = dockeroptions.TriggerDockerArgsProcessDeploy(appName, imageSourceType, processType)
	case "post-app-clone-setup":
		oldAppName := flag.Arg(0)
		newAppName := flag.Arg(1)
		err = dockeroptions.TriggerPostAppCloneSetup(oldAppName, newAppName)
	case "post-app-rename-setup":
		oldAppName := flag.Arg(0)
		newAppName := flag.Arg(1)
		err = dockeroptions.TriggerPostAppRenameSetup(oldAppName, newAppName)
	case "post-delete":
		appName := flag.Arg(0)
		err = dockeroptions.TriggerPostDelete(appName)
	case "report":
		appName := flag.Arg(0)
		err = dockeroptions.ReportSingleApp(appName, "", "")
	default:
		err = fmt.Errorf("Invalid plugin trigger call: %s", trigger)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
