package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/builds"
	"github.com/dokku/dokku/plugins/common"

	flag "github.com/spf13/pflag"
)

func main() {
	parts := strings.Split(os.Args[0], "/")
	subcommand := parts[len(parts)-1]

	var err error
	switch subcommand {
	case "cancel":
		args := flag.NewFlagSet("builds:cancel", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = builds.CommandCancel(args.Arg(0))
	case "info":
		args := flag.NewFlagSet("builds:info", flag.ExitOnError)
		format := args.String("format", "stdout", "format: [ stdout | json ]")
		args.Parse(os.Args[2:])
		err = builds.CommandInfo(args.Arg(0), args.Arg(1), *format)
	case "list":
		args := flag.NewFlagSet("builds:list", flag.ExitOnError)
		format := args.String("format", "stdout", "format: [ stdout | json ]")
		kind := args.String("kind", "", "filter by kind: [ build | deploy ]")
		status := args.String("status", "", "filter by status: [ running | succeeded | failed | canceled | abandoned ]")
		args.Parse(os.Args[2:])
		err = builds.CommandList(args.Arg(0), *format, *kind, *status)
	case "output":
		args := flag.NewFlagSet("builds:output", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = builds.CommandOutput(args.Arg(0), args.Arg(1))
	case "prune":
		args := flag.NewFlagSet("builds:prune", flag.ExitOnError)
		allApps := args.Bool("all-apps", false, "--all-apps: prune every app")
		args.Parse(os.Args[2:])
		err = builds.CommandPrune(args.Arg(0), *allApps)
	case "report":
		args := flag.NewFlagSet("builds:report", flag.ExitOnError)
		format := args.String("format", "stdout", "format: [ stdout | json ]")
		reportArgs, flagErr := common.ParseReportArgs("builds", os.Args[2:])
		if flagErr == nil {
			args.Parse(reportArgs.OSArgs)
			appName := args.Arg(0)
			if reportArgs.IsGlobal {
				appName = "--global"
			}
			err = builds.CommandReport(appName, *format, reportArgs.InfoFlag)
		}
	case "set":
		args := flag.NewFlagSet("builds:set", flag.ExitOnError)
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
		err = builds.CommandSet(appName, property, value)
	default:
		err = fmt.Errorf("Invalid plugin subcommand call: %s", subcommand)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
