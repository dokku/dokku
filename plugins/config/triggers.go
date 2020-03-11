package config

import "fmt"

// TriggerConfigGet returns an app config value by key
func TriggerConfigGet(appName string, key string) error {
	value, ok := Get(appName, key)
	if ok {
		fmt.Println(value)
	}

	return nil
}

// TriggerConfigGet returns a global config value by key
func TriggerConfigGetGlobal(key string) error {
	value, ok := Get("--global", key)
	if ok {
		fmt.Println(value)
	}

	return nil
}
