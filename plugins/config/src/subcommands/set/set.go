package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"strings"

	common "github.com/dokku/dokku/plugins/common"
	config "github.com/dokku/dokku/plugins/config"
)

// set the given entries from the specified environment
func main() {
	args := flag.NewFlagSet("config:set", flag.ExitOnError)
	global := args.Bool("global", false, "--global: use the global environment")
	encoded := args.Bool("encoded", false, "--encoded: interpret VALUEs as base64")
	noRestart := args.Bool("no-restart", false, "--no-restart: no restart")
	args.Parse(os.Args[2:])
	appName, pairs := config.GetCommonArgs(*global, args.Args())

	updated := make(map[string]string)
	for _, e := range pairs {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 1 {
			common.LogFail("Invalid env pair: " + e)
		}
		key, value := parts[0], parts[1]
		if *encoded {
			decoded, err := base64.StdEncoding.DecodeString(value)
			if err != nil {
				common.LogFail(fmt.Sprintf("%s for key '%s'", err.Error(), key))
			}
			value = string(decoded)
		}
		updated[key] = value
	}
	config.SetMany(appName, updated, !*noRestart)
}
