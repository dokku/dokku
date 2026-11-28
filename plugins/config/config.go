package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// Get retrieves a value from a config. If appName is empty the global config is used.
func Get(appName string, key string) (value string, ok bool) {
	env, err := loadAppOrGlobalEnv(appName)
	if err != nil {
		return "", false
	}
	if err = ValidateKey(key); err != nil {
		return "", false
	}
	return env.Get(key)
}

// GetWithDefault gets a value from a config. If appName is empty the global config is used. If the appName or key do not exist defaultValue is returned.
func GetWithDefault(appName string, key string, defaultValue string) (value string) {
	value, ok := Get(appName, key)
	if !ok {
		return defaultValue
	}
	return value
}

// SetMany variables in the environment. If appName is empty the global config is used. If restart is true the app is restarted.
// When replace is true, every key not in entries is removed, as with ReplaceMany without preserved keys.
func SetMany(appName string, entries map[string]string, replace bool, restart bool) (err error) {
	if replace {
		return ReplaceMany(appName, entries, nil, restart)
	}

	global := appName == "" || appName == "--global"
	name, filename := appOrGlobalTarget(appName)
	changed, err := SetEnvFileEntries(name, filename, entries, false)
	if err != nil {
		return err
	}
	if changed {
		keys := make([]string, 0, len(entries))
		for k := range entries {
			keys = append(keys, k)
		}
		triggerUpdate(appName, "set", keys)
	}
	if !global && restart {
		RestartApp(appName)
	}
	return
}

// ReplaceMany replaces the whole environment with entries, keeping the current
// value of any key in preserve that entries does not set. If appName is empty
// the global config is used. Removed keys are reported to post-config-update
// before the set keys. Nothing is written, triggered or restarted when the
// environment would not change, otherwise the app is restarted if restart is true.
func ReplaceMany(appName string, entries map[string]string, preserve []string, restart bool) error {
	global := appName == "" || appName == "--global"
	name, filename := appOrGlobalTarget(appName)
	changed, removed, err := ReplaceEnvFileEntries(name, filename, entries, preserve)
	if err != nil {
		return err
	}

	if !changed {
		common.LogInfo1Quiet("No changes to config vars")
		return nil
	}

	if len(removed) > 0 {
		triggerUpdate(appName, "unset", removed)
	}
	if len(entries) > 0 {
		keys := make([]string, 0, len(entries))
		for k := range entries {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		triggerUpdate(appName, "set", keys)
	}

	if !global && restart {
		RestartApp(appName)
	}
	return nil
}

// UnsetMany a value in a config. If appName is empty the global config is used. If restart is true the app is restarted.
func UnsetMany(appName string, keys []string, restart bool) (err error) {
	global := appName == "" || appName == "--global"
	name, filename := appOrGlobalTarget(appName)
	changed, err := UnsetEnvFileKeys(name, filename, keys)
	if err != nil {
		return err
	}
	if changed {
		triggerUpdate(appName, "unset", keys)
	}
	if !global && restart {
		RestartApp(appName)
	}
	return
}

// UnsetAll removes all config keys
func UnsetAll(appName string, restart bool) (err error) {
	global := appName == "" || appName == "--global"
	name, filename := appOrGlobalTarget(appName)
	changed, err := ClearEnvFile(name, filename)
	if err != nil {
		return err
	}
	if changed {
		triggerUpdate(appName, "clear", []string{})
	}
	if !global && restart {
		RestartApp(appName)
	}
	return
}

// SetEnvFileEntries sets the given entries in the environment file at
// filename, clearing it first when replace is true, and returns whether the
// file was changed. Every key is validated before anything is written.
// When replace is true the file is left untouched if it already holds exactly
// the given entries.
func SetEnvFileEntries(name string, filename string, entries map[string]string, replace bool) (bool, error) {
	if replace {
		changed, _, err := ReplaceEnvFileEntries(name, filename, entries, nil)
		return changed, err
	}

	for k := range entries {
		if err := ValidateKey(k); err != nil {
			return false, err
		}
	}

	return UpdateEnvFile(name, filename, func(env *Env) bool {
		for k, v := range entries {
			env.Set(k, v)
		}
		if len(entries) == 0 {
			return false
		}

		common.LogInfo1Quiet("Setting config vars")
		if os.Getenv("DOKKU_QUIET_OUTPUT") == "" {
			fmt.Println(prettyPrintEnvEntries("       ", entries))
		}
		return true
	})
}

// ReplaceEnvFileEntries replaces the contents of the environment file at
// filename with entries, keeping the current value of any key in preserve that
// entries does not set. It returns whether the file was changed and the sorted
// keys that were removed. The file is not written when it already holds the
// resulting environment. Every key is validated before anything is written.
func ReplaceEnvFileEntries(name string, filename string, entries map[string]string, preserve []string) (bool, []string, error) {
	for k := range entries {
		if err := ValidateKey(k); err != nil {
			return false, nil, err
		}
	}

	removed := []string{}
	changed, err := UpdateEnvFile(name, filename, func(env *Env) bool {
		current := env.Map()
		desired := make(map[string]string, len(entries))
		for k, v := range entries {
			desired[k] = v
		}

		kept := []string{}
		for _, k := range preserve {
			if _, given := desired[k]; given {
				continue
			}
			if v, ok := current[k]; ok {
				desired[k] = v
				kept = append(kept, k)
			}
		}

		for k := range current {
			if _, ok := desired[k]; !ok {
				removed = append(removed, k)
			}
		}
		sort.Strings(removed)

		differs := len(removed) > 0
		for k, v := range desired {
			if cv, ok := current[k]; !ok || cv != v {
				differs = true
				break
			}
		}
		if !differs {
			return false
		}

		sort.Strings(kept)
		for _, k := range kept {
			common.LogInfo1Quiet(fmt.Sprintf("Keeping dokku-managed config var %s", k))
		}
		for _, k := range removed {
			common.LogInfo1Quiet(fmt.Sprintf("Unsetting %s", k))
		}

		env.Clear()
		for k, v := range desired {
			env.Set(k, v)
		}

		if len(entries) > 0 {
			common.LogInfo1Quiet("Setting config vars")
			if os.Getenv("DOKKU_QUIET_OUTPUT") == "" {
				fmt.Println(prettyPrintEnvEntries("       ", entries))
			}
		}
		return true
	})
	if err != nil || !changed {
		return false, nil, err
	}

	return true, removed, nil
}

// UnsetEnvFileKeys removes the given keys from the environment file at
// filename and returns whether the file was changed. Every key is validated
// before anything is written.
func UnsetEnvFileKeys(name string, filename string, keys []string) (bool, error) {
	for _, k := range keys {
		if err := ValidateKey(k); err != nil {
			return false, err
		}
	}

	return UpdateEnvFile(name, filename, func(env *Env) bool {
		changed := false
		for _, k := range keys {
			if _, hasKey := env.Map()[k]; hasKey {
				common.LogInfo1Quiet(fmt.Sprintf("Unsetting %s", k))
				env.Unset(k)
				changed = true
			} else {
				common.LogInfo1Quiet(fmt.Sprintf("Skipping %s, it is not set in the environment", k))
			}
		}
		return changed
	})
}

// ClearEnvFile removes every key from the environment file at filename and
// returns whether the file was changed
func ClearEnvFile(name string, filename string) (bool, error) {
	return UpdateEnvFile(name, filename, func(env *Env) bool {
		changed := false
		for k := range env.Map() {
			common.LogInfo1Quiet(fmt.Sprintf("Unsetting %s", k))
			env.Unset(k)
			changed = true
		}
		return changed
	})
}

// UpdateEnvFile loads the environment file at filename, passes it to update
// and writes it back when update reports a change, returning whether it did.
// The directory holding the environment file stays locked from the load to
// the write, so concurrent config changes cannot drop each other's keys.
// Triggers reacting to the change must run after this returns, as a trigger
// that changes the same environment would otherwise wait on the lock forever.
func UpdateEnvFile(name string, filename string, update func(env *Env) bool) (bool, error) {
	changed := false
	apply := func() error {
		env, err := loadFromFile(name, filename)
		if err != nil {
			return err
		}

		if !update(env) {
			return nil
		}

		if err := env.Write(); err != nil {
			return fmt.Errorf("Unable to write config vars: %s", err.Error())
		}
		changed = true
		return nil
	}

	dir := filepath.Dir(filename)
	if !common.DirectoryExists(dir) {
		return changed, apply()
	}

	err := common.LockDirectory(dir, apply)
	return changed, err
}

// RestartApp redeploys a deployed app so that it picks up a change to its
// environment, unless the app's ps restore property is set to false
func RestartApp(appName string) {
	if shouldRestart(appName) {
		triggerRestart(appName)
	}
}

// appOrGlobalTarget returns the name and environment file path for an app, or
// for the global environment when appName is empty or --global
func appOrGlobalTarget(appName string) (string, string) {
	if appName == "" || appName == "--global" {
		return "<global>", getGlobalFile()
	}

	filename, _ := getAppFile(appName)
	return appName, filename
}

// ManagedKeys returns the app config keys that other plugins manage, as
// reported by the config-managed-keys trigger. The global environment has
// no managed keys. Lines that are not valid key names are ignored.
func ManagedKeys(appName string) ([]string, error) {
	if appName == "" || appName == "--global" {
		return nil, nil
	}

	output, err := managedKeysLoader(appName)
	if err != nil {
		return nil, fmt.Errorf("Unable to fetch dokku-managed config keys: %s", err.Error())
	}

	seen := map[string]bool{}
	keys := []string{}
	for _, line := range strings.Split(output, "\n") {
		key := strings.TrimSpace(line)
		if key == "" || seen[key] || ValidateKey(key) != nil {
			continue
		}
		seen[key] = true
		keys = append(keys, key)
	}
	return keys, nil
}

// managedKeysLoader returns the raw output of the config-managed-keys trigger
// for an app. It is a variable so that tests can run without plugn.
var managedKeysLoader = loadManagedKeys

// loadManagedKeys calls the config-managed-keys trigger for an app
func loadManagedKeys(appName string) (string, error) {
	results, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "config-managed-keys",
		Args:    []string{appName},
	})
	return results.StdoutContents(), err
}

func shouldRestart(appName string) bool {
	results, _ := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "ps-get-property",
		Args:    []string{appName, "restore"},
	})
	return results.StdoutContents() != "false"
}

func triggerRestart(appName string) {
	if !common.IsDeployed(appName) {
		return
	}

	common.LogInfo1(fmt.Sprintf("Restarting app %s", appName))
	_, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "release-and-deploy",
		Args:        []string{appName, "", "config-redeploy"},
		StreamStdio: true,
	})
	if err != nil {
		common.LogWarn(fmt.Sprintf("Failure while restarting app: %s", err))
	}
}

func triggerUpdate(appName string, operation string, args []string) {
	args = append([]string{appName, operation}, args...)
	_, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "post-config-update",
		Args:        args,
		StreamStdio: true,
	})
	if err != nil {
		common.LogWarn(fmt.Sprintf("Failure while triggering post-config-update: %s", err))
	}
}

// getEnvironment for the given app (global config if appName is empty). Merge with group and global environments if merged is true.
func getEnvironment(appName string, merged bool) (env *Env) {
	var err error
	if appName != "" && appName != "--global" && merged {
		env, err = LoadMergedAppEnv(appName)
	} else {
		env, err = loadAppOrGlobalEnv(appName)
	}

	if err != nil {
		common.LogFailWithError(err)
	}

	return env
}

func getAppNameOrGlobal(appName string, global bool) (string, error) {
	if global {
		return appName, nil
	}

	if err := common.VerifyAppName(appName); err != nil {
		return appName, err
	}

	return appName, nil
}

func loadAppOrGlobalEnv(appName string) (env *Env, err error) {
	if appName == "" || appName == "--global" {
		return LoadGlobalEnv()
	}
	return LoadAppEnv(appName)
}

// ValidateKey returns an error when key is not a valid environment variable name
func ValidateKey(key string) error {
	r, _ := regexp.Compile("^[a-zA-Z_][a-zA-Z0-9_]*$")
	if !r.MatchString(key) {
		return fmt.Errorf("Invalid key name: '%s'", key)
	}
	return nil
}
