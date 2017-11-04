package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dokku/dokku/plugins/network"
)

// writes true or false to stdout whether a given app has network config
func main() {
	flag.Parse()
	appName := flag.Arg(0)

	if network.HasNetworkConfig(appName) {
		fmt.Fprintln(os.Stdout, "true")
		return
	}

	fmt.Fprintln(os.Stdout, "false")
}
