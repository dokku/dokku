package logs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dokku/dokku/plugins/common"
)

// TriggerInstall initializes app restart policies
func TriggerInstall() error {
	if err := common.PropertySetup("logs"); err != nil {
		return fmt.Errorf("Unable to install the logs plugin: %s", err.Error())
	}

	directory := filepath.Join(common.MustGetEnv("DOKKU_LIB_ROOT"), "data", "logs")
	if err := os.MkdirAll(directory, 0755); err != nil {
		return err
	}

	if err := common.SetPermissions(directory, 0755); err != nil {
		return err
	}

	logDirectory := filepath.Join(common.MustGetEnv("DOKKU_LOGS_DIR"), "apps")
	if err := os.MkdirAll(logDirectory, 0755); err != nil {
		return err
	}

	if err := common.SetPermissions(logDirectory, 0755); err != nil {
		return err
	}

	return nil
}

// TriggerPostDelete destroys the network property for a given app container
func TriggerPostDelete(appName string) error {
	return common.PropertyDestroy("logs", appName)
}
