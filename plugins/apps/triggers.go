package apps

import (
	"fmt"
	"os"

	"github.com/dokku/dokku/plugins/common"
)

// TriggerAppCreate is a trigger to create an app
func TriggerAppCreate(appName string) error {
	return createApp(appName)
}

// TriggerAppDestroy is a trigger to destroy an app
func TriggerAppDestroy(appName string) error {
	return destroyApp(appName)
}

// TriggerAppExists is a trigger to check if an app exists
func TriggerAppExists(appName string) error {
	return appExists(appName)
}

// TriggerAppMaybeCreate is a trigger to allow gated app creation
func TriggerAppMaybeCreate(appName string) error {
	return maybeCreateApp(appName)
}

// TriggerPostDelete is the apps post-delete plugin trigger
func TriggerPostDelete(appName string) error {
	imageRepo := common.GetAppImageRepo(appName)

	// remove contents for apps that are symlinks to other folders
	if err := os.RemoveAll(fmt.Sprintf("%v/", common.AppRoot(appName))); err != nil {
		common.LogWarn(err.Error())
	}

	// then remove the folder and/or the symlink
	if err := os.RemoveAll(common.AppRoot(appName)); err != nil {
		common.LogWarn(err.Error())
	}

	imagesByAppLabel, err := listImagesByAppLabel(appName)
	if err != nil {
		common.LogWarn(err.Error())
	}

	imagesByRepo, err := listImagesByImageRepo(imageRepo)
	if err != nil {
		common.LogWarn(err.Error())
	}

	images := append(imagesByAppLabel, imagesByRepo...)
	common.RemoveImages(images)

	return nil
}
