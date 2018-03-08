package repo

import (
	"fmt"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

func TriggerInstall() error {
	if err := common.PropertySetup("repo"); err != nil {
		return fmt.Errorf("Unable to install the repo plugin: %s", err.Error())
	}

	return nil
}

func TriggerPostDelete(appName string) error {
	return common.PropertyDestroy("repo", appName)
}

func TriggerPostExtract(appName string, tmpWorkDir string) error {
	appRoot := strings.Join([]string{common.MustGetEnv("DOKKU_ROOT"), appName}, "/")

	hostCopyFolder := common.PropertyGet("repo", appName, "host-copy-folder")
	containerCopyFolder := common.PropertyGet("repo", appName, "container-copy-folder")

	if err := copyDirectory(tmpWorkDir, hostCopyFolder, appRoot); err != nil {
		return err
	}

	if err := copyDirectory(tmpWorkDir, containerCopyFolder, tmpWorkDir); err != nil {
		return err
	}

	return nil
}

func TriggerPreDelete(appName string) error {
	return PurgeCache(appName)
}
