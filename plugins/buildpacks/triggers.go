package buildpacks

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dokku/dokku/plugins/common"
)

// TriggerInstall runs the install step for the buildpacks plugin
func TriggerInstall() error {
	if err := common.PropertySetup("buildpacks"); err != nil {
		return fmt.Errorf("Unable to install the buildpacks plugin: %s", err.Error())
	}

	return migrateStackProperty()
}

// migrateStackProperty moves any legacy buildpacks stack values to the
// appropriate builder plugin, detecting herokuish vs pack stacks. It is
// idempotent: once a value is migrated the buildpacks stack key is removed.
func migrateStackProperty() error {
	appNames := []string{"--global"}
	apps, err := common.DokkuApps()
	if err != nil && !errors.Is(err, common.NoAppsExist) {
		return err
	}
	appNames = append(appNames, apps...)

	for _, appName := range appNames {
		if !common.PropertyExists("buildpacks", appName, "stack") {
			continue
		}

		value := common.PropertyGet("buildpacks", appName, "stack")
		if value != "" {
			builder := builderForStack(value)
			if !common.PropertyExists(builder, appName, "stack") {
				if err := common.PropertyWrite(builder, appName, "stack", value); err != nil {
					return err
				}
			}
		}

		if err := common.PropertyDelete("buildpacks", appName, "stack"); err != nil {
			return err
		}
	}

	return nil
}

// TriggerPostAppCloneSetup creates new buildpacks files
func TriggerPostAppCloneSetup(oldAppName string, newAppName string) error {
	err := common.PropertyClone("buildpacks", oldAppName, newAppName)
	if err != nil {
		return err
	}

	return nil
}

// TriggerPostAppRenameSetup renames buildpacks files
func TriggerPostAppRenameSetup(oldAppName string, newAppName string) error {
	if err := common.PropertyClone("buildpacks", oldAppName, newAppName); err != nil {
		return err
	}

	if err := common.PropertyDestroy("buildpacks", oldAppName); err != nil {
		return err
	}

	return nil
}

// TriggerPostDelete destroys the buildpacks property for a given app container
func TriggerPostDelete(appName string) error {
	return common.PropertyDestroy("buildpacks", appName)
}

// TriggerPostExtract writes a .buildpacks file into the app
func TriggerPostExtract(appName string, sourceWorkDir string) error {
	buildpacks, err := getBuildpacks(appName)
	if err != nil {
		return nil
	}

	if len(buildpacks) == 0 {
		return rewriteBuildpacksFile(sourceWorkDir)
	}

	buildpacksPath := filepath.Join(sourceWorkDir, ".buildpacks")
	file, err := os.OpenFile(buildpacksPath, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("Error writing .buildpacks file: %s", err.Error())
	}

	w := bufio.NewWriter(file)
	for _, buildpack := range buildpacks {
		buildpack, err = validBuildpackURL(buildpack)
		if err != nil {
			return err
		}

		fmt.Fprintln(w, buildpack)
	}

	if err = w.Flush(); err != nil {
		return fmt.Errorf("Error writing .buildpacks file: %s", err.Error())
	}
	if err = file.Chmod(0600); err != nil {
		return fmt.Errorf("Error setting .buildpacks file permissions: %s", err.Error())
	}

	return nil
}
