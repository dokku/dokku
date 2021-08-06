package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/registry"

	flag "github.com/spf13/pflag"
)

// main entrypoint to all subcommands
func main() {
	parts := strings.Split(os.Args[0], "/")
	subcommand := parts[len(parts)-1]

	var err error
	switch subcommand {
	case "login":
		args := flag.NewFlagSet("registry:login", flag.ExitOnError)
		passwordStdin := args.Bool("password-stdin", false, "--password-stdin: read password from stdin")
		args.Parse(os.Args[2:])
		server := args.Arg(0)
		username := args.Arg(1)
		password := args.Arg(2)
		err = registry.CommandLogin(server, username, password, *passwordStdin)
	case "report":
		args := flag.NewFlagSet("registry:report", flag.ExitOnError)
		format := args.String("format", "stdout", "format: [ stdout | json ]")
		osArgs, infoFlag, flagErr := common.ParseReportArgs("registry", os.Args[2:])
		if flagErr == nil {
			args.Parse(osArgs)
			appName := args.Arg(0)
			err = registry.CommandReport(appName, *format, infoFlag)
		}
	case "set":
		args := flag.NewFlagSet("registry:set", flag.ExitOnError)
		global := args.Bool("global", false, "--global: set a global property")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		property := args.Arg(1)
		value := args.Arg(2)
		if *global {
			appName = "--global"
			property = args.Arg(0)
			value = args.Arg(1)
		}
		err = registry.CommandSet(appName, property, value)
	default:
		err = fmt.Errorf("Invalid plugin subcommand call: %s", subcommand)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
