package logs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/dokku/dokku/plugins/common"
)

// TriggerDockerArgsProcessDeploy outputs the logs plugin docker options for an app
func TriggerDockerArgsProcessDeploy(appName string) error {
	stdin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	maxSize := common.PropertyGet("logs", appName, "max-size")
	if maxSize == "" {
		maxSize = common.PropertyGetDefault("logs", "--global", "max-size", MaxSize)
	}

	if maxSize != "unlimited" {
		fmt.Printf(" --log-opt max-size=%s ", maxSize)
	}

	fmt.Print(string(stdin))
	return nil
}

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
