package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/apps"
	"github.com/dokku/dokku/plugins/common"
)

// main entrypoint to all subcommands
func main() {
	parts := strings.Split(os.Args[0], "/")
	subcommand := parts[len(parts)-1]

	var err error
	switch subcommand {
	case "clone":
		args := flag.NewFlagSet("apps:clone", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = apps.CommandClone(args.Args())
	case "create":
		args := flag.NewFlagSet("apps:create", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = apps.CommandCreate(args.Args())
	case "destroy":
		args := flag.NewFlagSet("apps:destroy", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = apps.CommandDestroy(args.Args())
	case "exists":
		args := flag.NewFlagSet("apps:exists", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = apps.CommandExists(args.Args())
	case "list":
		args := flag.NewFlagSet("apps:list", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = apps.CommandList(args.Args())
	case "lock":
		args := flag.NewFlagSet("apps:lock", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = apps.CommandLock(args.Args())
	case "locked":
		args := flag.NewFlagSet("apps:locked", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = apps.CommandLocked(args.Args())
	case "rename":
		args := flag.NewFlagSet("apps:rename", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = apps.CommandRename(args.Args())
	case "report":
		args := flag.NewFlagSet("apps:report", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = apps.CommandReport(args.Args())
	case "unlock":
		args := flag.NewFlagSet("apps:unlock", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = apps.CommandUnlock(args.Args())
	default:
		common.LogFail(fmt.Sprintf("Invalid plugin subcommand call: %s", subcommand))
	}

	if err != nil {
		common.LogFail(err.Error())
	}
}
