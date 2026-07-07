package main

import (
	"fmt"
	"os"
	"strings"

	caddyvhosts "github.com/dokku/dokku/plugins/caddy-vhosts"
	"github.com/dokku/dokku/plugins/common"

	flag "github.com/spf13/pflag"
)

// main entrypoint to all subcommands
func main() {
	parts := strings.Split(os.Args[0], "/")
	subcommand := parts[len(parts)-1]

	var err error
	switch subcommand {
	case "report":
		args := flag.NewFlagSet("caddy-vhosts:report", flag.ExitOnError)
		format := args.String("format", "stdout", "format: [ stdout | json ]")
		reportArgs, flagErr := common.ParseReportArgs("caddy-vhosts", os.Args[2:])
		if flagErr == nil {
			args.Parse(reportArgs.OSArgs)
			appName := args.Arg(0)
			if reportArgs.IsGlobal {
				appName = "--global"
			}
			err = caddyvhosts.CommandReport(appName, *format, reportArgs.InfoFlag)
		}
	default:
		err = fmt.Errorf("Invalid plugin subcommand call: %s", subcommand)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
