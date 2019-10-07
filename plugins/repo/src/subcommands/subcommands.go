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
	default:
		common.LogFail(fmt.Sprintf("Invalid plugin subcommand call: %s", subcommand))
	}

	if err != nil {
		common.LogFail(err.Error())
	}
}
