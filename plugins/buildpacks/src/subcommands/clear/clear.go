package main

import (
	"flag"
	"os"

	"github.com/dokku/dokku/plugins/buildpacks"
	"github.com/dokku/dokku/plugins/common"
)

func main() {
	args := flag.NewFlagSet("buildpacks:clear", flag.ExitOnError)
	args.Parse(os.Args[2:])
	err := buildpacks.CommandClear(args.Args())
	if err != nil {
		common.LogFail(err.Error())
	}
}
