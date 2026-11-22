package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	certs "github.com/dokku/dokku/plugins/certs"
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
		err = certs.TriggerBackupAppExport(flag.Arg(0), flag.Arg(1))
	case "backup-app-import":
		err = certs.TriggerBackupAppImport(flag.Arg(0), flag.Arg(1))
	case "report":
		appName := flag.Arg(0)
		err = certs.ReportSingleApp(appName, "", "")
	default:
		err = fmt.Errorf("Invalid plugin trigger call: %s", trigger)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
