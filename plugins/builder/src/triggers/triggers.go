package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/builder"
	"github.com/dokku/dokku/plugins/common"
)

// main entrypoint to all triggers
func main() {
	parts := strings.Split(os.Args[0], "/")
	trigger := parts[len(parts)-1]
	flag.Parse()

	var err error
	switch trigger {
	case "builder-detect":
		appName := flag.Arg(0)
		err = builder.TriggerBuilderDetect(appName)
	case "install":
		err = builder.TriggerInstall()
	case "post-delete":
		appName := flag.Arg(0)
		err = builder.TriggerPostDelete(appName)
	case "report":
		appName := flag.Arg(0)
		err = builder.ReportSingleApp(appName, "", "")
	case "core-post-extract":
		appName := flag.Arg(0)
		sourceWorkDir := flag.Arg(1)
		err = builder.TriggerCorePostExtract(appName, sourceWorkDir)
	default:
		err = fmt.Errorf("Invalid plugin trigger call: %s", trigger)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
