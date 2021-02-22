package appjson

import (
	"path/filepath"

	"github.com/dokku/dokku/plugins/common"
)

// AppJSON is a struct that represents an app.json file as understood by Dokku
type AppJSON struct {
	Cron    []CronCommand `json:"cron"`
	Scripts struct {
		Dokku struct {
			Predeploy  string `json:"predeploy"`
			Postdeploy string `json:"postdeploy"`
		} `json:"dokku"`
		Postdeploy string `json:"postdeploy"`
	} `json:"scripts"`
}

// CronCommand is a struct that represents a single cron command from an app.json file
type CronCommand struct {
	Command  string `json:"command"`
	Schedule string `json:"schedule"`
}

// GetAppjsonDirectory returns the directory containing a given app's extracted app.json file
func GetAppjsonDirectory(appName string) string {
	return filepath.Join(common.MustGetEnv("DOKKU_LIB_ROOT"), "data", "app-json", appName)
}

// GetAppjsonPath returns the path to a given app's extracted app.json file
func GetAppjsonPath(appName string) string {
	return filepath.Join(GetAppjsonDirectory(appName), "app.json")
}
