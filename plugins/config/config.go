package config

import (
	"fmt"

	"github.com/dokku/dokku/plugins/common"
	configenv "github.com/dokku/dokku/plugins/config/src/configenv"
	columnize "github.com/ryanuber/columnize"
)

//GetWithDefault gets a value from a config. If appName is empty the global config is used.
func GetWithDefault(appName string, key string, defaultValue string) string {
	env, err := loadConfig(appName)
	if err != nil {
		return defaultValue
	}
	return env.GetDefault(key, defaultValue)
}

// GetWithDefault returns the value set for a given key, returning defaultValue if none found
// func GetWithDefault(appName string, key string, defaultValue string) (value string) {
// 	value = defaultValue

// 	envFile := strings.Join([]string{common.MustGetEnv("DOKKU_ROOT"), appName, "ENV"}, "/")
// 	lines, err := common.FileToSlice(envFile)
// 	if err != nil {
// 		return
// 	}
// 	prefix := fmt.Sprintf("export %v=", key)
// 	for _, line := range lines {
// 		if !strings.HasPrefix(line, prefix) {
// 			continue
// 		}
// 		value = strings.TrimPrefix(line, prefix)
// 		if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
// 			value = strings.TrimPrefix(strings.TrimSuffix(value, "'"), "'")
// 		}
// 	}
// 	return
// }

//HasKey determines if the config given by appName has a value for the given key
func HasKey(appName string, key string) bool {
	env, err := loadConfig(appName)
	if err != nil {
		return false
	}
	_, ok := env.Get(key)
	return ok
}

//SetMany variables in the environment. If appName is empty the global config is used. If restart is true the app is restarted.
func SetMany(appName string, entries map[string]string, restart bool) {
	global := appName == ""
	env := GetConfig(appName, false)

	keys := make([]string, 0, len(entries))

	for k, v := range entries {
		env.Set(k, v)
		keys = append(keys, k)
	}

	if len(entries) != 0 {
		common.LogInfo1("Setting config vars")
		fmt.Println(PrettyPrintEnvEntries("       ", entries))
		env.Write()
		args := append([]string{appName, "set"}, keys...)
		common.PlugnTrigger("post-config-update", args...)
	}

	if !global && restart && env.GetBoolDefault("DOKKU_APP_RESTORE", true) {
		Restart(appName)
	}
}

//UnsetMany a value in a config. If appName is empty the global config is used. If restart is true the app is restarted.
func UnsetMany(appName string, keys []string, restart bool) {
	global := appName == ""
	env := GetConfig(appName, false)
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
		Restart(appName)
	}
}

//PrettyPrintEnvEntries in columns
func PrettyPrintEnvEntries(prefix string, entries map[string]string) string {
	colConfig := columnize.DefaultConfig()
	colConfig.Prefix = prefix
	colConfig.Delim = "\x00"
	lines := make([]string, 0, len(entries))
	for k, v := range entries {
		lines = append(lines, fmt.Sprintf("%s:\x00%s", k, v))
	}
	return columnize.Format(lines, colConfig)
}

//GetCommonArgs extracts common positional args (appName and keys)
func GetCommonArgs(global bool, args []string) (string, []string) {
	nextArg := 0
	appName := ""
	if !global {
		if len(args) > 0 {
			appName = args[0]
		}
		if appName == "" {
			common.LogFail("Please specify an app or --global")
		} else {
			nextArg++
		}
	}
	keys := args[nextArg:]
	return appName, keys
}

//GetConfig for the given app (global config if appName is empty). Merge with global config if merged is true.
func GetConfig(appName string, merged bool) *configenv.Env {
	env, err := loadConfig(appName)
	if err != nil {
		common.LogFail(err.Error())
	}
	if appName != "" && merged {
		global, err := configenv.LoadGlobal()
		if err != nil {
			common.LogFail(err.Error())
		}
		global.Merge(env)
		return global
	}
	return env
}

//Restart trigger restart on app
func Restart(appName string) {
	common.LogInfo1(fmt.Sprintf("Restarting app %s", appName))
	common.PlugnTrigger("app-restart", appName)
}

func loadConfig(appName string) (*configenv.Env, error) {
	if appName == "" || appName == "--global" {
		return configenv.LoadGlobal()
	}
	return configenv.LoadApp(appName)
}
