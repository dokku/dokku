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
	helpHeader = `Usage: dokku app-json[:COMMAND]

Manage app-json settings for an app

Additional commands:`

	helpContent = `
    app-json:report [<app>] [<flag>], Displays a app-json report for one or more apps
    app-json:set <app> <property> (<value>), Set or clear a app-json property for an app
`
)

func main() {
	flag.Usage = usage
	flag.Parse()

	cmd := flag.Arg(0)
	switch cmd {
	case "app-json", "app-json:help":
		usage()
	case "help":
		command := common.NewShellCmd(fmt.Sprintf("ps -o command= %d", os.Getppid()))
		command.ShowOutput = false
		output, err := command.Output()

		if err == nil && strings.Contains(string(output), "--all") {
			fmt.Println(helpContent)
		} else {
			fmt.Print("\n    app-json, Manage app-json settings for an app\n")
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
