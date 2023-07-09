package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/ports"
)

// main entrypoint to all triggers
func main() {
	parts := strings.Split(os.Args[0], "/")
	trigger := parts[len(parts)-1]
	flag.Parse()

	var err error
	switch trigger {
	case "install":
		err = ports.TriggerInstall()
	case "ports-clear":
		appName := flag.Arg(0)
		err = ports.TriggerPortsClear(appName)
	case "ports-configure":
		appName := flag.Arg(0)
		err = ports.TriggerPortsConfigure(appName)
	case "ports-dockerfile-raw-tcp-ports":
		appName := flag.Arg(0)
		err = ports.TriggerPortsDockerfileRawTCPPorts(appName)
	case "ports-get":
		appName := flag.Arg(0)
		err = ports.TriggerPortsGet(appName)
	case "ports-get-available":
		err = ports.TriggerPortsGetAvailable()
	case "post-app-clone-setup":
		oldAppName := flag.Arg(0)
		newAppName := flag.Arg(1)
		err = ports.TriggerPostAppCloneSetup(oldAppName, newAppName)
	case "post-app-rename-setup":
		oldAppName := flag.Arg(0)
		newAppName := flag.Arg(1)
		err = ports.TriggerPostAppRenameSetup(oldAppName, newAppName)
	case "post-certs-remove":
		appName := flag.Arg(0)
		err = ports.TriggerPostCertsRemove(appName)
	case "post-certs-update":
		appName := flag.Arg(0)
		err = ports.TriggerPostCertsUpdate(appName)
	case "post-delete":
		appName := flag.Arg(0)
		err = ports.TriggerPostDelete(appName)
	case "report":
		appName := flag.Arg(0)
		err = ports.ReportSingleApp(appName, "", "")
	default:
		err = fmt.Errorf("Invalid plugin trigger call: %s", trigger)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
