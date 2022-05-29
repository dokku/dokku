package main

import (
	"fmt"
	"os"

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
	action := os.Args[1]

	var err error
	appName := "--global"
	switch action {
	case "bundle":
		args := flag.NewFlagSet("bundle", flag.ExitOnError)
		global := args.Bool("global", false, "--global: use the global environment")
		merged := args.Bool("merged", false, "--merged: merge app environment and global environment")
		args.Parse(os.Args[2:])
		if !*global {
			appName = args.Arg(0)
		}
		err = config.SubBundle(appName, *merged)
	case "clear":
		args := flag.NewFlagSet("clear", flag.ExitOnError)
		global := args.Bool("global", false, "--global: use the global environment")
		noRestart := args.Bool("no-restart", false, "--no-restart: no restart")
		args.Parse(os.Args[2:])
		if !*global {
			appName = args.Arg(0)
		}
		err = config.SubClear(appName, *noRestart)
	case "export":
		args := flag.NewFlagSet("export", flag.ExitOnError)
		global := args.Bool("global", false, "--global: use the global environment")
		merged := args.Bool("merged", false, "--merged: merge app environment and global environment")
		format := args.String("format", "exports", "--format: [ docker-args | docker-args-keys | exports | envfile | json | json-list | pack-keys | pretty | shell ] which format to export as)")
		args.Parse(os.Args[2:])
		if !*global {
			appName = args.Arg(0)
		}
		err = config.SubExport(appName, *merged, *format)
	case "get":
		args := flag.NewFlagSet("get", flag.ExitOnError)
		global := args.Bool("global", false, "--global: use the global environment")
		quoted := args.Bool("quoted", false, "--quoted: get the value quoted")
		args.Parse(os.Args[2:])
		if !*global {
			appName = args.Arg(0)
		}
		keys := getKeys(args.Args(), *global)
		err = config.SubGet(appName, keys, *quoted)
	case "keys":
		args := flag.NewFlagSet("keys", flag.ExitOnError)
		global := args.Bool("global", false, "--global: use the global environment")
		merged := args.Bool("merged", false, "--merged: merge app environment and global environment")
		args.Parse(os.Args[2:])
		if !*global {
			appName = args.Arg(0)
		}
		err = config.SubKeys(appName, *merged)
	case "show":
		args := flag.NewFlagSet("show", flag.ExitOnError)
		global := args.Bool("global", false, "--global: use the global environment")
		merged := args.Bool("merged", false, "--merged: display the app's environment merged with the global environment")
		args.Parse(os.Args[2:])
		if !*global {
			appName = args.Arg(0)
		}
		err = config.SubShow(appName, *merged, false, false)
	case "set":
		args := flag.NewFlagSet("set", flag.ExitOnError)
		global := args.Bool("global", false, "--global: use the global environment")
		encoded := args.Bool("encoded", false, "--encoded: interpret VALUEs as base64")
		noRestart := args.Bool("no-restart", false, "--no-restart: no restart")
		args.Parse(os.Args[2:])
		if !*global {
			appName = args.Arg(0)
		}
		pairs := getKeys(args.Args(), *global)
		err = config.SubSet(appName, pairs, *noRestart, *encoded)
	case "unset":
		args := flag.NewFlagSet("unset", flag.ExitOnError)
		global := args.Bool("global", false, "--global: use the global environment")
		noRestart := args.Bool("no-restart", false, "--no-restart: no restart")
		args.Parse(os.Args[2:])
		if !*global {
			appName = args.Arg(0)
		}
		keys := getKeys(args.Args(), *global)
		err = config.SubUnset(appName, keys, *noRestart)
	default:
		err = fmt.Errorf("Invalid plugin config_sub call: %s", action)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
