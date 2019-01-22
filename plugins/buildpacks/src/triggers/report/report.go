package main

import (
	"flag"

	"github.com/dokku/dokku/plugins/buildpacks"
)

// displays a buildpacks report for one or more apps
func main() {
	flag.Parse()
	appName := flag.Arg(0)

	buildpacks.ReportSingleApp(appName, "")
}
