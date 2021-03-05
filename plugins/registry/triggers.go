package registry

import (
	"fmt"

	"github.com/dokku/dokku/plugins/common"
)

// TriggerInstall runs the install step for the registry plugin
func TriggerInstall() error {
	if err := common.PropertySetup("registry"); err != nil {
		return fmt.Errorf("Unable to install the registry plugin: %s", err.Error())
	}

	return nil
}

// TriggerPostDelete destroys the registry property for a given app container
func TriggerPostDelete(appName string) error {
	return common.PropertyDestroy("registry", appName)
}
