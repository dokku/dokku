package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/scheduler"
)

// main entrypoint to all triggers
func main() {
	parts := strings.Split(os.Args[0], "/")
	trigger := parts[len(parts)-1]
	flag.Parse()

	var err error
	switch trigger {
	case "scheduler-detect":
		appName := flag.Arg(0)
		err = scheduler.TriggerSchedulerDetect(appName)
	case "install":
		err = scheduler.TriggerInstall()
	case "post-delete":
		appName := flag.Arg(0)
		err = scheduler.TriggerPostDelete(appName)
	case "report":
		appName := flag.Arg(0)
		err = scheduler.ReportSingleApp(appName, "", "")
	default:
		err = fmt.Errorf("Invalid plugin trigger call: %s", trigger)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
