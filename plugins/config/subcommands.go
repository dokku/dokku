package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

//CommandShow implements config:show
func CommandShow(args []string, global bool, shell bool, export bool, merged bool) {
	appName, _ := getCommonArgs(global, args)
	env := getEnvironment(appName, merged)
	if shell && export {
		common.LogFail("Only one of --shell and --export can be given")
	}
	if shell {
		fmt.Print(env.Export(ExportFormatShell))
	} else if export {
		fmt.Println(env.Export(ExportFormatExports))
	} else {
		contextName := "global"
		if appName != "" {
			contextName = appName
		}
		common.LogInfo2Quiet(contextName + " env vars")
		fmt.Println(env.Export(ExportFormatPretty))
	}
}

//CommandGet implements config:get
func CommandGet(args []string, global bool, quoted bool) {
	appName, keys := getCommonArgs(global, args)
	if len(keys) > 1 {
		common.LogFail(fmt.Sprintf("Unexpected argument(s): %v", keys[1:]))
	}
	if len(keys) == 0 {
		common.LogFail("Expected: key")
	}
	if value, ok := Get(appName, keys[0]); !ok {
		os.Exit(1)
	} else {
		if quoted {
			fmt.Printf("'%s'\n", singleQuoteEscape(value))
		} else {
			fmt.Printf("%s\n", value)
		}
	}
}

//CommandUnset implements config:unset
func CommandUnset(args []string, global bool, noRestart bool) {
	appName, keys := getCommonArgs(global, args)
	err := UnsetMany(appName, keys, !noRestart)
	if err != nil {
		common.LogFail(err.Error())
	}
}

//CommandSet implements config:set
func CommandSet(args []string, global bool, noRestart bool, encoded bool) {
	appName, pairs := getCommonArgs(global, args)
	updated := make(map[string]string)
	for _, e := range pairs {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 1 {
			common.LogFail("Invalid env pair: " + e)
		}
		key, value := parts[0], parts[1]
		if encoded {
			decoded, err := base64.StdEncoding.DecodeString(value)
			if err != nil {
				common.LogFail(fmt.Sprintf("%s for key '%s'", err.Error(), key))
			}
			value = string(decoded)
		}
		updated[key] = value
	}
	err := SetMany(appName, updated, !noRestart)
	if err != nil {
		common.LogFail(err.Error())
	}
}

//CommandKeys implements config:keys
func CommandKeys(args []string, global bool, merged bool) {
	appName, trailingArgs := getCommonArgs(global, args)
	if len(trailingArgs) > 0 {
		common.LogFail(fmt.Sprintf("Trailing argument(s): %v", trailingArgs))
	}
	env := getEnvironment(appName, merged)
	for _, k := range env.Keys() {
		fmt.Println(k)
	}
}

//CommandExport implements config:export
func CommandExport(args []string, global bool, merged bool, format string) {
	appName, trailingArgs := getCommonArgs(global, args)
	if len(trailingArgs) > 0 {
		common.LogFail(fmt.Sprintf("Trailing argument(s): %v", trailingArgs))
	}
	env := getEnvironment(appName, merged)
	exportType := ExportFormatExports
	suffix := "\n"
	switch format {
	case "exports":
		exportType = ExportFormatExports
	case "envfile":
		exportType = ExportFormatEnvfile
	case "docker-args":
		exportType = ExportFormatDockerArgs
	case "shell":
		exportType = ExportFormatShell
		suffix = " "
	case "pretty":
		exportType = ExportFormatPretty
	case "json":
		exportType = ExportFormatJSON
	case "json-list":
		exportType = ExportFormatJSONList
	default:
		common.LogFail(fmt.Sprintf("Unknown export format: %v", format))
	}
	exported := env.Export(exportType)
	fmt.Print(exported + suffix)
}

//CommandBundle implements config:bundle
func CommandBundle(args []string, global bool, merged bool) {
	appName, trailingArgs := getCommonArgs(global, args)
	if len(trailingArgs) > 0 {
		common.LogFail(fmt.Sprintf("Trailing argument(s): %v", trailingArgs))
	}
	env := getEnvironment(appName, merged)
	env.ExportBundle(os.Stdout)
}

//getEnvironment for the given app (global config if appName is empty). Merge with global environment if merged is true.
func getEnvironment(appName string, merged bool) (env *Env) {
	var err error
	if appName != "" && merged {
		env, err = LoadMergedAppEnv(appName)
	} else {
		env, err = loadAppOrGlobalEnv(appName)
	}
	if err != nil {
		common.LogFail(err.Error())
	}
	return env
}

//getCommonArgs extracts common positional args (appName and keys)
func getCommonArgs(global bool, args []string) (appName string, keys []string) {
	keys = args
	if !global {
		if len(args) > 0 {
			appName = args[0]
		}
		if appName == "" {
			common.LogFail("Please specify an app or --global")
		} else {
			keys = args[1:]
		}
	}
	return appName, keys
}
