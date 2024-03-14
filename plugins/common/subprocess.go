package common

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/codeskyblue/go-sh"
)

// PlugnTrigger fire the given plugn trigger with the given args
//
// Deprecated: use CallPlugnTrigger instead
func PlugnTrigger(triggerName string, args ...string) error {
	LogDebug(fmt.Sprintf("plugn trigger %s %v", triggerName, args))
	return PlugnTriggerSetup(triggerName, args...).Run()
}

// PlugnTriggerOutput fire the given plugn trigger with the given args
//
// Deprecated: use CallPlugnTrigger with CaptureOutput=true instead
func PlugnTriggerOutput(triggerName string, args ...string) ([]byte, error) {
	LogDebug(fmt.Sprintf("plugn trigger %s %v", triggerName, args))
	rE, wE, _ := os.Pipe()
	rO, wO, _ := os.Pipe()
	session := PlugnTriggerSetup(triggerName, args...)
	session.Stderr = wE
	session.Stdout = wO
	err := session.Run()
	wE.Close()
	wO.Close()

	readStderr, _ := io.ReadAll(rE)
	readStdout, _ := io.ReadAll(rO)

	stderr := string(readStderr[:])
	if err != nil {
		err = fmt.Errorf(stderr)
	}

	if os.Getenv("DOKKU_TRACE") == "1" {
		for _, line := range strings.Split(stderr, "\n") {
			LogDebug(fmt.Sprintf("plugn trigger %s stderr: %s", triggerName, line))
		}
		for _, line := range strings.Split(string(readStdout[:]), "\n") {
			LogDebug(fmt.Sprintf("plugn trigger %s stdout: %s", triggerName, line))
		}
	}

	return readStdout, err
}

// PlugnTriggerSetup sets up a plugn trigger call
//
// Deprecated: use CallPlugnTrigger instead
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
