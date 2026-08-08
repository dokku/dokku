package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/dokku/dokku/plugins/common"
)

// TriggerConfigExport returns a global config value by key
func TriggerConfigExport(appName string, global string, merged string, format string) error {
	g, err := strconv.ParseBool(global)
	if err != nil {
		return err
	}

	m, err := strconv.ParseBool(merged)
	if err != nil {
		return err
	}

	appName, err = getAppNameOrGlobal(appName, g)
	if err != nil {
		return err
	}

	return export(appName, m, format)
}

// TriggerConfigGet returns an app config value by key
func TriggerConfigGet(appName string, key string) error {
	value, ok := Get(appName, key)
	if ok {
		fmt.Print(value)
	}

	return nil
}

// TriggerConfigGetGlobal returns a global config value by key
func TriggerConfigGetGlobal(key string) error {
	value, ok := Get("--global", key)
	if ok {
		fmt.Print(value)
	}

	return nil
}

// TriggerConfigSet sets config values for an app
func TriggerConfigSet(appName string, noRestart bool, pairs ...string) error {
	return SubSet(appName, pairs, noRestart, false)
}

// TriggerConfigUnset unsets an app config value by key
func TriggerConfigUnset(appName string, key string, restart bool) error {
	UnsetMany(appName, []string{key}, restart)
	return nil
}

func setupAppConfigDir(appName string) error {
	appFile, err := getAppFile(appName)
	if err != nil {
		return err
	}

	appConfigDir := filepath.Dir(appFile)
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		return err
	}

	return common.SetPermissions(common.SetPermissionInput{
		Filename: appConfigDir,
		Mode:     os.FileMode(0755),
	})
}

// TriggerConfigMigrateEnv drains the pre-0.38 ENV files into the config
// property path. Exposed as a trigger so plugins that install before config can
// force the migration before reading a deprecated config var.
func TriggerConfigMigrateEnv() error {
	return MigrateEnvFiles()
}

// TriggerInstall runs the install step for the config plugin
func TriggerInstall() error {
	if err := common.PropertySetup("config"); err != nil {
		return fmt.Errorf("Unable to install the config plugin: %s", err.Error())
	}

	return MigrateEnvFiles()
}

// TriggerPostAppCloneSetup creates new buildpacks files
func TriggerPostAppCloneSetup(oldAppName string, newAppName string) error {
	oldEnv, err := LoadAppEnv(oldAppName)
	if err != nil {
		return fmt.Errorf("Unable to load old environment: %s", err.Error())
	}

	newEnv, err := LoadAppEnv(newAppName)
	if err != nil {
		return fmt.Errorf("Unable to load new environment: %s", err.Error())
	}

	newEnv.Merge(oldEnv)
	if err := newEnv.Write(); err != nil {
		return fmt.Errorf("Unable to write new environment: %s", err.Error())
	}

	return nil
}

// TriggerPostAppRenameSetup renames buildpacks files
func TriggerPostAppRenameSetup(oldAppName string, newAppName string) error {
	oldEnv, err := LoadAppEnv(oldAppName)
	if err != nil {
		return fmt.Errorf("Unable to load old environment: %s", err.Error())
	}

	newEnv, err := LoadAppEnv(newAppName)
	if err != nil {
		return fmt.Errorf("Unable to load new environment: %s", err.Error())
	}

	newEnv.Merge(oldEnv)
	if err := newEnv.Write(); err != nil {
		return fmt.Errorf("Unable to write new environment: %s", err.Error())
	}

	return nil
}

// TriggerPostCreate ensures apps have the correct config structure
func TriggerPostCreate(appName string) error {
	if err := common.PropertySetupApp("config", appName); err != nil {
		return err
	}

	return setupAppConfigDir(appName)
}

// TriggerPostDelete destroys the config data for a given app container
func TriggerPostDelete(appName string) error {
	appFile, err := getAppFile(appName)
	if err == nil {
		os.RemoveAll(filepath.Dir(appFile))
	}

	return common.PropertyDestroy("config", appName)
}
