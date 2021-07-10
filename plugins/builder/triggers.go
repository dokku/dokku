package builder

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/otiai10/copy"
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

// TriggerCorePostExtract moves a configured build-dir to be in the app root dir
func TriggerCorePostExtract(appName string, sourceWorkDir string) error {
	buildDir := strings.Trim(reportComputedBuildDir(appName), "/")
	if buildDir == "" {
		return nil
	}

	newSourceWorkDir := filepath.Join(sourceWorkDir, buildDir)
	if !common.DirectoryExists(newSourceWorkDir) {
		return fmt.Errorf("Specified build-dir not found in sourcecode working directory: %v", buildDir)
	}

	tmpWorkDir, err := ioutil.TempDir(os.TempDir(), fmt.Sprintf("dokku-%s-%s", common.MustGetEnv("DOKKU_PID"), "CorePostExtract"))
	if err != nil {
		return fmt.Errorf("Unable to create temporary working directory: %v", err.Error())
	}

	if err := removeAllContents(tmpWorkDir); err != nil {
		return fmt.Errorf("Unable to clear out temporary working directory for rewrite: %v", err.Error())
	}

	if err := copy.Copy(newSourceWorkDir, tmpWorkDir); err != nil {
		return fmt.Errorf("Unable to move build-dir to temporary working directory: %v", err.Error())
	}

	if err := removeAllContents(sourceWorkDir); err != nil {
		return fmt.Errorf("Unable to clear out sourcecode working directory for rewrite: %v", err.Error())
	}

	if err := copy.Copy(tmpWorkDir, sourceWorkDir); err != nil {
		return fmt.Errorf("Unable to move build-dir to sourcecode working directory: %v", err.Error())
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
