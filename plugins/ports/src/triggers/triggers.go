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
	case "ports-configure":
		appName := flag.Arg(0)
		err = ports.TriggerPortsConfigure(appName)
	case "ports-dockerfile-raw-tcp-ports":
		appName := flag.Arg(0)
		err = ports.TriggerPortsDockerfileRawTCPPorts(appName)
	case "post-certs-remove":
		appName := flag.Arg(0)
		err = ports.TriggerPostCertsRemove(appName)
	case "post-certs-update":
		appName := flag.Arg(0)
		err = ports.TriggerPostCertsUpdate(appName)
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
