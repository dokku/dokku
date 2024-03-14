package common

import (
	"path/filepath"
	"strings"
)

// PlugnTriggerExists returns whether a plugin trigger exists (ignoring the existence of any within the 20_events plugin)
func PlugnTriggerExists(triggerName string) bool {
	pluginPath := MustGetEnv("PLUGIN_PATH")
	pluginPathPrefix := filepath.Join(pluginPath, "enabled")
	glob := filepath.Join(pluginPathPrefix, "*", triggerName)
	exists := false
	files, _ := filepath.Glob(glob)
	for _, file := range files {
		plugin := strings.Trim(strings.TrimPrefix(strings.TrimSuffix(file, "/"+triggerName), pluginPathPrefix), "/")
		if plugin != "20_events" {
			exists = true
			break
		}
	}
	return exists
}
