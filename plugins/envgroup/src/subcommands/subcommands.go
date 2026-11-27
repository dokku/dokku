package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/envgroup"

	flag "github.com/spf13/pflag"
)

// main entrypoint to all subcommands
func main() {
	parts := strings.Split(os.Args[0], "/")
	subcommand := parts[len(parts)-1]

	var err error
	switch subcommand {
	case "add":
		args := flag.NewFlagSet("envgroup:add", flag.ExitOnError)
		index := args.Int("index", 0, "--index: the 1-based position to add the env group at")
		restart := args.Bool("restart", false, "--restart: redeploy the app")
		args.Parse(os.Args[2:])
		err = envgroup.CommandAdd(args.Arg(0), args.Arg(1), *index, *restart)
	case "bundle":
		args := flag.NewFlagSet("envgroup:bundle", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = envgroup.CommandBundle(args.Arg(0))
	case "clear":
		args := flag.NewFlagSet("envgroup:clear", flag.ExitOnError)
		restart := args.Bool("restart", false, "--restart: redeploy the apps in the env group")
		args.Parse(os.Args[2:])
		err = envgroup.CommandClear(args.Arg(0), *restart)
	case "create":
		args := flag.NewFlagSet("envgroup:create", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = envgroup.CommandCreate(args.Arg(0))
	case "destroy":
		args := flag.NewFlagSet("envgroup:destroy", flag.ExitOnError)
		force := args.Bool("force", false, "--force: remove the env group from its apps and destroy it")
		restart := args.Bool("restart", false, "--restart: redeploy the apps in the env group")
		args.Parse(os.Args[2:])
		err = envgroup.CommandDestroy(args.Arg(0), *force, *restart)
	case "exists":
		args := flag.NewFlagSet("envgroup:exists", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = envgroup.CommandExists(args.Arg(0))
	case "export":
		args := flag.NewFlagSet("envgroup:export", flag.ExitOnError)
		format := args.String("format", "exports", "--format: [ docker-args | docker-args-keys | exports | envfile | json | json-list | pack-keys | pretty | shell ] which format to export as)")
		args.Parse(os.Args[2:])
		err = envgroup.CommandExport(args.Arg(0), *format)
	case "get":
		args := flag.NewFlagSet("envgroup:get", flag.ExitOnError)
		quoted := args.Bool("quoted", false, "--quoted: get the value quoted")
		args.Parse(os.Args[2:])
		err = envgroup.CommandGet(args.Arg(0), common.VarArgs(args.Args(), 1), *quoted)
		if errors.Is(err, envgroup.ErrKeyNotSet) {
			os.Exit(1)
		}
	case "import":
		args := flag.NewFlagSet("envgroup:import", flag.ExitOnError)
		format := args.String("format", "envfile", "--format: [ envfile | json ] which format to import from")
		replace := args.Bool("replace", false, "--replace: replace existing config vars")
		restart := args.Bool("restart", false, "--restart: redeploy the apps in the env group")
		args.Parse(os.Args[2:])
		err = envgroup.CommandImport(args.Arg(0), args.Arg(1), *format, *replace, *restart)
	case "info":
		args := flag.NewFlagSet("envgroup:info", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = envgroup.CommandInfo(args.Arg(0))
	case "keys":
		args := flag.NewFlagSet("envgroup:keys", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = envgroup.CommandKeys(args.Arg(0))
	case "list":
		args := flag.NewFlagSet("envgroup:list", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = envgroup.CommandList()
	case "remove":
		args := flag.NewFlagSet("envgroup:remove", flag.ExitOnError)
		restart := args.Bool("restart", false, "--restart: redeploy the app")
		args.Parse(os.Args[2:])
		err = envgroup.CommandRemove(args.Arg(0), args.Arg(1), *restart)
	case "report":
		args := flag.NewFlagSet("envgroup:report", flag.ExitOnError)
		format := args.String("format", "stdout", "format: [ stdout | json ]")
		reportArgs, flagErr := common.ParseReportArgs("envgroup", os.Args[2:])
		if flagErr == nil {
			args.Parse(reportArgs.OSArgs)
			err = envgroup.CommandReport(args.Arg(0), *format, reportArgs.InfoFlag)
		}
	case "set":
		args := flag.NewFlagSet("envgroup:set", flag.ExitOnError)
		encoded := args.Bool("encoded", false, "--encoded: interpret VALUEs as base64")
		restart := args.Bool("restart", false, "--restart: redeploy the apps in the env group")
		args.Parse(os.Args[2:])
		err = envgroup.CommandSet(args.Arg(0), common.VarArgs(args.Args(), 1), *encoded, *restart)
	case "show":
		args := flag.NewFlagSet("envgroup:show", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = envgroup.CommandShow(args.Arg(0))
	case "unset":
		args := flag.NewFlagSet("envgroup:unset", flag.ExitOnError)
		restart := args.Bool("restart", false, "--restart: redeploy the apps in the env group")
		args.Parse(os.Args[2:])
		err = envgroup.CommandUnset(args.Arg(0), common.VarArgs(args.Args(), 1), *restart)
	default:
		err = fmt.Errorf("Invalid plugin subcommand call: %s", subcommand)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
