package main

import (
	"fmt"
	"os"
	"strings"

	flag "github.com/spf13/pflag"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/scheduler"
)

// main entrypoint to all triggers
func main() {
	parts := strings.Split(os.Args[0], "/")
	trigger := parts[len(parts)-1]
	global := flag.Bool("global", false, "--global: use the global environment")
	flag.Parse()

	var err error
	switch trigger {
	case "install":
		err = scheduler.TriggerInstall()
	case "post-app-clone-setup":
		oldAppName := flag.Arg(0)
		newAppName := flag.Arg(1)
		err = scheduler.TriggerPostAppCloneSetup(oldAppName, newAppName)
	case "post-app-rename-setup":
		oldAppName := flag.Arg(0)
		newAppName := flag.Arg(1)
		err = scheduler.TriggerPostAppRenameSetup(oldAppName, newAppName)
	case "post-delete":
		appName := flag.Arg(0)
		err = scheduler.TriggerPostDelete(appName)
	case "report":
		appName := flag.Arg(0)
		err = scheduler.ReportSingleApp(appName, "", "")
	case "scheduler-detect":
		appName := flag.Arg(0)
		if *global {
			appName = "--global"
		}
		err = scheduler.TriggerSchedulerDetect(appName)
	default:
		err = fmt.Errorf("Invalid plugin trigger call: %s", trigger)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
