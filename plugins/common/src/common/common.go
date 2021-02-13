package main

import (
	"fmt"
	"os"

	"github.com/dokku/dokku/plugins/common"
	flag "github.com/spf13/pflag"
)

func main() {
	quiet := flag.Bool("quiet", false, "--quiet: set DOKKU_QUIET_OUTPUT=1")
	flag.Parse()
	cmd := flag.Arg(0)

	if *quiet {
		os.Setenv("DOKKU_QUIET_OUTPUT", "1")
	}

	var err error
	switch cmd {
	case "is-deployed":
		appName := flag.Arg(1)
		if !common.IsDeployed(appName) {
			err = fmt.Errorf("App %v not deployed", appName)
		}
	default:
		err = fmt.Errorf("Invalid common command call: %v", cmd)
	}

	if err != nil {
		common.LogFailQuiet(err.Error())
	}
}
