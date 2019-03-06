package main

import (
	"flag"
	"os"

	"github.com/dokku/dokku/plugins/buildpacks"
	"github.com/dokku/dokku/plugins/common"
)

func main() {
	args := flag.NewFlagSet("buildpacks:remove", flag.ExitOnError)
	index := args.Int("index", 0, "--index: the 1-based index of the URL in the list of URLs")
	args.Parse(os.Args[2:])
	err := buildpacks.CommandRemove(args.Args(), *index)
	if err != nil {
		common.LogFail(err.Error())
	}
}
