package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/proxy"
)

// main entrypoint to all triggers
func main() {
	parts := strings.Split(os.Args[0], "/")
	trigger := parts[len(parts)-1]
	flag.Parse()

	var err error
	switch trigger {
	case "proxy-is-enabled":
		appName := flag.Arg(0)
		err = proxy.TriggerProxyIsEnabled(appName)
	case "proxy-type":
		appName := flag.Arg(0)
		err = proxy.TriggerProxyType(appName)
	case "post-certs-remove":
		appName := flag.Arg(0)
		err = proxy.TriggerPostCertsRemove(appName)
	case "post-certs-update":
		appName := flag.Arg(0)
		err = proxy.TriggerPostCertsUpdate(appName)
	case "report":
		appName := flag.Arg(0)
		err = proxy.ReportSingleApp(appName, "", "")
	default:
		err = fmt.Errorf("Invalid plugin trigger call: %s", trigger)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
