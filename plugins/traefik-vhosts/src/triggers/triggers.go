package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	traefikvhosts "github.com/dokku/dokku/plugins/traefik-vhosts"
)

// main entrypoint to all triggers
func main() {
	parts := strings.Split(os.Args[0], "/")
	trigger := parts[len(parts)-1]
	flag.Parse()

	var err error
	switch trigger {
	case "backup-app-export":
		err = traefikvhosts.TriggerBackupAppExport(flag.Arg(0), flag.Arg(1))
	case "backup-app-import":
		err = traefikvhosts.TriggerBackupAppImport(flag.Arg(0), flag.Arg(1))
	case "backup-global-export":
		err = traefikvhosts.TriggerBackupGlobalExport(flag.Arg(0))
	case "backup-global-import":
		err = traefikvhosts.TriggerBackupGlobalImport(flag.Arg(0))
	case "report":
		appName := flag.Arg(0)
		err = traefikvhosts.ReportSingleApp(appName, "", "")
	default:
		err = fmt.Errorf("Invalid plugin trigger call: %s", trigger)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
