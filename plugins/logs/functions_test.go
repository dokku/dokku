package logs

import (
	"encoding/json"
	"strings"
	"testing"
)

func appScope(sink string, cronSink string) vectorAppSinks {
	return vectorAppSinks{
		SourceID:      "docker-source:myapp",
		IncludeLabels: []string{"com.dokku.app-name=myapp"},
		SinkID:        "docker-sink:myapp",
		CronSinkID:    "docker-cron-sink:myapp",
		RouterID:      "docker-router:myapp",
		CronRemapID:   "docker-cron-remap:myapp",
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
	_, decoded := marshalConfig(t, []vectorAppSinks{{
		SourceID:      "docker-global-source",
		IncludeLabels: []string{"com.dokku.app-name"},
		SinkID:        "docker-global-sink",
		CronSinkID:    "docker-global-cron-sink",
		RouterID:      "docker-global-router",
		CronRemapID:   "docker-global-cron-remap",
		Sink:          "console://?encoding[codec]=json",
		CronSink:      "console://?encoding[codec]=text",
	}})

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
