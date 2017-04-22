package main

import (
	"flag"
	"fmt"
	"os"

	common "github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config/src/configenv"
)

// print the environment to stdout
func main() {
	const defaultPrefix = "export "
	const defaultSeparator = "\n"
	args := flag.NewFlagSet("config:export", flag.ExitOnError)
	global := args.Bool("global", false, "--global: use the global environment")
	merged := args.Bool("merged", false, "--merged: merge app environment and global environment")
	keys := args.Bool("keys", false, "--keys: export keys only")
	bundle := args.Bool("bundle", false, "--bundle: export as tar bundle")
	envfile := args.Bool("envfile", false, "--envfile: export as envfile rather than bash exports (--prefix='')")
	prefix := args.String("prefix", defaultPrefix, "--prefix: prefix")
	separator := args.String("separator", defaultSeparator, "--separator: separator")
	escapeNewlines := args.Bool("escape-newlines", false, "--escape-newlines: replace literal newlines with $'\n'")
	args.Parse(os.Args[2:])

	appName := args.Arg(0)
	if appName == "" && !*global {
		common.LogFail("Please specify an app or --global")
	}

	if *global {
		if appName != "" {
			common.LogFail("Trailing argument: " + appName)
		}
		if *merged == true {
			common.LogFail("Only app environments can be merged")
		}
		appName = "--global"
	}

	env, err := configenv.NewFromTarget(appName)
	if err != nil {
		common.LogFail(err.Error())
	}

	if *merged {
		global, err := configenv.LoadGlobal()
		if err != nil {
			common.LogFail(err.Error())
		}
		global.Merge(env)
		env = global
	}

	if *keys {
		for _, k := range env.Keys() {
			fmt.Println(k)
		}
		return
	}

	if *escapeNewlines {
		env.EscapeNewlines = true
	}

	if *bundle {
		if *prefix != defaultPrefix || *separator != defaultSeparator || *envfile {
			common.LogFail("--bundle cannot be given with --envfile, --prefix, or --separator")
		}
		env.ExportBundle(os.Stdout)
	}
	if *envfile {
		if *prefix != defaultPrefix {
			common.LogFail("--prefix and --envfile cannot both be given")
			return
		}
		*prefix = ""
	}
	fmt.Println(env.StringWithPrefixAndSeparator(*prefix, *separator))
}
