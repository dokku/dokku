package resource

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// TriggerDockerArgsProcessDeploy outputs the process-specific docker options
func TriggerDockerArgsProcessDeploy(appName string, processType string) error {
	stdin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	if os.Getenv("DOKKU_OMIT_RESOURCE_ARGS") == "1" {
		fmt.Print(string(stdin))
		return nil
	}

	resources, err := common.PropertyGetAll("resource", appName)
	if err != nil {
		fmt.Print(string(stdin))
		return nil
	}

	limits := make(map[string]string)
	reservations := make(map[string]string)

	validLimits := map[string]bool{
		"cpu":         true,
		"memory":      true,
		"memory-swap": true,
	}
	validReservations := map[string]bool{
		"memory": true,
	}
	validPrefixes := []string{"_default_.", fmt.Sprintf("%s.", processType)}
	for _, validPrefix := range validPrefixes {
		for key, value := range resources {
			if !strings.HasPrefix(key, validPrefix) {
				continue
			}
			parts := strings.SplitN(strings.TrimPrefix(key, validPrefix), ".", 2)
			if parts[0] == "limit" {
				if !validLimits[parts[1]] {
					continue
				}

				if parts[1] == "cpu" {
					parts[1] = "cpus"
				}

				limits[parts[1]] = value
			}
			if parts[0] == "reserve" {
				if !validReservations[parts[1]] {
					continue
				}

				reservations[parts[1]] = value
			}
		}
	}

	for key, value := range limits {
		if value == "" {
			continue
		}
		fmt.Printf(" --%s=%s ", key, value)
	}

	for key, value := range reservations {
		if value == "" {
			continue
		}
		fmt.Printf(" --%s-reservation=%s ", key, value)
	}

	fmt.Print(string(stdin))
	return nil
}

// TriggerInstall runs the install step for the resource plugin
func TriggerInstall() error {
	if err := common.PropertySetup("resource"); err != nil {
		return fmt.Errorf("Unable to install the resource plugin: %v", err)
	}
	return nil
}

// TriggerPostAppCloneSetup creates new resource files
func TriggerPostAppCloneSetup(oldAppName string, newAppName string) error {
	err := common.PropertyClone("resource", oldAppName, newAppName)
	if err != nil {
		return err
	}

	return nil
}

// TriggerPostAppRenameSetup renames resource files
func TriggerPostAppRenameSetup(oldAppName string, newAppName string) error {
	if err := common.PropertyClone("resource", oldAppName, newAppName); err != nil {
		return err
	}

	if err := common.PropertyDestroy("resource", oldAppName); err != nil {
		return err
	}

	return nil
}

// TriggerPostDelete destroys the resource property for a given app container
func TriggerPostDelete(appName string) error {
	return common.PropertyDestroy("resource", appName)
}

// TriggerResourceGetProperty writes the resource key to stdout for a given app container
func TriggerResourceGetProperty(appName string, processType string, resourceType string, key string) error {
	value, err := GetResourceValue(appName, processType, resourceType, key)
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, value)
	return nil
}
