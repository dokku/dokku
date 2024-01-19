package appjson

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

var (
	// DefaultProperties is a map of all valid app-json properties with corresponding default property values
	DefaultProperties = map[string]string{
		"appjson-path": "",
	}

	// GlobalProperties is a map of all valid global app-json properties
	GlobalProperties = map[string]bool{
		"appjson-path": true,
	}
)

// AppJSON is a struct that represents an app.json file as understood by Dokku
type AppJSON struct {
	Cron         []CronCommand            `json:"cron"`
	Formation    map[string]Formation     `json:"formation"`
	Healthchecks map[string][]Healthcheck `json:"healthchecks"`
	Scripts      struct {
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

type Healthcheck struct {
	Attempts     int32           `json:"attempts,omitempty"`
	Command      []string        `json:"command,omitempty"`
	Content      string          `json:"content,omitempty"`
	HTTPHeaders  []HTTPHeader    `json:"httpHeaders,omitempty"`
	InitialDelay int32           `json:"initialDelay,omitempty"`
	Listening    bool            `json:"listening,omitempty"`
	Name         string          `json:"name,omitempty"`
	Path         string          `json:"path,omitempty"`
	Port         int             `json:"port,omitempty"`
	Scheme       string          `json:"scheme,omitempty"`
	Timeout      int32           `json:"timeout,omitempty"`
	Type         HealthcheckType `json:"type,omitempty"`
	Uptime       int32           `json:"uptime,omitempty"`
	Wait         int32           `json:"wait,omitempty"`
	Warn         bool            `json:"warn,omitempty"`
	OnFailure    *OnFailure      `json:"onFailure,omitempty"`
}

type HealthcheckType string

const (
	HealthcheckType_Liveness  HealthcheckType = "liveness"
	HealthcheckType_Readiness HealthcheckType = "readiness"
	HealthcheckType_Startup   HealthcheckType = "startup"
)

type HTTPHeader struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

type OnFailure struct {
	Command []string `json:"command,omitempty"`
	Url     string   `json:"url,omitempty"`
}

// GetAppjsonDirectory returns the directory containing a given app's extracted app.json file
func GetAppjsonDirectory(appName string) string {
	return common.GetAppDataDirectory("app-json", appName)
}

// GetAppjsonPath returns the path to a given app's extracted app.json file for use by other plugins
func GetAppjsonPath(appName string) string {
	return getProcessSpecificAppJSONPath(appName)
}

// GetAppJSON returns the parsed app.json file for a given app
func GetAppJSON(appName string) (AppJSON, error) {
	if !hasAppJSON(appName) {
		return AppJSON{}, nil
	}

	b, err := os.ReadFile(getProcessSpecificAppJSONPath(appName))
	if err != nil {
		return AppJSON{}, fmt.Errorf("Cannot read app.json file: %v", err)
	}

	if strings.TrimSpace(string(b)) == "" {
		return AppJSON{}, nil
	}

	var appJSON AppJSON
	if err = json.Unmarshal(b, &appJSON); err != nil {
		return AppJSON{}, fmt.Errorf("Cannot parse app.json: %v", err)
	}

	return appJSON, nil
}
