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
func CommandPropertySet(pluginName, appName, property, value string, properties map[string]string) {
	if err := VerifyAppName(appName); err != nil {
		LogFail(err.Error())
	}
	if property == "" {
		LogFail("No property specified")
	}

	if _, ok := properties[property]; !ok {
		properties := reflect.ValueOf(properties).MapKeys()
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
	propertyPath := getPropertyPath(pluginName, appName, property)
	if err := os.Remove(propertyPath); err != nil {
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
	propertyPath := getPropertyPath(pluginName, appName, property)
	_, err := os.Stat(propertyPath)
	return !os.IsNotExist(err)
}

// PropertyGet returns the value for a given property
func PropertyGet(pluginName string, appName string, property string) string {
	return PropertyGetDefault(pluginName, appName, property, "")
}

// PropertyGetDefault returns the value for a given property with a specified default value
func PropertyGetDefault(pluginName, appName, property, defaultValue string) (val string) {
	if !PropertyExists(pluginName, appName, property) {
		return
	}

	propertyPath := getPropertyPath(pluginName, appName, property)
	b, err := ioutil.ReadFile(propertyPath)
	if err != nil {
		LogWarn(fmt.Sprintf("Unable to read %s property %s.%s", pluginName, appName, property))
		return
	}
	val = string(b)
	return
}

// PropertyWrite writes a value for a given application property
func PropertyWrite(pluginName string, appName string, property string, value string) {
	if err := makePropertyPath(pluginName, appName); err != nil {
		LogFail(fmt.Sprintf("Unable to create %s config directory for %s: %s", pluginName, appName, err.Error()))
	}

	propertyPath := getPropertyPath(pluginName, appName, property)
	file, err := os.Create(propertyPath)
	if err != nil {
		LogFail(fmt.Sprintf("Unable to write %s config value %s.%s: %s", pluginName, appName, property, err.Error()))
	}
	defer file.Close()

	fmt.Fprintf(file, value)
	file.Chmod(0600)
	setPermissions(propertyPath, 0600)
}

// PropertySetup creates the plugin config root
func PropertySetup(pluginName string) (err error) {
	pluginConfigRoot := getPluginConfigPath(pluginName)
	if err = os.MkdirAll(pluginConfigRoot, 0755); err != nil {
		return
	}
	return setPermissions(pluginConfigRoot, 0755)
}

func getPropertyPath(pluginName string, appName string, property string) string {
	pluginAppConfigRoot := getPluginAppPropertyPath(pluginName, appName)
	return strings.Join([]string{pluginAppConfigRoot, property}, "/")
}

// getPluginAppPropertyPath returns the plugin property path for a given plugin/app combination
func getPluginAppPropertyPath(pluginName string, appName string) string {
	return strings.Join([]string{getPluginConfigPath(pluginName), appName}, "/")
}

// getPluginConfigPath returns the plugin property path for a given plugin
func getPluginConfigPath(pluginName string) string {
	return strings.Join([]string{MustGetEnv("DOKKU_LIB_ROOT"), "config", pluginName}, "/")
}

// makePluginAppPropertyPath ensures that a property path exists
func makePluginAppPropertyPath(pluginName string, appName string) (err error) {
	pluginAppConfigRoot := getPluginAppPropertyPath(pluginName, appName)
	if err = os.MkdirAll(pluginAppConfigRoot, 0755); err != nil {
		return
	}
	return setPermissions(pluginAppConfigRoot, 0755)
}

// setPermissions sets the proper owner and filemode for a given file
func setPermissions(path string, fileMode os.FileMode) (err error) {
	if err = os.Chmod(path, fileMode); err != nil {
		return err
	}

	systemGroup := os.Getenv("DOKKU_SYSTEM_GROUP")
	systemUser := os.Getenv("DOKKU_SYSTEM_USER")
	if systemGroup == "" {
		systemGroup = "dokku"
	}
	if systemUser == "" {
		systemUser = "dokku"
	}

	group, err := user.LookupGroup(systemGroup)
	if err != nil {
		return
	}
	user, err := user.Lookup(systemUser)
	if err != nil {
		return
	}

	uid, err := strconv.Atoi(user.Uid)
	if err != nil {
		return
	}

	gid, err := strconv.Atoi(group.Gid)
	if err != nil {
		return
	}
	return os.Chown(path, uid, gid)
}
