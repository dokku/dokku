package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
)

func main() {
	args := flag.NewFlagSet("config:get", flag.ExitOnError)
	global := args.Bool("global", false, "--global: use the global environment")
	merged := args.Bool("merged", false, "--merged: merge app environment and global environment")
	args.Parse(os.Args[2:])

	appName, trailingArgs := config.GetCommonArgs(*global, args.Args())
	if len(trailingArgs) > 0 {
		common.LogFail(fmt.Sprintf("Trailing argument(s): %v", trailingArgs))
	}
	config := config.GetConfig(appName, *merged)
	for _, k := range config.Keys() {
		fmt.Println(k)
	}
}
