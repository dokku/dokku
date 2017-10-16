package config

import (
	"fmt"

	"github.com/dokku/dokku/plugins/common"
)

//Get retreives a value from a config. If appName is empty the global config is used.
func Get(appName string, key string) (value string, ok bool) {
	env, err := loadAppOrGlobalEnv(appName)
	if err != nil {
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
	for k, v := range entries {
		env.Set(k, v)
		keys = append(keys, k)
	}
	if len(entries) != 0 {
		common.LogInfo1("Setting config vars")
		fmt.Println(prettyPrintEnvEntries("       ", entries))
		env.Write()
		args := append([]string{appName, "set"}, keys...)
		common.PlugnTrigger("post-config-update", args...)
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
		common.LogInfo1(fmt.Sprintf("Unsetting %s", k))
		env.Unset(k)
		changed = true
	}
	if changed {
		env.Write()
		args := append([]string{appName, "unset"}, keys...)
		common.PlugnTrigger("post-config-update", args...)
	}
	if !global && restart && env.GetBoolDefault("DOKKU_APP_RESTORE", true) {
		triggerRestart(appName)
	}
	return
}

func triggerRestart(appName string) {
	common.LogInfo1(fmt.Sprintf("Restarting app %s", appName))
	common.PlugnTrigger("app-restart", appName)
}

func loadAppOrGlobalEnv(appName string) (env *Env, err error) {
	if appName == "" || appName == "--global" {
		return LoadGlobalEnv()
	}
	return LoadAppEnv(appName)
}
