package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/ps"

	flag "github.com/spf13/pflag"
)

// main entrypoint to all subcommands
func main() {
	parts := strings.Split(os.Args[0], "/")
	subcommand := parts[len(parts)-1]

	var err error
	switch subcommand {
	case "inspect":
		args := flag.NewFlagSet("ps:inspect", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = ps.CommandInspect(appName)
	case "rebuild":
		args := flag.NewFlagSet("ps:rebuild", flag.ExitOnError)
		allApps := args.Bool("all", false, "--all: rebuild all apps")
		parallelCount := args.Int("parallel", ps.RunInSerial, "--parallel: number of apps to rebuild in parallel, -1 to match cpu count")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = ps.CommandRebuild(appName, *allApps, *parallelCount)
	case "report":
		args := flag.NewFlagSet("ps:report", flag.ExitOnError)
		format := args.String("format", "stdout", "format: [ stdout | json ]")
		osArgs, infoFlag, flagErr := common.ParseReportArgs("ps", os.Args[2:])
		if flagErr == nil {
			args.Parse(osArgs)
			appName := args.Arg(0)
			err = ps.CommandReport(appName, *format, infoFlag)
		}
	case "restart":
		args := flag.NewFlagSet("ps:restart", flag.ExitOnError)
		allApps := args.Bool("all", false, "--all: restart all apps")
		parallelCount := args.Int("parallel", ps.RunInSerial, "--parallel: number of apps to restart in parallel, -1 to match cpu count")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = ps.CommandRestart(appName, *allApps, *parallelCount)
	case "restore":
		args := flag.NewFlagSet("ps:restore", flag.ExitOnError)
		allApps := args.Bool("all", false, "--all: restore all apps")
		parallelCount := args.Int("parallel", ps.RunInSerial, "--parallel: number of apps to restore in parallel, -1 to match cpu count")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = ps.CommandRestore(appName, *allApps, *parallelCount)
	case "retire":
		args := flag.NewFlagSet("ps:retire", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = ps.CommandRetire(appName)
	case "scale":
		args := flag.NewFlagSet("ps:scale", flag.ExitOnError)
		skipDeploy := args.Bool("skip-deploy", false, "--skip-deploy: skip deploy of the app")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		_, processTuples := common.ShiftString(args.Args())
		err = ps.CommandScale(appName, *skipDeploy, processTuples)
	case "set":
		args := flag.NewFlagSet("ps:set", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		property := args.Arg(1)
		value := args.Arg(2)
		err = ps.CommandSet(appName, property, value)
	case "start":
		args := flag.NewFlagSet("ps:start", flag.ExitOnError)
		allApps := args.Bool("all", false, "--all: start all apps")
		parallelCount := args.Int("parallel", ps.RunInSerial, "--parallel: number of apps to start in parallel, -1 to match cpu count")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = ps.CommandStart(appName, *allApps, *parallelCount)
	case "stop":
		args := flag.NewFlagSet("ps:stop", flag.ExitOnError)
		allApps := args.Bool("all", false, "--all: stop all apps")
		parallelCount := args.Int("parallel", ps.RunInSerial, "--parallel: number of apps to stop in parallel, -1 to match cpu count")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = ps.CommandStop(appName, *allApps, *parallelCount)
	default:
		err = fmt.Errorf("Invalid plugin subcommand call: %s", subcommand)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
