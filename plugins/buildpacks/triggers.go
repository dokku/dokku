package buildpacks

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// TriggerBuildpackStackName echos the stack name for the app
func TriggerBuildpackStackName(appName string) error {
	if stack := common.PropertyGetDefault("buildpacks", appName, "stack", ""); stack != "" {
		fmt.Println(stack)
		return nil
	}

	if stack := common.PropertyGetDefault("buildpacks", "--global", "stack", ""); stack != "" {
		fmt.Println(stack)
		return nil
	}

	b, _ := common.PlugnTriggerOutput("config-get", []string{appName, "DOKKU_IMAGE"}...)
	dokkuImage := strings.TrimSpace(string(b[:]))
	if dokkuImage != "" {
		common.LogWarn("Deprecated: use buildpacks:stack-set instead of specifying DOKKU_IMAGE environment variable")
		fmt.Println(dokkuImage)
		return nil
	}

	return nil
}

// TriggerInstall runs the install step for the buildpacks plugin
func TriggerInstall() error {
	if err := common.PropertySetup("buildpacks"); err != nil {
		return fmt.Errorf("Unable to install the buildpacks plugin: %s", err.Error())
	}

	return nil
}

// TriggerPostAppCloneSetup creates new buildpacks files
func TriggerPostAppCloneSetup(oldAppName string, newAppName string) error {
	err := common.PropertyClone("buildpacks", oldAppName, newAppName)
	if err != nil {
		return err
	}

	return nil
}

// TriggerPostAppRenameSetup renames buildpacks files
func TriggerPostAppRenameSetup(oldAppName string, newAppName string) error {
	if err := common.PropertyClone("buildpacks", oldAppName, newAppName); err != nil {
		return err
	}

	if err := common.PropertyDestroy("buildpacks", oldAppName); err != nil {
		return err
	}

	return nil
}

// TriggerPostDelete destroys the buildpacks property for a given app container
func TriggerPostDelete(appName string) error {
	return common.PropertyDestroy("buildpacks", appName)
}

// TriggerPostExtract writes a .buildpacks file into the app
func TriggerPostExtract(appName string, sourceWorkDir string) error {
	buildpacks, err := common.PropertyListGet("buildpacks", appName, "buildpacks")
	if err != nil {
		return nil
	}

	if len(buildpacks) == 0 {
		return nil
	}

	buildpacksPath := filepath.Join(sourceWorkDir, ".buildpacks")
	file, err := os.OpenFile(buildpacksPath, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("Error writing .buildpacks file: %s", err.Error())
	}

	w := bufio.NewWriter(file)
	for _, buildpack := range buildpacks {
		fmt.Fprintln(w, buildpack)
	}

	if err = w.Flush(); err != nil {
		return fmt.Errorf("Error writing .buildpacks file: %s", err.Error())
	}
	file.Chmod(0600)

	return nil
}
