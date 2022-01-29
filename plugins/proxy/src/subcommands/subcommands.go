package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/proxy"

	flag "github.com/spf13/pflag"
)

// main entrypoint to all subcommands
func main() {
	parts := strings.Split(os.Args[0], "/")
	subcommand := parts[len(parts)-1]

	var err error
	switch subcommand {
	case "build-config":
		args := flag.NewFlagSet("proxy:build-config", flag.ExitOnError)
		allApps := args.Bool("all", false, "--all: build-config for all apps")
		parallelCount := args.Int("parallel", proxy.RunInSerial, "--parallel: number of apps to build-config for in parallel, -1 to match cpu count")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = proxy.CommandBuildConfig(appName, *allApps, *parallelCount)
	case "clear-config":
		args := flag.NewFlagSet("proxy:clear-config", flag.ExitOnError)
		allApps := args.Bool("all", false, "--all: build-config for all apps")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = proxy.CommandClearConfig(appName, *allApps)
	case "disable":
		args := flag.NewFlagSet("proxy:disable", flag.ExitOnError)
		allApps := args.Bool("all", false, "--all: disable proxy for all apps")
		parallelCount := args.Int("parallel", proxy.RunInSerial, "--parallel: number of apps to disable proxy for in parallel, -1 to match cpu count")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = proxy.CommandDisable(appName, *allApps, *parallelCount)
	case "enable":
		args := flag.NewFlagSet("proxy:enable", flag.ExitOnError)
		allApps := args.Bool("all", false, "--all: enable proxy for all apps")
		parallelCount := args.Int("parallel", proxy.RunInSerial, "--parallel: number of apps to enable proxy for in parallel, -1 to match cpu count")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = proxy.CommandEnable(appName, *allApps, *parallelCount)
	case "ports":
		args := flag.NewFlagSet("proxy:ports", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = proxy.CommandPorts(appName)
	case "ports-add":
		args := flag.NewFlagSet("proxy:ports-add", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		_, portMaps := common.ShiftString(args.Args())
		err = proxy.CommandPortsAdd(appName, portMaps)
	case "ports-clear":
		args := flag.NewFlagSet("proxy:ports-clear", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = proxy.CommandPortsClear(appName)
	case "ports-remove":
		args := flag.NewFlagSet("proxy:ports-remove", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		_, portMaps := common.ShiftString(args.Args())
		err = proxy.CommandPortsRemove(appName, portMaps)
	case "ports-set":
		args := flag.NewFlagSet("proxy:ports-set", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		_, portMaps := common.ShiftString(args.Args())
		err = proxy.CommandPortsSet(appName, portMaps)
	case "report":
		args := flag.NewFlagSet("proxy:report", flag.ExitOnError)
		format := args.String("format", "stdout", "format: [ stdout | json ]")
		osArgs, infoFlag, flagErr := common.ParseReportArgs("proxy", os.Args[2:])
		if flagErr == nil {
			args.Parse(osArgs)
			appName := args.Arg(0)
			err = proxy.CommandReport(appName, *format, infoFlag)
		}
	case "set":
		args := flag.NewFlagSet("proxy:set", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		proxyType := args.Arg(1)
		err = proxy.CommandSet(appName, proxyType)
	default:
		err = fmt.Errorf("Invalid plugin subcommand call: %s", subcommand)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
