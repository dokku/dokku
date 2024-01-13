package scheduler_k3s

import (
	"fmt"
	"os"

	"github.com/dokku/dokku/plugins/common"
	"github.com/rancher/wharfie/pkg/registries"
	"gopkg.in/yaml.v3"
)

// TriggerInstall runs the install step for the scheduler-k3s plugin
func TriggerInstall() error {
	if err := common.PropertySetup("scheduler-k3s"); err != nil {
		return fmt.Errorf("Unable to install the scheduler-k3s plugin: %s", err.Error())
	}

	return nil
}

// TriggerPostAppCloneSetup creates new scheduler-k3s files
func TriggerPostAppCloneSetup(oldAppName string, newAppName string) error {
	err := common.PropertyClone("scheduler-k3s", oldAppName, newAppName)
	if err != nil {
		return err
	}

	return nil
}

// TriggerPostAppRenameSetup renames scheduler-k3s files
func TriggerPostAppRenameSetup(oldAppName string, newAppName string) error {
	if err := common.PropertyClone("scheduler-k3s", oldAppName, newAppName); err != nil {
		return err
	}

	if err := common.PropertyDestroy("scheduler-k3s", oldAppName); err != nil {
		return err
	}

	return nil
}

// TriggerPostDelete destroys the scheduler-k3s data for a given app container
func TriggerPostDelete(appName string) error {
	dataErr := common.RemoveAppDataDirectory("scheduler-k3s", appName)
	propertyErr := common.PropertyDestroy("scheduler-k3s", appName)

	if dataErr != nil {
		return dataErr
	}

	return propertyErr
}

// TriggerPostRegistryLogin updates the `/etc/rancher/k3s/registries.yaml` to include
// auth information for the registry. Note that if the file does not exist, it won't be updated.
func TriggerPostRegistryLogin(server string, username string, password string) error {
	if !common.FileExists("/usr/local/bin/k3s") {
		return nil
	}

	registry := registries.Registry{}
	registryFile := "/etc/rancher/k3s/registries.yaml"
	yamlFile, err := os.ReadFile(registryFile)
	if err != nil {
		return fmt.Errorf("Unable to read existing registries.yaml: %w", err)
	}

	err = yaml.Unmarshal(yamlFile, registry)
	if err != nil {
		return fmt.Errorf("Unable to unmarshal registry configuration from yaml: %w", err)
	}

	common.LogInfo1("Updating k3s configuration")
	if registry.Auths == nil {
		registry.Auths = map[string]registries.AuthConfig{}
	}

	if server == "docker.io" {
		server = "registry-1.docker.io"
	}

	registry.Auths[server] = registries.AuthConfig{
		Username: username,
		Password: password,
	}

	data, err := yaml.Marshal(&registry)
	if err != nil {
		return fmt.Errorf("Unable to marshal registry configuration to yaml: %w", err)
	}

	if err := os.WriteFile(registryFile, data, os.FileMode(644)); err != nil {
		return fmt.Errorf("Unable to write registry configuration to file: %w", err)
	}

	// todo: auth against all nodes in cluster
	return nil
}
