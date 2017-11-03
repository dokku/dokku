package main

import (
	"flag"
	"os"

	"github.com/dokku/dokku/plugins/config"
)

// set the given entries to the specified environment
func main() {
	args := flag.NewFlagSet("config:set", flag.ExitOnError)
	global := args.Bool("global", false, "--global: use the global environment")
	encoded := args.Bool("encoded", false, "--encoded: interpret VALUEs as base64")
	noRestart := args.Bool("no-restart", false, "--no-restart: no restart")
	args.Parse(os.Args[2:])
	config.CommandSet(args.Args(), *global, *noRestart, *encoded)
}
