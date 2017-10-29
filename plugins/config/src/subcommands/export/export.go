package main

import (
	"flag"
	"os"

	"github.com/dokku/dokku/plugins/config"
)

// print the environment to stdout
func main() {
	const defaultPrefix = "export "
	const defaultSeparator = "\n"
	args := flag.NewFlagSet("config:export", flag.ExitOnError)
	global := args.Bool("global", false, "--global: use the global environment")
	merged := args.Bool("merged", false, "--merged: merge app environment and global environment")
	format := args.String("format", "exports", "--format: [ exports | envfile | docker-args | shell ] which format to export as)")
	args.Parse(os.Args[2:])
	config.CommandExport(args.Args(), *global, *merged, *format)
}
