package appjson

import (
	"path/filepath"

	"github.com/dokku/dokku/plugins/common"
)

var (
	// DefaultProperties is a map of all valid network properties with corresponding default property values
	DefaultProperties = map[string]string{
		"appjson-path": "",
	}

	// GlobalProperties is a map of all valid global network properties
	GlobalProperties = map[string]bool{
		"appjson-path": true,
	}
)

// AppJSON is a struct that represents an app.json file as understood by Dokku
type AppJSON struct {
	Cron      []CronCommand        `json:"cron"`
	Formation map[string]Formation `json:"formation"`
	Scripts   struct {
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

// Formation is a struct that represents the scale for a process from an app.json file
type Formation struct {
	Quantity    *int `json:"quantity"`
	MaxParallel *int `json:"max_parallel"`
}

// GetAppjsonDirectory returns the directory containing a given app's extracted app.json file
func GetAppjsonDirectory(appName string) string {
	return common.GetAppDataDirectory("app-json", appName)
}

// GetAppjsonPath returns the path to a given app's extracted app.json file
func GetAppjsonPath(appName string) string {
	return filepath.Join(GetAppjsonDirectory(appName), "app.json")
}
