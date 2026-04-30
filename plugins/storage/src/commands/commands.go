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
	helpHeader = `Usage: dokku storage[:COMMAND]

Manage mounted volumes

Additional commands:`

	helpContent = `
    storage:create <name> [<path>] [flags], Register a named storage entry
    storage:destroy <name>, Remove a named storage entry (must be unmounted from every app first)
    storage:ensure-directory [--chown option] <directory>, [DEPRECATED] use storage:create instead
    storage:exec <name> [-- <cmd>...], Run a command (or shell) in a temporary container that mounts the entry
    storage:info <name> [--format text|json], Show details for one storage entry
    storage:list <app> [--format text|json], List bind mounts for app's container(s) (host:container)
    storage:list-entries [--scheduler s] [--format text|json], List registered storage entries
    storage:migrate [<app>|--all], Re-run the legacy -v to attachment migration for an app
    storage:mount <app> <host-dir:container-dir>, Create a new bind mount
    storage:report [<app>] [<flag>], Displays a storage report for one or more apps
    storage:set <name> [flags], Update a storage entry in place
    storage:unmount <app> <host-dir:container-dir>, Remove an existing bind mount
    storage:wait <name>, Wait for a storage entry's PVC to be bound (k3s)`
)

func main() {
	flag.Usage = usage
	flag.Parse()

	cmd := flag.Arg(0)
	switch cmd {
	case "storage":
		usage()
	case "storage:help":
		usage()
	case "help":
		result, err := common.CallExecCommand(common.ExecCommandInput{
			Command: "ps",
			Args:    []string{"-o", "command=", strconv.Itoa(os.Getppid())},
		})
		if err == nil && strings.Contains(result.StdoutContents(), "--all") {
			fmt.Println(helpContent)
		} else {
			fmt.Print("\n    storage, Manage mounted volumes\n")
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
