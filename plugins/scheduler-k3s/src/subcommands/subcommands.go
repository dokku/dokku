package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	scheduler_k3s "github.com/dokku/dokku/plugins/scheduler-k3s"

	flag "github.com/spf13/pflag"
)

// main entrypoint to all subcommands
func main() {
	parts := strings.Split(os.Args[0], "/")
	subcommand := parts[len(parts)-1]

	var err error
	switch subcommand {
	case "initialize":
		args := flag.NewFlagSet("scheduler-k3s:initialize", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = scheduler_k3s.CommandInitialize()
	case "report":
		args := flag.NewFlagSet("scheduler-k3s:report", flag.ExitOnError)
		format := args.String("format", "stdout", "format: [ stdout | json ]")
		osArgs, infoFlag, flagErr := common.ParseReportArgs("scheduler-k3s", os.Args[2:])
		if flagErr == nil {
			args.Parse(osArgs)
			appName := args.Arg(0)
			err = scheduler_k3s.CommandReport(appName, *format, infoFlag)
		}
	case "set":
		args := flag.NewFlagSet("scheduler-k3s:set", flag.ExitOnError)
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
		err = scheduler_k3s.CommandSet(appName, property, value)
	default:
		err = fmt.Errorf("Invalid plugin subcommand call: %s", subcommand)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
