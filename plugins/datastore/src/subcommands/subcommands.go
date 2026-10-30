package main

import (
	"os"
	"strings"

	"github.com/dokku/dokku-datastore/sdk"
)

// main entrypoint to all subcommands
func main() {
	parts := strings.Split(os.Args[0], "/")
	subcommand := parts[len(parts)-1]

	args := []string{}
	if len(os.Args) > 2 {
		args = os.Args[2:]
	}

	os.Exit(sdk.RunSubcommand(subcommand, args))
}
