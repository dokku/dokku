package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/proxy"
)

// main entrypoint to all subcommands
func main() {
	parts := strings.Split(os.Args[0], "/")
	subcommand := parts[len(parts)-1]

	var err error
	switch subcommand {
	case "build-config":
		args := flag.NewFlagSet("proxy:build-config", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = proxy.CommandBuildConfig(appName)
	case "disable":
		args := flag.NewFlagSet("proxy:disable", flag.ExitOnError)
		skipRestart := args.Bool("no-restart", false, "--no-restart: skip restart of the app")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = proxy.CommandDisable(appName, *skipRestart)
	case "enable":
		args := flag.NewFlagSet("proxy:enable", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = proxy.CommandEnable(appName)
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
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		infoFlag := args.Arg(1)
		err = proxy.CommandReport(appName, infoFlag)
	case "set":
		args := flag.NewFlagSet("proxy:set", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		proxyType := args.Arg(1)
		err = proxy.CommandSet(appName, proxyType)
	default:
		common.LogFail(fmt.Sprintf("Invalid plugin subcommand call: %s", subcommand))
	}

	if err != nil {
		common.LogFail(err.Error())
	}
}
