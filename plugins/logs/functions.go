package logs

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/dokku/dokku/plugins/common"
	"github.com/joncalhoun/qson"
)

type vectorConfig struct {
	Sources map[string]vectorSource `json:"sources"`
	Sinks   map[string]vectorSink   `json:"sinks"`
}

type vectorSource struct {
	Type          string   `json:"type"`
	IncludeLabels []string `json:"include_labels,omitempty"`
}

type vectorSink map[string]interface{}

const vectorContainerName = "vector"

func killVectorContainer() error {
	if !common.ContainerExists(vectorContainerName) {
		return nil
	}

	if err := stopVectorContainer(); err != nil {
		return err
	}

	time.Sleep(10 * time.Second)
	if err := removeVectorContainer(); err != nil {
		return err
	}

	return nil
}

func removeVectorContainer() error {
	if !common.ContainerExists(vectorContainerName) {
		return nil
	}

	cmd := common.NewShellCmd(strings.Join([]string{
		common.DockerBin(), "container", "rm", "-f", vectorContainerName}, " "))

	return common.SuppressOutput(func() error {
		if cmd.Execute() {
			return nil
		}

		if common.ContainerExists(vectorContainerName) {
			return errors.New("Unable to remove vector container")
		}

		return nil
	})
}

func startVectorContainer(vectorImage string) error {
	cmd := common.NewShellCmd(strings.Join([]string{
		common.DockerBin(),
		"container",
		"run", "--detach", "--name", vectorContainerName, common.MustGetEnv("DOKKU_GLOBAL_RUN_ARGS"),
		"--volume", "/var/lib/dokku/data/logs/vector.json:/etc/vector/vector.json",
		"--volume", "/var/run/docker.sock:/var/run/docker.sock",
		"--volume", common.MustGetEnv("DOKKU_LOGS_HOST_DIR") + ":/var/logs/dokku/apps",
		vectorImage,
		"--config", "/etc/vector/vector.json", "--watch-config"}, " "))

	if !cmd.Execute() {
		return errors.New("Unable to start vector container")
	}

	return nil
}

func stopVectorContainer() error {
	if !common.ContainerExists(vectorContainerName) {
		return nil
	}

	if !common.ContainerIsRunning(vectorContainerName) {
		return nil
	}

	cmd := common.NewShellCmd(strings.Join([]string{
		common.DockerBin(), "container", "stop", vectorContainerName}, " "))

	return common.SuppressOutput(func() error {
		if cmd.Execute() {
			return nil
		}

		if common.ContainerIsRunning(vectorContainerName) {
			return errors.New("Unable to stop vector container")
		}

		return nil
	})
}

func valueToConfig(appName string, value string) (vectorSink, error) {
	var data vectorSink
	u, err := url.Parse(value)
	if err != nil {
		return data, err
	}

	if u.Query().Get("sinks") != "" {
		return data, errors.New("Invalid option sinks")
	}

	t := fmt.Sprintf("type=%s", u.Scheme)
	i := fmt.Sprintf("inputs[]=docker-source:%s", appName)
	if appName == "--global" {
		i = "inputs[]=docker-global-source"
	}
	if appName == "--null" {
		i = "inputs[]=docker-null-source"
	}

	initialQuery := fmt.Sprintf("%s&%s", t, i)
	query := u.RawQuery
	if query == "" {
		query = initialQuery
	} else if strings.HasPrefix(query, "&") {
		query = fmt.Sprintf("%s%s", initialQuery, query)
	} else {
		query = fmt.Sprintf("%s&%s", initialQuery, query)
	}

	b, err := qson.ToJSON(query)
	if err != nil {
		return data, err
	}

	if err := json.Unmarshal(b, &data); err != nil {
		return data, err
	}

	return data, nil
}

func writeVectorConfig() error {
	apps, err := common.DokkuApps()
	if err != nil {
		return err
	}

	data := vectorConfig{
		Sources: map[string]vectorSource{},
		Sinks:   map[string]vectorSink{},
	}
	for _, appName := range apps {
		value := common.PropertyGet("logs", appName, "vector-sink")
		if value == "" {
			continue
		}

		sink, err := valueToConfig(appName, value)
		if err != nil {
			return err
		}

		data.Sources[fmt.Sprintf("docker-source:%s", appName)] = vectorSource{
			Type:          "docker_logs",
			IncludeLabels: []string{fmt.Sprintf("com.dokku.app-name=%s", appName)},
		}

		data.Sinks[fmt.Sprintf("docker-sink:%s", appName)] = sink
	}

	value := common.PropertyGet("logs", "--global", "vector-sink")
	if value != "" {
		sink, err := valueToConfig("--global", value)
		if err != nil {
			return err
		}

		data.Sources["docker-global-source"] = vectorSource{
			Type:          "docker_logs",
			IncludeLabels: []string{"com.dokku.app-name"},
		}

		data.Sinks["docker-global-sink"] = sink
	}

	if len(data.Sources) == 0 {
		// pull from no containers
		data.Sources["docker-null-source"] = vectorSource{
			Type:          "docker_logs",
			IncludeLabels: []string{"com.dokku.vector-null"},
		}
	}

	if len(data.Sinks) == 0 {
		// write logs to a blackhole
		sink, err := valueToConfig("--null", "blackhole://?print_amount=1")
		if err != nil {
			return err
		}

		data.Sinks["docker-null-sink"] = sink
	}

	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	vectorConfig := filepath.Join(common.MustGetEnv("DOKKU_LIB_ROOT"), "data", "logs", "vector.json")
	if err := common.WriteSliceToFile(vectorConfig, []string{string(b)}); err != nil {
		return err
	}

	return nil
}
