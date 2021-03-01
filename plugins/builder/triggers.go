package builder

import (
	"fmt"

	"github.com/dokku/dokku/plugins/common"
)

// TriggerBuilderDetect outputs a manually selected builder for the app
func TriggerBuilderDetect(appName string) error {
	if builder := common.PropertyGet("builder", appName, "selected"); builder != "" {
		fmt.Println(builder)
		return nil
	}

	if builder := common.PropertyGet("builder", "--global", "selected"); builder != "" {
		fmt.Println(builder)
		return nil
	}

	return nil
}

// TriggerInstall runs the install step for the builder plugin
func TriggerInstall() error {
	if err := common.PropertySetup("builder"); err != nil {
		return fmt.Errorf("Unable to install the builder plugin: %s", err.Error())
	}

	return nil
}

// TriggerPostDelete destroys the builder property for a given app container
func TriggerPostDelete(appName string) error {
	return common.PropertyDestroy("builder", appName)
}
