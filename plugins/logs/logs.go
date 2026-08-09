package logs

import (
	"embed"
	"encoding/base64"
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

// ContainerTypeLabel is the docker label holding the type of a dokku container
const ContainerTypeLabel = "com.dokku.container-type"

// CronContainerType is the ContainerTypeLabel value used for cron task containers
const CronContainerType = "cron"

// CronIDLabel is the docker label holding the cron task id
const CronIDLabel = "com.dokku.cron-id"

// CronRouteName is the vector route output carrying cron task logs
const CronRouteName = "cron"

var (
	// DefaultProperties is a map of all valid logs properties with corresponding default property values
	DefaultProperties = map[string]string{
		"app-label-alias":  AppLabelAlias,
		"max-size":         MaxSize,
		"vector-cron-sink": "",
		"vector-sink":      "",
	}

	// GlobalProperties is a map of all valid global logs properties
	GlobalProperties = map[string]bool{
		"app-label-alias":  true,
		"max-size":         true,
		"vector-cron-sink": true,
		"vector-image":     true,
		"vector-networks":  true,
		"vector-sink":      true,
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

// SinkValueToConfigInput is the input for the SinkValueToConfig function
type SinkValueToConfigInput struct {
	// SinkValue is the sink DSN to convert
	SinkValue string

	// Inputs are the vector component ids feeding the sink. When empty, the
	// inputs key is omitted, which is appropriate for callers that only need
	// the parsed sink for validation or redaction.
	Inputs []string
}

// SinkValueToConfig converts a sink DSN value to a VectorSink
func SinkValueToConfig(input SinkValueToConfigInput) (VectorSink, error) {
	var data VectorSink
	sinkValue := input.SinkValue
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
	if len(input.Inputs) > 0 {
		data["inputs"] = input.Inputs
	}

	// add special support for `base64enc:VAL` fields
	for key, value := range data {
		valueString, ok := value.(string)
		if !ok {
			continue
		}

		if encodedValue, found := strings.CutPrefix(valueString, "base64enc:"); found {
			decodedValue, err := base64.StdEncoding.DecodeString(encodedValue)
			if err != nil {
				return data, fmt.Errorf("Error decoding base64: %w", err)
			}
			data[key] = strings.TrimSpace(string(decodedValue))
		}
	}

	return data, nil
}
