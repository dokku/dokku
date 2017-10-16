package main

import (
	"flag"
	"os"

	"github.com/dokku/dokku/plugins/config"
)

//unset the given entries from the given environment
func main() {
	args := flag.NewFlagSet("config:unset", flag.ExitOnError)
	global := args.Bool("global", false, "--global: use the global environment")
	noRestart := args.Bool("no-restart", false, "--no-restart: no restart")
	args.Parse(os.Args[2:])
	config.CommandUnset(args.Args(), *global, *noRestart)
}
