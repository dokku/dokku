package main

import (
	"flag"

	"github.com/dokku/dokku/plugins/common"
)

// destroys the network property for a given app container
func main() {
	flag.Parse()
	appName := flag.Arg(0)

	err := common.PropertyDestroy("network", appName)
	if err != nil {
		common.LogFail(err.Error())
	}
}
