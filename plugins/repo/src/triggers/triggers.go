package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/repo"
)

// main entrypoint to all triggers
func main() {
	parts := strings.Split(os.Args[0], "/")
	trigger := parts[len(parts)-1]
	flag.Parse()

	var err error
	switch trigger {
	case "install":
		err = repo.TriggerInstall()
	case "post-delete":
		appName := flag.Arg(0)
		err = repo.TriggerPostDelete(appName)
	case "post-extract":
		appName := flag.Arg(0)
		tmpWorkDir := flag.Arg(1)
		err = repo.TriggerPostExtract(appName, tmpWorkDir)
	case "pre-delete":
		appName := flag.Arg(0)
		err = repo.TriggerPreDelete(appName)
	default:
		common.LogFail(fmt.Sprintf("Invalid plugin trigger call: %s", trigger))
	}

	if err != nil {
		common.LogFail(err.Error())
	}
}
