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
// byte label cap, is a fixed 40-character sha1 hex digest, and is
// deterministic. Without this the dokku.com/cron-hash label can exceed the
// cap and the Kubernetes API server rejects the manifest.
func TestCronIDLabelValue(t *testing.T) {
	cases := []string{
		"short",
		"app===echo hello-from-cron===5 5 5 5 5",
		strings.Repeat("a-very-long-cron-id-that-would-far-exceed-the-label-cap-", 10),
	}
	for _, in := range cases {
		out := cronIDLabelValue(in)
		if len(out) != 40 {
			t.Errorf("cronIDLabelValue(%q) = %q (len %d); want 40 hex chars", in, out, len(out))
		}
		if len(out) > 63 {
			t.Errorf("cronIDLabelValue(%q) = %q (len %d); must be <= 63", in, out, len(out))
		}
		if out != cronIDLabelValue(in) {
			t.Errorf("cronIDLabelValue(%q) is not deterministic", in)
		}
	}
}

// TestCronJobTemplateQuotesAllDigitSuffix asserts that the rendered cron-job
// manifest produces string-typed annotation and label values even when the
// suffix and cron-id consist entirely of digits. Without `| quote` in the
// template, YAML would coerce these to numbers and the manifest would be
// rejected by the Kubernetes API server.
func TestCronJobTemplateQuotesAllDigitSuffix(t *testing.T) {
	manifest := renderCronJobTemplate(t, map[string]interface{}{
		"id":                 "1234567890",
		"hash":               "0123456789abcdef0123456789abcdef01234567",
		"schedule":           "5 5 5 5 5",
		"suffix":             "1234567890",
		"suspend":            false,
		"concurrency_policy": "Allow",
	})

	forEachManifestMetadata(t, manifest, func(metadata map[string]interface{}) {
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
	})
}

// TestCronJobTemplateEmitsHashLabel asserts that the rendered manifest
// emits the sha1 hash as the dokku.com/cron-hash label and matching
// annotation, while keeping the original base36 cron-id in the
// dokku.com/cron-id annotation so cron:list can still surface it to users.
func TestCronJobTemplateEmitsHashLabel(t *testing.T) {
	originalID := "app===echo hello-from-cron===5 5 5 5 5"
	hashedID := cronIDLabelValue(originalID)

	manifest := renderCronJobTemplate(t, map[string]interface{}{
		"id":                 originalID,
		"hash":               hashedID,
		"schedule":           "5 5 5 5 5",
		"suffix":             "abcde",
		"suspend":            false,
		"concurrency_policy": "Allow",
	})

	sawHashLabel := false
	sawIDAnnotation := false
	forEachManifestMetadata(t, manifest, func(metadata map[string]interface{}) {
		if labels, ok := metadata["labels"].(map[string]interface{}); ok {
			if value, ok := labels["dokku.com/cron-hash"]; ok {
				sawHashLabel = true
				if value != hashedID {
					t.Errorf("dokku.com/cron-hash label = %q, want %q", value, hashedID)
				}
			}
		}
		annotations, _ := metadata["annotations"].(map[string]interface{})
		if value, ok := annotations["dokku.com/cron-hash"]; ok {
			if value != hashedID {
				t.Errorf("dokku.com/cron-hash annotation = %q, want %q (must mirror the label)", value, hashedID)
			}
		}
		if value, ok := annotations["dokku.com/cron-id"]; ok {
			sawIDAnnotation = true
			if value != originalID {
				t.Errorf("dokku.com/cron-id annotation = %q, want %q", value, originalID)
			}
		}
	})
	if !sawHashLabel {
		t.Errorf("rendered manifest did not include a dokku.com/cron-hash label")
	}
	if !sawIDAnnotation {
		t.Errorf("rendered manifest did not include a dokku.com/cron-id annotation")
	}
}

// TestCronJobTemplateRendersLongCronID is the direct regression test for
// dokku/dokku#8594: a cron-id well over the 63-byte Kubernetes label cap
// must render cleanly and every resulting label value must stay under the
// cap.
func TestCronJobTemplateRendersLongCronID(t *testing.T) {
	longCronID := strings.Repeat("a-very-long-cron-id-that-would-far-exceed-the-label-cap-", 10)
	manifest := renderCronJobTemplate(t, map[string]interface{}{
		"id":                 longCronID,
		"hash":               cronIDLabelValue(longCronID),
		"schedule":           "5 5 5 5 5",
		"suffix":             "abcde",
		"suspend":            false,
		"concurrency_policy": "Allow",
	})

	forEachManifestMetadata(t, manifest, func(metadata map[string]interface{}) {
		labels, _ := metadata["labels"].(map[string]interface{})
		for key, value := range labels {
			str, ok := value.(string)
			if !ok {
				t.Errorf("label %q has non-string value %v (type %T)", key, value, value)
				continue
			}
			if len(str) > 63 {
				t.Errorf("label %q value %q exceeds Kubernetes' 63-byte cap (len %d)", key, str, len(str))
			}
		}
	})
}

// renderCronJobTemplate renders templates/chart/cron-job.yaml with the
// supplied cron-process values and returns the YAML manifest as a string.
// The chart is materialised in a tempdir alongside the shared _helpers.tpl
// so the helm engine resolves named templates the same way it does in
// production.
func renderCronJobTemplate(t *testing.T, cronValues map[string]interface{}) string {
	t.Helper()

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
			"cron-process": map[string]interface{}{
				"args": []interface{}{"echo", "hello"},
				"cron": cronValues,
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

	for name, content := range rendered {
		if filepath.Base(name) == "cron-job.yaml" {
			return content
		}
	}
	t.Fatalf("cron-job.yaml not rendered; got: %v", rendered)
	return ""
}

// forEachManifestMetadata invokes fn on every metadata block in a multi-doc
// YAML manifest, including the nested jobTemplate and pod template metadata
// blocks inside a CronJob spec.
func forEachManifestMetadata(t *testing.T, manifest string, fn func(metadata map[string]interface{})) {
	t.Helper()

	decoder := yaml.NewDecoder(strings.NewReader(manifest))
	for {
		var doc map[string]interface{}
		if err := decoder.Decode(&doc); err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			t.Fatalf("yaml decode failed (would also fail in Kubernetes API): %v\nrendered:\n%s", err, manifest)
		}
		if doc == nil {
			continue
		}
		walkMetadata(doc, fn)
	}
}

func walkMetadata(node interface{}, fn func(metadata map[string]interface{})) {
	switch typed := node.(type) {
	case map[string]interface{}:
		if metadata, ok := typed["metadata"].(map[string]interface{}); ok {
			fn(metadata)
		}
		for _, value := range typed {
			walkMetadata(value, fn)
		}
	case []interface{}:
		for _, value := range typed {
			walkMetadata(value, fn)
		}
	}
}
