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
	args := flag.NewFlagSet("config:export", flag.ExitOnError)
	global := args.Bool("global", false, "--global: use the global environment")
	envfile := args.Bool("envfile", false, "--envfile: export as envfile rather than bash exports")
	args.Parse(os.Args[2:])

	appName := args.Arg(0)
	if appName == "" && !*global {
		common.LogFail("Please specify an app or --global")
	}

	if *global {
		if appName != "" {
			common.LogFail("Trailing argument: " + appName)
		}
		appName = "--global"
	}

	env, err := configenv.NewFromTarget(appName)
	if err != nil {
		common.LogFail(err.Error())
	}
	if *envfile {
		fmt.Println(env.EnvfileString())
	} else {
		fmt.Println(env.ExportfileString())
	}
}
