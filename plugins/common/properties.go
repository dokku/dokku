package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"reflect"
	"strconv"
	"strings"
)

// CommandPropertySet is a generic function that will set a property for a given plugin/app combination
func CommandPropertySet(pluginName string, appName string, property string, value string, validProperties map[string]bool) {
	err := VerifyAppName(appName)
	if err != nil {
		LogFail(err.Error())
	}

	if property == "" {
		LogFail("No property specified")
	}

	if !isValidProperty(validProperties, property) {
		properties := reflect.ValueOf(validProperties).MapKeys()
		validPropertyList := make([]string, len(properties))
		for i := 0; i < len(properties); i++ {
			validPropertyList[i] = properties[i].String()
		}

		LogFail(fmt.Sprintf("Invalid property specified, valid properties include: %s", strings.Join(validPropertyList, ", ")))
	}

	if value != "" {
		LogInfo2Quiet(fmt.Sprintf("Setting %s to %s", property, value))
		PropertyWrite(pluginName, appName, property, value)
	} else {
		LogInfo2Quiet(fmt.Sprintf("Unsetting %s", property))
		PropertyDelete(pluginName, appName, property)
	}
}

// PropertyDelete deletes a property from the plugin properties for an app
func PropertyDelete(pluginName string, appName string, property string) {
	pluginAppConfigRoot := getPluginAppPropertyPath(pluginName, appName)
	propertyPath := strings.Join([]string{pluginAppConfigRoot, property}, "/")
	err := os.Remove(propertyPath)
	if err != nil {
		LogFail(fmt.Sprintf("Unable to remove %s property %s.%s", pluginName, appName, property))
	}
}

// PropertyDestroy destroys the plugin properties for an app
func PropertyDestroy(pluginName string, appName string) {
	if appName == "_all_" {
		pluginConfigPath := getPluginConfigPath(pluginName)
		os.RemoveAll(pluginConfigPath)
	} else {
		pluginAppConfigRoot := getPluginAppPropertyPath(pluginName, appName)
		os.RemoveAll(pluginAppConfigRoot)
	}
}

// PropertyExists returns whether a property exists or not
func PropertyExists(pluginName string, appName string, property string) bool {
	pluginAppConfigRoot := getPluginAppPropertyPath(pluginName, appName)
	propertyPath := strings.Join([]string{pluginAppConfigRoot, property}, "/")
	_, err := os.Stat(propertyPath)
	return !os.IsNotExist(err)
}

// PropertyGet returns the value for a given property
func PropertyGet(pluginName string, appName string, property string) string {
	return PropertyGetDefault(pluginName, appName, property, "")
}

// PropertyGetDefault returns the value for a given property with a specified default value
func PropertyGetDefault(pluginName string, appName string, property string, defaultValue string) string {
	if !PropertyExists(pluginName, appName, property) {
		return ""
	}

	pluginAppConfigRoot := getPluginAppPropertyPath(pluginName, appName)
	propertyPath := strings.Join([]string{pluginAppConfigRoot, property}, "/")

	b, err := ioutil.ReadFile(propertyPath)
	if err != nil {
		LogWarn(fmt.Sprintf("Unable to read %s property %s.%s", pluginName, appName, property))
		return ""
	}

	return string(b)
}

// PropertyWrite writes a value for a given application property
func PropertyWrite(pluginName string, appName string, property string, value string) {
	err := makePropertyPath(pluginName, appName)
	if err != nil {
		LogFail(fmt.Sprintf("Unable to create %s config directory for %s: %s", pluginName, appName, err.Error()))
	}

	pluginAppConfigRoot := getPluginAppPropertyPath(pluginName, appName)
	propertyPath := strings.Join([]string{pluginAppConfigRoot, property}, "/")
	file, err := os.Create(propertyPath)
	if err != nil {
		LogFail(fmt.Sprintf("Unable to write %s config value %s.%s: %s", pluginName, appName, property, err.Error))
	}
	defer file.Close()

	fmt.Fprintf(file, value)
	file.Chmod(0600)
	setPermissions(propertyPath, 0600)
}

// PropertySetup creates the plugin config root
func PropertySetup(pluginName string) error {
	pluginConfigRoot := getPluginConfigPath(pluginName)
	err := os.MkdirAll(pluginConfigRoot, 0755)
	if err != nil {
		return err
	}
	return setPermissions(pluginConfigRoot, 0755)
}

// isValidProperty returns whether a property is a valid property or not
func isValidProperty(validProperties map[string]bool, property string) bool {
	return validProperties[property]
}

// getPluginAppPropertyPath returns the plugin property path for a given plugin/app combination
func getPluginAppPropertyPath(pluginName string, appName string) string {
	return strings.Join([]string{getPluginConfigPath(pluginName), appName}, "/")
}

// getPluginConfigPath returns the plugin property path for a given plugin
func getPluginConfigPath(pluginName string) string {
	return strings.Join([]string{MustGetEnv("DOKKU_LIB_ROOT"), "config", pluginName}, "/")
}

// makePropertyPath ensures that a property path exists
func makePropertyPath(pluginName string, appName string) error {
	pluginAppConfigRoot := getPluginAppPropertyPath(pluginName, appName)
	err := os.MkdirAll(pluginAppConfigRoot, 0755)
	if err != nil {
		return err
	}
	return setPermissions(pluginAppConfigRoot, 0755)
}

// setPermissions sets the proper owner and filemode for a given file
func setPermissions(path string, fileMode os.FileMode) error {
	err := os.Chmod(path, fileMode)
	if err != nil {
		return err
	}

	group, err := user.LookupGroup("dokku")
	if err != nil {
		return err
	}
	user, err := user.Lookup("dokku")
	if err != nil {
		return err
	}

	uid, err := strconv.Atoi(user.Uid)
	if err != nil {
		return err
	}

	gid, err := strconv.Atoi(group.Gid)
	if err != nil {
		return err
	}

	err = os.Chown(path, uid, gid)
	if err != nil {
		return err
	}
	return nil
}
