package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/run"

	flag "github.com/spf13/pflag"
)

// main entrypoint to all subcommands
func main() {
	parts := strings.Split(os.Args[0], "/")
	subcommand := parts[len(parts)-1]

	var err error
	switch subcommand {
	case "detached":
		args := flag.NewFlagSet("run:detached", flag.ExitOnError)
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
		err = run.CommandDetached(appName, command, *env, *noTty, *forceTty, *cronID)
	case "list":
		args := flag.NewFlagSet("run:list", flag.ExitOnError)
		format := args.String("format", "stdout", "--format: [ stdout | json ]")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = run.CommandList(appName, *format)
	case "logs":
		args := flag.NewFlagSet("run:logs", flag.ExitOnError)
		container := args.String("container", "", "--container: container id")
		num := args.IntP("num", "n", 100, "--num: number of lines to display")
		quiet := args.BoolP("quiet", "q", false, "--quiet: only display log output")
		tail := args.BoolP("tail", "t", false, "--tail: follow log output")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = run.CommandLogs(appName, *container, *num, *quiet, *tail)
	case "stop":
		args := flag.NewFlagSet("run:stop", flag.ExitOnError)
		container := args.String("container", "", "--container: container id")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = run.CommandStop(appName, *container)
	default:
		err = fmt.Errorf("Invalid plugin subcommand call: %s", subcommand)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
