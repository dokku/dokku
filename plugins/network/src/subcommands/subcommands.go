package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/network"
)

// main entrypoint to all subcommands
func main() {
	parts := strings.Split(os.Args[0], "/")
	subcommand := parts[len(parts)-1]
	flag.Parse()

	switch subcommand {
	case "create":
		networkName := flag.Arg(1)
		err := network.CommandCreate(networkName)
		if err != nil {
			common.LogFail(err.Error())
		}
	case "destroy":
		networkName := flag.Arg(1)
		forceDestroy := false
		if flag.Arg(2) == "force" {
			forceDestroy = true
		}
		err := network.CommandDestroy(networkName, forceDestroy)
		if err != nil {
			common.LogFail(err.Error())
		}
	case "exists":
		networkName := flag.Arg(1)
		err := network.CommandExists(networkName)
		if err != nil {
			common.LogFail(err.Error())
		}
	case "info":
		err := network.CommandInfo()
		if err != nil {
			common.LogFail(err.Error())
		}
	case "list":
		err := network.CommandList()
		if err != nil {
			common.LogFail(err.Error())
		}
	case "rebuild":
		appName := flag.Arg(1)
		network.BuildConfig(appName)
	case "rebuildall":
		network.CommandRebuildall()
	case "report":
		appName := flag.Arg(1)
		infoFlag := flag.Arg(2)
		network.CommandReport(appName, infoFlag)
	case "set":
		appName := flag.Arg(1)
		property := flag.Arg(2)
		value := flag.Arg(3)
		network.CommandSet(appName, property, value)
	default:
		common.LogFail(fmt.Sprintf("Invalid plugin subcommand call: %s", subcommand))
	}
}
