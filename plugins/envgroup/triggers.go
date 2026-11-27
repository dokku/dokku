package envgroup

import (
	"fmt"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
)

// TriggerEnvgroupAppEnv outputs the merged environment of an app's groups in
// the given format, defaulting to json
func TriggerEnvgroupAppEnv(appName string, format string) error {
	if format == "" {
		format = "json"
	}

	exportType, err := config.ParseExportFormat(format)
	if err != nil {
		return err
	}

	env, err := MergedAppGroupEnv(appName)
	if err != nil {
		return err
	}

	exported := env.Export(exportType)
	if exported != "" {
		fmt.Println(exported)
	}
	return nil
}

// TriggerInstall runs the install step for the envgroup plugin
func TriggerInstall() error {
	if err := common.PropertySetup(pluginName); err != nil {
		return fmt.Errorf("Unable to install the envgroup plugin: %s", err.Error())
	}

	if err := common.CreateDataDirectory(pluginName); err != nil {
		return fmt.Errorf("Unable to install the envgroup plugin: %s", err.Error())
	}

	return nil
}

// TriggerPostAppCloneSetup copies an app's group memberships to its clone
func TriggerPostAppCloneSetup(oldAppName string, newAppName string) error {
	return common.PropertyClone(pluginName, oldAppName, newAppName)
}

// TriggerPostAppRenameSetup moves an app's group memberships to its new name
func TriggerPostAppRenameSetup(oldAppName string, newAppName string) error {
	if err := common.PropertyClone(pluginName, oldAppName, newAppName); err != nil {
		return err
	}

	return common.PropertyDestroy(pluginName, oldAppName)
}

// TriggerPostDelete removes an app's group memberships
func TriggerPostDelete(appName string) error {
	return common.PropertyDestroy(pluginName, appName)
}
