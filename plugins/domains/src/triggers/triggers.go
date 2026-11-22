package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	domains "github.com/dokku/dokku/plugins/domains"
)

// main entrypoint to all triggers
func main() {
	parts := strings.Split(os.Args[0], "/")
	trigger := parts[len(parts)-1]
	flag.Parse()

	var err error
	switch trigger {
	case "backup-app-export":
		err = domains.TriggerBackupAppExport(flag.Arg(0), flag.Arg(1))
	case "backup-app-import":
		err = domains.TriggerBackupAppImport(flag.Arg(0), flag.Arg(1))
	case "backup-global-export":
		err = domains.TriggerBackupGlobalExport(flag.Arg(0))
	case "backup-global-import":
		err = domains.TriggerBackupGlobalImport(flag.Arg(0))
	case "report":
		appName := flag.Arg(0)
		err = domains.ReportSingleApp(appName, "", "")
	default:
		err = fmt.Errorf("Invalid plugin trigger call: %s", trigger)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
