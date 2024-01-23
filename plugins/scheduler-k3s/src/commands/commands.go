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
	helpHeader = `Usage: dokku scheduler-k3s[:COMMAND]

Manage scheduler-k3s settings for an app

Additional commands:`

	helpContent = `
    scheduler-k3s:initialize, Initializes a cluster
    scheduler-k3s:report [<app>] [<flag>], Displays a scheduler-k3s report for one or more apps
    scheduler-k3s:set <app> <property> (<value>), Set or clear a scheduler-k3s property for an app
`
)

func main() {
	flag.Usage = usage
	flag.Parse()

	cmd := flag.Arg(0)
	switch cmd {
	case "scheduler-k3s", "scheduler-k3s:help":
		usage()
	case "help":
		command := common.NewShellCmd(fmt.Sprintf("ps -o command= %d", os.Getppid()))
		command.ShowOutput = false
		output, err := command.Output()

		if err == nil && strings.Contains(string(output), "--all") {
			fmt.Println(helpContent)
		} else {
			fmt.Print("\n    scheduler-k3s, Manage scheduler-k3s settings for an app\n")
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
