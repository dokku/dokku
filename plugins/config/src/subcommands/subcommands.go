package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
)

// main entrypoint to all subcommands
func main() {
	parts := strings.Split(os.Args[0], "/")
	subcommand := parts[len(parts)-1]

	switch subcommand {
	case "bundle":
		args := flag.NewFlagSet("config:bundle", flag.ExitOnError)
		global := args.Bool("global", false, "--global: use the global environment")
		merged := args.Bool("merged", false, "--merged: merge app environment and global environment")
		args.Parse(os.Args[2:])
		config.CommandBundle(args.Args(), *global, *merged)
	case "clear":
		args := flag.NewFlagSet("config:clear", flag.ExitOnError)
		global := args.Bool("global", false, "--global: use the global environment")
		noRestart := args.Bool("no-restart", false, "--no-restart: no restart")
		args.Parse(os.Args[2:])
		config.CommandClear(args.Args(), *global, *noRestart)
	case "export":
		args := flag.NewFlagSet("config:export", flag.ExitOnError)
		global := args.Bool("global", false, "--global: use the global environment")
		merged := args.Bool("merged", false, "--merged: merge app environment and global environment")
		format := args.String("format", "exports", "--format: [ exports | envfile | docker-args | shell | pretty | json | json-list ] which format to export as)")
		args.Parse(os.Args[2:])
		config.CommandExport(args.Args(), *global, *merged, *format)
	case "get":
		args := flag.NewFlagSet("config:get", flag.ExitOnError)
		global := args.Bool("global", false, "--global: use the global environment")
		quoted := args.Bool("quoted", false, "--quoted: get the value quoted")
		args.Parse(os.Args[2:])
		config.CommandGet(args.Args(), *global, *quoted)
	case "keys":
		args := flag.NewFlagSet("config:keys", flag.ExitOnError)
		global := args.Bool("global", false, "--global: use the global environment")
		merged := args.Bool("merged", false, "--merged: merge app environment and global environment")
		args.Parse(os.Args[2:])
		config.CommandKeys(args.Args(), *global, *merged)
	case "set":
		args := flag.NewFlagSet("config:set", flag.ExitOnError)
		global := args.Bool("global", false, "--global: use the global environment")
		encoded := args.Bool("encoded", false, "--encoded: interpret VALUEs as base64")
		noRestart := args.Bool("no-restart", false, "--no-restart: no restart")
		args.Parse(os.Args[2:])
		config.CommandSet(args.Args(), *global, *noRestart, *encoded)
	case "unset":
		args := flag.NewFlagSet("config:unset", flag.ExitOnError)
		global := args.Bool("global", false, "--global: use the global environment")
		noRestart := args.Bool("no-restart", false, "--no-restart: no restart")
		args.Parse(os.Args[2:])
		config.CommandUnset(args.Args(), *global, *noRestart)
	default:
		common.LogFail(fmt.Sprintf("Invalid plugin subcommand call: %s", subcommand))
	}

}
