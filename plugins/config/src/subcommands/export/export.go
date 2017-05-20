package main

import (
	"flag"
	"fmt"
	"os"

	common "github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
	"github.com/dokku/dokku/plugins/config/src/configenv"
)

// print the environment to stdout
func main() {
	const defaultPrefix = "export "
	const defaultSeparator = "\n"
	args := flag.NewFlagSet("config:export", flag.ExitOnError)
	global := args.Bool("global", false, "--global: use the global environment")
	merged := args.Bool("merged", false, "--merged: merge app environment and global environment")
	format := args.String("format", "exports", "--format: [ exports | envfile | docker-args ] which format to export as)")
	args.Parse(os.Args[2:])

	appName, trailingArgs := config.GetCommonArgs(*global, args.Args())
	if len(trailingArgs) > 0 {
		common.LogFail(fmt.Sprintf("Trailing argument(s): %v", trailingArgs))
	}

	env := config.GetConfig(appName, *merged)
	exportType := configenv.Exports
	switch *format {
	case "exports":
		exportType = configenv.Exports
	case "envfile":
		exportType = configenv.Envfile
	case "docker-args":
		exportType = configenv.DockerArgs
	default:
		common.LogFail(fmt.Sprintf("Unknown export format: %v", *format))
	}
	exported := env.Export(exportType)
	fmt.Println(exported)
}
