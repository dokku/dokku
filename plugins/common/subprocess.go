package common

import (
	"path/filepath"
	"strings"

	"github.com/codeskyblue/go-sh"
)

// PlugnTriggerSetup sets up a plugn trigger call
//
// Deprecated: use CallPlugnTrigger with Stdin instead
func PlugnTriggerSetup(triggerName string, args ...string) *sh.Session {
	shellArgs := make([]interface{}, len(args)+2)
	shellArgs[0] = "trigger"
	shellArgs[1] = triggerName
	for i, arg := range args {
		shellArgs[i+2] = arg
	}
	return sh.Command("plugn", shellArgs...)
}

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
