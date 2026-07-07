package main

import (
	"fmt"
	"os"
	"strings"

	buildernixpacks "github.com/dokku/dokku/plugins/builder-nixpacks"
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
		args := flag.NewFlagSet("builder-nixpacks:report", flag.ExitOnError)
		format := args.String("format", "stdout", "format: [ stdout | json ]")
		reportArgs, flagErr := common.ParseReportArgs("builder-nixpacks", os.Args[2:])
		if flagErr == nil {
			args.Parse(reportArgs.OSArgs)
			appName := args.Arg(0)
			if reportArgs.IsGlobal {
				appName = "--global"
			}
			err = buildernixpacks.CommandReport(appName, *format, reportArgs.InfoFlag)
		}
	default:
		err = fmt.Errorf("Invalid plugin subcommand call: %s", subcommand)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
