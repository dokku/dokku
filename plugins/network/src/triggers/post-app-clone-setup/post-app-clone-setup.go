package main

import (
	"flag"
	"os"

	"github.com/dokku/dokku/plugins/network"
)

// cleanup network files for a new app clone
func main() {
	flag.Parse()
	appName := flag.Arg(1)

	success := network.PostAppCloneSetup(appName)
	if !success {
		os.Exit(1)
	}
}
