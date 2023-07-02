package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/ports"

	flag "github.com/spf13/pflag"
)

// main entrypoint to all subcommands
func main() {
	parts := strings.Split(os.Args[0], "/")
	subcommand := parts[len(parts)-1]

	var err error
	switch subcommand {
	case "list":
		args := flag.NewFlagSet("ports:list", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = ports.CommandList(appName)
	case "add":
		args := flag.NewFlagSet("ports:add", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		_, portMaps := common.ShiftString(args.Args())
		err = ports.CommandAdd(appName, portMaps)
	case "clear":
		args := flag.NewFlagSet("ports:clear", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = ports.CommandClear(appName)
	case "remove":
		args := flag.NewFlagSet("ports:remove", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		_, portMaps := common.ShiftString(args.Args())
		err = ports.CommandRemove(appName, portMaps)
	case "set":
		args := flag.NewFlagSet("ports:set", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		_, portMaps := common.ShiftString(args.Args())
		err = ports.CommandSet(appName, portMaps)
	case "report":
		args := flag.NewFlagSet("ports:report", flag.ExitOnError)
		format := args.String("format", "stdout", "format: [ stdout | json ]")
		osArgs, infoFlag, flagErr := common.ParseReportArgs("ports", os.Args[2:])
		if flagErr == nil {
			args.Parse(osArgs)
			appName := args.Arg(0)
			err = ports.CommandReport(appName, *format, infoFlag)
		}
	default:
		err = fmt.Errorf("Invalid plugin subcommand call: %s", subcommand)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
