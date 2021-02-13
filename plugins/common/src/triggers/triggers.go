package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// main entrypoint to all triggers
func main() {
	parts := strings.Split(os.Args[0], "/")
	trigger := parts[len(parts)-1]
	flag.Parse()

	var err error
	switch trigger {
	case "core-post-deploy":
		appName := flag.Arg(0)
		err = common.TriggerCorePostDeploy(appName)
	case "install":
		err = common.TriggerInstall()
	case "post-delete":
		appName := flag.Arg(0)
		err = common.TriggerPostDelete(appName)
	default:
		common.LogFail(fmt.Sprintf("Invalid plugin trigger call: %s", trigger))
	}

	if err != nil {
		common.LogFail(err.Error())
	}
}
