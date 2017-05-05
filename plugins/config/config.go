package config

import configenv "github.com/dokku/dokku/plugins/config/src/configenv"
import columnize "github.com/ryanuber/columnize"
import "fmt"

//GetWithDefault gets value from app config
func GetWithDefault(appName string, key string, defaultValue string) string {
	env, err := configenv.NewFromTarget(appName)
	if err != nil {
		return defaultValue
	}
	return env.GetDefault(key, defaultValue)
}

//PrettyPrintLogEntries in columns
func PrettyPrintLogEntries(prefix string, entries map[string]string) string {
	colConfig := columnize.DefaultConfig()
	colConfig.Prefix = prefix
	colConfig.Delim = "\x00"
	lines := make([]string, 0, len(entries))
	for k, v := range entries {
		lines = append(lines, fmt.Sprintf("%s:\x00%s", k, v))
	}
	return columnize.Format(lines, colConfig)
}
