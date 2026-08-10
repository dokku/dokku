package logs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dokku/dokku/plugins/common"
)

func appScope(sink string, cronSink string) vectorAppSinks {
	return vectorAppSinks{
		SourceID:      "docker-source:myapp",
		IncludeLabels: []string{"com.dokku.app-name=myapp"},
		SinkID:        "docker-sink:myapp",
		CronSinkID:    "docker-cron-sink:myapp",
		RouterID:      "docker-router:myapp",
		CronRemapID:   "docker-cron-remap:myapp",
		RelabelID:     "docker-relabel:myapp",
		Sink:          sink,
		CronSink:      cronSink,
	}
}

func aliasScope(sink string, cronSink string, alias string) vectorAppSinks {
	scope := appScope(sink, cronSink)
	scope.LabelAlias = alias
	return scope
}

func globalScope(sink string, cronSink string) vectorAppSinks {
	return vectorAppSinks{
		SourceID:      "docker-global-source",
		IncludeLabels: []string{"com.dokku.app-name"},
		SinkID:        "docker-global-sink",
		CronSinkID:    "docker-global-cron-sink",
		RouterID:      "docker-global-router",
		CronRemapID:   "docker-global-cron-remap",
		RelabelID:     "docker-global-relabel",
		Sink:          sink,
		CronSink:      cronSink,
	}
}

func marshalConfig(t *testing.T, scopes []vectorAppSinks) (string, map[string]interface{}) {
	t.Helper()

	config, err := buildVectorConfig(scopes)
	if err != nil {
		t.Fatalf("buildVectorConfig() error = %v", err)
	}

	b, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() error = %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	return string(b), decoded
}

func lookup(t *testing.T, decoded map[string]interface{}, path ...string) interface{} {
	t.Helper()

	var current interface{} = decoded
	for _, key := range path {
		asMap, ok := current.(map[string]interface{})
		if !ok {
			t.Fatalf("path %v: %q is not a map", path, key)
		}
		current, ok = asMap[key]
		if !ok {
			t.Fatalf("path %v: missing key %q", path, key)
		}
	}

	return current
}

// TestBuildVectorConfigOmitsTransforms is the backwards compatibility guard:
// a config without a cron sink must marshal exactly as it did before cron
// routing existed, which means no transforms key at all.
func TestBuildVectorConfigOmitsTransforms(t *testing.T) {
	raw, decoded := marshalConfig(t, []vectorAppSinks{appScope("console://?encoding[codec]=json", "")})

	if strings.Contains(raw, "transforms") {
		t.Errorf("config should not contain a transforms key:\n%s", raw)
	}

	inputs := lookup(t, decoded, "sinks", "docker-sink:myapp", "inputs")
	if got := inputs.([]interface{})[0]; got != "docker-source:myapp" {
		t.Errorf("sink inputs[0] = %v, want docker-source:myapp", got)
	}
}

func TestBuildVectorConfigCronSinkOnly(t *testing.T) {
	_, decoded := marshalConfig(t, []vectorAppSinks{appScope("", "console://?encoding[codec]=json")})

	if got := lookup(t, decoded, "transforms", "docker-router:myapp", "type"); got != "route" {
		t.Errorf("router type = %v, want route", got)
	}

	// nothing consumes the unmatched output when there is no plain sink
	if got := lookup(t, decoded, "transforms", "docker-router:myapp", "reroute_unmatched"); got != false {
		t.Errorf("reroute_unmatched = %v, want false", got)
	}

	condition := lookup(t, decoded, "transforms", "docker-router:myapp", "route", "cron", "source")
	want := `.label."com.dokku.container-type" == "cron"`
	if condition != want {
		t.Errorf("route condition = %v, want %v", condition, want)
	}

	remapSource := lookup(t, decoded, "transforms", "docker-cron-remap:myapp", "source").(string)
	for _, fragment := range []string{".dokku_app", ".dokku_cron_id", `.label."com.dokku.cron-id"`} {
		if !strings.Contains(remapSource, fragment) {
			t.Errorf("remap source %q missing %q", remapSource, fragment)
		}
	}

	inputs := lookup(t, decoded, "sinks", "docker-cron-sink:myapp", "inputs")
	if got := inputs.([]interface{})[0]; got != "docker-cron-remap:myapp" {
		t.Errorf("cron sink inputs[0] = %v, want docker-cron-remap:myapp", got)
	}

	sinks := lookup(t, decoded, "sinks").(map[string]interface{})
	if _, ok := sinks["docker-sink:myapp"]; ok {
		t.Error("plain sink should not exist when only a cron sink is set")
	}
}

func TestBuildVectorConfigBothSinks(t *testing.T) {
	_, decoded := marshalConfig(t, []vectorAppSinks{
		appScope("console://?encoding[codec]=json", "console://?encoding[codec]=text"),
	})

	if got := lookup(t, decoded, "transforms", "docker-router:myapp", "reroute_unmatched"); got != true {
		t.Errorf("reroute_unmatched = %v, want true", got)
	}

	inputs := lookup(t, decoded, "sinks", "docker-sink:myapp", "inputs")
	if got := inputs.([]interface{})[0]; got != "docker-router:myapp._unmatched" {
		t.Errorf("plain sink inputs[0] = %v, want docker-router:myapp._unmatched", got)
	}

	cronInputs := lookup(t, decoded, "sinks", "docker-cron-sink:myapp", "inputs")
	if got := cronInputs.([]interface{})[0]; got != "docker-cron-remap:myapp" {
		t.Errorf("cron sink inputs[0] = %v, want docker-cron-remap:myapp", got)
	}
}

func TestBuildVectorConfigGlobalScope(t *testing.T) {
	_, decoded := marshalConfig(t, []vectorAppSinks{
		globalScope("console://?encoding[codec]=json", "console://?encoding[codec]=text"),
	})

	lookup(t, decoded, "transforms", "docker-global-router")
	lookup(t, decoded, "transforms", "docker-global-cron-remap")
	lookup(t, decoded, "sinks", "docker-global-cron-sink")

	inputs := lookup(t, decoded, "sinks", "docker-global-sink", "inputs")
	if got := inputs.([]interface{})[0]; got != "docker-global-router._unmatched" {
		t.Errorf("global sink inputs[0] = %v, want docker-global-router._unmatched", got)
	}
}

func TestBuildVectorConfigNoSinks(t *testing.T) {
	raw, decoded := marshalConfig(t, []vectorAppSinks{appScope("", "")})

	if strings.Contains(raw, "transforms") {
		t.Errorf("config should not contain a transforms key:\n%s", raw)
	}

	lookup(t, decoded, "sources", "docker-null-source")

	inputs := lookup(t, decoded, "sinks", "docker-null-sink", "inputs")
	if got := inputs.([]interface{})[0]; got != "docker-null-source" {
		t.Errorf("null sink inputs[0] = %v, want docker-null-source", got)
	}
}

func TestBuildVectorConfigInvalidSink(t *testing.T) {
	if _, err := buildVectorConfig([]vectorAppSinks{appScope("console://?sinks=nope", "")}); err == nil {
		t.Fatal("buildVectorConfig() expected an error for an invalid sink DSN")
	}
}

// TestBuildVectorConfigAliasKeepsSourceFilter is the guard for the bug this
// alias handling replaced: filtering the source on the alias matched no
// container at all, because dokku only ever labels containers with the literal
// key, so setting the property silently collected nothing.
func TestBuildVectorConfigAliasKeepsSourceFilter(t *testing.T) {
	_, decoded := marshalConfig(t, []vectorAppSinks{
		aliasScope("console://?encoding[codec]=json", "", "app_name"),
	})

	labels := lookup(t, decoded, "sources", "docker-source:myapp", "include_labels")
	if got := labels.([]interface{})[0]; got != "com.dokku.app-name=myapp" {
		t.Errorf("include_labels[0] = %v, want com.dokku.app-name=myapp", got)
	}
}

func TestBuildVectorConfigAliasRelabelsPlainBranch(t *testing.T) {
	_, decoded := marshalConfig(t, []vectorAppSinks{
		aliasScope("console://?encoding[codec]=json", "", "app_name"),
	})

	if got := lookup(t, decoded, "transforms", "docker-relabel:myapp", "type"); got != "remap" {
		t.Errorf("relabel type = %v, want remap", got)
	}

	inputs := lookup(t, decoded, "transforms", "docker-relabel:myapp", "inputs")
	if got := inputs.([]interface{})[0]; got != "docker-source:myapp" {
		t.Errorf("relabel inputs[0] = %v, want docker-source:myapp", got)
	}

	source := lookup(t, decoded, "transforms", "docker-relabel:myapp", "source")
	want := `.label."app_name" = del(.label."com.dokku.app-name")`
	if source != want {
		t.Errorf("relabel source = %v, want %v", source, want)
	}

	sinkInputs := lookup(t, decoded, "sinks", "docker-sink:myapp", "inputs")
	if got := sinkInputs.([]interface{})[0]; got != "docker-relabel:myapp" {
		t.Errorf("sink inputs[0] = %v, want docker-relabel:myapp", got)
	}
}

// TestBuildVectorConfigAliasRelabelsBothBranches pins the rename onto the tail
// of the cron remap. dokku_app is captured from the literal label first, so it
// keeps its value no matter which alias the event is shipped under.
func TestBuildVectorConfigAliasRelabelsBothBranches(t *testing.T) {
	_, decoded := marshalConfig(t, []vectorAppSinks{
		aliasScope("console://?encoding[codec]=json", "console://?encoding[codec]=text", "app_name"),
	})

	remapSource := lookup(t, decoded, "transforms", "docker-cron-remap:myapp", "source")
	want := ".dokku_app = to_string(.label.\"com.dokku.app-name\") ?? \"\"\n" +
		".dokku_cron_id = to_string(.label.\"com.dokku.cron-id\") ?? \"\"\n" +
		".label.\"app_name\" = del(.label.\"com.dokku.app-name\")"
	if remapSource != want {
		t.Errorf("cron remap source = %v, want %v", remapSource, want)
	}

	inputs := lookup(t, decoded, "transforms", "docker-relabel:myapp", "inputs")
	if got := inputs.([]interface{})[0]; got != "docker-router:myapp._unmatched" {
		t.Errorf("relabel inputs[0] = %v, want docker-router:myapp._unmatched", got)
	}

	sinkInputs := lookup(t, decoded, "sinks", "docker-sink:myapp", "inputs")
	if got := sinkInputs.([]interface{})[0]; got != "docker-relabel:myapp" {
		t.Errorf("sink inputs[0] = %v, want docker-relabel:myapp", got)
	}
}

// TestBuildVectorConfigAliasCronSinkOnly covers the branch with nothing
// downstream to consume a relabel component: vector rejects a transform whose
// output no sink reads, so the rename has to stay inside the cron remap.
func TestBuildVectorConfigAliasCronSinkOnly(t *testing.T) {
	_, decoded := marshalConfig(t, []vectorAppSinks{
		aliasScope("", "console://?encoding[codec]=text", "app_name"),
	})

	transforms := lookup(t, decoded, "transforms").(map[string]interface{})
	if _, ok := transforms["docker-relabel:myapp"]; ok {
		t.Error("relabel transform should not exist without a plain sink")
	}

	remapSource := lookup(t, decoded, "transforms", "docker-cron-remap:myapp", "source").(string)
	if !strings.Contains(remapSource, `.label."app_name" = del(.label."com.dokku.app-name")`) {
		t.Errorf("cron remap source %q missing the rename", remapSource)
	}
}

func setupScopesTest(t *testing.T, apps []string) {
	t.Helper()
	t.Setenv("PLUGIN_PATH", "/var/lib/dokku/plugins")
	t.Setenv("PLUGIN_ENABLED_PATH", "/var/lib/dokku/plugins/enabled")
	t.Setenv("DOKKU_LIB_ROOT", t.TempDir())
	t.Setenv("DOKKU_ROOT", t.TempDir())
	t.Setenv("DOKKU_SYSTEM_USER", "root")
	t.Setenv("DOKKU_SYSTEM_GROUP", "root")

	if err := common.PropertySetup("logs"); err != nil {
		t.Fatalf("PropertySetup: %v", err)
	}

	for _, appName := range apps {
		if err := os.MkdirAll(filepath.Join(os.Getenv("DOKKU_ROOT"), appName), 0755); err != nil {
			t.Fatalf("MkdirAll %s: %v", appName, err)
		}
	}
}

func scopeByID(t *testing.T, scopes []vectorAppSinks, sourceID string) vectorAppSinks {
	t.Helper()

	for _, scope := range scopes {
		if scope.SourceID == sourceID {
			return scope
		}
	}

	t.Fatalf("no scope with source id %q", sourceID)
	return vectorAppSinks{}
}

// TestVectorScopesGlobalAliasIgnoresAppNamedGlobal pins the global scope to the
// --global property. It used to resolve against an app literally named global,
// which would have let that app's own alias drive every other app's shipping.
func TestVectorScopesGlobalAliasIgnoresAppNamedGlobal(t *testing.T) {
	setupScopesTest(t, []string{"global", "myapp"})

	if err := common.PropertyWrite("logs", "--global", "app-label-alias", "gname"); err != nil {
		t.Fatalf("PropertyWrite --global: %v", err)
	}
	if err := common.PropertyWrite("logs", "global", "app-label-alias", "hijack"); err != nil {
		t.Fatalf("PropertyWrite global: %v", err)
	}

	scopes := vectorScopes()

	globalScope := scopeByID(t, scopes, "docker-global-source")
	if globalScope.LabelAlias != "gname" {
		t.Errorf("global LabelAlias = %q, want gname", globalScope.LabelAlias)
	}

	// the app named global differs from the global alias, so it is the only
	// scope that should be listed as an override
	want := []vectorLabelAliasOverride{{AppName: "global", Alias: "hijack"}}
	if len(globalScope.LabelAliasOverrides) != 1 || globalScope.LabelAliasOverrides[0] != want[0] {
		t.Errorf("global LabelAliasOverrides = %v, want %v", globalScope.LabelAliasOverrides, want)
	}

	appScope := scopeByID(t, scopes, "docker-source:myapp")
	if appScope.LabelAlias != "gname" {
		t.Errorf("myapp LabelAlias = %q, want gname", appScope.LabelAlias)
	}
	if got := appScope.IncludeLabels[0]; got != "com.dokku.app-name=myapp" {
		t.Errorf("myapp include_labels[0] = %q, want com.dokku.app-name=myapp", got)
	}
}

func TestRelabelVRLDefaultAlias(t *testing.T) {
	for _, alias := range []string{"", AppLabelAlias} {
		if got := relabelVRL(aliasScope("console://", "", alias)); got != "" {
			t.Errorf("relabelVRL(%q) = %q, want an empty string", alias, got)
		}
	}
}

// TestRelabelVRLGlobalOverrides covers the global scope, which collects every
// app: an app whose own alias differs from the global one needs a branch here,
// or its alias would be silently dropped whenever it ships through the global
// sink.
func TestRelabelVRLGlobalOverrides(t *testing.T) {
	prelude := "app = to_string(.label.\"com.dokku.app-name\") ?? \"\"\n"

	tests := []struct {
		name      string
		alias     string
		overrides []vectorLabelAliasOverride
		want      string
	}{
		{
			name:      "renaming override under a default global alias",
			alias:     AppLabelAlias,
			overrides: []vectorLabelAliasOverride{{AppName: "appa", Alias: "foo"}},
			want:      prelude + "if app == \"appa\" {\n  .label.\"foo\" = del(.label.\"com.dokku.app-name\")\n}",
		},
		{
			name:      "override pinned back to the default label",
			alias:     "gname",
			overrides: []vectorLabelAliasOverride{{AppName: "appb", Alias: AppLabelAlias}},
			want:      prelude + "if app != \"appb\" {\n  .label.\"gname\" = del(.label.\"com.dokku.app-name\")\n}",
		},
		{
			name:  "both kinds of override at once",
			alias: "gname",
			overrides: []vectorLabelAliasOverride{
				{AppName: "appa", Alias: "foo"},
				{AppName: "appb", Alias: AppLabelAlias},
			},
			want: prelude + "if app == \"appa\" {\n  .label.\"foo\" = del(.label.\"com.dokku.app-name\")\n}" +
				" else if app != \"appb\" {\n  .label.\"gname\" = del(.label.\"com.dokku.app-name\")\n}",
		},
	}

	for _, test := range tests {
		scope := globalScope("console://", "")
		scope.LabelAlias = test.alias
		scope.LabelAliasOverrides = test.overrides

		if got := relabelVRL(scope); got != test.want {
			t.Errorf("%s: relabelVRL() = %q, want %q", test.name, got, test.want)
		}
	}
}
