package logs

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	dockeroptions "github.com/dokku/dokku/plugins/docker-options"
)

// TriggerDockerArgsProcessDeploy outputs the logs plugin docker options for an app
func TriggerDockerArgsProcessDeploy(appName string) error {
	stdin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	allowedDrivers := map[string]bool{
		"local":     true,
		"json-file": true,
	}

	ignoreMaxSize := false
	options, err := dockeroptions.GetDockerOptionsForPhase(appName, "deploy")
	if err != nil {
		return err
	}

	for _, option := range options {
		if !strings.HasPrefix(option, "--log-driver=") {
			continue
		}

		logDriver := strings.TrimPrefix(option, "--log-driver=")
		if !allowedDrivers[logDriver] {
			ignoreMaxSize = true
		}
		break
	}

	if !ignoreMaxSize {
		maxSize := common.PropertyGet("logs", appName, "max-size")
		if maxSize == "" {
			maxSize = common.PropertyGetDefault("logs", "--global", "max-size", MaxSize)
		}

		if maxSize != "unlimited" {
			fmt.Printf(" --log-opt=max-size=%s ", maxSize)
		}
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

// TriggerLogsGetProperty writes the logs key to stdout for a given app container
func TriggerLogsGetProperty(appName string, key string) error {
	if key != "max-size" {
		return errors.New("Invalid logs property specified")
	}

	value := common.PropertyGet("logs", appName, "max-size")
	if value == "" {
		value = common.PropertyGetDefault("logs", "--global", "max-size", MaxSize)
	}

	fmt.Println(value)
	return nil
}

// TriggerPostDelete destroys the network property for a given app container
func TriggerPostDelete(appName string) error {
	return common.PropertyDestroy("logs", appName)
}
