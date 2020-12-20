package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/logs"

	flag "github.com/spf13/pflag"
)

// main entrypoint to all subcommands
func main() {
	parts := strings.Split(os.Args[0], "/")
	subcommand := parts[len(parts)-1]

	var err error
	switch subcommand {
	case "failed":
		args := flag.NewFlagSet("logs:failed", flag.ExitOnError)
		allApps := args.Bool("all", false, "--all: restore all apps")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = logs.CommandFailed(appName, *allApps)
	default:
		common.LogFail(fmt.Sprintf("Invalid plugin subcommand call: %s", subcommand))
	}

	if err != nil {
		common.LogFail(err.Error())
	}
}
