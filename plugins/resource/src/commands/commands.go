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
	helpHeader = `Usage: dokku resource[:COMMAND]

Manage resource settings for an app

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
		result, err := common.CallExecCommand(common.ExecCommandInput{
			Command:       "ps",
			Args:          []string{"-o", "command=", strconv.Itoa(os.Getppid())},
			CaptureOutput: true,
			StreamStdio:   false,
		})
		if err == nil && strings.Contains(result.StdoutContents(), "--all") {
			fmt.Println(helpContent)
		} else {
			fmt.Print("\n    resource, Manage resource settings for an app\n")
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
