package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/ps"
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
		allApps := args.Bool("all", false, "--all: restart all apps")
		parallelCount := args.Int("parallel", ps.RUN_IN_SERIAL, "--parallel: number of apps to restart in parallel, -1 to match cpu count")
		runInSerial := args.Bool("serial", false, "--serial: restart apps serially")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = ps.CommandRebuild(appName, *allApps, *runInSerial, *parallelCount)
	case "report":
		args := flag.NewFlagSet("ps:report", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		infoFlag := args.Arg(1)
		err = ps.CommandReport(appName, infoFlag)
	case "restart":
		args := flag.NewFlagSet("ps:restart", flag.ExitOnError)
		allApps := args.Bool("all", false, "--all: restart all apps")
		parallelCount := args.Int("parallel", ps.RUN_IN_SERIAL, "--parallel: number of apps to restart in parallel, -1 to match cpu count")
		runInSerial := args.Bool("serial", false, "--serial: restart apps serially")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = ps.CommandRestart(appName, *allApps, *runInSerial, *parallelCount)
	case "restore":
		args := flag.NewFlagSet("ps:restore", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = ps.CommandRestore(appName)
	case "retire":
		args := flag.NewFlagSet("ps:retire", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = ps.CommandRetire()
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
		allApps := args.Bool("all", false, "--all: restart all apps")
		parallelCount := args.Int("parallel", ps.RUN_IN_SERIAL, "--parallel: number of apps to restart in parallel, -1 to match cpu count")
		runInSerial := args.Bool("serial", false, "--serial: restart apps serially")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = ps.CommandStart(appName, *allApps, *runInSerial, *parallelCount)
	case "stop":
		args := flag.NewFlagSet("ps:stop", flag.ExitOnError)
		allApps := args.Bool("all", false, "--all: restart all apps")
		parallelCount := args.Int("parallel", ps.RUN_IN_SERIAL, "--parallel: number of apps to restart in parallel, -1 to match cpu count")
		runInSerial := args.Bool("serial", false, "--serial: restart apps serially")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = ps.CommandStop(appName, *allApps, *runInSerial, *parallelCount)
	default:
		common.LogFail(fmt.Sprintf("Invalid plugin subcommand call: %s", subcommand))
	}

	if err != nil {
		common.LogFail(err.Error())
	}
}
