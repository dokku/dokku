package main

import (
	"flag"
	"fmt"
	"os"

	common "github.com/dokku/dokku/plugins/common"
	network "github.com/dokku/dokku/plugins/network"
)

// write the ipaddress to stdout for a given app container
func main() {
	flag.Parse()
	appName := flag.Arg(0)
	procType := flag.Arg(1)
	isHerokuishContainer := common.ToBool(flag.Arg(2))
	containerID := flag.Arg(3)

	ipAddress := network.GetContainerIpaddress(appName, procType, isHerokuishContainer, containerID)
	fmt.Fprintln(os.Stdout, ipAddress)
}
