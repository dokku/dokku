package logs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

type vectorConfig struct {
	Sources map[string]vectorSource `json:"sources"`

	// Transforms is left nil unless a cron sink is configured, so that configs
	// without one marshal identically to those generated before cron routing
	// existed
	Transforms map[string]any `json:"transforms,omitempty"`

	Sinks map[string]VectorSink `json:"sinks"`
}

type vectorSource struct {
	Type          string   `json:"type"`
	IncludeLabels []string `json:"include_labels,omitempty"`
}

type vectorRouteTransform struct {
	Type             string                     `json:"type"`
	Inputs           []string                   `json:"inputs"`
	RerouteUnmatched bool                       `json:"reroute_unmatched"`
	Route            map[string]vectorCondition `json:"route"`
}

type vectorCondition struct {
	Type   string `json:"type"`
	Source string `json:"source"`
}

type vectorRemapTransform struct {
	Type   string   `json:"type"`
	Inputs []string `json:"inputs"`
	Source string   `json:"source"`
}

type vectorTemplateData struct {
	DokkuLibRoot   string
	DokkuLogsDir   string
	VectorImage    string
	VectorNetworks []string
}

const vectorContainerName = "vector-vector-1"
const vectorOldContainerName = "vector"

func getComposeFile() ([]byte, error) {
	result, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "vector-template-source",
	})
	if err == nil && result.ExitCode == 0 && strings.TrimSpace(result.Stdout) != "" {
		contents, err := os.ReadFile(strings.TrimSpace(result.Stdout))
		if err != nil {
			return []byte{}, fmt.Errorf("Unable to read compose template: %s", err)
		}

		return contents, nil
	}

	contents, err := templates.ReadFile("templates/compose.yml.tmpl")
	if err != nil {
		return []byte{}, fmt.Errorf("Unable to read compose template: %s", err)
	}

	return contents, nil
}

func startVectorContainer(vectorImage string) error {
	if !common.IsComposeInstalled() {
		return errors.New("Required docker compose plugin is not installed")
	}

	if common.ContainerExists(vectorOldContainerName) {
		return errors.New("Vector container %s already exists in old format, run 'dokku logs:vector-stop' once to remove it")
	}

	tmpFile, err := os.CreateTemp(os.TempDir(), "vector-compose-*.yml")
	if err != nil {
		return fmt.Errorf("Unable to create temporary file: %s", err)
	}
	defer os.Remove(tmpFile.Name())

	contents, err := getComposeFile()
	if err != nil {
		return fmt.Errorf("Unable to read compose template: %s", err)
	}

	tmpl, err := template.New("compose.yml").Parse(string(contents))
	if err != nil {
		return fmt.Errorf("Unable to parse compose template: %s", err)
	}

	dokkuLibRoot := os.Getenv("DOKKU_LIB_HOST_ROOT")
	if dokkuLibRoot == "" {
		dokkuLibRoot = os.Getenv("DOKKU_LIB_ROOT")
	}

	dokkuLogsDir := os.Getenv("DOKKU_LOGS_HOST_DIR")
	if dokkuLogsDir == "" {
		dokkuLogsDir = os.Getenv("DOKKU_LOGS_DIR")
	}

	data := vectorTemplateData{
		DokkuLibRoot:   dokkuLibRoot,
		DokkuLogsDir:   dokkuLogsDir,
		VectorImage:    vectorImage,
		VectorNetworks: getVectorNetworks(),
	}

	if err := tmpl.Execute(tmpFile, data); err != nil {
		return fmt.Errorf("Unable to execute compose template: %s", err)
	}

	return common.ComposeUp(common.ComposeUpInput{
		ProjectName: "vector",
		ComposeFile: tmpFile.Name(),
	})
}

func getVectorNetworks() []string {
	value := common.PropertyGet("logs", "--global", "vector-networks")
	if value == "" {
		return nil
	}

	networks := []string{}
	for _, name := range strings.Split(value, ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		networks = append(networks, name)
	}

	return networks
}

func getComputedVectorImage() string {
	return common.PropertyGetDefault("logs", "--global", "vector-image", getDefaultVectorImage())
}

// getDefaultVectorImage returns the default image used for the vector container
func getDefaultVectorImage() string {
	contents := strings.TrimSpace(VectorDockerfile)
	parts := strings.SplitN(contents, " ", 2)
	return parts[1]
}

func stopVectorContainer() error {
	if !common.IsComposeInstalled() {
		return errors.New("Required docker compose plugin is not installed")
	}

	if common.ContainerExists(vectorOldContainerName) {
		common.ContainerRemove(vectorOldContainerName)
	}

	tmpFile, err := os.CreateTemp(os.TempDir(), "vector-compose-*.yml")
	if err != nil {
		return fmt.Errorf("Unable to create temporary file: %s", err)
	}
	defer os.Remove(tmpFile.Name())

	contents, err := getComposeFile()
	if err != nil {
		return fmt.Errorf("Unable to read compose template: %s", err)
	}

	tmpl, err := template.New("compose.yml").Parse(string(contents))
	if err != nil {
		return fmt.Errorf("Unable to parse compose template: %s", err)
	}

	dokkuLibRoot := os.Getenv("DOKKU_LIB_HOST_ROOT")
	if dokkuLibRoot == "" {
		dokkuLibRoot = os.Getenv("DOKKU_LIB_ROOT")
	}

	dokkuLogsDir := os.Getenv("DOKKU_LOGS_HOST_DIR")
	if dokkuLogsDir == "" {
		dokkuLogsDir = os.Getenv("DOKKU_LOGS_DIR")
	}

	data := vectorTemplateData{
		DokkuLibRoot:   dokkuLibRoot,
		DokkuLogsDir:   dokkuLogsDir,
		VectorImage:    getComputedVectorImage(),
		VectorNetworks: getVectorNetworks(),
	}

	if err := tmpl.Execute(tmpFile, data); err != nil {
		return fmt.Errorf("Unable to execute compose template: %s", err)
	}

	return common.ComposeDown(common.ComposeDownInput{
		ProjectName: "vector",
		ComposeFile: tmpFile.Name(),
	})
}

// vectorAppSinks holds the resolved sink configuration for a single app or for
// the global scope
type vectorAppSinks struct {
	// SourceID is the vector source component id
	SourceID string

	// IncludeLabels is the docker_logs label filter for the source
	IncludeLabels []string

	// SinkID is the component id for the sink receiving non-cron logs
	SinkID string

	// CronSinkID is the component id for the sink receiving cron task logs
	CronSinkID string

	// RouterID is the component id for the route transform splitting the source
	RouterID string

	// CronRemapID is the component id for the remap transform on the cron branch
	CronRemapID string

	// Sink is the DSN for non-cron logs, empty when unset
	Sink string

	// CronSink is the DSN for cron task logs, empty when unset
	CronSink string
}

// cronRouteTransforms returns the route and remap pair that splits a source
// into cron and non-cron branches. The remap flattens the cron labels into
// top-level fields because vector drops any event whose sink template
// references a missing field, and a nested quoted path is awkward to template.
//
// The route carries a single condition, so an event either matches it or falls
// through to the reserved _unmatched output. A second route added later would
// need a mutually exclusive condition, since route fans out to every match.
func cronRouteTransforms(routerID string, remapID string, sourceID string, hasSink bool) map[string]any {
	return map[string]any{
		routerID: vectorRouteTransform{
			Type:             "route",
			Inputs:           []string{sourceID},
			RerouteUnmatched: hasSink,
			Route: map[string]vectorCondition{
				CronRouteName: {
					Type:   "vrl",
					Source: fmt.Sprintf("%s == %q", vrlLabelPath(ContainerTypeLabel), CronContainerType),
				},
			},
		},
		remapID: vectorRemapTransform{
			Type:   "remap",
			Inputs: []string{fmt.Sprintf("%s.%s", routerID, CronRouteName)},
			Source: fmt.Sprintf(".dokku_app = to_string(%s) ?? \"\"\n.dokku_cron_id = to_string(%s) ?? \"\"",
				vrlLabelPath(AppLabelAlias), vrlLabelPath(CronIDLabel)),
		},
	}
}

// vrlLabelPath renders a docker label lookup as a VRL path. Segments holding
// characters outside [A-Za-z0-9_] must be double quoted.
func vrlLabelPath(label string) string {
	return fmt.Sprintf(".label.%q", label)
}

// buildVectorConfig assembles the vector configuration for the supplied scopes.
// It performs no IO so that the generated shape can be asserted directly.
func buildVectorConfig(scopes []vectorAppSinks) (vectorConfig, error) {
	data := vectorConfig{
		Sources: map[string]vectorSource{},
		Sinks:   map[string]VectorSink{},
	}

	for _, scope := range scopes {
		if scope.Sink == "" && scope.CronSink == "" {
			continue
		}

		data.Sources[scope.SourceID] = vectorSource{
			Type:          "docker_logs",
			IncludeLabels: scope.IncludeLabels,
		}

		sinkInputs := []string{scope.SourceID}
		if scope.CronSink != "" {
			if data.Transforms == nil {
				data.Transforms = map[string]any{}
			}
			for id, transform := range cronRouteTransforms(scope.RouterID, scope.CronRemapID, scope.SourceID, scope.Sink != "") {
				data.Transforms[id] = transform
			}

			sinkInputs = []string{fmt.Sprintf("%s._unmatched", scope.RouterID)}

			cronSink, err := SinkValueToConfig(SinkValueToConfigInput{
				SinkValue: scope.CronSink,
				Inputs:    []string{scope.CronRemapID},
			})
			if err != nil {
				return data, err
			}

			data.Sinks[scope.CronSinkID] = cronSink
		}

		if scope.Sink != "" {
			sink, err := SinkValueToConfig(SinkValueToConfigInput{
				SinkValue: scope.Sink,
				Inputs:    sinkInputs,
			})
			if err != nil {
				return data, err
			}

			data.Sinks[scope.SinkID] = sink
		}
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
		sink, err := SinkValueToConfig(SinkValueToConfigInput{
			SinkValue: VectorDefaultSink,
			Inputs:    []string{"docker-null-source"},
		})
		if err != nil {
			return data, err
		}

		data.Sinks["docker-null-sink"] = sink
	}

	return data, nil
}

// vectorScopes collects the sink configuration for every app plus the global scope
func vectorScopes() []vectorAppSinks {
	apps, _ := common.UnfilteredDokkuApps()
	scopes := []vectorAppSinks{}
	for _, appName := range apps {
		inflectedAppName := strings.ReplaceAll(appName, ".", "-")
		scopes = append(scopes, vectorAppSinks{
			SourceID:      fmt.Sprintf("docker-source:%s", inflectedAppName),
			IncludeLabels: []string{fmt.Sprintf("%s=%s", reportComputedAppLabelAlias(appName), appName)},
			SinkID:        fmt.Sprintf("docker-sink:%s", inflectedAppName),
			CronSinkID:    fmt.Sprintf("docker-cron-sink:%s", inflectedAppName),
			RouterID:      fmt.Sprintf("docker-router:%s", inflectedAppName),
			CronRemapID:   fmt.Sprintf("docker-cron-remap:%s", inflectedAppName),
			Sink:          common.PropertyGet("logs", appName, "vector-sink"),
			CronSink:      common.PropertyGet("logs", appName, "vector-cron-sink"),
		})
	}

	return append(scopes, vectorAppSinks{
		SourceID:      "docker-global-source",
		IncludeLabels: []string{reportComputedAppLabelAlias("global")},
		SinkID:        "docker-global-sink",
		CronSinkID:    "docker-global-cron-sink",
		RouterID:      "docker-global-router",
		CronRemapID:   "docker-global-cron-remap",
		Sink:          common.PropertyGet("logs", "--global", "vector-sink"),
		CronSink:      common.PropertyGet("logs", "--global", "vector-cron-sink"),
	})
}

func writeVectorConfig() error {
	data, err := buildVectorConfig(vectorScopes())
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	b = bytes.Replace(b, []byte("\\u0026"), []byte("&"), -1)
	b = bytes.Replace(b, []byte("\\u002B"), []byte("+"), -1)

	vectorConfig := filepath.Join(common.GetDataDirectory("logs"), "vector.json")
	if err := common.WriteBytesToFile(common.WriteBytesToFileInput{
		Bytes:    b,
		Filename: vectorConfig,
		Mode:     os.FileMode(0600),
	}); err != nil {
		return err
	}

	return nil
}
