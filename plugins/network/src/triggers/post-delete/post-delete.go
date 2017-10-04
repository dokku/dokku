package main

import (
	"flag"

	"github.com/dokku/dokku/plugins/common"
)

// write the port to stdout for a given app container
func main() {
	flag.Parse()
	appName := flag.Arg(0)

	common.PropertyDestroy("network", appName)
}
