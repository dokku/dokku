package main

import (
	"flag"
	"os"

	"github.com/dokku/dokku/plugins/config"
)

func main() {
	args := flag.NewFlagSet("config:keys", flag.ExitOnError)
	global := args.Bool("global", false, "--global: use the global environment")
	merged := args.Bool("merged", false, "--merged: merge app environment and global environment")
	args.Parse(os.Args[2:])
	config.CommandKeys(args.Args(), *global, *merged)
}
