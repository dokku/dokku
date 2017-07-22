package config

import (
	"fmt"
	"strings"

	common "github.com/dokku/dokku/plugins/common"
)

// GetWithDefault returns the value set for a given key, returning defaultValue if none found
func GetWithDefault(appName string, key string, defaultValue string) string {
	envFile := strings.Join([]string{common.MustGetEnv("DOKKU_ROOT"), appName, "ENV"}, "/")
	lines, err := common.FileToSlice(envFile)
	if err != nil {
		return defaultValue
	}

	value := defaultValue
	prefix := fmt.Sprintf("export %v=", key)
	for _, line := range lines {
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		value = strings.TrimPrefix(line, prefix)
		if strings.HasPrefix(line, "'") && strings.HasSuffix(line, "'") {
			value = strings.TrimPrefix(strings.TrimSuffix(value, "'"), "'")
		}
	}

	return value
}
