package appjson

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"k8s.io/utils/ptr"
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
	// Cron is a list of cron commands to execute
	Cron []CronCommand `json:"cron"`

	// Formation is a map of process types to scale
	Formation map[string]Formation `json:"formation"`

	// Healthchecks is a map of process types to healthchecks
	Healthchecks map[string][]Healthcheck `json:"healthchecks"`

	// Scripts is a map of scripts to execute
	Scripts struct {
		// Dokku is a map of scripts to execute for Dokku-specific events
		Dokku struct {
			// Predeploy is a script to execute before a deploy
			Predeploy string `json:"predeploy"`

			// Postdeploy is a script to execute after a deploy
			Postdeploy string `json:"postdeploy"`
		} `json:"dokku"`

		// Postdeploy is a script to execute after a deploy
		Postdeploy string `json:"postdeploy"`
	} `json:"scripts"`
}

// CronCommand is a struct that represents a single cron command from an app.json file
type CronCommand struct {
	// Command is the command to execute
	Command string `json:"command"`

	// Schedule is the cron schedule to execute the command on
	Schedule string `json:"schedule"`
}

// Formation is a struct that represents the scale for a process from an app.json file
type Formation struct {
	// Autoscaling is whether or not to enable autoscaling
	Autoscaling *FormationAutoscaling `json:"autoscaling"`

	// Quantity is the number of processes to run
	Quantity *int `json:"quantity"`

	// MaxParallel is the maximum number of processes to start in parallel
	MaxParallel *int `json:"max_parallel"`
}

// FormationAutoscaling is a struct that represents the autoscaling configuration for a process from an app.json file
type FormationAutoscaling struct {
	// CoolDownSeconds is the number of seconds to wait before scaling again
	CooldownPeriodSeconds *int `json:"cooldown_period_seconds,omitempty"`

	// MaxQuantity is the maximum number of processes to run
	MaxQuantity *int `json:"max_quantity,omitempty"`

	// MinQuantity is the minimum number of processes to run
	MinQuantity *int `json:"min_quantity,omitempty"`

	// PollingIntervalSeconds is the number of seconds to wait between autoscaling checks
	PollingIntervalSeconds *int `json:"polling_interval_seconds,omitempty"`

	// Triggers is a list of triggers to use for autoscaling
	Triggers []FormationAutoscalingTrigger `json:"triggers,omitempty"`
}

// FormationAutoscalingTrigger is a struct that represents a single autoscaling trigger from an app.json file
type FormationAutoscalingTrigger struct {
	// Name is the name of the trigger
	Name string `json:"name,omitempty"`

	// Type is the type of the trigger
	Type string `json:"type,omitempty"`

	// Metadata is a map of metadata to use for the trigger
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Healthcheck is a struct that represents a single healthcheck from an app.json file
type Healthcheck struct {
	// Attempts is the number of attempts to make before considering a healthcheck failed
	Attempts int32 `json:"attempts,omitempty"`

	// Command is the command to execute for the healthcheck
	Command []string `json:"command,omitempty"`

	// Content is the content to check for in the healthcheck response
	Content string `json:"content,omitempty"`

	// HTTPHeaders is a list of HTTP headers to send with the healthcheck request
	HTTPHeaders []HTTPHeader `json:"httpHeaders,omitempty"`

	// InitialDelay is the number of seconds to wait before starting healthchecks
	InitialDelay int32 `json:"initialDelay,omitempty"`

	// Listening is whether or not this is a listening check
	Listening bool `json:"listening,omitempty"`

	// Name is the name of the healthcheck
	Name string `json:"name,omitempty"`

	// Path is the path to check for in the healthcheck response
	Path string `json:"path,omitempty"`

	// Port is the port to check for in the healthcheck response
	Port int `json:"port,omitempty"`

	// Scheme is the scheme to use for the healthcheck request
	Scheme string `json:"scheme,omitempty"`

	// Timeout is the number of seconds to wait before considering a healthcheck failed
	Timeout int32 `json:"timeout,omitempty"`

	// Type is the type of healthcheck
	Type HealthcheckType `json:"type,omitempty"`

	// Uptime is the number of seconds to wait before considering a container running
	Uptime int32 `json:"uptime,omitempty"`

	// Wait is the number of seconds to wait between healthchecks
	Wait int32 `json:"wait,omitempty"`

	// Warn is whether or not to warn on a failed healthcheck instead of error out
	Warn bool `json:"warn,omitempty"`

	// OnFailure is the action to take on a failed healthcheck
	OnFailure *OnFailure `json:"onFailure,omitempty"`
}

// HealthcheckType is a string that represents the type of a healthcheck from an app.json file
type HealthcheckType string

const (
	// HealthcheckType_Liveness is a healthcheck type that represents a liveness check
	HealthcheckType_Liveness HealthcheckType = "liveness"

	// HealthcheckType_Readiness is a healthcheck type that represents a readiness check
	HealthcheckType_Readiness HealthcheckType = "readiness"

	// HealthcheckType_Startup is a healthcheck type that represents a startup check
	HealthcheckType_Startup HealthcheckType = "startup"
)

// HTTPHeader is a struct that represents a single HTTP header associated with a healthcheck
type HTTPHeader struct {
	// Name is the name of the HTTP header
	Name string `json:"name,omitempty"`

	// Value is the value of the HTTP header
	Value string `json:"value,omitempty"`
}

// OnFailure is a struct that represents the on failure action for a healthcheck
type OnFailure struct {
	// Command is the command to execute on failure
	Command []string `json:"command,omitempty"`

	// Url is the URL to call on failure
	Url string `json:"url,omitempty"`
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

func GetAutoscalingConfig(appName string, processType string, replicas int) (FormationAutoscaling, bool, error) {
	appJSON, err := GetAppJSON(appName)
	if err != nil {
		return FormationAutoscaling{}, false, err
	}

	common.LogWarn(fmt.Sprintf("appJSON: %v", appJSON))

	formation, ok := appJSON.Formation[processType]
	if !ok {
		return FormationAutoscaling{}, false, nil
	}

	if formation.Autoscaling == nil {
		return FormationAutoscaling{}, false, nil
	}

	autoscaling := *formation.Autoscaling
	if autoscaling.CooldownPeriodSeconds == nil {
		autoscaling.CooldownPeriodSeconds = ptr.To(300)
	}

	if autoscaling.MinQuantity == nil {
		autoscaling.MinQuantity = ptr.To(replicas)
	}

	if autoscaling.MaxQuantity == nil {
		defaultValue := math.Max(float64(replicas), float64(*autoscaling.MinQuantity))
		autoscaling.MaxQuantity = ptr.To(int(defaultValue))
	}

	if autoscaling.PollingIntervalSeconds == nil {
		autoscaling.PollingIntervalSeconds = ptr.To(30)
	}

	if len(autoscaling.Triggers) == 0 {
		return FormationAutoscaling{}, false, nil
	}

	return autoscaling, true, nil
}
