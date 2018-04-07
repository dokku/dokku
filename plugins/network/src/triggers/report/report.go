package main

import (
	"flag"

	"github.com/dokku/dokku/plugins/network"
)

// displays a network report for one or more apps
func main() {
	flag.Parse()
	appName := flag.Arg(0)

	network.ReportSingleApp(appName, "")
}
