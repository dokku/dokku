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
	appName := flag.Arg(1)
	procType := flag.Arg(2)
	isHerokuishContainer := common.ToBool(flag.Arg(3))
	containerId := flag.Arg(4)

	port := network.GetContainerPort(appName, procType, isHerokuishContainer, containerId)
	fmt.Fprintln(os.Stdout, port)
}
