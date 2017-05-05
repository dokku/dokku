package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"strings"

	common "github.com/dokku/dokku/plugins/common"
	config "github.com/dokku/dokku/plugins/config"
	"github.com/dokku/dokku/plugins/config/src/configenv"
)

// set the given entries from the specified environment
func main() {
	args := flag.NewFlagSet("config:set", flag.ExitOnError)
	global := args.Bool("global", false, "--global: use the global environment")
	encoded := args.Bool("encoded", false, "--encoded: interpret VALUEs as base64")
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

	entries := args.Args()[nextArg:]
	updated := make(map[string]string)
	for _, e := range entries {
		//log
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 1 {
			common.LogFail("Invalid env pair: " + e)
		}
		key, value := parts[0], parts[1]
		if *encoded {
			decoded, err := base64.StdEncoding.DecodeString(value)
			if err != nil {
				common.LogFail(fmt.Sprintf("%s for key '%s'", err.Error(), key))
			}
			value = string(decoded)
		}
		env.Set(key, value)
		updated[key] = value
	}

	if len(updated) != 0 {
		common.LogInfo1("Setting config vars")
		fmt.Println(config.PrettyPrintLogEntries("       ", updated))
		env.Write()
		args := append([]string{appName, "set"}, entries...)
		common.PlugnTrigger("post-config-update", args...)
	}

	if shouldRestart && env.GetBoolDefault("DOKKU_APP_RESTORE", true) {
		common.LogInfo1(fmt.Sprintf("Restarting app %s", appName))
		cmd := common.NewTokenizedShellCmd("dokku", "ps:restart", appName)
		cmd.Execute()
	}
}
