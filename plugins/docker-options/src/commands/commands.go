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
	helpHeader = `Usage: dokku docker-options[:COMMAND]

Manage docker options for an app

Additional commands:`

	helpContent = `
    docker-options:add <app> <phase(s)> OPTION, Add Docker option to app for phase (comma separated phase list)
    docker-options:clear <app> [phase(s)], Clear a docker options from application with an optional phase (comma separated phase list)
    docker-options:remove <app> <phase(s)> OPTION, Remove Docker option from app for phase (comma separated phase list)
    docker-options:report [<app>] [<flag>], Displays a docker options report for one or more apps`
)

func main() {
	flag.Usage = usage
	flag.Parse()

	cmd := flag.Arg(0)
	switch cmd {
	case "docker-options", "docker-options:help":
		usage()
	case "help":
		result, err := common.CallExecCommand(common.ExecCommandInput{
			Command: "ps",
			Args:    []string{"-o", "command=", strconv.Itoa(os.Getppid())},
		})
		if err == nil && strings.Contains(result.StdoutContents(), "--all") {
			fmt.Println(helpContent)
		} else {
			fmt.Print("\n    docker-options, Manage docker options for an app\n")
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
