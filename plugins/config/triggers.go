package config

import "fmt"

func TriggerConfigGet(appName string, key string) error {
	value, ok := Get(appName, key)
	if ok {
		fmt.Println(value)
	}

	return nil
}

func TriggerConfigGetGlobal(key string) error {
	value, ok := Get("--global", key)
	if ok {
		fmt.Println(value)
	}

	return nil
}
