package common

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

// CommandPropertySet is a generic function that will set a property for a given plugin/app combination
func CommandPropertySet(pluginName, appName, property, value string, properties map[string]string, globalProperties map[string]bool) {
	if appName != "--global" {
		if err := VerifyAppName(appName); err != nil {
			LogFailWithError(err)
		}
	}
	if appName == "--global" && !globalProperties[property] {
		LogFail("Property cannot be specified globally")
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

		sort.Strings(validPropertyList)
		LogFail(fmt.Sprintf("Invalid property specified, valid properties include: %s", strings.Join(validPropertyList, ", ")))
	}

	if value != "" {
		LogInfo2Quiet(fmt.Sprintf("Setting %s to %s", property, value))
		PropertyWrite(pluginName, appName, property, value)
	} else {
		LogInfo2Quiet(fmt.Sprintf("Unsetting %s", property))
		if err := PropertyDelete(pluginName, appName, property); err != nil {
			LogFailWithError(err)
		}
	}
}

// PropertyClone clones a set of properties from one app to another
func PropertyClone(pluginName string, oldAppName string, newAppName string) error {
	properties, err := PropertyGetAll(pluginName, oldAppName)
	if err != nil {
		return nil
	}

	for property, value := range properties {
		if err := PropertyWrite(pluginName, newAppName, property, value); err != nil {
			return err
		}
	}

	return nil
}

// PropertyDelete deletes a property from the plugin properties for an app
func PropertyDelete(pluginName string, appName string, property string) error {
	propertyPath := getPropertyPath(pluginName, appName, property)
	if err := os.Remove(propertyPath); err != nil {
		if !PropertyExists(pluginName, appName, property) {
			return nil
		}

		return fmt.Errorf("Unable to remove %s property %s.%s", pluginName, appName, property)
	}

	return nil
}

// PropertyDestroy destroys the plugin properties for an app
func PropertyDestroy(pluginName string, appName string) error {
	if appName == "_all_" {
		pluginConfigPath := getPluginConfigPath(pluginName)
		return os.RemoveAll(pluginConfigPath)
	}

	pluginAppConfigRoot := getPluginAppPropertyPath(pluginName, appName)
	return os.RemoveAll(pluginAppConfigRoot)
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

// PropertyGetAll returns a map of all properties for a given app
func PropertyGetAll(pluginName string, appName string) (map[string]string, error) {
	properties := make(map[string]string)
	pluginAppConfigRoot := getPluginAppPropertyPath(pluginName, appName)

	fi, err := os.Stat(pluginAppConfigRoot)
	if err != nil {
		return properties, nil
	}

	if !fi.IsDir() {
		return properties, errors.New("Specified property path is not a directory")
	}

	files, err := ioutil.ReadDir(pluginAppConfigRoot)
	if err != nil {
		return properties, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		property := file.Name()
		properties[property] = PropertyGet(pluginName, appName, property)
	}

	return properties, nil
}

// PropertyGetDefault returns the value for a given property with a specified default value
func PropertyGetDefault(pluginName, appName, property, defaultValue string) (val string) {
	if !PropertyExists(pluginName, appName, property) {
		val = defaultValue
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

// PropertyListAdd adds a property to a list at an optionally specified index
func PropertyListAdd(pluginName string, appName string, property string, value string, index int) error {
	if err := propertyTouch(pluginName, appName, property); err != nil {
		return err
	}

	scannedLines, err := PropertyListGet(pluginName, appName, property)
	if err != nil {
		return err
	}

	value = strings.TrimSpace(value)

	var lines []string
	for i, line := range scannedLines {
		if index != 0 && i == (index-1) {
			lines = append(lines, value)
		}
		lines = append(lines, line)
	}

	if index == 0 || index > len(scannedLines) {
		lines = append(lines, value)
	}

	return PropertyListWrite(pluginName, appName, property, lines)
}

// PropertyListWrite completely rewrites a list property
func PropertyListWrite(pluginName string, appName string, property string, values []string) error {
	if err := propertyTouch(pluginName, appName, property); err != nil {
		return err
	}

	propertyPath := getPropertyPath(pluginName, appName, property)
	file, err := os.OpenFile(propertyPath, os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	w := bufio.NewWriter(file)
	for _, line := range values {
		fmt.Fprintln(w, line)
	}
	if err = w.Flush(); err != nil {
		return fmt.Errorf("Unable to write %s config value %s.%s: %s", pluginName, appName, property, err.Error())
	}

	file.Chmod(0600)
	SetPermissions(propertyPath, 0600)
	return nil
}

// PropertyListGet returns a property list
func PropertyListGet(pluginName string, appName string, property string) (lines []string, err error) {
	if !PropertyExists(pluginName, appName, property) {
		return lines, nil
	}

	propertyPath := getPropertyPath(pluginName, appName, property)
	file, err := os.Open(propertyPath)
	if err != nil {
		return lines, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err = scanner.Err(); err != nil {
		return lines, fmt.Errorf("Unable to read %s config value for %s.%s: %s", pluginName, appName, property, err.Error())
	}

	return lines, nil
}

// PropertyListLength returns the length of a property list
func PropertyListLength(pluginName string, appName string, property string) (length int, err error) {
	if !PropertyExists(pluginName, appName, property) {
		return length, nil
	}

	propertyPath := getPropertyPath(pluginName, appName, property)
	file, err := os.Open(propertyPath)
	if err != nil {
		return length, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err = scanner.Err(); err != nil {
		return length, fmt.Errorf("Unable to read %s config value for %s.%s: %s", pluginName, appName, property, err.Error())
	}

	length = len(lines)
	return length, nil
}

// PropertyListGetByIndex returns an entry within property list by index
func PropertyListGetByIndex(pluginName string, appName string, property string, index int) (propertyValue string, err error) {
	lines, err := PropertyListGet(pluginName, appName, property)
	if err != nil {
		return
	}

	found := false
	for i, line := range lines {
		if i == index {
			propertyValue = line
			found = true
		}
	}

	if !found {
		err = errors.New("Index not found")
	}

	return
}

// PropertyListGetByValue returns an entry within property list by value
func PropertyListGetByValue(pluginName string, appName string, property string, value string) (propertyValue string, err error) {
	lines, err := PropertyListGet(pluginName, appName, property)
	if err != nil {
		return
	}

	found := false
	for _, line := range lines {
		if line == value {
			propertyValue = line
			found = true
		}
	}

	if !found {
		err = errors.New("Value not found")
	}

	return
}

// PropertyListRemove removes a value from a property list
func PropertyListRemove(pluginName string, appName string, property string, value string) error {
	lines, err := PropertyListGet(pluginName, appName, property)
	if err != nil {
		return err
	}

	propertyPath := getPropertyPath(pluginName, appName, property)
	file, err := os.OpenFile(propertyPath, os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	found := false
	w := bufio.NewWriter(file)
	for _, line := range lines {
		if line == value {
			found = true
			continue
		}
		fmt.Fprintln(w, line)
	}
	if err = w.Flush(); err != nil {
		return fmt.Errorf("Unable to write %s config value %s.%s: %s", pluginName, appName, property, err.Error())
	}

	file.Chmod(0600)
	SetPermissions(propertyPath, 0600)

	if !found {
		return errors.New("Property not found, nothing was removed")
	}

	return nil
}

// PropertyListRemoveByPrefix removes a value by prefix from a property list
func PropertyListRemoveByPrefix(pluginName string, appName string, property string, prefix string) error {
	lines, err := PropertyListGet(pluginName, appName, property)
	if err != nil {
		return err
	}

	propertyPath := getPropertyPath(pluginName, appName, property)
	file, err := os.OpenFile(propertyPath, os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	found := false
	w := bufio.NewWriter(file)
	for _, line := range lines {
		if strings.HasPrefix(line, prefix) {
			found = true
			continue
		}
		fmt.Fprintln(w, line)
	}
	if err = w.Flush(); err != nil {
		return fmt.Errorf("Unable to write %s config value %s.%s: %s", pluginName, appName, property, err.Error())
	}

	file.Chmod(0600)
	SetPermissions(propertyPath, 0600)

	if !found {
		return errors.New("Property not found, nothing was removed")
	}

	return nil
}

// PropertyListSet sets a value within a property list at a specified index
func PropertyListSet(pluginName string, appName string, property string, value string, index int) error {
	if err := propertyTouch(pluginName, appName, property); err != nil {
		return err
	}

	scannedLines, err := PropertyListGet(pluginName, appName, property)
	if err != nil {
		return err
	}

	value = strings.TrimSpace(value)

	var lines []string
	if index >= len(scannedLines) {
		for _, line := range scannedLines {
			lines = append(lines, line)
		}
		lines = append(lines, value)
	} else {
		for i, line := range scannedLines {
			if i == index {
				lines = append(lines, value)
			} else {
				lines = append(lines, line)
			}
		}
	}

	return PropertyListWrite(pluginName, appName, property, lines)
}

// propertyTouch ensures a given application property file exists
func propertyTouch(pluginName string, appName string, property string) error {
	if err := makePluginAppPropertyPath(pluginName, appName); err != nil {
		return fmt.Errorf("Unable to create %s config directory for %s: %s", pluginName, appName, err.Error())
	}

	propertyPath := getPropertyPath(pluginName, appName, property)
	if PropertyExists(pluginName, appName, property) {
		return nil
	}

	file, err := os.Create(propertyPath)
	if err != nil {
		return fmt.Errorf("Unable to write %s config value %s.%s: %s", pluginName, appName, property, err.Error())
	}
	defer file.Close()

	return nil
}

// PropertyWrite writes a value for a given application property
func PropertyWrite(pluginName string, appName string, property string, value string) error {
	if err := propertyTouch(pluginName, appName, property); err != nil {
		return err
	}

	propertyPath := getPropertyPath(pluginName, appName, property)
	file, err := os.Create(propertyPath)
	if err != nil {
		return fmt.Errorf("Unable to write %s config value %s.%s: %s", pluginName, appName, property, err.Error())
	}
	defer file.Close()

	fmt.Fprint(file, value)
	file.Chmod(0600)
	SetPermissions(propertyPath, 0600)
	return nil
}

// PropertySetup creates the plugin config root
func PropertySetup(pluginName string) error {
	pluginConfigRoot := getPluginConfigPath(pluginName)
	if err := os.MkdirAll(pluginConfigRoot, 0755); err != nil {
		return err
	}
	if err := SetPermissions(filepath.Join(MustGetEnv("DOKKU_LIB_ROOT"), "config"), 0755); err != nil {
		return err
	}
	return SetPermissions(pluginConfigRoot, 0755)
}

func getPropertyPath(pluginName string, appName string, property string) string {
	pluginAppConfigRoot := getPluginAppPropertyPath(pluginName, appName)
	return filepath.Join(pluginAppConfigRoot, property)
}

// getPluginAppPropertyPath returns the plugin property path for a given plugin/app combination
func getPluginAppPropertyPath(pluginName string, appName string) string {
	return filepath.Join(getPluginConfigPath(pluginName), appName)
}

// getPluginConfigPath returns the plugin property path for a given plugin
func getPluginConfigPath(pluginName string) string {
	return filepath.Join(MustGetEnv("DOKKU_LIB_ROOT"), "config", pluginName)
}

// makePluginAppPropertyPath ensures that a property path exists
func makePluginAppPropertyPath(pluginName string, appName string) error {
	pluginAppConfigRoot := getPluginAppPropertyPath(pluginName, appName)
	if err := os.MkdirAll(pluginAppConfigRoot, 0755); err != nil {
		return err
	}
	return SetPermissions(pluginAppConfigRoot, 0755)
}
