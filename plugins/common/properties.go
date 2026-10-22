package common

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

// CommandPropertySet is a generic function that will set a property for a given plugin/app combination
func CommandPropertySet(pluginName, appName, property, value string, validProperties map[string]string, validGlobalProperties map[string]bool) {
	if appName != "--global" {
		if err := VerifyAppName(appName); err != nil {
			LogFailWithError(err)
		}
	}
	if appName == "--global" && !validGlobalProperties[property] {
		LogFail("Property cannot be specified globally")
	}
	if property == "" {
		LogFail("No property specified")
	}

	for k := range validGlobalProperties {
		if _, ok := validProperties[k]; !ok {
			validProperties[k] = ""
		}
	}

	if _, ok := validProperties[property]; !ok {
		properties := reflect.ValueOf(validProperties).MapKeys()
		validPropertyList := make([]string, len(properties))
		for i := 0; i < len(properties); i++ {
			validPropertyList[i] = properties[i].String()
		}

		sort.Strings(validPropertyList)
		LogFail(fmt.Sprintf("Invalid property specified, valid properties include: %s", strings.Join(validPropertyList, ", ")))
	}

	if value != "" {
		LogInfo2Quiet(fmt.Sprintf("Setting %s to %s", property, value))
		if err := PropertyWrite(pluginName, appName, property, value); err != nil {
			LogFailWithError(err)
		}
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

	files, err := os.ReadDir(pluginAppConfigRoot)
	if err != nil {
		return properties, err
	}

	for _, file := range files {
		if file.IsDir() || strings.HasPrefix(file.Name(), ".") {
			continue
		}
		property := file.Name()
		properties[property] = PropertyGet(pluginName, appName, property)
	}

	return properties, nil
}

// PropertyGetAllByPrefix returns a map of all properties for a given app with a specified prefix
func PropertyGetAllByPrefix(pluginName string, appName string, prefix string) (map[string]string, error) {
	properties := make(map[string]string)
	pluginAppConfigRoot := getPluginAppPropertyPath(pluginName, appName)

	fi, err := os.Stat(pluginAppConfigRoot)
	if err != nil {
		return properties, nil
	}

	if !fi.IsDir() {
		return properties, errors.New("Specified property path is not a directory")
	}

	files, err := os.ReadDir(pluginAppConfigRoot)
	if err != nil {
		return properties, err
	}

	for _, file := range files {
		if file.IsDir() || strings.HasPrefix(file.Name(), ".") {
			continue
		}
		property := file.Name()
		if !strings.HasPrefix(property, prefix) {
			continue
		}

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
	b, err := os.ReadFile(propertyPath)
	if err != nil {
		LogWarn(fmt.Sprintf("Unable to read %s property %s.%s", pluginName, appName, property))
		return
	}
	val = string(b)
	return
}

// PropertyListAdd adds a property to a list at an optionally specified index
func PropertyListAdd(pluginName string, appName string, property string, value string, index int) error {
	value = strings.TrimSpace(value)
	return PropertyListUpdate(pluginName, appName, property, func(scannedLines []string) ([]string, error) {
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

		return lines, nil
	})
}

// PropertyListUpdate reads a list property, passes it to fn and writes back the
// list fn returns, holding the plugin/app property lock throughout so a
// concurrent update cannot be lost between the read and the write. Nothing is
// written when fn returns an error.
func PropertyListUpdate(pluginName string, appName string, property string, fn func(lines []string) ([]string, error)) error {
	return PropertyLock(pluginName, appName, func(p *LockedProperties) error {
		lines, err := PropertyListGet(pluginName, appName, property)
		if err != nil {
			return err
		}

		lines, err = fn(lines)
		if err != nil {
			return err
		}

		return p.ListWrite(property, lines)
	})
}

// PropertyListWrite completely rewrites a list property
func PropertyListWrite(pluginName string, appName string, property string, values []string) error {
	return PropertyLock(pluginName, appName, func(p *LockedProperties) error {
		return p.ListWrite(property, values)
	})
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
	return propertyListRemoveMatching(pluginName, appName, property, func(line string) bool {
		return line == value
	})
}

// PropertyListRemoveByPrefix removes a value by prefix from a property list
func PropertyListRemoveByPrefix(pluginName string, appName string, property string, prefix string) error {
	return propertyListRemoveMatching(pluginName, appName, property, func(line string) bool {
		return strings.HasPrefix(line, prefix)
	})
}

// propertyListRemoveMatching drops every line of a list property that matches,
// returning an error when nothing matched. A missing property is reported as
// not found without taking the lock, so removing from an app with no
// properties does not create its property directory.
func propertyListRemoveMatching(pluginName string, appName string, property string, matches func(line string) bool) error {
	errNotFound := errors.New("Property not found, nothing was removed")
	if !PropertyExists(pluginName, appName, property) {
		return errNotFound
	}

	return PropertyListUpdate(pluginName, appName, property, func(lines []string) ([]string, error) {
		found := false
		kept := []string{}
		for _, line := range lines {
			if matches(line) {
				found = true
				continue
			}
			kept = append(kept, line)
		}

		if !found {
			return nil, errNotFound
		}
		return kept, nil
	})
}

// PropertyListSet sets a value within a property list at a specified index
func PropertyListSet(pluginName string, appName string, property string, value string, index int) error {
	value = strings.TrimSpace(value)
	return PropertyListUpdate(pluginName, appName, property, func(scannedLines []string) ([]string, error) {
		var lines []string
		if index >= len(scannedLines) {
			lines = append(lines, scannedLines...)
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

		return lines, nil
	})
}

// LockedProperties writes the properties of one plugin/app pair while the
// caller holds that pair's property lock. It is handed to the PropertyLock
// callback and is only valid until the callback returns.
type LockedProperties struct {
	pluginName string
	appName    string
}

// Write writes a value for a property
func (p *LockedProperties) Write(property string, value string) error {
	return writePropertyFile(p.pluginName, p.appName, property, []byte(value))
}

// Delete deletes a property
func (p *LockedProperties) Delete(property string) error {
	return PropertyDelete(p.pluginName, p.appName, property)
}

// ListWrite completely rewrites a list property
func (p *LockedProperties) ListWrite(property string, values []string) error {
	var b strings.Builder
	for _, line := range values {
		b.WriteString(line)
		b.WriteString("\n")
	}
	return writePropertyFile(p.pluginName, p.appName, property, []byte(b.String()))
}

// MapWrite persists a string→string map for a property as a single JSON file.
// A nil map is persisted as "{}".
func (p *LockedProperties) MapWrite(property string, m map[string]string) error {
	if m == nil {
		m = map[string]string{}
	}

	b, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("Unable to marshal %s map property %s.%s: %w", p.pluginName, p.appName, property, err)
	}

	return writePropertyFile(p.pluginName, p.appName, property, b)
}

// PropertyLock runs fn while holding an exclusive lock on the properties of a
// plugin/app pair, creating the property directory when it does not exist yet.
// Use it to read properties and write values derived from them without a
// concurrent writer slipping in between. Reads inside fn go through the normal
// getters; writes must go through the LockedProperties handle, since the public
// writers take this same lock and would block forever. Locks for other
// plugin/app pairs may be taken inside fn.
func PropertyLock(pluginName string, appName string, fn func(p *LockedProperties) error) error {
	if err := makePluginAppPropertyPath(pluginName, appName); err != nil {
		return fmt.Errorf("Unable to create %s config directory for %s: %s", pluginName, appName, err.Error())
	}

	return LockDirectory(getPluginAppPropertyPath(pluginName, appName), func() error {
		return fn(&LockedProperties{pluginName: pluginName, appName: appName})
	})
}

// writePropertyFile atomically replaces a property's contents. The data is
// written to a hidden temporary file in the property directory and renamed
// into place, so a concurrent reader sees either the old or the new contents
// and never a truncated file.
func writePropertyFile(pluginName string, appName string, property string, data []byte) error {
	propertyPath := getPropertyPath(pluginName, appName, property)
	file, err := os.CreateTemp(filepath.Dir(propertyPath), fmt.Sprintf(".%s.tmp-*", property))
	if err != nil {
		return fmt.Errorf("Unable to write %s config value %s.%s: %s", pluginName, appName, property, err.Error())
	}
	defer os.Remove(file.Name())

	if _, err := file.Write(data); err != nil {
		file.Close()
		return fmt.Errorf("Unable to write %s config value %s.%s: %s", pluginName, appName, property, err.Error())
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("Unable to close %s config value %s.%s: %s", pluginName, appName, property, err.Error())
	}

	if err := SetPermissions(SetPermissionInput{
		Filename: file.Name(),
		Mode:     os.FileMode(0600),
	}); err != nil {
		return fmt.Errorf("Unable to set permissions for %s config value %s.%s: %s", pluginName, appName, property, err.Error())
	}

	if err := os.Rename(file.Name(), propertyPath); err != nil {
		return fmt.Errorf("Unable to write %s config value %s.%s: %s", pluginName, appName, property, err.Error())
	}

	return nil
}

// PropertyWrite writes a value for a given application property
func PropertyWrite(pluginName string, appName string, property string, value string) error {
	return PropertyLock(pluginName, appName, func(p *LockedProperties) error {
		return p.Write(property, value)
	})
}

// PropertyMapWrite persists a string→string map for a given property as a single
// JSON file. Unlike PropertyWrite/PropertyListWrite, map *keys* may contain any
// bytes (including '/' or '\n') because they live inside file content rather
// than in the on-disk filename. A nil map is persisted as "{}".
func PropertyMapWrite(pluginName string, appName string, property string, m map[string]string) error {
	return PropertyLock(pluginName, appName, func(p *LockedProperties) error {
		return p.MapWrite(property, m)
	})
}

// PropertyMapGet returns the string→string map persisted for a property. A
// missing or empty-file property is reported as an empty map without error.
func PropertyMapGet(pluginName string, appName string, property string) (map[string]string, error) {
	m := map[string]string{}
	if !PropertyExists(pluginName, appName, property) {
		return m, nil
	}

	b, err := os.ReadFile(getPropertyPath(pluginName, appName, property))
	if err != nil {
		return nil, fmt.Errorf("Unable to read %s map property %s.%s: %w", pluginName, appName, property, err)
	}

	if len(b) == 0 {
		return m, nil
	}

	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("Unable to parse %s map property %s.%s as JSON: %w", pluginName, appName, property, err)
	}
	return m, nil
}

// PropertyMapSet reads the current map, sets one key, and writes the map back.
func PropertyMapSet(pluginName string, appName string, property string, key string, value string) error {
	return PropertyLock(pluginName, appName, func(p *LockedProperties) error {
		m, err := PropertyMapGet(pluginName, appName, property)
		if err != nil {
			return err
		}
		m[key] = value
		return p.MapWrite(property, m)
	})
}

// PropertyMapDelete reads the current map, removes one key, and writes the map
// back. Removing a key that is not present is a no-op.
func PropertyMapDelete(pluginName string, appName string, property string, key string) error {
	if !PropertyExists(pluginName, appName, property) {
		return nil
	}
	return PropertyLock(pluginName, appName, func(p *LockedProperties) error {
		m, err := PropertyMapGet(pluginName, appName, property)
		if err != nil {
			return err
		}
		if _, ok := m[key]; !ok {
			return nil
		}
		delete(m, key)
		return p.MapWrite(property, m)
	})
}

// PropertyMapLength returns the number of entries in the map persisted for a
// property. A missing property is reported as length 0 without error.
func PropertyMapLength(pluginName string, appName string, property string) (int, error) {
	m, err := PropertyMapGet(pluginName, appName, property)
	if err != nil {
		return 0, err
	}
	return len(m), nil
}

// PropertySetup creates the plugin config root
func PropertySetup(pluginName string) error {
	configRoot := filepath.Join(MustGetEnv("DOKKU_LIB_ROOT"), "config")
	pluginConfigRoot := getPluginConfigPath(pluginName)
	if err := os.MkdirAll(pluginConfigRoot, 0755); err != nil {
		return err
	}

	// check if configRoot is a symlink
	if !IsSymlink(configRoot) {
		input := SetPermissionInput{
			Filename: configRoot,
			Mode:     os.FileMode(0755),
		}

		if err := SetPermissions(input); err != nil {
			return err
		}
	}

	return SetPermissions(SetPermissionInput{
		Filename: pluginConfigRoot,
		Mode:     os.FileMode(0755),
	})
}

func PropertySetupApp(pluginName string, appName string) error {
	if err := PropertySetup(pluginName); err != nil {
		return err
	}

	if err := makePluginAppPropertyPath(pluginName, appName); err != nil {
		return fmt.Errorf("Unable to create %s config directory for %s: %s", pluginName, appName, err.Error())
	}

	return nil
}

// MigrateConfigEntry describes a single config-to-property migration
type MigrateConfigEntry struct {
	// ConfigVar is the per-app environment variable name (e.g. "DOKKU_APP_PROXY_TYPE")
	ConfigVar string

	// GlobalConfigVar is the global environment variable name (e.g. "DOKKU_PROXY_TYPE"), empty if none
	GlobalConfigVar string

	// Property is the target property name (e.g. "type")
	Property string

	// Transform is an optional value transformation function applied before writing
	Transform func(value string) string

	// ListProperty indicates the value should be written as a list property via PropertyListWrite
	ListProperty bool
}

// MigrateConfigToProperties migrates config variables to properties for a given plugin
// across all apps and optionally globally. It is idempotent: if the property already
// exists, the migration is skipped for that app/entry.
func MigrateConfigToProperties(pluginName string, entries []MigrateConfigEntry) error {
	migrateLegacyEnvFiles()

	apps, err := UnfilteredDokkuApps()
	if err != nil && !errors.Is(err, NoAppsExist) {
		return nil
	}

	for _, entry := range entries {
		if entry.GlobalConfigVar != "" {
			if err := migrateConfigEntry(pluginName, "--global", entry.GlobalConfigVar, entry); err != nil {
				return err
			}
		}

		for _, appName := range apps {
			if entry.ConfigVar == "" {
				continue
			}

			if err := migrateConfigEntry(pluginName, appName, entry.ConfigVar, entry); err != nil {
				return err
			}
		}
	}

	return nil
}

// migrateLegacyEnvFiles drains the pre-0.38 ENV files into the config property
// path before any config var is read.
//
// Install triggers fire in lexicographic order of the enabled-plugin directory
// names, so plugins sorting before "config" - apps, builder and checks among
// them - would otherwise read an environment the config plugin has not
// relocated yet, find it empty, and migrate nothing without reporting anything.
// A failure here is not fatal: the migration is retried on the next install, so
// an unavailable trigger degrades to the previous behavior rather than aborting
// every plugin's install.
//
// The trigger's stderr is streamed rather than captured. This is the first call
// to reach the migration on an upgrade, so anything it has to report about the
// files it found is only ever said once, and a captured warning would be
// discarded with the response.
func migrateLegacyEnvFiles() {
	if _, err := CallPlugnTrigger(PlugnTriggerInput{
		Trigger:      "config-migrate-env",
		StreamStderr: true,
	}); err != nil {
		LogWarn(fmt.Sprintf("Unable to migrate legacy env files: %s", err.Error()))
	}
}

// migrateConfigEntry migrates a single config variable to a property for a given app
func migrateConfigEntry(pluginName string, appName string, configVar string, entry MigrateConfigEntry) error {
	if entry.ListProperty {
		if exists, _ := PropertyListLength(pluginName, appName, entry.Property); exists > 0 {
			return nil
		}
	} else if PropertyExists(pluginName, appName, entry.Property) {
		return nil
	}

	triggerName := "config-get"
	triggerArgs := []string{appName, configVar}
	if appName == "--global" {
		triggerName = "config-get-global"
		triggerArgs = []string{configVar}
	}

	results, _ := CallPlugnTrigger(PlugnTriggerInput{
		Trigger: triggerName,
		Args:    triggerArgs,
	})
	value := results.StdoutContents()
	if value == "" {
		return nil
	}

	if entry.Transform != nil {
		value = entry.Transform(value)
	}

	if appName == "--global" {
		LogInfo1(fmt.Sprintf("Migrating deprecated global %s to %s %s property. Use 'dokku %s:set --global %s <value>' to manage this going forward.", configVar, pluginName, entry.Property, pluginName, entry.Property))
	} else {
		LogInfo1(fmt.Sprintf("Migrating deprecated %s to %s %s property for %s. Use 'dokku %s:set %s %s <value>' to manage this going forward.", configVar, pluginName, entry.Property, appName, pluginName, appName, entry.Property))
	}

	if entry.ListProperty {
		values := strings.Split(value, " ")
		if err := PropertyListWrite(pluginName, appName, entry.Property, values); err != nil {
			return err
		}
	} else {
		if err := PropertyWrite(pluginName, appName, entry.Property, value); err != nil {
			return err
		}
	}

	unsetTrigger := "config-unset"
	unsetArgs := []string{appName, configVar}
	if appName == "--global" {
		unsetArgs = []string{"--global", configVar}
	}

	_, err := CallPlugnTrigger(PlugnTriggerInput{
		Trigger: unsetTrigger,
		Args:    unsetArgs,
		Env: map[string]string{
			"DOKKU_QUIET_OUTPUT": "1",
		},
	})
	if err != nil {
		LogWarn(err.Error())
	}

	return nil
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
	return SetPermissions(SetPermissionInput{
		Filename: pluginAppConfigRoot,
		Mode:     os.FileMode(0755),
	})
}
