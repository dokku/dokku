package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

func export(appName string, merged bool, format string) error {
	env := getEnvironment(appName, merged)
	exportType, err := ParseExportFormat(format)
	if err != nil {
		return err
	}

	suffix := "\n"
	if exportType == ExportFormatShell {
		suffix = " "
	}

	exported := env.Export(exportType)
	fmt.Print(exported + suffix)
	return nil
}

// SubBundle implements the logic for config:bundle without app name validation
func SubBundle(appName string, merged bool) error {
	env := getEnvironment(appName, merged)
	return env.ExportBundle(os.Stdout)
}

// SubClear implements the logic for config:clear without app name validation
func SubClear(appName string, noRestart bool) error {
	return UnsetAll(appName, !noRestart)
}

// SubExport implements the logic for config:export without app name validation
func SubExport(appName string, merged bool, format string) error {
	return export(appName, merged, format)
}

// SubGet implements the logic for config:get without app name validation
func SubGet(appName string, keys []string, quoted bool) error {
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
		fmt.Printf("'%s'\n", SingleQuoteEscape(value))
	} else {
		fmt.Printf("%s\n", value)
	}

	return nil
}

// SubImport imports environment variables from a file
func SubImport(appName string, replace bool, noRestart bool, format string, filename string) error {
	if filename == "-" {
		common.LogInfo1Quiet("Reading config vars from stdin")
	}

	if format == "" {
		format = "envfile"
	}

	env, err := LoadImportEnv(appName, filename, format)
	if err != nil {
		return err
	}

	if len(env.Map()) == 0 {
		return errors.New("No config vars to set")
	}

	for k, v := range env.Map() {
		common.LogInfo1Quiet(fmt.Sprintf("Setting config var %s=%s", k, v))
	}

	if replace {
		return replaceAppEnv(appName, env.Map(), noRestart)
	}

	return SetMany(appName, env.Map(), false, !noRestart)
}

// replaceAppEnv replaces an app or global environment with entries, keeping
// the dokku-managed keys of an app that entries does not set
func replaceAppEnv(appName string, entries map[string]string, noRestart bool) error {
	preserve, err := ManagedKeys(appName)
	if err != nil {
		return err
	}

	return ReplaceMany(appName, entries, preserve, !noRestart)
}

// SubKeys implements the logic for config:keys without app name validation
func SubKeys(appName string, merged bool) error {
	env := getEnvironment(appName, merged)
	for _, k := range env.Keys() {
		fmt.Println(k)
	}
	return nil
}

// SubSet implements the logic for config:set without app name validation.
// When replace is true every key not given is unset.
func SubSet(appName string, pairs []string, noRestart bool, encoded bool, replace bool) error {
	if len(pairs) == 0 {
		return errors.New("At least one env pair must be given")
	}

	updated, err := ParseEnvPairs(pairs, encoded)
	if err != nil {
		return err
	}

	if replace {
		return replaceAppEnv(appName, updated, noRestart)
	}

	return SetMany(appName, updated, false, !noRestart)
}

// ParseEnvPairs parses KEY=VALUE pairs into a map, base64-decoding each value
// when encoded is true
func ParseEnvPairs(pairs []string, encoded bool) (map[string]string, error) {
	updated := make(map[string]string)
	for _, e := range pairs {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 1 {
			return nil, fmt.Errorf("Invalid env pair: %v", e)
		}

		key, value := parts[0], parts[1]
		if encoded {
			decoded, err := base64.StdEncoding.DecodeString(value)
			if err != nil {
				return nil, fmt.Errorf("%s for key '%s'", err.Error(), key)
			}
			value = string(decoded)
		}
		updated[key] = value
	}

	return updated, nil
}

// SubShow implements the logic for config:show without app name validation
func SubShow(appName string, merged bool, shell bool, export bool) error {
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

// SubUnset implements the logic for config:unset without app name validation
func SubUnset(appName string, keys []string, noRestart bool) error {
	if len(keys) == 0 {
		return fmt.Errorf("At least one key must be given")
	}

	return UnsetMany(appName, keys, !noRestart)
}
