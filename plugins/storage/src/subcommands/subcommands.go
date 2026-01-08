package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/storage"

	flag "github.com/spf13/pflag"
)

func main() {
	parts := strings.Split(os.Args[0], "/")
	subcommand := parts[len(parts)-1]

	var err error
	switch subcommand {
	case "default":
		err = storage.CommandHelp()
	case "ensure-directory":
		args := flag.NewFlagSet("storage:ensure-directory", flag.ExitOnError)
		chown := args.String("chown", "herokuish", "--chown: chown option (herokuish, heroku, paketo, root, false)")
		args.Parse(os.Args[2:])
		directory := args.Arg(0)
		err = storage.CommandEnsureDirectory(directory, *chown)
	case "list":
		args := flag.NewFlagSet("storage:list", flag.ExitOnError)
		format := args.String("format", "text", "--format: output format (text, json)")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = storage.CommandList(appName, *format)
	case "mount":
		args := flag.NewFlagSet("storage:mount", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		mountPath := args.Arg(1)
		err = storage.CommandMount(appName, mountPath)
	case "report":
		args := flag.NewFlagSet("storage:report", flag.ExitOnError)
		format := args.String("format", "stdout", "--format: output format (stdout, json)")
		args.Parse(os.Args[2:])

		osArgs, infoFlag, parseErr := common.ParseReportArgs("storage", args.Args())
		if parseErr != nil {
			err = parseErr
		} else {
			appName := ""
			if len(osArgs) > 0 {
				appName = osArgs[0]
			}
			if *format == "stdout" {
				err = storage.CommandReport(appName, infoFlag)
			} else {
				err = storage.CommandReportSingleApp(appName, infoFlag, *format)
			}
		}
	case "unmount":
		args := flag.NewFlagSet("storage:unmount", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		mountPath := args.Arg(1)
		err = storage.CommandUnmount(appName, mountPath)
	default:
		err = fmt.Errorf("Invalid plugin subcommand call: %s", subcommand)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
