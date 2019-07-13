package main

import (
	"flag"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/repo"
)

// deletes the contents of the build cache stored in the repository
func main() {
	flag.Parse()
	appName := flag.Arg(1)
	if appName == "" {
		common.LogFail("Please specify an app to run the command on")
	}

	err := repo.PurgeCache(appName)
	if err != nil {
		common.LogWarn(err.Error())
	}
}
