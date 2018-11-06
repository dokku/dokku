package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/network"
)

// returns the listeners (host:port combinations) for a given app container
func main() {
	flag.Parse()
	appName := flag.Arg(0)

	listeners := network.GetListeners(appName)
	fmt.Fprint(os.Stdout, strings.Join(listeners, " "))
}
