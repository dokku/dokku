package logs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
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

	// RelabelID is the component id for the remap transform renaming the app
	// label on the non-cron branch
	RelabelID string

	// LabelAlias is the label key the app name is shipped under for this scope.
	// It is the AppLabelAlias constant unless the app-label-alias property is set
	LabelAlias string

	// LabelAliasOverrides holds the apps within this scope whose own alias
	// differs from LabelAlias. Only the global scope carries these, since a
	// per-app source only ever covers one app
	LabelAliasOverrides []vectorLabelAliasOverride

	// Sink is the DSN for non-cron logs, empty when unset
	Sink string

	// CronSink is the DSN for cron task logs, empty when unset
	CronSink string
}

// vectorLabelAliasOverride is the alias a single app ships its name under when
// it differs from the alias of the scope collecting it
type vectorLabelAliasOverride struct {
	// AppName is the app the override applies to
	AppName string

	// Alias is the label key the app name is shipped under
	Alias string
}

// cronRouteTransforms returns the route and remap pair that splits a source
// into cron and non-cron branches. The remap flattens the cron labels into
// top-level fields because vector drops any event whose sink template
// references a missing field, and a nested quoted path is awkward to template.
//
// The route carries a single condition, so an event either matches it or falls
// through to the reserved _unmatched output. A second route added later would
// need a mutually exclusive condition, since route fans out to every match.
//
// Any label rename is appended to the remap rather than given a component of
// its own, so that dokku_app is captured from the literal label before the
// rename runs and keeps its value regardless of the configured alias.
func cronRouteTransforms(scope vectorAppSinks, relabel string) map[string]any {
	source := fmt.Sprintf(".dokku_app = to_string(%s) ?? \"\"\n.dokku_cron_id = to_string(%s) ?? \"\"",
		vrlLabelPath(AppLabelAlias), vrlLabelPath(CronIDLabel))
	if relabel != "" {
		source = fmt.Sprintf("%s\n%s", source, relabel)
	}

	return map[string]any{
		scope.RouterID: vectorRouteTransform{
			Type:             "route",
			Inputs:           []string{scope.SourceID},
			RerouteUnmatched: scope.Sink != "",
			Route: map[string]vectorCondition{
				CronRouteName: {
					Type:   "vrl",
					Source: fmt.Sprintf("%s == %q", vrlLabelPath(ContainerTypeLabel), CronContainerType),
				},
			},
		},
		scope.CronRemapID: vectorRemapTransform{
			Type:   "remap",
			Inputs: []string{fmt.Sprintf("%s.%s", scope.RouterID, CronRouteName)},
			Source: source,
		},
	}
}

// vrlLabelPath renders a docker label lookup as a VRL path. Segments holding
// characters outside [A-Za-z0-9_] must be double quoted.
func vrlLabelPath(label string) string {
	return fmt.Sprintf(".label.%q", label)
}

// vrlRenameLabel renders the assignment moving the dokku app label onto the
// supplied alias
func vrlRenameLabel(alias string) string {
	return fmt.Sprintf("%s = del(%s)", vrlLabelPath(alias), vrlLabelPath(AppLabelAlias))
}

// vrlClause is one branch of a generated VRL conditional. An empty Condition
// renders as a trailing else
type vrlClause struct {
	Condition string
	Statement string
}

// vrlIfChain renders clauses as a single if/else if/else statement
func vrlIfChain(clauses []vrlClause) string {
	if len(clauses) == 1 && clauses[0].Condition == "" {
		return clauses[0].Statement
	}

	var chain strings.Builder
	for i, clause := range clauses {
		if i > 0 {
			chain.WriteString(" else ")
		}
		if clause.Condition != "" {
			chain.WriteString(fmt.Sprintf("if %s ", clause.Condition))
		}
		chain.WriteString(fmt.Sprintf("{\n  %s\n}", clause.Statement))
	}

	return chain.String()
}

// relabelVRL renders the program renaming the dokku app label to the alias
// configured for the scope, or an empty string when there is nothing to rename.
//
// The rename happens on the event rather than on the container because dokku
// only ever labels containers com.dokku.app-name: pointing the source filter at
// any other label - which is what this property used to do - matches nothing at
// all and silently collects no logs.
//
// The global scope collects every app, so an app whose own alias differs from
// the global one gets a branch of its own here. Without that, a per-app alias
// would be silently ignored for any app shipping through the global sink.
func relabelVRL(scope vectorAppSinks) string {
	alias := scope.LabelAlias
	if alias == "" {
		alias = AppLabelAlias
	}

	if len(scope.LabelAliasOverrides) == 0 {
		if alias == AppLabelAlias {
			return ""
		}

		return vrlRenameLabel(alias)
	}

	clauses := []vrlClause{}
	retained := []string{}
	for _, override := range scope.LabelAliasOverrides {
		if override.Alias == AppLabelAlias {
			retained = append(retained, fmt.Sprintf("app != %q", override.AppName))
			continue
		}

		clauses = append(clauses, vrlClause{
			Condition: fmt.Sprintf("app == %q", override.AppName),
			Statement: vrlRenameLabel(override.Alias),
		})
	}

	if alias != AppLabelAlias {
		clauses = append(clauses, vrlClause{
			Condition: strings.Join(retained, " && "),
			Statement: vrlRenameLabel(alias),
		})
	}

	if len(clauses) == 0 {
		return ""
	}

	return fmt.Sprintf("app = to_string(%s) ?? \"\"\n%s", vrlLabelPath(AppLabelAlias), vrlIfChain(clauses))
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

		relabel := relabelVRL(scope)

		sinkInputs := []string{scope.SourceID}
		if scope.CronSink != "" {
			if data.Transforms == nil {
				data.Transforms = map[string]any{}
			}
			for id, transform := range cronRouteTransforms(scope, relabel) {
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
			// the cron branch renames within its own remap, so this transform
			// only exists when there is a non-cron sink downstream to consume it
			if relabel != "" {
				if data.Transforms == nil {
					data.Transforms = map[string]any{}
				}
				data.Transforms[scope.RelabelID] = vectorRemapTransform{
					Type:   "remap",
					Inputs: sinkInputs,
					Source: relabel,
				}

				sinkInputs = []string{scope.RelabelID}
			}

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
	globalAlias := reportComputedAppLabelAlias("--global")

	scopes := []vectorAppSinks{}
	overrides := []vectorLabelAliasOverride{}
	for _, appName := range apps {
		inflectedAppName := strings.ReplaceAll(appName, ".", "-")
		appAlias := reportComputedAppLabelAlias(appName)
		if appAlias != globalAlias {
			overrides = append(overrides, vectorLabelAliasOverride{
				AppName: appName,
				Alias:   appAlias,
			})
		}

		scopes = append(scopes, vectorAppSinks{
			SourceID:      fmt.Sprintf("docker-source:%s", inflectedAppName),
			IncludeLabels: []string{fmt.Sprintf("%s=%s", AppLabelAlias, appName)},
			SinkID:        fmt.Sprintf("docker-sink:%s", inflectedAppName),
			CronSinkID:    fmt.Sprintf("docker-cron-sink:%s", inflectedAppName),
			RouterID:      fmt.Sprintf("docker-router:%s", inflectedAppName),
			CronRemapID:   fmt.Sprintf("docker-cron-remap:%s", inflectedAppName),
			RelabelID:     fmt.Sprintf("docker-relabel:%s", inflectedAppName),
			LabelAlias:    appAlias,
			Sink:          common.PropertyGet("logs", appName, "vector-sink"),
			CronSink:      common.PropertyGet("logs", appName, "vector-cron-sink"),
		})
	}

	sort.Slice(overrides, func(i int, j int) bool {
		return overrides[i].AppName < overrides[j].AppName
	})

	return append(scopes, vectorAppSinks{
		SourceID:            "docker-global-source",
		IncludeLabels:       []string{AppLabelAlias},
		SinkID:              "docker-global-sink",
		CronSinkID:          "docker-global-cron-sink",
		RouterID:            "docker-global-router",
		CronRemapID:         "docker-global-cron-remap",
		RelabelID:           "docker-global-relabel",
		LabelAlias:          globalAlias,
		LabelAliasOverrides: overrides,
		Sink:                common.PropertyGet("logs", "--global", "vector-sink"),
		CronSink:            common.PropertyGet("logs", "--global", "vector-cron-sink"),
	})
}

// regenerateVectorConfig rewrites the generated vector config so that it matches
// the current set of apps and their properties. App lifecycle triggers call this
// instead of writeVectorConfig because the config is derived state: failing to
// rewrite it should not abort the app operation that changed the state it is
// derived from.
func regenerateVectorConfig() {
	if err := writeVectorConfig(); err != nil {
		common.LogWarn(fmt.Sprintf("Unable to write updated vector config: %s", err.Error()))
	}
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
