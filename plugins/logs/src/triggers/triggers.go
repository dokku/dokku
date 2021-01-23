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
	case "post-delete":
		appName := flag.Arg(0)
		err = logs.TriggerPostDelete(appName)
	case "report":
		appName := flag.Arg(0)
		err = logs.ReportSingleApp(appName, "")
	default:
		common.LogFail(fmt.Sprintf("Invalid plugin trigger call: %s", trigger))
	}

	if err != nil {
		common.LogFail(err.Error())
	}
}
