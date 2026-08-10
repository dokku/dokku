package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// envMigratedProperty records that an ENV file has been drained out of the
// pre-0.38 location into the config property path.
const envMigratedProperty = "env-migrated"

// MigrateEnvFiles drains the pre-0.38 ENV files - $DOKKU_ROOT/ENV and
// $DOKKU_ROOT/<app>/ENV - into the config property path, removing each file as
// soon as it has been drained.
//
// It is safe to call from any plugin's install trigger, and does so via the
// config-migrate-env plugin trigger. Plugins whose enabled-directory name sorts
// before "config" - apps, builder and checks among them - run their deprecated
// DOKKU_* config var to property migrations before the config plugin's own
// install trigger would have relocated these files, and would otherwise read an
// empty environment and silently migrate nothing. Nothing here may assume the
// config install trigger has already run.
func MigrateEnvFiles() error {
	if err := common.PropertySetup("config"); err != nil {
		return fmt.Errorf("Unable to setup config properties: %s", err.Error())
	}

	if err := migrateGlobalEnv(); err != nil {
		return fmt.Errorf("Unable to migrate global environment: %s", err.Error())
	}

	apps, err := common.UnfilteredDokkuApps()
	if err != nil {
		return nil
	}

	for _, appName := range apps {
		if err := migrateAppEnv(appName); err != nil {
			return fmt.Errorf("Unable to migrate environment for %s: %s", appName, err.Error())
		}
	}

	return nil
}

// migrateGlobalEnv drains $DOKKU_ROOT/ENV into the global config property path.
func migrateGlobalEnv() error {
	if err := common.PropertySetup("--global"); err != nil {
		return fmt.Errorf("Unable to setup global environment: %s", err.Error())
	}

	oldGlobalEnvFile := filepath.Join(common.MustGetEnv("DOKKU_ROOT"), "ENV")
	globalEnv, err := LoadGlobalEnv()
	if err != nil {
		return fmt.Errorf("Unable to load global environment: %s", err.Error())
	}

	return drainLegacyEnvFile("--global", oldGlobalEnvFile, globalEnv)
}

// migrateAppEnv drains $DOKKU_ROOT/<app>/ENV into the app's config property path.
func migrateAppEnv(appName string) error {
	if err := common.PropertySetupApp("config", appName); err != nil {
		return fmt.Errorf("Unable to setup app environment: %s", err.Error())
	}

	if err := setupAppConfigDir(appName); err != nil {
		return fmt.Errorf("Unable to setup app config directory: %s", err.Error())
	}

	env, err := LoadAppEnv(appName)
	if err != nil {
		return fmt.Errorf("Unable to load app environment: %s", err.Error())
	}

	oldEnvFile := filepath.Join(common.AppRoot(appName), "ENV")
	return drainLegacyEnvFile(appName, oldEnvFile, env)
}

// drainLegacyEnvFile merges the legacy ENV file at oldEnvFile into env, records
// the migration against name, and removes the legacy file. The file is only
// removed once the merged environment has been written successfully, so a parse
// or write failure leaves the original in place.
//
// The merge happens once and only once. A legacy file that is still there when
// the migration is already on record is handed to preserveStaleEnvFile instead:
// releases 0.38.0 through 0.38.25 recorded the migration and left the file
// behind on purpose, so importing it would replay the environment as it stood
// at that upgrade over every config:set and config:unset made since.
func drainLegacyEnvFile(name string, oldEnvFile string, env *Env) error {
	migrated := common.PropertyGetDefault("config", name, envMigratedProperty, "") == "true"

	if !common.FileExists(oldEnvFile) {
		if migrated {
			return nil
		}
		return writeEnvMigrated(name)
	}

	if migrated {
		return preserveStaleEnvFile(name, oldEnvFile, env)
	}

	oldEnv, err := loadFromFile(name, oldEnvFile)
	if err != nil {
		return fmt.Errorf("Unable to load old environment: %s", err.Error())
	}

	env.Merge(oldEnv)
	if err := env.Write(); err != nil {
		return fmt.Errorf("Unable to write environment: %s", err.Error())
	}

	if err := common.SetPermissions(common.SetPermissionInput{
		Filename: env.Filename(),
		Mode:     os.FileMode(0600),
	}); err != nil {
		return fmt.Errorf("Unable to set permissions on environment: %s", err.Error())
	}

	if err := writeEnvMigrated(name); err != nil {
		return err
	}

	if err := os.Remove(oldEnvFile); err != nil {
		return fmt.Errorf("Unable to remove migrated file %s: %s", oldEnvFile, err.Error())
	}

	return nil
}

// preserveStaleEnvFile clears a legacy ENV file that outlived its own migration
// out of the way without importing any of it. A file whose values all match the
// current environment has nothing left to give and is removed outright;
// otherwise it is moved alongside itself with a .migrated suffix and the keys it
// disagrees on are named, so an operator who did write it by hand can still
// recover the values through `dokku config:set`.
func preserveStaleEnvFile(name string, oldEnvFile string, env *Env) error {
	staleEnv, err := loadFromFile(name, oldEnvFile)
	if err != nil {
		return fmt.Errorf("Unable to load stale environment: %s", err.Error())
	}

	diverged := divergedKeys(staleEnv, env)
	if len(diverged) == 0 {
		if err := os.Remove(oldEnvFile); err != nil {
			return fmt.Errorf("Unable to remove stale file %s: %s", oldEnvFile, err.Error())
		}
		return nil
	}

	preservedFile := oldEnvFile + ".migrated"
	common.LogWarn(fmt.Sprintf("Not importing already-migrated ENV file %s: the current config takes precedence for %s", oldEnvFile, strings.Join(diverged, " ")))
	common.LogWarn(fmt.Sprintf("Preserved at %s; apply any values still needed with dokku config:set", preservedFile))

	if err := os.Rename(oldEnvFile, preservedFile); err != nil {
		return fmt.Errorf("Unable to preserve stale file %s: %s", oldEnvFile, err.Error())
	}

	return nil
}

// divergedKeys returns the keys stale holds that env either does not have at all
// or holds a different value for.
func divergedKeys(stale *Env, env *Env) []string {
	diverged := []string{}
	for _, key := range stale.Keys() {
		staleValue, _ := stale.Get(key)
		if value, ok := env.Get(key); !ok || value != staleValue {
			diverged = append(diverged, key)
		}
	}

	return diverged
}

func writeEnvMigrated(name string) error {
	if err := common.PropertyWrite("config", name, envMigratedProperty, "true"); err != nil {
		return fmt.Errorf("Unable to set %s property: %s", envMigratedProperty, err.Error())
	}
	return nil
}
