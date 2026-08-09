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

func renderIngressTemplate(t *testing.T, values map[string]interface{}) []map[string]interface{} {
	t.Helper()

	chartDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(chartDir, "templates"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	chartYAML := []byte("apiVersion: v2\nname: test\nversion: 0.0.1\n")
	if err := os.WriteFile(filepath.Join(chartDir, "Chart.yaml"), chartYAML, 0o644); err != nil {
		t.Fatalf("write Chart.yaml: %v", err)
	}

	ingressTpl, err := templates.ReadFile("templates/chart/ingress.yaml")
	if err != nil {
		t.Fatalf("read ingress template: %v", err)
	}
	if err := os.WriteFile(filepath.Join(chartDir, "templates", "ingress.yaml"), ingressTpl, 0o644); err != nil {
		t.Fatalf("write ingress template: %v", err)
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

	manifest := rendered["test/templates/ingress.yaml"]

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

func testIngressValues(tlsEnabled bool, domains ...map[string]interface{}) map[string]interface{} {
	domainValues := []interface{}{}
	for _, domain := range domains {
		domainValues = append(domainValues, domain)
	}

	web := map[string]interface{}{
		"domains": domainValues,
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
				"ingress_class": "nginx",
			},
		},
		"processes": map[string]interface{}{
			"web": map[string]interface{}{
				"web": web,
			},
		},
	}
}

func ingressHost(t *testing.T, doc map[string]interface{}) string {
	t.Helper()

	rules, ok := doc["spec"].(map[string]interface{})["rules"].([]interface{})
	if !ok || len(rules) != 1 {
		t.Fatalf("expected ingress to contain a single rule, got %#v", doc["spec"])
	}

	host, ok := rules[0].(map[string]interface{})["host"].(string)
	if !ok {
		t.Fatalf("expected rule to contain a string host, got %#v", rules[0])
	}

	return host
}

func TestIngressTemplateExactDomainUsesSlugForObjectName(t *testing.T) {
	values := testIngressValues(false, map[string]interface{}{"name": "app.example.com", "slug": "app-example-com"})

	docs := renderIngressTemplate(t, values)
	if len(docs) != 1 {
		t.Fatalf("expected 1 ingress, got %d", len(docs))
	}

	doc := findDocByName(t, docs, "myapp-web-app-example-com")
	if got := ingressHost(t, doc); got != "app.example.com" {
		t.Fatalf("expected host app.example.com, got %q", got)
	}
}

func TestIngressTemplateWildcardDomainRendersWildcardHost(t *testing.T) {
	values := testIngressValues(true, map[string]interface{}{"name": "*.example.com", "slug": "wildcard-example-com"})

	docs := renderIngressTemplate(t, values)
	if len(docs) != 1 {
		t.Fatalf("expected 1 ingress, got %d", len(docs))
	}

	doc := findDocByName(t, docs, "myapp-web-wildcard-example-com")
	if got := ingressHost(t, doc); got != "*.example.com" {
		t.Fatalf("expected host *.example.com, got %q", got)
	}

	tls, ok := doc["spec"].(map[string]interface{})["tls"].([]interface{})
	if !ok || len(tls) != 1 {
		t.Fatalf("expected ingress to contain a single tls entry, got %#v", doc["spec"])
	}

	hosts, ok := tls[0].(map[string]interface{})["hosts"].([]interface{})
	if !ok || len(hosts) != 1 || hosts[0] != "*.example.com" {
		t.Fatalf("expected tls hosts [*.example.com], got %#v", tls[0])
	}
	if got := tls[0].(map[string]interface{})["secretName"]; got != "tls-myapp-web" {
		t.Fatalf("expected tls secretName tls-myapp-web, got %#v", got)
	}
}

func TestIngressTemplateWildcardAndApexRenderDistinctObjects(t *testing.T) {
	values := testIngressValues(false,
		map[string]interface{}{"name": "*.example.com", "slug": "wildcard-example-com"},
		map[string]interface{}{"name": "example.com", "slug": "example-com"},
	)

	docs := renderIngressTemplate(t, values)
	if len(docs) != 2 {
		t.Fatalf("expected 2 ingresses (one per domain), got %d", len(docs))
	}

	wildcard := findDocByName(t, docs, "myapp-web-wildcard-example-com")
	if got := ingressHost(t, wildcard); got != "*.example.com" {
		t.Fatalf("expected wildcard host *.example.com, got %q", got)
	}

	apex := findDocByName(t, docs, "myapp-web-example-com")
	if got := ingressHost(t, apex); got != "example.com" {
		t.Fatalf("expected apex host example.com, got %q", got)
	}
}

func TestIngressTemplateNonNginxIngressClassRendersNothing(t *testing.T) {
	values := testIngressValues(false, map[string]interface{}{"name": "app.example.com", "slug": "app-example-com"})
	values["global"].(map[string]interface{})["network"].(map[string]interface{})["ingress_class"] = "traefik"

	docs := renderIngressTemplate(t, values)
	if len(docs) != 0 {
		t.Fatalf("expected 0 ingresses when ingress_class is not nginx, got %d: %#v", len(docs), docs)
	}
}
