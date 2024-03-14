package logs

import (
	"embed"
	"fmt"

	"github.com/dokku/dokku/plugins/common"
)

// MaxSize is the default max retention size for docker logs
const MaxSize = "10m"

var (
	// DefaultProperties is a map of all valid logs properties with corresponding default property values
	DefaultProperties = map[string]string{
		"max-size":    "",
		"vector-sink": "",
	}

	// GlobalProperties is a map of all valid global logs properties
	GlobalProperties = map[string]bool{
		"max-size":     true,
		"vector-image": true,
		"vector-sink":  true,
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
