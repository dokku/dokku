package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/logs"

	flag "github.com/spf13/pflag"
)

// main entrypoint to all subcommands
func main() {
	parts := strings.Split(os.Args[0], "/")
	subcommand := parts[len(parts)-1]

	var err error
	switch subcommand {
	case "failed":
		args := flag.NewFlagSet("logs:failed", flag.ExitOnError)
		allApps := args.Bool("all", false, "--all: restore all apps")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = logs.CommandFailed(appName, *allApps)
	case "report":
		args := flag.NewFlagSet("logs:report", flag.ExitOnError)
		osArgs, infoFlag, flagErr := common.ParseReportArgs("logs", os.Args[2:])
		if flagErr == nil {
			args.Parse(osArgs)
			appName := args.Arg(0)
			err = logs.CommandReport(appName, infoFlag)
		}
	case "set":
		args := flag.NewFlagSet("logs:set", flag.ExitOnError)
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
		err = logs.CommandSet(appName, property, value)
	case "vector-logs":
		args := flag.NewFlagSet("logs:vector-logs", flag.ExitOnError)
		num := args.Int("num", 100, "the number of lines to display")
		tail := args.Bool("tail", false, "continually stream logs")
		args.Parse(os.Args[2:])
		err = logs.CommandVectorLogs(*num, *tail)
	case "vector-start":
		args := flag.NewFlagSet("logs:vector-start", flag.ExitOnError)
		vectorImage := args.String("vector-image", logs.VectorImage, "--vector-image: the name of the docker image to run for vector")
		args.Parse(os.Args[2:])
		err = logs.CommandVectorStart(*vectorImage)
	case "vector-stop":
		args := flag.NewFlagSet("logs:vector-stop", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = logs.CommandVectorStop()
	default:
		common.LogFail(fmt.Sprintf("Invalid plugin subcommand call: %s", subcommand))
	}

	if err != nil {
		common.LogFail(err.Error())
	}
}
