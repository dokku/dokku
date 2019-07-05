package main

import (
	"flag"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/repo"
)

// destroys the buildpacks property for a given app container
func main() {
	flag.Parse()
	appName := flag.Arg(0)

	err := repo.PurgeCache(appName)
	if err != nil {
		common.LogFail(err.Error())
	}
}
