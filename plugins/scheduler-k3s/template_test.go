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

func TestToCoreV1PodSecurityContext(t *testing.T) {
	t.Run("nil when no sysctls are configured", func(t *testing.T) {
		securityContext := SecurityContext{Privileged: true}
		if got := securityContext.ToCoreV1PodSecurityContext(); got != nil {
			t.Errorf("ToCoreV1PodSecurityContext() = %v, want nil", got)
		}
	})

	t.Run("nil for an empty sysctl slice", func(t *testing.T) {
		securityContext := SecurityContext{Sysctls: []Sysctl{}}
		if got := securityContext.ToCoreV1PodSecurityContext(); got != nil {
			t.Errorf("ToCoreV1PodSecurityContext() = %v, want nil", got)
		}
	})

	t.Run("preserves order and values", func(t *testing.T) {
		securityContext := SecurityContext{
			Sysctls: []Sysctl{
				{Name: "kernel.sem", Value: "250"},
				{Name: "net.core.somaxconn", Value: "1024"},
			},
		}

		got := securityContext.ToCoreV1PodSecurityContext()
		if got == nil {
			t.Fatal("ToCoreV1PodSecurityContext() = nil, want a security context")
		}
		if len(got.Sysctls) != 2 {
			t.Fatalf("ToCoreV1PodSecurityContext() has %d sysctls, want 2", len(got.Sysctls))
		}
		if got.Sysctls[0].Name != "kernel.sem" || got.Sysctls[0].Value != "250" {
			t.Errorf("ToCoreV1PodSecurityContext() sysctls[0] = %v, want kernel.sem=250", got.Sysctls[0])
		}
		if got.Sysctls[1].Name != "net.core.somaxconn" || got.Sysctls[1].Value != "1024" {
			t.Errorf("ToCoreV1PodSecurityContext() sysctls[1] = %v, want net.core.somaxconn=1024", got.Sysctls[1])
		}
	})
}

func renderDeploymentTemplate(t *testing.T, globalValues map[string]interface{}) string {
	t.Helper()

	chartDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(chartDir, "templates"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	chartYAML := []byte("apiVersion: v2\nname: test\nversion: 0.0.1\n")
	if err := os.WriteFile(filepath.Join(chartDir, "Chart.yaml"), chartYAML, 0o644); err != nil {
		t.Fatalf("write Chart.yaml: %v", err)
	}

	for _, name := range []string{"deployment.yaml", "_helpers.tpl"} {
		contents, err := templates.ReadFile("templates/chart/" + name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if err := os.WriteFile(filepath.Join(chartDir, "templates", name), contents, 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	loaded, err := loader.Load(chartDir)
	if err != nil {
		t.Fatalf("load chart: %v", err)
	}

	global := map[string]interface{}{
		"app_name":      "myapp",
		"deployment_id": "1",
		"namespace":     "myapp",
		"image": map[string]interface{}{
			"name": "myapp:latest",
			"type": "dockerfile",
		},
	}
	for key, value := range globalValues {
		global[key] = value
	}

	values := map[string]interface{}{
		"global": global,
		"processes": map[string]interface{}{
			"worker": map[string]interface{}{
				"args":     []interface{}{"echo", "hello"},
				"replicas": 1,
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
		if filepath.Base(name) == "deployment.yaml" {
			return content
		}
	}
	t.Fatalf("deployment.yaml not rendered; got: %v", rendered)
	return ""
}

// TestDeploymentSysctlsRendering asserts the pod-level securityContext.sysctls
// block only appears when sysctls are configured, and that numeric values are
// quoted. Kubernetes types sysctls[].value as a string, so an unquoted 1024
// renders as a YAML integer and the API server rejects the manifest.
func TestDeploymentSysctlsRendering(t *testing.T) {
	t.Run("absent when no security context is set", func(t *testing.T) {
		manifest := renderDeploymentTemplate(t, map[string]interface{}{})
		if strings.Contains(manifest, "sysctls:") {
			t.Errorf("rendered deployment unexpectedly contains sysctls:\n%s", manifest)
		}
	})

	t.Run("absent when the security context has no sysctls", func(t *testing.T) {
		manifest := renderDeploymentTemplate(t, map[string]interface{}{
			"security_context": map[string]interface{}{"privileged": true},
		})
		if strings.Contains(manifest, "sysctls:") {
			t.Errorf("rendered deployment unexpectedly contains sysctls:\n%s", manifest)
		}
	})

	t.Run("numeric values are quoted", func(t *testing.T) {
		manifest := renderDeploymentTemplate(t, map[string]interface{}{
			"security_context": map[string]interface{}{
				"sysctls": []interface{}{
					map[string]interface{}{"name": "net.core.somaxconn", "value": "1024"},
				},
			},
		})

		if !strings.Contains(manifest, "- name: net.core.somaxconn") {
			t.Errorf("rendered deployment missing sysctl name:\n%s", manifest)
		}
		if !strings.Contains(manifest, `value: "1024"`) {
			t.Errorf("rendered deployment did not quote the sysctl value:\n%s", manifest)
		}
	})

	t.Run("sysctls land on the pod spec, not the container", func(t *testing.T) {
		manifest := renderDeploymentTemplate(t, map[string]interface{}{
			"security_context": map[string]interface{}{
				"sysctls": []interface{}{
					map[string]interface{}{"name": "net.core.somaxconn", "value": "1024"},
				},
			},
		})

		var doc map[string]interface{}
		decoder := yaml.NewDecoder(strings.NewReader(manifest))
		for {
			if err := decoder.Decode(&doc); err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				t.Fatalf("decode manifest: %v", err)
			}
			if doc == nil {
				continue
			}

			spec := doc["spec"].(map[string]interface{})
			template := spec["template"].(map[string]interface{})
			podSpec := template["spec"].(map[string]interface{})

			securityContext, ok := podSpec["securityContext"].(map[string]interface{})
			if !ok {
				t.Fatalf("pod spec has no securityContext:\n%s", manifest)
			}
			if _, ok := securityContext["sysctls"]; !ok {
				t.Errorf("pod securityContext has no sysctls:\n%s", manifest)
			}
		}
	})
}
