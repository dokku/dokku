package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/cron"

	flag "github.com/spf13/pflag"
)

// main entrypoint to all subcommands
func main() {
	parts := strings.Split(os.Args[0], "/")
	subcommand := parts[len(parts)-1]

	var err error
	switch subcommand {
	case "list":
		args := flag.NewFlagSet("cron:list", flag.ExitOnError)
		global := args.Bool("global", false, "--global: set a global property")
		format := args.String("format", "stdout", "format: [ stdout | json ]")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		if *global {
			appName = "--global"
		}
		err = cron.CommandList(appName, *format)
	case "report":
		args := flag.NewFlagSet("cron:report", flag.ExitOnError)
		format := args.String("format", "stdout", "format: [ stdout | json ]")
		osArgs, infoFlag, flagErr := common.ParseReportArgs("cron", os.Args[2:])
		if flagErr == nil {
			args.Parse(osArgs)
			appName := args.Arg(0)
			err = cron.CommandReport(appName, *format, infoFlag)
		}
	case "resume":
		args := flag.NewFlagSet("cron:resume", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		cronID := args.Arg(1)
		err = cron.CommandResume(appName, cronID)
	case "run":
		args := flag.NewFlagSet("cron:run", flag.ExitOnError)
		detached := args.Bool("detach", false, "--detach: run the container in a detached mode")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		cronID := args.Arg(1)
		err = cron.CommandRun(appName, cronID, *detached)
	case "set":
		args := flag.NewFlagSet("cron:set", flag.ExitOnError)
		global := args.Bool("global", false, "--global: set a global property")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		property := args.Arg(1)
		value := args.Arg(2)
		if *global {
			appName = "--global"
			property = args.Arg(0)
			value = args.Arg(1)
		}
		err = cron.CommandSet(appName, property, value)
	case "suspend":
		args := flag.NewFlagSet("cron:suspend", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		cronID := args.Arg(1)
		err = cron.CommandSuspend(appName, cronID)
	default:
		err = fmt.Errorf("Invalid plugin subcommand call: %s", subcommand)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
