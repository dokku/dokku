package envgroup

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
)

// CommandAdd adds an app to a group at an optional 1-based index
func CommandAdd(groupName string, appName string, index int, restart bool) error {
	if err := verifyGroup(groupName); err != nil {
		return err
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if index < 0 {
		return errors.New("The index must be a positive number")
	}

	if err := addAppToGroup(groupName, appName, index); err != nil {
		return err
	}

	common.LogInfo1Quiet(fmt.Sprintf("Added app %s to env group %s", appName, groupName))
	if groupHasKeys(groupName) {
		handleAffectedApps([]string{appName}, restart)
	}
	return nil
}

// CommandBundle writes a tarball of a group's environment to stdout
func CommandBundle(groupName string) error {
	env, err := loadVerifiedGroupEnv(groupName)
	if err != nil {
		return err
	}

	return env.ExportBundle(os.Stdout)
}

// CommandClear unsets every key in a group
func CommandClear(groupName string, restart bool) error {
	if err := verifyGroup(groupName); err != nil {
		return err
	}

	changed, err := config.ClearEnvFile(groupName, groupFile(groupName))
	if err != nil {
		return err
	}

	return afterGroupChange(groupName, changed, restart)
}

// CommandCreate creates a group
func CommandCreate(groupName string) error {
	if err := validateGroupName(groupName); err != nil {
		return err
	}

	if GroupExists(groupName) {
		return fmt.Errorf("Env group %s already exists", groupName)
	}

	if err := createGroup(groupName); err != nil {
		return fmt.Errorf("Unable to create env group %s: %s", groupName, err.Error())
	}

	common.LogInfo1Quiet(fmt.Sprintf("Created env group %s", groupName))
	return nil
}

// CommandDestroy destroys a group. A group that apps still belong to is only
// destroyed when force is true, in which case it is removed from those apps.
func CommandDestroy(groupName string, force bool, restart bool) error {
	if err := verifyGroup(groupName); err != nil {
		return err
	}

	members, err := GroupMembers(groupName)
	if err != nil {
		return err
	}

	if len(members) > 0 && !force {
		return fmt.Errorf("Env group %s is in use by the following apps: %s. Pass --force to remove it from them and destroy it", groupName, strings.Join(members, ", "))
	}

	hasKeys := groupHasKeys(groupName)
	for _, appName := range members {
		if err := removeAppFromGroup(groupName, appName); err != nil {
			return err
		}
		common.LogInfo1Quiet(fmt.Sprintf("Removed app %s from env group %s", appName, groupName))
	}

	if err := os.RemoveAll(groupDirectory(groupName)); err != nil {
		return fmt.Errorf("Unable to destroy env group %s: %s", groupName, err.Error())
	}

	common.LogInfo1Quiet(fmt.Sprintf("Destroyed env group %s", groupName))
	if hasKeys {
		handleAffectedApps(members, restart)
	}
	return nil
}

// CommandExists exits non-zero when a group does not exist
func CommandExists(groupName string) error {
	return verifyGroup(groupName)
}

// CommandExport outputs a group's environment in the given format
func CommandExport(groupName string, format string) error {
	env, err := loadVerifiedGroupEnv(groupName)
	if err != nil {
		return err
	}

	exportType, err := config.ParseExportFormat(format)
	if err != nil {
		return err
	}

	suffix := "\n"
	if exportType == config.ExportFormatShell {
		suffix = " "
	}

	fmt.Print(env.Export(exportType) + suffix)
	return nil
}

// CommandGet outputs the value of a key in a group
func CommandGet(groupName string, keys []string, quoted bool) error {
	if len(keys) == 0 {
		return errors.New("Expected: key")
	}

	if len(keys) != 1 {
		return fmt.Errorf("Unexpected argument(s): %v", keys[1:])
	}

	env, err := loadVerifiedGroupEnv(groupName)
	if err != nil {
		return err
	}

	if err := config.ValidateKey(keys[0]); err != nil {
		return err
	}

	value, ok := env.Get(keys[0])
	if !ok {
		return ErrKeyNotSet
	}

	if quoted {
		fmt.Printf("'%s'\n", config.SingleQuoteEscape(value))
	} else {
		fmt.Printf("%s\n", value)
	}

	return nil
}

// CommandImport sets the keys in a group from an envfile or json file, or
// from stdin when filename is "-"
func CommandImport(groupName string, filename string, format string, replace bool, restart bool) error {
	if err := verifyGroup(groupName); err != nil {
		return err
	}

	if filename == "-" {
		common.LogInfo1Quiet("Reading config vars from stdin")
	}

	env, err := config.LoadImportEnv(groupName, filename, format)
	if err != nil {
		return err
	}

	if env.Len() == 0 {
		return errors.New("No config vars to set")
	}

	changed, err := config.SetEnvFileEntries(groupName, groupFile(groupName), env.Map(), replace)
	if err != nil {
		return err
	}

	return afterGroupChange(groupName, changed, restart)
}

// CommandInfo outputs the apps that belong to a group and its number of keys
func CommandInfo(groupName string) error {
	env, err := loadVerifiedGroupEnv(groupName)
	if err != nil {
		return err
	}

	members, err := GroupMembers(groupName)
	if err != nil {
		return err
	}

	common.LogInfo2Quiet(fmt.Sprintf("%s env group information", groupName))
	common.LogVerbose(fmt.Sprintf("Env group apps:  %s", strings.Join(members, " ")))
	common.LogVerbose(fmt.Sprintf("Env group keys:  %d", env.Len()))
	return nil
}

// CommandKeys outputs the keys set in a group
func CommandKeys(groupName string) error {
	env, err := loadVerifiedGroupEnv(groupName)
	if err != nil {
		return err
	}

	for _, k := range env.Keys() {
		fmt.Println(k)
	}
	return nil
}

// CommandList outputs every group
func CommandList() error {
	groups, err := ListGroups()
	if err != nil {
		return err
	}

	common.LogInfo2Quiet("Env groups")
	for _, groupName := range groups {
		fmt.Println(groupName)
	}
	return nil
}

// CommandRemove removes an app from a group
func CommandRemove(groupName string, appName string, restart bool) error {
	if err := validateGroupName(groupName); err != nil {
		return err
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if err := removeAppFromGroup(groupName, appName); err != nil {
		return err
	}

	common.LogInfo1Quiet(fmt.Sprintf("Removed app %s from env group %s", appName, groupName))
	if groupHasKeys(groupName) {
		handleAffectedApps([]string{appName}, restart)
	}
	return nil
}

// CommandReport displays an envgroup report for one or more apps
func CommandReport(appName string, format string, infoFlag string) error {
	if len(appName) == 0 {
		apps, err := common.DokkuApps()
		if err != nil {
			if errors.Is(err, common.NoAppsExist) {
				common.LogWarn(err.Error())
				return nil
			}
			return err
		}
		for _, appName := range apps {
			if err := ReportSingleApp(appName, format, infoFlag); err != nil {
				return err
			}
		}
		return nil
	}

	return ReportSingleApp(appName, format, infoFlag)
}

// CommandSet sets one or more keys in a group
func CommandSet(groupName string, pairs []string, encoded bool, restart bool) error {
	if err := verifyGroup(groupName); err != nil {
		return err
	}

	if len(pairs) == 0 {
		return errors.New("At least one env pair must be given")
	}

	entries, err := config.ParseEnvPairs(pairs, encoded)
	if err != nil {
		return err
	}

	changed, err := config.SetEnvFileEntries(groupName, groupFile(groupName), entries, false)
	if err != nil {
		return err
	}

	return afterGroupChange(groupName, changed, restart)
}

// CommandShow pretty-prints a group's environment
func CommandShow(groupName string) error {
	env, err := loadVerifiedGroupEnv(groupName)
	if err != nil {
		return err
	}

	common.LogInfo2Quiet(groupName + " env vars")
	fmt.Println(env.Export(config.ExportFormatPretty))
	return nil
}

// CommandUnset unsets one or more keys in a group
func CommandUnset(groupName string, keys []string, restart bool) error {
	if err := verifyGroup(groupName); err != nil {
		return err
	}

	if len(keys) == 0 {
		return errors.New("At least one key must be given")
	}

	changed, err := config.UnsetEnvFileKeys(groupName, groupFile(groupName), keys)
	if err != nil {
		return err
	}

	return afterGroupChange(groupName, changed, restart)
}

// afterGroupChange handles the apps that belong to a group once its
// environment has been changed
func afterGroupChange(groupName string, changed bool, restart bool) error {
	if !changed {
		return nil
	}

	members, err := GroupMembers(groupName)
	if err != nil {
		return err
	}

	handleAffectedApps(members, restart)
	return nil
}

// loadVerifiedGroupEnv loads a group's environment after checking that the group exists
func loadVerifiedGroupEnv(groupName string) (*config.Env, error) {
	if err := verifyGroup(groupName); err != nil {
		return nil, err
	}

	return LoadGroupEnv(groupName)
}
