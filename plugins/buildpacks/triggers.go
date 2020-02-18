package buildpacks

import (
	"bufio"
	"fmt"
	"os"
	"path"

	"github.com/dokku/dokku/plugins/common"
)

// TriggerInstall runs the install step for the buildpacks plugin
func TriggerInstall() {
	if err := common.PropertySetup("buildpacks"); err != nil {
		common.LogFail(fmt.Sprintf("Unable to install the buildpacks plugin: %s", err.Error()))
	}
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
func TriggerPostDelete(appName string) {
	err := common.PropertyDestroy("buildpacks", appName)
	if err != nil {
		common.LogFail(err.Error())
	}
}

// TriggerPostExtract writes a .buildpacks file into the app
func TriggerPostExtract(appName string, sourceWorkDir string) {
	buildpacks, err := common.PropertyListGet("buildpacks", appName, "buildpacks")
	if err != nil {
		return
	}

	if len(buildpacks) == 0 {
		return
	}

	buildpacksPath := path.Join(sourceWorkDir, ".buildpacks")
	file, err := os.OpenFile(buildpacksPath, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		common.LogFail(fmt.Sprintf("Error writing .buildpacks file: %s", err.Error()))
		return
	}

	w := bufio.NewWriter(file)
	for _, buildpack := range buildpacks {
		fmt.Fprintln(w, buildpack)
	}

	if err = w.Flush(); err != nil {
		common.LogFail(fmt.Sprintf("Error writing .buildpacks file: %s", err.Error()))
		return
	}
	file.Chmod(0600)
}
