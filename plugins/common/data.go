package common

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/otiai10/copy"
)

// CreateAppDataDirectory creates a data directory for the given plugin/app combination with the correct permissions
func CreateAppDataDirectory(pluginName, appName string) error {
	directory := GetAppDataDirectory(pluginName, appName)
	if err := os.MkdirAll(directory, 0755); err != nil {
		return err
	}

	if err := SetPermissions(directory, 0755); err != nil {
		return err
	}

	return nil
}

// CreateDataDirectory creates a data directory for the given plugin/app combination with the correct permissions
func CreateDataDirectory(pluginName string) error {
	directory := GetDataDirectory(pluginName)
	if err := os.MkdirAll(directory, 0755); err != nil {
		return err
	}

	if err := SetPermissions(directory, 0755); err != nil {
		return err
	}

	return nil
}

// GetAppDataDirectory returns the path to the data directory for the given plugin/app combination
func GetAppDataDirectory(pluginName string, appName string) string {
	return filepath.Join(GetDataDirectory(pluginName), appName)
}

// GetDataDirectory returns the path to the data directory for the specified plugin
func GetDataDirectory(pluginName string) string {
	return filepath.Join(MustGetEnv("DOKKU_LIB_ROOT"), "data", pluginName)
}

// RemoveAppDataDirectory removes the path to the data directory for the given plugin/app combination
func RemoveAppDataDirectory(pluginName, appName string) error {
	return os.RemoveAll(GetAppDataDirectory(pluginName, appName))
}

func CloneAppData(pluginName string, oldAppName string, newAppName string) error {
	oldDataDir := GetAppDataDirectory(pluginName, oldAppName)
	newDataDir := GetAppDataDirectory(pluginName, newAppName)
	if err := copy.Copy(oldDataDir, newDataDir); err != nil {
		return fmt.Errorf("Unable to clone app data to new location: %v", err.Error())
	}

	return nil
}

// SetupAppData creates
func SetupAppData(pluginName string) error {
	if err := CreateDataDirectory(pluginName); err != nil {
		return err
	}

	apps, err := UnfilteredDokkuApps()
	if err != nil {
		return nil
	}

	for _, appName := range apps {
		if err := CreateAppDataDirectory(pluginName, appName); err != nil {
			return err
		}
	}

	return nil
}
