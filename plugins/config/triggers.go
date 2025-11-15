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

// TriggerConfigUnset unsets an app config value by key
func TriggerConfigUnset(appName string, key string, restart bool) error {
	UnsetMany(appName, []string{key}, restart)
	return nil
}

func migrateGlobalEnv() error {
	if err := common.PropertySetup("--global"); err != nil {
		return fmt.Errorf("Unable to setup global environment: %s", err.Error())
	}

	oldGlobalEnvFile := filepath.Join(common.MustGetEnv("DOKKU_ROOT"), "ENV")
	isGlobalMigrated := common.PropertyGetDefault("config", "--global", "env-migrated", "")
	if isGlobalMigrated == "true" {
		return nil
	}

	oldGlobalEnv, err := loadFromFile("--global", oldGlobalEnvFile)
	if err != nil {
		return fmt.Errorf("Unable to load old global environment: %s", err.Error())
	}

	globalEnv, err := LoadGlobalEnv()
	if err != nil {
		return fmt.Errorf("Unable to load global environment: %s", err.Error())
	}

	globalEnv.Merge(oldGlobalEnv)
	if err := globalEnv.Write(); err != nil {
		return fmt.Errorf("Unable to write global environment: %s", err.Error())
	}

	if err := common.PropertyWrite("config", "--global", "env-migrated", "true"); err != nil {
		return fmt.Errorf("Unable to set env-migrated property: %s", err.Error())
	}

	return nil
}

// TriggerInstall runs the install step for the config plugin
func TriggerInstall() error {
	if err := common.PropertySetup("config"); err != nil {
		return fmt.Errorf("Unable to install the config plugin: %s", err.Error())
	}

	apps, err := common.UnfilteredDokkuApps()
	if err != nil {
		return nil
	}

	if err := migrateGlobalEnv(); err != nil {
		return fmt.Errorf("Unable to migrate global environment: %s", err.Error())
	}

	// migrate all app ENV files to config path
	for _, appName := range apps {
		if err := common.PropertySetupApp("config", appName); err != nil {
			return fmt.Errorf("Unable to setup app environment: %s", err.Error())
		}

		oldEnvFile := filepath.Join(common.AppRoot(appName) + "ENV")
		isMigrated := common.PropertyGetDefault("config", appName, "env-migrated", "")
		// delete the old file on the next install
		if isMigrated == "true" {
			if err := os.RemoveAll(oldEnvFile); err != nil {
				return fmt.Errorf("Unable to remove old ENV file: %s", err.Error())
			}
			continue
		}

		// skip if the file doesn't exist
		if _, err := os.Stat(oldEnvFile); err != nil {
			if err := common.PropertyWrite("config", appName, "env-migrated", "true"); err != nil {
				return fmt.Errorf("Unable to set env-migrated property: %s", err.Error())
			}
			continue
		}

		// merge in the old env into the new env
		oldEnv, err := loadFromFile(appName, oldEnvFile)
		if err != nil {
			return fmt.Errorf("Unable to load old environment: %s", err.Error())
		}

		env, err := LoadAppEnv(appName)
		if err != nil {
			return fmt.Errorf("Unable to load app environment: %s", err.Error())
		}

		env.Merge(oldEnv)
		if err := env.Write(); err != nil {
			return fmt.Errorf("Unable to write app environment: %s", err.Error())
		}

		if err := common.PropertyWrite("config", appName, "env-migrated", "true"); err != nil {
			return fmt.Errorf("Unable to set env-migrated property: %s", err.Error())
		}
	}

	return nil
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
	return common.PropertySetupApp("config", appName)
}

// TriggerPostDelete destroys the config data for a given app container
func TriggerPostDelete(appName string) error {
	return common.PropertyDestroy("config", appName)
}
