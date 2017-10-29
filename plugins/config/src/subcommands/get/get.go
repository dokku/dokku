package main

import (
	"flag"
	"os"

	"github.com/dokku/dokku/plugins/config"
)

// get the given entries from the specified environment
func main() {
	args := flag.NewFlagSet("config:get", flag.ExitOnError)
	global := args.Bool("global", false, "--global: use the global environment")
	quoted := args.Bool("quoted", false, "--quoted: get the value quoted")
	args.Parse(os.Args[2:])
	config.CommandGet(args.Args(), *global, *quoted)
}
