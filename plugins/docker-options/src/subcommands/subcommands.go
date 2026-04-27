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
		processes := args.StringSlice("process", []string{}, "process types to scope this option to (reserved for future use)")
		global := args.Bool("global", false, "explicitly mark as a global option (reserved for future use)")
		args.Parse(os.Args[2:])

		if err = dockeroptions.ErrIfReservedFlagsUsed(*processes, *global); err != nil {
			break
		}

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
		err = dockeroptions.CommandAdd(appName, phases, strings.Join(optionParts, " "))
	case "remove":
		args := flag.NewFlagSet("docker-options:remove", flag.ExitOnError)
		args.SetInterspersed(false)
		processes := args.StringSlice("process", []string{}, "process types to scope this option to (reserved for future use)")
		global := args.Bool("global", false, "explicitly mark as a global option (reserved for future use)")
		args.Parse(os.Args[2:])

		if err = dockeroptions.ErrIfReservedFlagsUsed(*processes, *global); err != nil {
			break
		}

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
		err = dockeroptions.CommandRemove(appName, phases, strings.Join(optionParts, " "))
	case "clear":
		args := flag.NewFlagSet("docker-options:clear", flag.ExitOnError)
		args.SetInterspersed(false)
		processes := args.StringSlice("process", []string{}, "process types to scope this option to (reserved for future use)")
		global := args.Bool("global", false, "explicitly mark as a global option (reserved for future use)")
		args.Parse(os.Args[2:])

		if err = dockeroptions.ErrIfReservedFlagsUsed(*processes, *global); err != nil {
			break
		}

		positional := args.Args()
		appName := ""
		phases := ""
		if len(positional) > 0 {
			appName = positional[0]
		}
		if len(positional) > 1 {
			phases = positional[1]
		}
		err = dockeroptions.CommandClear(appName, phases)
	case "report":
		args := flag.NewFlagSet("docker-options:report", flag.ExitOnError)
		format := args.String("format", "stdout", "format: [ stdout | json ]")
		osArgs, infoFlag, flagErr := common.ParseReportArgs("docker-options", os.Args[2:])
		if flagErr == nil {
			args.Parse(osArgs)
			appName := args.Arg(0)
			err = dockeroptions.CommandReport(appName, *format, infoFlag)
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
