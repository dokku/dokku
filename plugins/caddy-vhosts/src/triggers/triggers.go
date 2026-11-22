package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	caddyvhosts "github.com/dokku/dokku/plugins/caddy-vhosts"
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
		err = caddyvhosts.TriggerBackupAppExport(flag.Arg(0), flag.Arg(1))
	case "backup-app-import":
		err = caddyvhosts.TriggerBackupAppImport(flag.Arg(0), flag.Arg(1))
	case "backup-global-export":
		err = caddyvhosts.TriggerBackupGlobalExport(flag.Arg(0))
	case "backup-global-import":
		err = caddyvhosts.TriggerBackupGlobalImport(flag.Arg(0))
	case "report":
		appName := flag.Arg(0)
		err = caddyvhosts.ReportSingleApp(appName, "", "")
	default:
		err = fmt.Errorf("Invalid plugin trigger call: %s", trigger)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
