package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"

	flag "github.com/spf13/pflag"
)

func getKeys(args []string, global bool) []string {
	keys := args
	if !global && len(keys) > 0 {
		keys = keys[1:]
	}
	return keys
}

// main entrypoint to all subcommands
func main() {
	parts := strings.Split(os.Args[0], "/")
	subcommand := parts[len(parts)-1]

	var err error
	var appName string
	switch subcommand {
	case "bundle":
		args := flag.NewFlagSet("config:bundle", flag.ExitOnError)
		global := args.Bool("global", false, "--global: use the global environment")
		merged := args.Bool("merged", false, "--merged: merge app environment and global environment")
		args.Parse(os.Args[2:])
		if !*global {
			appName = args.Arg(0)
		}
		err = config.CommandBundle(appName, *global, *merged)
	case "clear":
		args := flag.NewFlagSet("config:clear", flag.ExitOnError)
		global := args.Bool("global", false, "--global: use the global environment")
		noRestart := args.Bool("no-restart", false, "--no-restart: no restart")
		args.Parse(os.Args[2:])
		if !*global {
			appName = args.Arg(0)
		}
		err = config.CommandClear(appName, *global, *noRestart)
	case "export":
		args := flag.NewFlagSet("config:export", flag.ExitOnError)
		global := args.Bool("global", false, "--global: use the global environment")
		merged := args.Bool("merged", false, "--merged: merge app environment and global environment")
		format := args.String("format", "exports", "--format: [ docker-args | docker-args-keys | exports | envfile | json | json-list | pack-keys | pretty | shell ] which format to export as)")
		args.Parse(os.Args[2:])
		if !*global {
			appName = args.Arg(0)
		}
		err = config.CommandExport(appName, *global, *merged, *format)
	case "get":
		args := flag.NewFlagSet("config:get", flag.ExitOnError)
		global := args.Bool("global", false, "--global: use the global environment")
		quoted := args.Bool("quoted", false, "--quoted: get the value quoted")
		args.Parse(os.Args[2:])
		if !*global {
			appName = args.Arg(0)
		}
		keys := getKeys(args.Args(), *global)
		err = config.CommandGet(appName, keys, *global, *quoted)
	case "keys":
		args := flag.NewFlagSet("config:keys", flag.ExitOnError)
		global := args.Bool("global", false, "--global: use the global environment")
		merged := args.Bool("merged", false, "--merged: merge app environment and global environment")
		args.Parse(os.Args[2:])
		if !*global {
			appName = args.Arg(0)
		}
		err = config.CommandKeys(appName, *global, *merged)
	case "show":
		args := flag.NewFlagSet("config:show", flag.ExitOnError)
		global := args.Bool("global", false, "--global: use the global environment")
		merged := args.Bool("merged", false, "--merged: display the app's environment merged with the global environment")
		args.Parse(os.Args[2:])
		if !*global {
			appName = args.Arg(0)
		}
		err = config.CommandShow(appName, *global, *merged, false, false)
	case "set":
		args := flag.NewFlagSet("config:set", flag.ExitOnError)
		global := args.Bool("global", false, "--global: use the global environment")
		encoded := args.Bool("encoded", false, "--encoded: interpret VALUEs as base64")
		noRestart := args.Bool("no-restart", false, "--no-restart: no restart")
		args.Parse(os.Args[2:])
		if !*global {
			appName = args.Arg(0)
		}
		pairs := getKeys(args.Args(), *global)
		err = config.CommandSet(appName, pairs, *global, *noRestart, *encoded)
	case "unset":
		args := flag.NewFlagSet("config:unset", flag.ExitOnError)
		global := args.Bool("global", false, "--global: use the global environment")
		noRestart := args.Bool("no-restart", false, "--no-restart: no restart")
		args.Parse(os.Args[2:])
		if !*global {
			appName = args.Arg(0)
		}
		keys := getKeys(args.Args(), *global)
		err = config.CommandUnset(appName, keys, *global, *noRestart)
	default:
		err = fmt.Errorf("Invalid plugin subcommand call: %s", subcommand)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
