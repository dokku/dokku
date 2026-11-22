package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	builderlambda "github.com/dokku/dokku/plugins/builder-lambda"
	"github.com/dokku/dokku/plugins/common"
)

// main entrypoint to all triggers
func main() {
	parts := strings.Split(os.Args[0], "/")
	trigger := parts[len(parts)-1]
	flag.Parse()

	var err error
	switch trigger {
	case "backup-app-export":
		err = builderlambda.TriggerBackupAppExport(flag.Arg(0), flag.Arg(1))
	case "backup-app-import":
		err = builderlambda.TriggerBackupAppImport(flag.Arg(0), flag.Arg(1))
	case "backup-global-export":
		err = builderlambda.TriggerBackupGlobalExport(flag.Arg(0))
	case "backup-global-import":
		err = builderlambda.TriggerBackupGlobalImport(flag.Arg(0))
	case "report":
		appName := flag.Arg(0)
		err = builderlambda.ReportSingleApp(appName, "", "")
	default:
		err = fmt.Errorf("Invalid plugin trigger call: %s", trigger)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
