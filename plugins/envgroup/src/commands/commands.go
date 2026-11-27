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
	helpHeader = `Usage: dokku envgroup[:COMMAND]

Manage environment variable groups shared by apps

Additional commands:`

	helpContent = `
    envgroup:add [--index <index>] [--restart] <group> <app>, Adds an app to an env group
    envgroup:bundle <group>, Bundles an env group's environment into a tarfile
    envgroup:clear [--restart] <group>, Clears an env group's environment variables
    envgroup:create <group>, Creates an env group
    envgroup:destroy [--force] [--restart] <group>, Destroys an env group
    envgroup:exists <group>, Checks if an env group exists
    envgroup:export [--format <format>] <group>, Exports an env group's environment
    envgroup:get [--quoted] <group> KEY, Displays an env group config value
    envgroup:import [--format <format>] [--replace] [--restart] <group> [FILE|-], Imports an env group's environment from a file
    envgroup:info <group>, Displays the apps in an env group and its number of keys
    envgroup:keys <group>, Shows the keys set in an env group
    envgroup:list, Lists all env groups
    envgroup:remove [--restart] <group> <app>, Removes an app from an env group
    envgroup:report [<app>] [<flag>], Displays an envgroup report for one or more apps
    envgroup:set [--encoded] [--restart] <group> KEY1=VALUE1 [KEY2=VALUE2 ...], Sets one or more env group config vars
    envgroup:show <group>, Pretty-prints an env group's environment
    envgroup:unset [--restart] <group> KEY1 [KEY2 ...], Unsets one or more env group config vars`
)

func main() {
	flag.Usage = usage
	flag.Parse()

	cmd := flag.Arg(0)
	switch cmd {
	case "envgroup", "envgroup:help":
		usage()
	case "help":
		result, err := common.CallExecCommand(common.ExecCommandInput{
			Command: "ps",
			Args:    []string{"-o", "command=", strconv.Itoa(os.Getppid())},
		})
		if err == nil && strings.Contains(result.StdoutContents(), "--all") {
			fmt.Println(helpContent)
		} else {
			fmt.Print("\n    envgroup, Manage environment variable groups shared by apps\n")
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
