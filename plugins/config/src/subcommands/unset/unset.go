package main

import (
	"flag"
	"os"

	common "github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config/src/configenv"
)

//unset the given entries from the given environment
func main() {
	args := flag.NewFlagSet("config:unset", flag.ExitOnError)
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
	keys := args.Args()[nextArg:]
	for _, k := range keys {
		//log
		env.Unset(k)
		changed = true
	}

	if changed {
		//log
		env.Write()
		args := append([]string{appName, "unset"}, keys...)
		common.PlugnTrigger("post-config-update", args...)
	}

	if shouldRestart && env.GetBoolDefault("DOKKU_APP_RESTORE", true) {
		cmd := common.NewTokenizedShellCmd("dokku", "ps:restart", appName)
		cmd.Execute()
	}
}
