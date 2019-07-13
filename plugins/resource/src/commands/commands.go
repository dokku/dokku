package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	columnize "github.com/ryanuber/columnize"
)

const (
	helpHeader = `Usage: dokku resource[:COMMAND]

Manages resource settings for an app

Additional commands:`

	helpContent = `
    resource:limit [--process-type <process-type>] [RESOURCE_OPTS...] <app>, Limit resources for a given app/process-type combination
    resource:limit-clear [--process-type <process-type>] <app>, Limit resources for a given app/process-type combination
    resource:report [<app>] [<flag>], Displays a resource report for one or more apps
    resource:reserve [--process-type <process-type>] [RESOURCE_OPTS...] <app>, Reserve resources for a given app/process-type combination
    resource:reserve-clear [--process-type <process-type>] <app>, Reserve resources for a given app/process-type combination
`
)

func main() {
	flag.Usage = usage
	flag.Parse()

	cmd := flag.Arg(0)
	switch cmd {
	case "resource", "resource:help":
		usage()
	case "help":
		command := common.NewShellCmd(fmt.Sprintf("ps -o command= %d", os.Getppid()))
		command.ShowOutput = false
		output, err := command.Output()

		if err == nil && strings.Contains(string(output), "--all") {
			fmt.Println(helpContent)
		} else {
			fmt.Print("\n    resource, Manages resource settings for an app\n")
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
	config := columnize.DefaultConfig()
	config.Delim = ","
	config.Prefix = "    "
	config.Empty = ""
	content := strings.Split(helpContent, "\n")[1:]
	fmt.Println(helpHeader)
	fmt.Println(columnize.Format(content, config))
}
