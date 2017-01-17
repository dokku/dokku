package main

import (
	"flag"

	"strings"

	"os"

	common "github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config/src/configenv"
)

// set the given entries from the specified environment
func main() {
	args := flag.NewFlagSet("config:set", flag.ExitOnError)
	global := args.Bool("global", false, "--global: use the global environment")
	noRestart := args.Bool("no-restart", false, "--no-restart: no restart")
	args.Parse(os.Args[2:])
	shouldRestart := !*global && !*noRestart
	var nextArg = 0
	appName := args.Arg(0)
	if appName == "" && !*global {
		common.LogFail("Please specify an app or --global")
	}

	if *global {
		appName = "--global"
	} else {
		nextArg = 1
	}

	env, err := configenv.NewFromTarget(appName)
	if err != nil {
		common.LogFail(err.Error())
	}

	var changed = false
	entries := args.Args()[nextArg:]
	for _, e := range entries {
		//log
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 1 {
			common.LogFail("Invalid env pair: " + e)
		}
		env.Set(parts[0], parts[1])
		changed = true
	}

	if changed {
		//log
		env.Write()
		args := append([]string{appName, "set"}, entries...)
		common.PlugnTrigger("post-config-update", args...)
	}

	if shouldRestart && env.GetBoolDefault("DOKKU_APP_RESTORE", true) {
		cmd := common.NewTokenizedShellCmd("dokku", "ps:restart", appName)
		cmd.Execute()
	}
}
