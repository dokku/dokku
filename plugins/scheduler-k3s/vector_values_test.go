package scheduler_k3s

import (
	"testing"

	"github.com/dokku/dokku/plugins/common"
	"gopkg.in/yaml.v3"
)

// baseVectorValues mirrors the shipped templates/helm-config/vector.yaml
// closely enough to exercise the sink and transform merging
const baseVectorValues = `
customConfig:
  sources:
    kubernetes_logs:
      type: kubernetes_logs
    internal_metrics:
      type: internal_metrics
  transforms:
    kubernetes_container_logs:
      type: remap
      inputs:
        - kubernetes_logs
      source: |
        .container = .kubernetes.container_name
  sinks:
    default_global_sink:
      type: console
      inputs:
        - kubernetes_container_logs
      encoding:
        codec: json
    prom_exporter:
      type: prometheus_exporter
      inputs:
        - internal_metrics
      address: 0.0.0.0:9090
`

func setupVectorValuesTest(t *testing.T, sink string, cronSink string) map[string]interface{} {
	t.Helper()
	t.Setenv("PLUGIN_PATH", "/var/lib/dokku/plugins")
	t.Setenv("PLUGIN_ENABLED_PATH", "/var/lib/dokku/plugins/enabled")
	t.Setenv("DOKKU_LIB_ROOT", t.TempDir())
	t.Setenv("DOKKU_SYSTEM_USER", "root")
	t.Setenv("DOKKU_SYSTEM_GROUP", "root")

	if err := common.PropertySetup("logs"); err != nil {
		t.Fatalf("PropertySetup: %v", err)
	}
	if sink != "" {
		if err := common.PropertyWrite("logs", "--global", "vector-sink", sink); err != nil {
			t.Fatalf("PropertyWrite vector-sink: %v", err)
		}
	}
	if cronSink != "" {
		if err := common.PropertyWrite("logs", "--global", "vector-cron-sink", cronSink); err != nil {
			t.Fatalf("PropertyWrite vector-cron-sink: %v", err)
		}
	}

	values := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(baseVectorValues), &values); err != nil {
		t.Fatalf("Unmarshal base values: %v", err)
	}

	return values
}

func vectorSinks(t *testing.T, values map[string]interface{}) map[string]interface{} {
	t.Helper()
	customConfig, ok := values["customConfig"].(map[string]interface{})
	if !ok {
		t.Fatal("customConfig is not a map")
	}
	sinks, ok := customConfig["sinks"].(map[string]interface{})
	if !ok {
		t.Fatal("sinks is not a map")
	}
	return sinks
}

func vectorTransforms(t *testing.T, values map[string]interface{}) map[string]interface{} {
	t.Helper()
	customConfig, ok := values["customConfig"].(map[string]interface{})
	if !ok {
		t.Fatal("customConfig is not a map")
	}
	transforms, ok := customConfig["transforms"].(map[string]interface{})
	if !ok {
		t.Fatal("transforms is not a map")
	}
	return transforms
}

func sinkInput(t *testing.T, sinks map[string]interface{}, sinkID string) string {
	t.Helper()
	sink, ok := sinks[sinkID].(map[string]interface{})
	if !ok {
		t.Fatalf("sink %q is missing or not a map", sinkID)
	}

	switch inputs := sink["inputs"].(type) {
	case []string:
		return inputs[0]
	case []interface{}:
		return inputs[0].(string)
	default:
		t.Fatalf("sink %q has unexpected inputs type %T", sinkID, sink["inputs"])
		return ""
	}
}

func TestUpdateVectorValuesNoSinks(t *testing.T) {
	values := setupVectorValuesTest(t, "", "")

	updated, err := updateVectorValues(values)
	if err != nil {
		t.Fatalf("updateVectorValues() error = %v", err)
	}

	sinks := vectorSinks(t, updated)
	if _, ok := sinks[kubernetesDefaultSink]; !ok {
		t.Error("default_global_sink should remain when no sink is configured")
	}
	if _, ok := sinks[kubernetesGlobalSink]; ok {
		t.Error("kubernetes_global_sink should not exist when no sink is configured")
	}
}

// TestUpdateVectorValuesPreservesPromExporter guards against the sinks map
// being replaced wholesale, which previously dropped the prometheus exporter
// that the chart still exposes on port 9090
func TestUpdateVectorValuesPreservesPromExporter(t *testing.T) {
	values := setupVectorValuesTest(t, "console://?encoding[codec]=json", "")

	updated, err := updateVectorValues(values)
	if err != nil {
		t.Fatalf("updateVectorValues() error = %v", err)
	}

	sinks := vectorSinks(t, updated)
	if _, ok := sinks["prom_exporter"]; !ok {
		t.Error("prom_exporter should survive configuring a vector-sink")
	}
	if _, ok := sinks[kubernetesDefaultSink]; ok {
		t.Error("default_global_sink should be superseded by the configured sink")
	}
	if got := sinkInput(t, sinks, kubernetesGlobalSink); got != kubernetesLogsTransform {
		t.Errorf("global sink input = %v, want %v", got, kubernetesLogsTransform)
	}
}

func TestUpdateVectorValuesCronSinkOnly(t *testing.T) {
	values := setupVectorValuesTest(t, "", "console://?encoding[codec]=text")

	updated, err := updateVectorValues(values)
	if err != nil {
		t.Fatalf("updateVectorValues() error = %v", err)
	}

	transforms := vectorTransforms(t, updated)
	if _, ok := transforms[kubernetesRouterTransform]; !ok {
		t.Error("router transform should be added")
	}
	if _, ok := transforms[kubernetesLogsTransform]; !ok {
		t.Error("base container logs transform should be preserved")
	}

	sinks := vectorSinks(t, updated)
	if got := sinkInput(t, sinks, kubernetesCronSink); got != kubernetesCronRemapTransform {
		t.Errorf("cron sink input = %v, want %v", got, kubernetesCronRemapTransform)
	}

	// without a configured plain sink, the console default takes the unmatched branch
	if got := sinkInput(t, sinks, kubernetesDefaultSink); got != kubernetesRouterTransform+"._unmatched" {
		t.Errorf("default sink input = %v, want %v._unmatched", got, kubernetesRouterTransform)
	}
}

func TestUpdateVectorValuesBothSinks(t *testing.T) {
	values := setupVectorValuesTest(t, "console://?encoding[codec]=json", "console://?encoding[codec]=text")

	updated, err := updateVectorValues(values)
	if err != nil {
		t.Fatalf("updateVectorValues() error = %v", err)
	}

	sinks := vectorSinks(t, updated)
	if got := sinkInput(t, sinks, kubernetesGlobalSink); got != kubernetesRouterTransform+"._unmatched" {
		t.Errorf("global sink input = %v, want %v._unmatched", got, kubernetesRouterTransform)
	}
	if got := sinkInput(t, sinks, kubernetesCronSink); got != kubernetesCronRemapTransform {
		t.Errorf("cron sink input = %v, want %v", got, kubernetesCronRemapTransform)
	}
	if _, ok := sinks["prom_exporter"]; !ok {
		t.Error("prom_exporter should survive configuring both sinks")
	}
}

func TestUpdateVectorValuesInvalidCustomConfig(t *testing.T) {
	values := setupVectorValuesTest(t, "console://", "")
	delete(values, "customConfig")

	if _, err := updateVectorValues(values); err == nil {
		t.Fatal("updateVectorValues() expected an error for missing customConfig")
	}
}
