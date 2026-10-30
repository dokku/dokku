package main

import (
	"os"
	"strings"

	"github.com/dokku/dokku-datastore/sdk"
)

// main entrypoint to all triggers
func main() {
	parts := strings.Split(os.Args[0], "/")
	trigger := parts[len(parts)-1]

	os.Exit(sdk.RunTrigger(trigger, os.Args[1:]))
}
