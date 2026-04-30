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
	helpHeader = `Usage: dokku builds[:COMMAND]

Manage running and historical builds

Additional commands:`

	helpContent = `
    builds:cancel <app>, Cancel a running build for an app
    builds:info <app> <build-id> [--format json|stdout], Show details for a single build
    builds:list [<app>] [--format json] [--kind build|deploy] [--status <status>], List builds
    builds:output <app> [<build-id>|current], Show build output
    builds:prune <app> [--all-apps], Prune build records to retention
    builds:report [<app>] [<flag>], Display a build report for one or more apps
    builds:set [--global|<app>] <key> [<value>], Set or clear a builds property`
)

func main() {
	flag.Usage = usage
	flag.Parse()

	cmd := flag.Arg(0)
	switch cmd {
	case "builds", "builds:help":
		usage()
	case "help":
		result, err := common.CallExecCommand(common.ExecCommandInput{
			Command: "ps",
			Args:    []string{"-o", "command=", strconv.Itoa(os.Getppid())},
		})
		if err == nil && strings.Contains(result.StdoutContents(), "--all") {
			fmt.Println(helpContent)
		} else {
			fmt.Print("\n    builds, Manage running and historical builds\n")
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
