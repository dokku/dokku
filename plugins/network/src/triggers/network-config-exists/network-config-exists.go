package main

import (
	"flag"
	"fmt"
	"os"

	network "github.com/dokku/dokku/plugins/network"
)

// write the ipaddress to stdout for a given app container
func main() {
	flag.Parse()
	appName := flag.Arg(0)

	if network.HasNetworkConfig(appName) {
		fmt.Fprintln(os.Stdout, "true")
		return
	}

	fmt.Fprintln(os.Stdout, "false")
}
