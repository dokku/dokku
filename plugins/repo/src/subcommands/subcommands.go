package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/repo"
)

// main entrypoint to all subcommands
func main() {
	parts := strings.Split(os.Args[0], "/")
	subcommand := parts[len(parts)-1]
	flag.Parse()

	var err error
	switch subcommand {
	case "gc":
		flag.Parse()
		appName := flag.Arg(1)
		err = repo.CommandGc(appName)
	case "purge-cache":
		appName := flag.Arg(1)
		err = repo.CommandPurgeCache(appName)
	case "report":
		appName := flag.Arg(1)
		infoFlag := flag.Arg(2)
		repo.CommandReport(appName, infoFlag)
	case "set":
		appName := flag.Arg(1)
		property := flag.Arg(2)
		value := flag.Arg(3)
		err = repo.CommandSet(appName, property, value)
	default:
		common.LogFail(fmt.Sprintf("Invalid plugin subcommand call: %s", subcommand))
	}

	if err != nil {
		common.LogFail(err.Error())
	}
}
