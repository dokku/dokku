package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/apps"
	"github.com/dokku/dokku/plugins/common"

	flag "github.com/spf13/pflag"
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
		oldAppName := args.Arg(0)
		newAppName := args.Arg(1)
		err = apps.CommandClone(oldAppName, newAppName, *skipDeploy, *ignoreExisting)
	case "create":
		args := flag.NewFlagSet("apps:create", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = apps.CommandCreate(appName)
	case "destroy":
		args := flag.NewFlagSet("apps:destroy", flag.ExitOnError)
		force := args.Bool("force", false, "--force: force destroy without confirmation")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = apps.CommandDestroy(appName, *force)
	case "exists":
		args := flag.NewFlagSet("apps:exists", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = apps.CommandExists(appName)
	case "list":
		args := flag.NewFlagSet("apps:list", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = apps.CommandList()
	case "lock":
		args := flag.NewFlagSet("apps:lock", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = apps.CommandLock(appName)
	case "locked":
		args := flag.NewFlagSet("apps:locked", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = apps.CommandLocked(appName)
	case "rename":
		args := flag.NewFlagSet("apps:rename", flag.ExitOnError)
		skipDeploy := args.Bool("skip-deploy", false, "--skip-deploy: skip deploy of the new app")
		args.Parse(os.Args[2:])
		oldAppName := args.Arg(0)
		newAppName := args.Arg(1)
		err = apps.CommandRename(oldAppName, newAppName, *skipDeploy)
	case "report":
		args := flag.NewFlagSet("apps:report", flag.ExitOnError)
		osArgs, infoFlag, err := common.ParseReportArgs("apps", os.Args[2:])
		if err == nil {
			args.Parse(osArgs)
			appName := args.Arg(0)
			err = apps.CommandReport(appName, infoFlag)
		}
	case "unlock":
		args := flag.NewFlagSet("apps:unlock", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = apps.CommandUnlock(appName)
	default:
		common.LogFail(fmt.Sprintf("Invalid plugin subcommand call: %s", subcommand))
	}

	if err != nil {
		common.LogFail(err.Error())
	}
}
