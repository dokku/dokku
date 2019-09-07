package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/buildpacks"
	"github.com/dokku/dokku/plugins/common"
)

// main entrypoint to all subcommands
func main() {
	parts := strings.Split(os.Args[0], "/")
	subcommand := parts[len(parts)-1]

	var err error
	switch subcommand {
	case "add":
		args := flag.NewFlagSet("buildpacks:add", flag.ExitOnError)
		index := args.Int("index", 0, "--index: the 1-based index of the URL in the list of URLs")
		args.Parse(os.Args[2:])
		err = buildpacks.CommandAdd(args.Args(), *index)
	case "clear":
		args := flag.NewFlagSet("buildpacks:clear", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = buildpacks.CommandClear(args.Args())
	case "list":
		args := flag.NewFlagSet("buildpacks:list", flag.ExitOnError)
		args.Parse(os.Args[2:])
		buildpacks.CommandList(args.Args())
	case "remove":
		args := flag.NewFlagSet("buildpacks:remove", flag.ExitOnError)
		index := args.Int("index", 0, "--index: the 1-based index of the URL in the list of URLs")
		args.Parse(os.Args[2:])
		err = buildpacks.CommandRemove(args.Args(), *index)
	case "report":
		flag.Parse()
		appName := flag.Arg(1)
		infoFlag := flag.Arg(2)
		buildpacks.CommandReport(appName, infoFlag)
	case "set":
		args := flag.NewFlagSet("buildpacks:set", flag.ExitOnError)
		index := args.Int("index", 0, "--index: the 1-based index of the URL in the list of URLs")
		args.Parse(os.Args[2:])
		err = buildpacks.CommandSet(args.Args(), *index)
	default:
		common.LogFail(fmt.Sprintf("Invalid plugin subcommand call: %s", subcommand))
	}

	if err != nil {
		common.LogFail(err.Error())
	}
}
