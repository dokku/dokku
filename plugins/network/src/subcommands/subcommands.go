package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/network"

	flag "github.com/spf13/pflag"
)

// main entrypoint to all subcommands
func main() {
	parts := strings.Split(os.Args[0], "/")
	subcommand := parts[len(parts)-1]

	var err error
	switch subcommand {
	case "create":
		args := flag.NewFlagSet("network:create", flag.ExitOnError)
		args.Parse(os.Args[2:])
		networkName := args.Arg(0)
		err = network.CommandCreate(networkName)
	case "destroy":
		args := flag.NewFlagSet("network:destroy", flag.ExitOnError)
		force := args.Bool("force", false, "--force: force destroy without confirmation")
		args.Parse(os.Args[2:])
		networkName := args.Arg(0)
		err = network.CommandDestroy(networkName, *force)
	case "exists":
		args := flag.NewFlagSet("network:exists", flag.ExitOnError)
		args.Parse(os.Args[2:])
		networkName := args.Arg(0)
		err = network.CommandExists(networkName)
	case "info":
		args := flag.NewFlagSet("network:info", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = network.CommandInfo()
	case "list":
		args := flag.NewFlagSet("network:list", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = network.CommandList()
	case "rebuild":
		args := flag.NewFlagSet("network:rebuild", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = network.BuildConfig(appName)
	case "rebuildall":
		args := flag.NewFlagSet("network:rebuildall", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = network.CommandRebuildall()
	case "report":
		args := flag.NewFlagSet("network:report", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		infoFlag := args.Arg(1)
		err = network.CommandReport(appName, infoFlag)
	case "set":
		args := flag.NewFlagSet("network:set", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		property := args.Arg(1)
		value := args.Arg(2)
		err = network.CommandSet(appName, property, value)
	default:
		common.LogFail(fmt.Sprintf("Invalid plugin subcommand call: %s", subcommand))
	}

	if err != nil {
		common.LogFail(err.Error())
	}
}
