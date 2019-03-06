package main

import (
	"flag"
	"os"

	"github.com/dokku/dokku/plugins/buildpacks"
)

func main() {
	args := flag.NewFlagSet("buildpacks:list", flag.ExitOnError)
	args.Parse(os.Args[2:])
	buildpacks.CommandList(args.Args())
}
