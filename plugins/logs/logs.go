package logs

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/fastfishio/qson"
)

// VectorSink is a map of vector sink properties
type VectorSink map[string]interface{}

// MaxSize is the default max retention size for docker logs
const MaxSize = "10m"

// AppLabelAlias is the property key for the app label alias
const AppLabelAlias = "com.dokku.app-name"

var (
	// DefaultProperties is a map of all valid logs properties with corresponding default property values
	DefaultProperties = map[string]string{
		"app-label-alias": AppLabelAlias,
		"max-size":        MaxSize,
		"vector-sink":     "",
	}

	// GlobalProperties is a map of all valid global logs properties
	GlobalProperties = map[string]bool{
		"app-label-alias": true,
		"max-size":        true,
		"vector-image":    true,
		"vector-sink":     true,
	}
)

// VectorDockerfile is the contents of the default Dockerfile
// containing the version of vector Dokku uses
//
//go:embed Dockerfile
var VectorDockerfile string

// VectorDefaultSink contains the default sink in use for vector log shipping
const VectorDefaultSink = "blackhole://?print_interval_secs=1"

//go:embed templates/*
var templates embed.FS

// GetFailedLogs outputs failed deploy logs for a given app
func GetFailedLogs(appName string) error {
	common.LogInfo2Quiet(fmt.Sprintf("%s failed deploy logs", appName))
	scheduler := common.GetAppScheduler(appName)
	_, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "scheduler-logs-failed",
		Args:        []string{scheduler, appName},
		StreamStdio: true,
	})
	return err
}

// SinkValueToConfig converts a sink DSN value to a VectorSink
func SinkValueToConfig(appName string, sinkValue string) (VectorSink, error) {
	var data VectorSink
	if strings.Contains(sinkValue, "://") {
		parts := strings.SplitN(sinkValue, "://", 2)
		parts[0] = strings.ReplaceAll(parts[0], "_", "-")
		sinkValue = strings.Join(parts, "://")
	}
	u, err := url.Parse(sinkValue)
	if err != nil {
		return data, err
	}

	if u.Query().Get("sinks") != "" {
		return data, errors.New("Invalid option sinks")
	}

	u.Scheme = strings.ReplaceAll(u.Scheme, "-", "_")

	query := u.RawQuery
	query = strings.TrimPrefix(query, "&")

	b, err := qson.ToJSON(query)
	if err != nil {
		return data, err
	}

	if err := json.Unmarshal(b, &data); err != nil {
		return data, err
	}

	data["type"] = u.Scheme
	data["inputs"] = []string{"docker-source:" + appName}
	if appName == "--global" {
		data["inputs"] = []string{"docker-global-source"}
	}
	if appName == "--null" {
		data["inputs"] = []string{"docker-null-source"}
	}

	return data, nil
}
