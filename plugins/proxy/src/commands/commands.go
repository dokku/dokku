package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

const (
	helpHeader = `Usage: dokku proxy[:COMMAND]

Manage the proxy integration for an app

Additional commands:`

	helpContent = `
    proxy:disable <app>, Disable proxy for app
    proxy:enable <app>, Enable proxy for app
    proxy:ports <app>, List proxy port mappings for app
    proxy:ports-add <app> [<scheme>:<host-port>:<container-port>...], Add proxy port mappings to an app
    proxy:ports-clear <app>, Clear all proxy port mappings for an app
    proxy:ports-remove <app> [<host-port>|<scheme>:<host-port>:<container-port>...], Remove specific proxy port mappings from an app
    proxy:ports-set <app> [<scheme>:<host-port>:<container-port>...], Set proxy port mappings for an app
    proxy:report [<app>] [<flag>], Displays a proxy report for one or more apps
    proxy:set <app> <proxy-type>, Set proxy type for app
`
)

func main() {
	flag.Usage = usage
	flag.Parse()

	cmd := flag.Arg(0)
	switch cmd {
	case "proxy", "proxy:help":
		usage()
	case "help":
		command := common.NewShellCmd(fmt.Sprintf("ps -o command= %d", os.Getppid()))
		command.ShowOutput = false
		output, err := command.Output()

		if err == nil && strings.Contains(string(output), "--all") {
			fmt.Println(helpContent)
		} else {
			fmt.Print("\n    proxy, Manage the proxy integration for an app\n")
		}
	default:
		dokkuNotImplementExitCode, err := strconv.Atoi(os.Getenv("DOKKU_NOT_IMPLEMENTED_EXIT"))
		if err != nil {
			fmt.Println("failed to retrieve DOKKU_NOT_IMPLEMENTED_EXIT environment variable")
			dokkuNotImplementExitCode = 10
		}
		os.Exit(dokkuNotImplementExitCode)
	}
}

func usage() {
	common.CommandUsage(helpHeader, helpContent)
}
