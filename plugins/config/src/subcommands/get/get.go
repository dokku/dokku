package main

import (
	"flag"
	"fmt"
	"os"

	common "github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config/src/configenv"
)

// get the given entries from the specified environment
func main() {
	args := flag.NewFlagSet("config:get", flag.ExitOnError)
	global := args.Bool("global", false, "--global: use the global environment")
	quoted := args.Bool("quoted", false, "--quoted: get the value quoted")
	args.Parse(os.Args[2:])
	appName := args.Arg(0)
	nextArg := 0
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

	if args.NArg() > nextArg+1 {
		common.LogFail(fmt.Sprintf("Unexpected argument(s): %v", args.Args()[nextArg+1:]))
	}
	key := args.Arg(nextArg)
	value := env.GetDefault(key, "")
	if *quoted {
		fmt.Printf("'%s'", configenv.SingleQuoteEscape(value))
	} else {
		fmt.Printf("%s", value)
	}
}
