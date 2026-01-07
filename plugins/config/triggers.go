package config

import (
	"fmt"
	"strconv"
)

// TriggerConfigExport returns a global config value by key
func TriggerConfigExport(appName string, global string, merged string, format string) error {
	g, err := strconv.ParseBool(global)
	if err != nil {
		return err
	}

	m, err := strconv.ParseBool(merged)
	if err != nil {
		return err
	}

	appName, err = getAppNameOrGlobal(appName, g)
	if err != nil {
		return err
	}

	return export(appName, m, format)
}

// TriggerConfigGet returns an app config value by key
func TriggerConfigGet(appName string, key string) error {
	value, ok := Get(appName, key)
	if ok {
		fmt.Print(value)
	}

	return nil
}

// TriggerConfigGetGlobal returns a global config value by key
func TriggerConfigGetGlobal(key string) error {
	value, ok := Get("--global", key)
	if ok {
		fmt.Print(value)
	}

	return nil
}

// TriggerConfigSet sets config values for an app
func TriggerConfigSet(appName string, noRestart bool, pairs ...string) error {
	return SubSet(appName, pairs, noRestart, false)
}

// TriggerConfigUnset unsets an app config value by key
func TriggerConfigUnset(appName string, key string, restart bool) error {
	UnsetMany(appName, []string{key}, restart)
	return nil
}

// TriggerPostAppCloneSetup creates new buildpacks files
func TriggerPostAppCloneSetup(oldAppName string, newAppName string) error {
	oldEnv, err := LoadAppEnv(oldAppName)
	if err != nil {
		return fmt.Errorf("Unable to load old environment: %s", err.Error())
	}

	newEnv, err := LoadAppEnv(newAppName)
	if err != nil {
		return fmt.Errorf("Unable to load new environment: %s", err.Error())
	}

	newEnv.Merge(oldEnv)
	if err := newEnv.Write(); err != nil {
		return fmt.Errorf("Unable to write new environment: %s", err.Error())
	}

	return nil
}

// TriggerPostAppRenameSetup renames buildpacks files
func TriggerPostAppRenameSetup(oldAppName string, newAppName string) error {
	oldEnv, err := LoadAppEnv(oldAppName)
	if err != nil {
		return fmt.Errorf("Unable to load old environment: %s", err.Error())
	}

	newEnv, err := LoadAppEnv(newAppName)
	if err != nil {
		return fmt.Errorf("Unable to load new environment: %s", err.Error())
	}

	newEnv.Merge(oldEnv)
	if err := newEnv.Write(); err != nil {
		return fmt.Errorf("Unable to write new environment: %s", err.Error())
	}

	return nil
}
