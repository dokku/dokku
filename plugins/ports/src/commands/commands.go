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
	helpHeader = `Usage: dokku ports[:COMMAND]

Manage ports for an app

Additional commands:`

	helpContent = `
    ports:list <app>, List port mappings for app
    ports:add <app> [<scheme>:<host-port>:<container-port>...], Add port mappings to an app
    ports:clear <app>, Clear all port mappings for an app
    ports:remove <app> [<host-port>|<scheme>:<host-port>:<container-port>...], Remove specific port mappings from an app
    ports:set <app> [<scheme>:<host-port>:<container-port>...], Set port mappings for an app
    ports:report [<app>] [<flag>], Displays a ports report for one or more apps
`
)

func main() {
	flag.Usage = usage
	flag.Parse()

	cmd := flag.Arg(0)
	switch cmd {
	case "ports", "ports:help":
		usage()
	case "help":
		command := common.NewShellCmd(fmt.Sprintf("ps -o command= %d", os.Getppid()))
		command.ShowOutput = false
		output, err := command.Output()

		if err == nil && strings.Contains(string(output), "--all") {
			fmt.Println(helpContent)
		} else {
			fmt.Print("\n    ports, Manage ports for an app\n")
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
