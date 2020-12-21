package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/buildpacks"
	"github.com/dokku/dokku/plugins/common"

	flag "github.com/spf13/pflag"
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
		appName := args.Arg(0)
		buildpack := args.Arg(1)
		err = buildpacks.CommandAdd(appName, buildpack, *index)
	case "clear":
		args := flag.NewFlagSet("buildpacks:clear", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = buildpacks.CommandClear(appName)
	case "list":
		args := flag.NewFlagSet("buildpacks:list", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = buildpacks.CommandList(appName)
	case "remove":
		args := flag.NewFlagSet("buildpacks:remove", flag.ExitOnError)
		index := args.Int("index", 0, "--index: the 1-based index of the URL in the list of URLs")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		buildpack := args.Arg(1)
		err = buildpacks.CommandRemove(appName, buildpack, *index)
	case "report":
		args := flag.NewFlagSet("buildpacks:report", flag.ExitOnError)
		osArgs, infoFlag, err := common.ParseReportArgs("buildpacks", os.Args[2:])
		if err == nil {
			args.Parse(osArgs)
			appName := args.Arg(0)
			err = buildpacks.CommandReport(appName, infoFlag)
		}
	case "set":
		args := flag.NewFlagSet("buildpacks:set", flag.ExitOnError)
		index := args.Int("index", 0, "--index: the 1-based index of the URL in the list of URLs")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		buildpack := args.Arg(1)
		err = buildpacks.CommandSet(appName, buildpack, *index)
	default:
		common.LogFail(fmt.Sprintf("Invalid plugin subcommand call: %s", subcommand))
	}

	if err != nil {
		common.LogFail(err.Error())
	}
}
