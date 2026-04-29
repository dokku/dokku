package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	dockeroptions "github.com/dokku/dokku/plugins/docker-options"

	flag "github.com/spf13/pflag"
)

func main() {
	parts := strings.Split(os.Args[0], "/")
	subcommand := parts[len(parts)-1]

	var err error
	switch subcommand {
	case "add":
		args := flag.NewFlagSet("docker-options:add", flag.ExitOnError)
		args.SetInterspersed(false)
		processes := args.StringSlice("process", []string{}, "process types to scope this option to (deploy phase only)")
		args.Parse(os.Args[2:])

		positional := args.Args()
		appName := ""
		phases := ""
		var optionParts []string
		if len(positional) > 0 {
			appName = positional[0]
		}
		if len(positional) > 1 {
			phases = positional[1]
		}
		if len(positional) > 2 {
			optionParts = positional[2:]
		}
		err = dockeroptions.CommandAdd(appName, *processes, phases, strings.Join(optionParts, " "))
	case "remove":
		args := flag.NewFlagSet("docker-options:remove", flag.ExitOnError)
		args.SetInterspersed(false)
		processes := args.StringSlice("process", []string{}, "process types to scope this option to (deploy phase only)")
		args.Parse(os.Args[2:])

		positional := args.Args()
		appName := ""
		phases := ""
		var optionParts []string
		if len(positional) > 0 {
			appName = positional[0]
		}
		if len(positional) > 1 {
			phases = positional[1]
		}
		if len(positional) > 2 {
			optionParts = positional[2:]
		}
		err = dockeroptions.CommandRemove(appName, *processes, phases, strings.Join(optionParts, " "))
	case "clear":
		args := flag.NewFlagSet("docker-options:clear", flag.ExitOnError)
		args.SetInterspersed(false)
		processes := args.StringSlice("process", []string{}, "process types to scope this option to (deploy phase only)")
		args.Parse(os.Args[2:])

		positional := args.Args()
		appName := ""
		phases := ""
		if len(positional) > 0 {
			appName = positional[0]
		}
		if len(positional) > 1 {
			phases = positional[1]
		}
		err = dockeroptions.CommandClear(appName, *processes, phases)
	case "list":
		args := flag.NewFlagSet("docker-options:list", flag.ExitOnError)
		processType := args.String("process", "", "process type to query (omit for the default scope)")
		phase := args.String("phase", "", "phase to query [build|deploy|run] (required)")
		args.Parse(os.Args[2:])

		appName := args.Arg(0)
		err = dockeroptions.CommandList(appName, *processType, *phase)
	case "report":
		args := flag.NewFlagSet("docker-options:report", flag.ExitOnError)
		format := args.String("format", "stdout", "format: [ stdout | json ]")
		reportArgs, flagErr := common.ParseReportArgs("docker-options", os.Args[2:])
		if flagErr == nil {
			args.Parse(reportArgs.OSArgs)
			appName := args.Arg(0)
			if reportArgs.IsGlobal {
				appName = "--global"
			}
			err = dockeroptions.CommandReport(appName, *format, reportArgs.InfoFlag)
		} else {
			err = flagErr
		}
	default:
		err = fmt.Errorf("Invalid plugin subcommand call: %s", subcommand)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
