package main

import (
	"os"

	"github.com/dokku/dokku-datastore/sdk"
)

// main answers the help dokku asks every plugin for, and runs any datastore
// command dokku found no subcommand for, which is what a datastore's own
// commands are
func main() {
	os.Exit(sdk.RunCommands(os.Args[1:]))
}
