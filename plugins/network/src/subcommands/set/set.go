package main

import (
	"flag"

	common "github.com/dokku/dokku/plugins/common"
	network "github.com/dokku/dokku/plugins/network"
)

// set or clear a network property for an app
func main() {
	flag.Parse()
	appName := flag.Arg(1)
	property := flag.Arg(2)
	value := flag.Arg(3)

	if property == "bind-all-interfaces" && value == "" {
		value = "false"
	}

	common.CommandPropertySet("network", appName, property, value, network.ValidProperties)
}
