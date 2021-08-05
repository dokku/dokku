package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/registry"
)

// main entrypoint to all triggers
func main() {
	parts := strings.Split(os.Args[0], "/")
	trigger := parts[len(parts)-1]
	flag.Parse()

	var err error
	switch trigger {
	case "deployed-app-image-repo":
		appName := flag.Arg(0)
		err = registry.TriggerDeployedAppImageRepo(appName)
	case "deployed-app-image-tag":
		appName := flag.Arg(0)
		err = registry.TriggerDeployedAppImageTag(appName)
	case "deployed-app-repository":
		appName := flag.Arg(0)
		err = registry.TriggerDeployedAppRepository(appName)
	case "install":
		err = registry.TriggerInstall()
	case "post-delete":
		appName := flag.Arg(0)
		err = registry.TriggerPostDelete(appName)
	case "post-release-builder":
		appName := flag.Arg(1)
		image := flag.Arg(2)
		err = registry.TriggerPostReleaseBuilder(appName, image)
	case "report":
		appName := flag.Arg(0)
		err = registry.ReportSingleApp(appName, "", "")
	default:
		err = fmt.Errorf("Invalid plugin trigger call: %s", trigger)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
