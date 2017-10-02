package config

import (
	"fmt"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// GetWithDefault returns the value set for a given key, returning defaultValue if none found
func GetWithDefault(appName string, key string, defaultValue string) (value string) {
	value = defaultValue

	envFile := strings.Join([]string{common.MustGetEnv("DOKKU_ROOT"), appName, "ENV"}, "/")
	lines, err := common.FileToSlice(envFile)
	if err != nil {
		return
	}
	prefix := fmt.Sprintf("export %v=", key)
	for _, line := range lines {
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		value = strings.TrimPrefix(line, prefix)
		if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
			value = strings.TrimPrefix(strings.TrimSuffix(value, "'"), "'")
		}
	}
	return
}
