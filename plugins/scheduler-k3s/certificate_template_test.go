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

// renderCertificateChart renders the certificate.yaml and issuer.yaml templates
// together and returns the parsed documents for each, keyed by template file name.
func renderCertificateChart(t *testing.T, values map[string]interface{}) map[string][]map[string]interface{} {
	t.Helper()

	chartDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(chartDir, "templates"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	chartYAML := []byte("apiVersion: v2\nname: test\nversion: 0.0.1\n")
	if err := os.WriteFile(filepath.Join(chartDir, "Chart.yaml"), chartYAML, 0o644); err != nil {
		t.Fatalf("write Chart.yaml: %v", err)
	}

	for _, name := range []string{"certificate", "issuer", "_helpers"} {
		ext := ".yaml"
		if name == "_helpers" {
			ext = ".tpl"
		}
		tpl, err := templates.ReadFile("templates/chart/" + name + ext)
		if err != nil {
			t.Fatalf("read %s template: %v", name, err)
		}
		if err := os.WriteFile(filepath.Join(chartDir, "templates", name+ext), tpl, 0o644); err != nil {
			t.Fatalf("write %s template: %v", name, err)
		}
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

	result := map[string][]map[string]interface{}{}
	for _, name := range []string{"certificate", "issuer"} {
		manifest := rendered["test/templates/"+name+".yaml"]
		var docs []map[string]interface{}
		decoder := yaml.NewDecoder(strings.NewReader(manifest))
		for {
			var doc map[string]interface{}
			if err := decoder.Decode(&doc); err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				t.Fatalf("yaml decode of %s failed: %v\nrendered:\n%s", name, err, manifest)
			}
			if doc != nil {
				docs = append(docs, doc)
			}
		}
		result[name] = docs
	}

	return result
}

func testCertificateValues(issuerKind string, issuerName string, issuer map[string]interface{}) map[string]interface{} {
	global := map[string]interface{}{
		"app_name":  "myapp",
		"namespace": "myns",
	}
	if issuer != nil {
		global["issuer"] = issuer
	}

	web := map[string]interface{}{
		"domains": []interface{}{
			map[string]interface{}{"name": "app.example.com"},
		},
		"tls": map[string]interface{}{
			"enabled":           true,
			"use_imported_cert": false,
			"issuer_kind":       issuerKind,
			"issuer_name":       issuerName,
		},
	}

	return map[string]interface{}{
		"global": global,
		"processes": map[string]interface{}{
			"web": map[string]interface{}{
				"web": web,
			},
		},
	}
}

func certificateIssuerRef(t *testing.T, docs map[string][]map[string]interface{}) map[string]interface{} {
	t.Helper()
	certs := docs["certificate"]
	if len(certs) != 1 {
		t.Fatalf("expected 1 certificate document, got %d: %#v", len(certs), certs)
	}
	spec, ok := certs[0]["spec"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected certificate spec, got %#v", certs[0]["spec"])
	}
	issuerRef, ok := spec["issuerRef"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected certificate issuerRef, got %#v", spec["issuerRef"])
	}
	return issuerRef
}

func TestCertificateTemplateUsesSharedClusterIssuerByDefault(t *testing.T) {
	docs := renderCertificateChart(t, testCertificateValues("ClusterIssuer", "letsencrypt-prod", map[string]interface{}{"enabled": false}))

	issuerRef := certificateIssuerRef(t, docs)
	if issuerRef["kind"] != "ClusterIssuer" {
		t.Fatalf("expected issuerRef.kind ClusterIssuer, got %#v", issuerRef["kind"])
	}
	if issuerRef["name"] != "letsencrypt-prod" {
		t.Fatalf("expected issuerRef.name letsencrypt-prod, got %#v", issuerRef["name"])
	}

	if len(docs["issuer"]) != 0 {
		t.Fatalf("expected no Issuer document for shared ClusterIssuer, got %#v", docs["issuer"])
	}
}

func TestCertificateTemplateRendersWithoutIssuerValues(t *testing.T) {
	// Reproduces the deploy failure where global.issuer is absent from values.yaml
	// (the common case, no per-app email). The chart must still render without a
	// nil-pointer error, use the shared ClusterIssuer, and emit no Issuer document.
	docs := renderCertificateChart(t, testCertificateValues("ClusterIssuer", "letsencrypt-prod", nil))

	issuerRef := certificateIssuerRef(t, docs)
	if issuerRef["kind"] != "ClusterIssuer" {
		t.Fatalf("expected issuerRef.kind ClusterIssuer, got %#v", issuerRef["kind"])
	}
	if issuerRef["name"] != "letsencrypt-prod" {
		t.Fatalf("expected issuerRef.name letsencrypt-prod, got %#v", issuerRef["name"])
	}

	if len(docs["issuer"]) != 0 {
		t.Fatalf("expected no Issuer document when global.issuer is absent, got %#v", docs["issuer"])
	}
}

func TestCertificateTemplateDefaultsKindWhenIssuerKindEmpty(t *testing.T) {
	docs := renderCertificateChart(t, testCertificateValues("", "letsencrypt-prod", map[string]interface{}{"enabled": false}))

	issuerRef := certificateIssuerRef(t, docs)
	if issuerRef["kind"] != "ClusterIssuer" {
		t.Fatalf("expected empty issuer_kind to default to ClusterIssuer, got %#v", issuerRef["kind"])
	}
}

func TestCertificateTemplateRendersPerAppNamespacedIssuer(t *testing.T) {
	issuer := map[string]interface{}{
		"enabled":       true,
		"name":          "myapp-letsencrypt-stag",
		"email":         "app@dokku.me",
		"server":        LetsencryptServerStag,
		"ingress_class": "nginx",
	}
	docs := renderCertificateChart(t, testCertificateValues("Issuer", "myapp-letsencrypt-stag", issuer))

	issuerRef := certificateIssuerRef(t, docs)
	if issuerRef["kind"] != "Issuer" {
		t.Fatalf("expected issuerRef.kind Issuer, got %#v", issuerRef["kind"])
	}
	if issuerRef["name"] != "myapp-letsencrypt-stag" {
		t.Fatalf("expected issuerRef.name myapp-letsencrypt-stag, got %#v", issuerRef["name"])
	}

	issuers := docs["issuer"]
	if len(issuers) != 1 {
		t.Fatalf("expected 1 Issuer document, got %d: %#v", len(issuers), issuers)
	}
	if issuers[0]["kind"] != "Issuer" {
		t.Fatalf("expected kind Issuer, got %#v", issuers[0]["kind"])
	}

	metadata, _ := issuers[0]["metadata"].(map[string]interface{})
	if metadata["name"] != "myapp-letsencrypt-stag" {
		t.Fatalf("expected Issuer name myapp-letsencrypt-stag, got %#v", metadata["name"])
	}
	if metadata["namespace"] != "myns" {
		t.Fatalf("expected Issuer namespace myns, got %#v", metadata["namespace"])
	}

	spec, _ := issuers[0]["spec"].(map[string]interface{})
	acme, _ := spec["acme"].(map[string]interface{})
	if acme["email"] != "app@dokku.me" {
		t.Fatalf("expected Issuer acme email app@dokku.me, got %#v", acme["email"])
	}
	if acme["server"] != LetsencryptServerStag {
		t.Fatalf("expected Issuer acme server %q, got %#v", LetsencryptServerStag, acme["server"])
	}
}
