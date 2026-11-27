// Package envgroup implements the dokku core envgroup plugin. An env group is
// a named set of environment variables that is shared by the apps added to
// it. An app's merged environment layers its groups, in order, between the
// global environment and the app's own environment.
package envgroup

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
)

const (
	pluginName     = "envgroup"
	groupsProperty = "groups"
)

// ErrKeyNotSet is returned when a requested key is not set in a group
var ErrKeyNotSet = errors.New("key is not set")

// groupDirectory returns the directory holding a group's environment
func groupDirectory(groupName string) string {
	return filepath.Join(common.GetDataDirectory(pluginName), groupName)
}

// groupFile returns the path to the envfile holding a group's environment
func groupFile(groupName string) string {
	return filepath.Join(groupDirectory(groupName), "ENV")
}

// validateGroupName checks that a group name follows the app naming rules
func validateGroupName(groupName string) error {
	if groupName == "" {
		return errors.New("Please specify an env group name")
	}

	if err := common.IsValidAppName(groupName); err != nil {
		return errors.New("Env group name must begin with lowercase alphanumeric character, and may only contain lowercase alphanumerics, dots, and hyphens")
	}

	return nil
}

// GroupExists returns whether the named group has been created
func GroupExists(groupName string) bool {
	if validateGroupName(groupName) != nil {
		return false
	}

	return common.DirectoryExists(groupDirectory(groupName))
}

// verifyGroup checks that the named group is valid and has been created
func verifyGroup(groupName string) error {
	if err := validateGroupName(groupName); err != nil {
		return err
	}

	if !GroupExists(groupName) {
		return fmt.Errorf("Env group %s does not exist", groupName)
	}

	return nil
}

// createGroup creates the directory holding a group's environment
func createGroup(groupName string) error {
	if err := common.CreateDataDirectory(pluginName); err != nil {
		return err
	}

	directory := groupDirectory(groupName)
	if err := os.MkdirAll(directory, 0755); err != nil {
		return err
	}

	return common.SetPermissions(common.SetPermissionInput{
		Filename: directory,
		Mode:     os.FileMode(0755),
	})
}

// ListGroups returns the names of every group, sorted
func ListGroups() ([]string, error) {
	groups := []string{}
	directory := common.GetDataDirectory(pluginName)
	if !common.DirectoryExists(directory) {
		return groups, nil
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		return groups, err
	}

	for _, entry := range entries {
		if !entry.IsDir() || validateGroupName(entry.Name()) != nil {
			continue
		}
		groups = append(groups, entry.Name())
	}

	sort.Strings(groups)
	return groups, nil
}

// LoadGroupEnv loads a group's environment
func LoadGroupEnv(groupName string) (*config.Env, error) {
	return config.LoadEnvFile(groupName, groupFile(groupName))
}

// AppGroups returns the groups an app belongs to, lowest precedence first
func AppGroups(appName string) ([]string, error) {
	groups, err := common.PropertyListGet(pluginName, appName, groupsProperty)
	if err != nil {
		return []string{}, err
	}

	if groups == nil {
		groups = []string{}
	}

	return groups, nil
}

// GroupMembers returns the apps that belong to a group, sorted by app name
func GroupMembers(groupName string) ([]string, error) {
	members := []string{}
	apps, err := common.UnfilteredDokkuApps()
	if err != nil {
		if errors.Is(err, common.NoAppsExist) {
			return members, nil
		}
		return members, err
	}

	for _, appName := range apps {
		groups, err := AppGroups(appName)
		if err != nil {
			return members, err
		}

		if containsString(groups, groupName) {
			members = append(members, appName)
		}
	}

	sort.Strings(members)
	return members, nil
}

// MergedAppGroupEnv merges the environments of an app's groups in order, so
// that a group added later overrides a group added earlier
func MergedAppGroupEnv(appName string) (*config.Env, error) {
	env := config.NewEnvFromMap(appName, map[string]string{})
	groups, err := AppGroups(appName)
	if err != nil {
		return env, err
	}

	for _, groupName := range groups {
		if !GroupExists(groupName) {
			continue
		}

		groupEnv, err := LoadGroupEnv(groupName)
		if err != nil {
			return env, err
		}
		env.Merge(groupEnv)
	}

	return env, nil
}

// insertAtIndex inserts value into lines at a 1-based index, appending it when
// index is 0 or past the end of the list
func insertAtIndex(lines []string, value string, index int) []string {
	if index <= 0 || index > len(lines) {
		return append(lines, value)
	}

	result := make([]string, 0, len(lines)+1)
	result = append(result, lines[:index-1]...)
	result = append(result, value)
	result = append(result, lines[index-1:]...)
	return result
}

// addAppToGroup adds a group to an app's group list at a 1-based index
func addAppToGroup(groupName string, appName string, index int) error {
	return common.PropertyListUpdate(pluginName, appName, groupsProperty, func(lines []string) ([]string, error) {
		if containsString(lines, groupName) {
			return lines, fmt.Errorf("App %s is already a member of env group %s", appName, groupName)
		}

		return insertAtIndex(lines, groupName, index), nil
	})
}

// removeAppFromGroup removes a group from an app's group list
func removeAppFromGroup(groupName string, appName string) error {
	return common.PropertyListUpdate(pluginName, appName, groupsProperty, func(lines []string) ([]string, error) {
		if !containsString(lines, groupName) {
			return lines, fmt.Errorf("App %s is not a member of env group %s", appName, groupName)
		}

		result := []string{}
		for _, line := range lines {
			if line != groupName {
				result = append(result, line)
			}
		}
		return result, nil
	})
}

// groupHasKeys returns whether a group's environment has any keys set
func groupHasKeys(groupName string) bool {
	if !GroupExists(groupName) {
		return false
	}

	env, err := LoadGroupEnv(groupName)
	if err != nil {
		return false
	}

	return env.Len() > 0
}

// handleAffectedApps redeploys the deployed apps among appNames when restart
// is true, and otherwise lists the deployed apps that need a redeploy for an
// environment change to take effect
func handleAffectedApps(appNames []string, restart bool) {
	deployed := []string{}
	for _, appName := range appNames {
		if common.IsDeployed(appName) {
			deployed = append(deployed, appName)
		}
	}

	if len(deployed) == 0 {
		return
	}

	if restart {
		for _, appName := range deployed {
			config.RestartApp(appName)
		}
		return
	}

	common.LogWarn(fmt.Sprintf("The following apps must be redeployed for the change to take effect: %s", strings.Join(deployed, ", ")))
	common.LogWarn("Pass --restart to redeploy them, or run ps:rebuild on each app")
}

// containsString returns whether value is in values
func containsString(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}
