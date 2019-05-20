package main

import (
	"flag"
	"os"

	"github.com/dokku/dokku/plugins/config"
)

//clear all entries from the given environment
func main() {
	args := flag.NewFlagSet("config:clear", flag.ExitOnError)
	global := args.Bool("global", false, "--global: use the global environment")
	noRestart := args.Bool("no-restart", false, "--no-restart: no restart")
	args.Parse(os.Args[2:])
	config.CommandClear(args.Args(), *global, *noRestart)
}
