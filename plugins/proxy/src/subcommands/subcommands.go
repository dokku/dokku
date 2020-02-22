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
	case "disable":
		args := flag.NewFlagSet("proxy:disable", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = proxy.CommandDisable(args.Args())
	case "enable":
		args := flag.NewFlagSet("proxy:enable", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = proxy.CommandEnable(args.Args())
	case "ports":
		args := flag.NewFlagSet("proxy:ports", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = proxy.CommandPorts(args.Args())
	case "ports-add":
		args := flag.NewFlagSet("proxy:ports-add", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = proxy.CommandPortsAdd(args.Args())
	case "ports-clear":
		args := flag.NewFlagSet("proxy:ports-clear", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = proxy.CommandPortsClear(args.Args())
	case "ports-remove":
		args := flag.NewFlagSet("proxy:ports-remove", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = proxy.CommandPortsRemove(args.Args())
	case "ports-set":
		args := flag.NewFlagSet("proxy:ports-set", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = proxy.CommandPortsSet(args.Args())
	case "report":
		args := flag.NewFlagSet("proxy:report", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = proxy.CommandReport(args.Args())
	case "set":
		args := flag.NewFlagSet("proxy:set", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = proxy.CommandSet(args.Args())
	default:
		common.LogFail(fmt.Sprintf("Invalid plugin subcommand call: %s", subcommand))
	}

	if err != nil {
		common.LogFail(err.Error())
	}
}
