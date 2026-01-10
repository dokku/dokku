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
		global := args.Bool("global", false, "--global: login globally instead of per-app")
		args.Parse(os.Args[2:])

		argCount := args.NArg()
		var appName, server, username, password string

		// When --password-stdin is used, password is not in args
		// Global login: 2 args (server, username) with --password-stdin, 3 args without
		// Per-app login: 3 args (app, server, username) with --password-stdin, 4 args without
		globalArgCount := 3
		perAppArgCount := 4
		if *passwordStdin {
			globalArgCount = 2
			perAppArgCount = 3
		}

		if *global {
			// --global: server, username, [password]
			server = args.Arg(0)
			username = args.Arg(1)
			if !*passwordStdin {
				password = args.Arg(2)
			}
		} else if argCount == globalArgCount {
			// global login without --global flag: warn and treat as global
			common.LogWarn("Deprecated: please use --global flag for global registry login")
			server = args.Arg(0)
			username = args.Arg(1)
			if !*passwordStdin {
				password = args.Arg(2)
			}
		} else if argCount >= perAppArgCount {
			// per-app login: app, server, username, [password]
			appName = args.Arg(0)
			server = args.Arg(1)
			username = args.Arg(2)
			if !*passwordStdin {
				password = args.Arg(3)
			}
		}

		err = registry.CommandLogin(appName, server, username, password, *passwordStdin)
	case "logout":
		args := flag.NewFlagSet("registry:logout", flag.ExitOnError)
		global := args.Bool("global", false, "--global: logout globally instead of per-app")
		args.Parse(os.Args[2:])

		var appName, server string
		if *global {
			server = args.Arg(0)
		} else if args.NArg() == 1 {
			// 1 arg: global logout (backwards compatible)
			server = args.Arg(0)
		} else {
			// 2 args: app, server
			appName = args.Arg(0)
			server = args.Arg(1)
		}

		err = registry.CommandLogout(appName, server)
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
