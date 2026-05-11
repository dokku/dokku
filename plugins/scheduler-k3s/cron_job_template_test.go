package scheduler_k3s

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
)

// TestCronIDLabelValue asserts the hashed cron ID fits inside Kubernetes' 63
// byte label cap and is deterministic. Without this the label can exceed the
// cap and the Kubernetes API server rejects the manifest.
func TestCronIDLabelValue(t *testing.T) {
	cases := []string{
		"short",
		"app===echo hello-from-cron===5 5 5 5 5",
		strings.Repeat("a-very-long-cron-id-that-would-far-exceed-the-label-cap-", 10),
	}
	for _, in := range cases {
		out := cronIDLabelValue(in)
		if len(out) > 63 {
			t.Errorf("cronIDLabelValue(%q) = %q (len %d); must be <= 63", in, out, len(out))
		}
		if out != cronIDLabelValue(in) {
			t.Errorf("cronIDLabelValue(%q) is not deterministic", in)
		}
	}
}

// TestCronJobTemplateQuotesAllDigitSuffix asserts that the rendered cron-job
// manifest produces string-typed annotation values even when the suffix and
// cron-id consist entirely of digits. Without `| quote` in the template, YAML
// would coerce these to numbers and the manifest would be rejected by the
// Kubernetes API server.
func TestCronJobTemplateQuotesAllDigitSuffix(t *testing.T) {
	chartDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(chartDir, "templates"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	chartYAML := []byte("apiVersion: v2\nname: test\nversion: 0.0.1\n")
	if err := os.WriteFile(filepath.Join(chartDir, "Chart.yaml"), chartYAML, 0o644); err != nil {
		t.Fatalf("write Chart.yaml: %v", err)
	}

	cronJobTpl, err := templates.ReadFile("templates/chart/cron-job.yaml")
	if err != nil {
		t.Fatalf("read cron-job template: %v", err)
	}
	if err := os.WriteFile(filepath.Join(chartDir, "templates", "cron-job.yaml"), cronJobTpl, 0o644); err != nil {
		t.Fatalf("write cron-job template: %v", err)
	}

	helpersTpl, err := templates.ReadFile("templates/chart/_helpers.tpl")
	if err != nil {
		t.Fatalf("read _helpers: %v", err)
	}
	if err := os.WriteFile(filepath.Join(chartDir, "templates", "_helpers.tpl"), helpersTpl, 0o644); err != nil {
		t.Fatalf("write _helpers: %v", err)
	}

	loaded, err := loader.Load(chartDir)
	if err != nil {
		t.Fatalf("load chart: %v", err)
	}

	values := map[string]interface{}{
		"global": map[string]interface{}{
			"app_name":      "myapp",
			"deployment_id": "1",
			"namespace":     "myapp",
			"image": map[string]interface{}{
				"name": "myapp:latest",
				"type": "dockerfile",
			},
		},
		"processes": map[string]interface{}{
			"cron-id-123": map[string]interface{}{
				"args": []interface{}{"echo", "hello"},
				"cron": map[string]interface{}{
					"id":                 "1234567890",
					"hash":               "abc123def",
					"schedule":           "5 5 5 5 5",
					"suffix":             "1234567890",
					"suspend":            false,
					"concurrency_policy": "Allow",
				},
			},
		},
	}

	renderValues, err := chartutil.ToRenderValues(loaded, values, chartutil.ReleaseOptions{Name: "test", Namespace: "default"}, nil)
	if err != nil {
		t.Fatalf("ToRenderValues: %v", err)
	}

	rendered, err := engine.Render(loaded, renderValues)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	var manifest string
	for name, content := range rendered {
		if filepath.Base(name) == "cron-job.yaml" {
			manifest = content
			break
		}
	}
	if manifest == "" {
		t.Fatalf("cron-job.yaml not rendered; got: %v", rendered)
	}

	decoder := yaml.NewDecoder(strings.NewReader(manifest))
	for {
		var doc map[string]interface{}
		if err := decoder.Decode(&doc); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			t.Fatalf("yaml decode failed (would also fail in Kubernetes API): %v\nrendered:\n%s", err, manifest)
		}
		if doc == nil {
			continue
		}
		metadata, ok := doc["metadata"].(map[string]interface{})
		if !ok {
			continue
		}
		annotations, _ := metadata["annotations"].(map[string]interface{})
		for key, value := range annotations {
			if _, isString := value.(string); !isString {
				t.Errorf("annotation %q has non-string value %v (type %T); helm template must apply | quote", key, value, value)
			}
		}
		labels, _ := metadata["labels"].(map[string]interface{})
		for key, value := range labels {
			if _, isString := value.(string); !isString {
				t.Errorf("label %q has non-string value %v (type %T); helm template must apply | quote", key, value, value)
			}
		}
	}
}
