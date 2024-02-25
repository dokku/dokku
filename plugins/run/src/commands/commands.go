package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/run"

	flag "github.com/spf13/pflag"
)

const (
	helpHeader = `Usage: dokku run[:COMMAND]

Run a one-off process inside a container

Additional commands:`

	helpContent = `
    run [-e|--env KEY=VALUE] [--no-tty] <app> <cmd>, Run a command in a new container using the current app image
    run:detached [-e|-env KEY=VALUE] [--no-tty|--tty] <app> <cmd>, Run a command in a new detached container using the current app image
    run:list [--format json|stdout] <app>, List all run containers for an app
    run:logs <app|--container CONTAINER> [-h] [-t] [-n num] [-q], Display recent log output for run containers
    run:stop <app|--container CONTAINER>, Stops all run containers for an app or a specified run container`
)

func main() {
	flag.Usage = usage
	flag.Parse()

	cmd := flag.Arg(0)
	switch cmd {
	case "run":
		args := flag.NewFlagSet("run", flag.ExitOnError)
		env := args.StringToStringP("env", "e", map[string]string{}, "--env: environment variables to set")
		noTty := args.Bool("no-tty", false, "--no-tty: do not allocate a pseudo-TTY")
		forceTty := args.Bool("tty", false, "--tty: force allocation of a pseudo-TTY")
		cronID := args.String("cron-id", "", "--cron-id: cron job id")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		allArgs := args.Args()
		command := []string{}
		if len(allArgs) > 1 {
			command = allArgs[1:]
		}
		if err := run.CommandDefault(appName, command, *env, *noTty, *forceTty, *cronID); err != nil {
			common.LogFailWithError(err)
		}
	case "run:help":
		usage()
	case "help":
		result, err := common.CallExecCommand(common.ExecCommandInput{
			Command: "ps",
			Args:    []string{"-o", "command=", strconv.Itoa(os.Getppid())},
		})
		if err == nil && strings.Contains(result.StdoutContents(), "--all") {
			fmt.Println(helpContent)
		} else {
			fmt.Print("\n    run, Run a one-off process inside a container\n")
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
