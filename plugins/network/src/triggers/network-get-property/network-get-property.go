package main

import (
	"flag"
	"fmt"
	"os"

	common "github.com/dokku/dokku/plugins/common"
	network "github.com/dokku/dokku/plugins/network"
)

// write the port to stdout for a given app container
func main() {
	flag.Parse()
	appName := flag.Arg(0)
	property := flag.Arg(1)

	defaultValue := network.GetDefaultValue(property)
	value := common.PropertyGetDefault("network", appName, property, defaultValue)
	fmt.Fprintln(os.Stdout, value)
}
