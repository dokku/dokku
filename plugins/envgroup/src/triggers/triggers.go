package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/envgroup"
)

// main entrypoint to all triggers
func main() {
	parts := strings.Split(os.Args[0], "/")
	trigger := parts[len(parts)-1]
	flag.Parse()

	var err error
	switch trigger {
	case "backup-app-export":
		err = envgroup.TriggerBackupAppExport(flag.Arg(0), flag.Arg(1))
	case "backup-app-import":
		err = envgroup.TriggerBackupAppImport(flag.Arg(0), flag.Arg(1))
	case "backup-global-export":
		err = envgroup.TriggerBackupGlobalExport(flag.Arg(0))
	case "backup-global-import":
		err = envgroup.TriggerBackupGlobalImport(flag.Arg(0))
	case "envgroup-app-env":
		err = envgroup.TriggerEnvgroupAppEnv(flag.Arg(0), flag.Arg(1))
	case "install":
		err = envgroup.TriggerInstall()
	case "post-app-clone-setup":
		err = envgroup.TriggerPostAppCloneSetup(flag.Arg(0), flag.Arg(1))
	case "post-app-rename-setup":
		err = envgroup.TriggerPostAppRenameSetup(flag.Arg(0), flag.Arg(1))
	case "post-delete":
		err = envgroup.TriggerPostDelete(flag.Arg(0))
	case "report":
		err = envgroup.ReportSingleApp(flag.Arg(0), "", "")
	default:
		err = fmt.Errorf("Invalid plugin trigger call: %s", trigger)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
