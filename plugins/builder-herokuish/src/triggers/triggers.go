package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	builderherokuish "github.com/dokku/dokku/plugins/builder-herokuish"
	"github.com/dokku/dokku/plugins/common"
)

// main entrypoint to all triggers
func main() {
	parts := strings.Split(os.Args[0], "/")
	trigger := parts[len(parts)-1]
	flag.Parse()

	var err error
	switch trigger {
	case "report":
		appName := flag.Arg(0)
		err = builderherokuish.ReportSingleApp(appName, "", "")
	default:
		err = fmt.Errorf("Invalid plugin trigger call: %s", trigger)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
