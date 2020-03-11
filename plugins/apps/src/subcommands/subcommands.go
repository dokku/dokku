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
		skipDeploy := args.Bool("skip-deploy", false, "--skip-deploy: skip deploy of the new app")
		ignoreExisting := args.Bool("ignore-existing", false, "--ignore-existing: exit 0 if new app already exists")
		args.Parse(os.Args[2:])
		err = apps.CommandClone(args.Args(), *skipDeploy, *ignoreExisting)
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
		skipDeploy := args.Bool("skip-deploy", false, "--skip-deploy: skip deploy of the new app")
		args.Parse(os.Args[2:])
		err = apps.CommandRename(args.Args(), *skipDeploy)
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
