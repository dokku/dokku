package dockeroptions

import (
	"fmt"
	"io"
	"os"

	"github.com/dokku/dokku/plugins/common"
)

// TriggerPostAppCloneSetup copies docker option files from the source app to the cloned app
func TriggerPostAppCloneSetup(oldAppName string, newAppName string) error {
	for _, phase := range availablePhases {
		if err := copyPhaseFile(oldAppName, newAppName, phase); err != nil {
			return err
		}
	}
	return nil
}

// TriggerPostAppRenameSetup copies docker option files from the old app name to the new app name
func TriggerPostAppRenameSetup(oldAppName string, newAppName string) error {
	for _, phase := range availablePhases {
		if err := copyPhaseFile(oldAppName, newAppName, phase); err != nil {
			return err
		}
	}
	return nil
}

// TriggerPostDelete deletes the docker option files for an app
func TriggerPostDelete(appName string) error {
	for _, phase := range availablePhases {
		if err := removePhaseFile(appName, phase); err != nil {
			return err
		}
	}
	return nil
}

func copyPhaseFile(srcApp string, dstApp string, phase string) error {
	srcPath := getPhaseFilePath(srcApp, phase)
	if !common.FileExists(srcPath) {
		return nil
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("Unable to open docker options phase file %s.%s: %s", srcApp, phase, err.Error())
	}
	defer src.Close()

	dstPath := getPhaseFilePath(dstApp, phase)
	dst, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("Unable to create docker options phase file %s.%s: %s", dstApp, phase, err.Error())
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("Unable to copy docker options phase file %s.%s to %s.%s: %s", srcApp, phase, dstApp, phase, err.Error())
	}

	return nil
}
