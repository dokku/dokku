package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// CommandBundle implements config:bundle
func CommandBundle(appName string, global bool, merged bool) error {
	appName, err := getAppNameOrGlobal(appName, global)
	if err != nil {
		return err
	}

	env := getEnvironment(appName, merged)
	return env.ExportBundle(os.Stdout)
}

// CommandClear implements config:clear
func CommandClear(appName string, global bool, noRestart bool) error {
	appName, err := getAppNameOrGlobal(appName, global)
	if err != nil {
		return err
	}

	return UnsetAll(appName, !noRestart)
}

// CommandExport implements config:export
func CommandExport(appName string, global bool, merged bool, format string) error {
	appName, err := getAppNameOrGlobal(appName, global)
	if err != nil {
		return err
	}

	env := getEnvironment(appName, merged)
	exportType := ExportFormatExports
	suffix := "\n"

	exportTypes := map[string]ExportFormat{
		"exports":          ExportFormatExports,
		"envfile":          ExportFormatEnvfile,
		"docker-args":      ExportFormatDockerArgs,
		"docker-args-keys": ExportFormatDockerArgsKeys,
		"shell":            ExportFormatShell,
		"pretty":           ExportFormatPretty,
		"json":             ExportFormatJSON,
		"json-list":        ExportFormatJSONList,
	}

	exportType, ok := exportTypes[format]
	if !ok {
		return fmt.Errorf("Unknown export format: %v", format)
	}

	if exportType == ExportFormatShell {
		suffix = " "
	}

	exported := env.Export(exportType)
	fmt.Print(exported + suffix)
	return nil
}

// CommandGet implements config:get
func CommandGet(appName string, keys []string, global bool, quoted bool) error {
	appName, err := getAppNameOrGlobal(appName, global)
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return errors.New("Expected: key")
	}

	if len(keys) != 1 {
		return fmt.Errorf("Unexpected argument(s): %v", keys[1:])
	}

	value, ok := Get(appName, keys[0])
	if !ok {
		os.Exit(1)
		return nil
	}

	if quoted {
		fmt.Printf("'%s'\n", singleQuoteEscape(value))
	} else {
		fmt.Printf("%s\n", value)
	}

	return nil
}

// CommandKeys implements config:keys
func CommandKeys(appName string, global bool, merged bool) error {
	appName, err := getAppNameOrGlobal(appName, global)
	if err != nil {
		return err
	}

	env := getEnvironment(appName, merged)
	for _, k := range env.Keys() {
		fmt.Println(k)
	}
	return nil
}

// CommandSet implements config:set
func CommandSet(appName string, pairs []string, global bool, noRestart bool, encoded bool) error {
	appName, err := getAppNameOrGlobal(appName, global)
	if err != nil {
		return err
	}

	if len(pairs) == 0 {
		return errors.New("At least one env pair must be given")
	}

	updated := make(map[string]string)
	for _, e := range pairs {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 1 {
			return fmt.Errorf("Invalid env pair: %v", e)
		}

		key, value := parts[0], parts[1]
		if encoded {
			decoded, err := base64.StdEncoding.DecodeString(value)
			if err != nil {
				return fmt.Errorf("%s for key '%s'", err.Error(), key)
			}
			value = string(decoded)
		}
		updated[key] = value
	}

	return SetMany(appName, updated, !noRestart)
}

// CommandShow implements config:show
func CommandShow(appName string, global bool, merged bool, shell bool, export bool) error {
	appName, err := getAppNameOrGlobal(appName, global)
	if err != nil {
		return err
	}

	env := getEnvironment(appName, merged)
	if shell && export {
		return errors.New("Only one of --shell and --export can be given")
	}
	if shell {
		common.LogWarn("Deprecated: Use 'config:export --format shell' instead")
		fmt.Print(env.Export(ExportFormatShell))
	} else if export {
		common.LogWarn("Deprecated: Use 'config:export --format exports' instead")
		fmt.Println(env.Export(ExportFormatExports))
	} else {
		contextName := "global"
		if appName != "" {
			contextName = appName
		}
		common.LogInfo2Quiet(contextName + " env vars")
		fmt.Println(env.Export(ExportFormatPretty))
	}

	return nil
}

// CommandUnset implements config:unset
func CommandUnset(appName string, keys []string, global bool, noRestart bool) error {
	appName, err := getAppNameOrGlobal(appName, global)
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return fmt.Errorf("At least one key must be given")
	}

	return UnsetMany(appName, keys, !noRestart)
}
