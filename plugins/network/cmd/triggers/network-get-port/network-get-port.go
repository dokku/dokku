package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dokku/dokku/plugins/common"
	network "github.com/dokku/dokku/plugins/network"
)

// write the port to stdout for a given app container
func main() {
	flag.Parse()
	appName := flag.Arg(0)
	procType := flag.Arg(1)
	isHerokuishContainer := common.ToBool(flag.Arg(2))
	containerID := flag.Arg(3)

	port := network.GetContainerPort(appName, procType, isHerokuishContainer, containerID)
	fmt.Fprintln(os.Stdout, port)
}
