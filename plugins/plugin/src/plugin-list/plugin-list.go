package main

import (
	"os"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/plugin"

	flag "github.com/spf13/pflag"
)

func main() {
	args := flag.NewFlagSet("plugin:list", flag.ExitOnError)
	format := args.String("format", "json", "format: [ json ]")
	args.Parse(os.Args[1:])

	if err := plugin.CommandList(*format); err != nil {
		common.LogFailWithError(err)
	}
}
