package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	cron "github.com/dokku/dokku/plugins/cron"
)

// main entrypoint to all triggers
func main() {
	parts := strings.Split(os.Args[0], "/")
	trigger := parts[len(parts)-1]
	flag.Parse()

	var err error
	switch trigger {
	case "cron-write":
		err = cron.TriggerCronWrite()
	case "post-delete":
		err = cron.TriggerPostDelete()
	case "post-deploy":
		err = cron.TriggerPostDeploy()
	case "report":
		appName := flag.Arg(0)
		err = cron.ReportSingleApp(appName, "", "")
	default:
		err = fmt.Errorf("Invalid plugin trigger call: %s", trigger)
	}

	if err != nil {
		common.LogFail(err.Error())
	}
}
