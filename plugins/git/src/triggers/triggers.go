package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	git "github.com/dokku/dokku/plugins/git"
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
		err = git.ReportSingleApp(appName, "", "")
	default:
		err = fmt.Errorf("Invalid plugin trigger call: %s", trigger)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
