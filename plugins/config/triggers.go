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
	return export(appName, g, m, format)
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
