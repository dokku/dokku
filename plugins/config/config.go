package config

import (
	"fmt"
	"os"
	"regexp"

	"github.com/dokku/dokku/plugins/common"
)

//Get retrieves a value from a config. If appName is empty the global config is used.
func Get(appName string, key string) (value string, ok bool) {
	env, err := loadAppOrGlobalEnv(appName)
	if err != nil {
		return "", false
	}
	if err = validateKey(key); err != nil {
		return "", false
	}
	return env.Get(key)
}

//GetWithDefault gets a value from a config. If appName is empty the global config is used. If the appName or key do not exist defaultValue is returned.
func GetWithDefault(appName string, key string, defaultValue string) (value string) {
	value, ok := Get(appName, key)
	if !ok {
		return defaultValue
	}
	return value
}

//SetMany variables in the environment. If appName is empty the global config is used. If restart is true the app is restarted.
func SetMany(appName string, entries map[string]string, restart bool) (err error) {
	global := appName == ""
	env, err := loadAppOrGlobalEnv(appName)
	if err != nil {
		return
	}
	keys := make([]string, 0, len(entries))
	for k := range entries {
		if err = validateKey(k); err != nil {
			return
		}
	}
	for k, v := range entries {
		env.Set(k, v)
		keys = append(keys, k)
	}
	if len(entries) != 0 {
		common.LogInfo1Quiet("Setting config vars")
		if os.Getenv("DOKKU_QUIET_OUTPUT") == "" {
			fmt.Println(prettyPrintEnvEntries("       ", entries))
		}
		env.Write()
		triggerUpdate(appName, "set", keys)
	}
	if !global && restart && env.GetBoolDefault("DOKKU_APP_RESTORE", true) {
		triggerRestart(appName)
	}
	return
}

//UnsetMany a value in a config. If appName is empty the global config is used. If restart is true the app is restarted.
func UnsetMany(appName string, keys []string, restart bool) (err error) {
	global := appName == ""
	env, err := loadAppOrGlobalEnv(appName)
	if err != nil {
		return
	}
	var changed = false
	for _, k := range keys {
		if err = validateKey(k); err != nil {
			return
		}
	}
	for _, k := range keys {
		if _, hasKey := env.Map()[k]; hasKey {
			common.LogInfo1Quiet(fmt.Sprintf("Unsetting %s", k))
			env.Unset(k)
			changed = true
		} else {
			common.LogInfo1Quiet(fmt.Sprintf("Skipping %s, it is not set in the environment", k))
		}
	}
	if changed {
		env.Write()
		triggerUpdate(appName, "unset", keys)
	}
	if !global && restart && env.GetBoolDefault("DOKKU_APP_RESTORE", true) {
		triggerRestart(appName)
	}
	return
}

func triggerRestart(appName string) {
	common.LogInfo1(fmt.Sprintf("Restarting app %s", appName))
	if err := common.PlugnTrigger("app-restart", appName); err != nil {
		common.LogWarn(fmt.Sprintf("Failure while restarting app: %s", err))
	}
}

func triggerUpdate(appName string, operation string, args []string) {
	args = append([]string{appName, operation}, args...)
	if err := common.PlugnTrigger("post-config-update", args...); err != nil {
		common.LogWarn(fmt.Sprintf("Failure while triggering post-config-update: %s", err))
	}
}

func loadAppOrGlobalEnv(appName string) (env *Env, err error) {
	if appName == "" || appName == "--global" {
		return LoadGlobalEnv()
	}
	return LoadAppEnv(appName)
}

func validateKey(key string) error {
	r, _ := regexp.Compile("^[a-zA-Z_][a-zA-Z0-9_]*$")
	if !r.MatchString(key) {
		return fmt.Errorf("Invalid key name: '%s'", key)
	}
	return nil
}
