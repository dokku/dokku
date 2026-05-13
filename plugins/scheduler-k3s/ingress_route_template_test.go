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

func renderIngressRouteTemplate(t *testing.T, values map[string]interface{}) []map[string]interface{} {
	t.Helper()

	chartDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(chartDir, "templates"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	chartYAML := []byte("apiVersion: v2\nname: test\nversion: 0.0.1\n")
	if err := os.WriteFile(filepath.Join(chartDir, "Chart.yaml"), chartYAML, 0o644); err != nil {
		t.Fatalf("write Chart.yaml: %v", err)
	}

	ingressRouteTpl, err := templates.ReadFile("templates/chart/ingress-route.yaml")
	if err != nil {
		t.Fatalf("read ingress-route template: %v", err)
	}
	if err := os.WriteFile(filepath.Join(chartDir, "templates", "ingress-route.yaml"), ingressRouteTpl, 0o644); err != nil {
		t.Fatalf("write ingress-route template: %v", err)
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

	renderValues, err := chartutil.ToRenderValues(loaded, values, chartutil.ReleaseOptions{Name: "test", Namespace: "default"}, nil)
	if err != nil {
		t.Fatalf("ToRenderValues: %v", err)
	}

	rendered, err := engine.Render(loaded, renderValues)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	manifest, ok := rendered["test/templates/ingress-route.yaml"]
	if !ok {
		t.Fatalf("ingress-route.yaml not rendered; got: %v", rendered)
	}

	var docs []map[string]interface{}
	decoder := yaml.NewDecoder(strings.NewReader(manifest))
	for {
		var doc map[string]interface{}
		if err := decoder.Decode(&doc); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			t.Fatalf("yaml decode failed: %v\nrendered:\n%s", err, manifest)
		}
		if doc != nil {
			docs = append(docs, doc)
		}
	}

	return docs
}

func testIngressRouteValues(tlsEnabled bool) map[string]interface{} {
	web := map[string]interface{}{
		"domains": []interface{}{
			map[string]interface{}{"name": "app.example.com"},
		},
		"port_maps": []interface{}{
			map[string]interface{}{
				"name":           "http-80-5000",
				"scheme":         "http",
				"container_port": 5000,
			},
		},
		"tls": map[string]interface{}{
			"enabled":           tlsEnabled,
			"use_imported_cert": false,
		},
	}

	return map[string]interface{}{
		"global": map[string]interface{}{
			"app_name":  "myapp",
			"namespace": "myns",
			"network": map[string]interface{}{
				"ingress_class": "traefik",
			},
		},
		"processes": map[string]interface{}{
			"web": map[string]interface{}{
				"web": web,
			},
		},
	}
}

func findDocByName(t *testing.T, docs []map[string]interface{}, name string) map[string]interface{} {
	t.Helper()
	for _, doc := range docs {
		metadata, _ := doc["metadata"].(map[string]interface{})
		if metadata["name"] == name {
			return doc
		}
	}
	t.Fatalf("document %q not found in %#v", name, docs)
	return nil
}

func TestIngressRouteTemplateTLSCreatesSeparateHTTPAndHTTPSRoutes(t *testing.T) {
	docs := renderIngressRouteTemplate(t, testIngressRouteValues(true))
	if len(docs) != 2 {
		t.Fatalf("expected 2 ingress routes when tls enabled, got %d", len(docs))
	}

	httpRoute := findDocByName(t, docs, "myapp-web-http-80-5000")
	httpSpec := httpRoute["spec"].(map[string]interface{})
	if got := httpSpec["entryPoints"]; len(got.([]interface{})) != 1 || got.([]interface{})[0] != "web" {
		t.Fatalf("expected HTTP route entryPoints [web], got %#v", got)
	}
	httpRoutes := httpSpec["routes"].([]interface{})
	httpRoute0 := httpRoutes[0].(map[string]interface{})
	if _, ok := httpRoute0["middlewares"]; !ok {
		t.Fatalf("expected HTTP route to contain redirect middleware, got %#v", httpRoute0)
	}
	if _, ok := httpSpec["tls"]; ok {
		t.Fatalf("expected HTTP route to omit tls block, got %#v", httpSpec["tls"])
	}

	httpsRoute := findDocByName(t, docs, "myapp-web-http-80-5000-websecure")
	httpsSpec := httpsRoute["spec"].(map[string]interface{})
	if got := httpsSpec["entryPoints"]; len(got.([]interface{})) != 1 || got.([]interface{})[0] != "websecure" {
		t.Fatalf("expected HTTPS route entryPoints [websecure], got %#v", got)
	}
	httpsRoutes := httpsSpec["routes"].([]interface{})
	httpsRoute0 := httpsRoutes[0].(map[string]interface{})
	if _, ok := httpsRoute0["middlewares"]; ok {
		t.Fatalf("expected HTTPS route to omit redirect middleware, got %#v", httpsRoute0["middlewares"])
	}
	tls, ok := httpsSpec["tls"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected HTTPS route to include tls block, got %#v", httpsSpec["tls"])
	}
	if tls["secretName"] != "tls-myapp-web" {
		t.Fatalf("expected HTTPS route secretName %q, got %#v", "tls-myapp-web", tls["secretName"])
	}
}

func TestIngressRouteTemplateWithoutTLSKeepsSingleHTTPRoute(t *testing.T) {
	docs := renderIngressRouteTemplate(t, testIngressRouteValues(false))
	if len(docs) != 1 {
		t.Fatalf("expected 1 ingress route when tls disabled, got %d", len(docs))
	}

	httpRoute := findDocByName(t, docs, "myapp-web-http-80-5000")
	httpSpec := httpRoute["spec"].(map[string]interface{})
	if got := httpSpec["entryPoints"]; len(got.([]interface{})) != 1 || got.([]interface{})[0] != "web" {
		t.Fatalf("expected HTTP route entryPoints [web], got %#v", got)
	}
	httpRoutes := httpSpec["routes"].([]interface{})
	httpRoute0 := httpRoutes[0].(map[string]interface{})
	if _, ok := httpRoute0["middlewares"]; ok {
		t.Fatalf("expected non-TLS route to omit redirect middleware, got %#v", httpRoute0["middlewares"])
	}
	if _, ok := httpSpec["tls"]; ok {
		t.Fatalf("expected non-TLS route to omit tls block, got %#v", httpSpec["tls"])
	}
}
