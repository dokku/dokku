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

func renderRedirectIngressTemplate(t *testing.T, values map[string]interface{}) []map[string]interface{} {
	t.Helper()

	chartDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(chartDir, "templates"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	chartYAML := []byte("apiVersion: v2\nname: test\nversion: 0.0.1\n")
	if err := os.WriteFile(filepath.Join(chartDir, "Chart.yaml"), chartYAML, 0o644); err != nil {
		t.Fatalf("write Chart.yaml: %v", err)
	}

	for _, name := range []string{"redirect-ingress.yaml", "_helpers.tpl"} {
		tpl, err := templates.ReadFile("templates/chart/" + name)
		if err != nil {
			t.Fatalf("read %s template: %v", name, err)
		}
		if err := os.WriteFile(filepath.Join(chartDir, "templates", name), tpl, 0o644); err != nil {
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

	manifest := rendered["test/templates/redirect-ingress.yaml"]

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

func testRedirectValues(ingressClass string, tlsEnabled bool, redirects ...map[string]interface{}) map[string]interface{} {
	redirectValues := []interface{}{}
	for _, redirect := range redirects {
		redirectValues = append(redirectValues, redirect)
	}

	web := map[string]interface{}{
		"domains": []interface{}{
			map[string]interface{}{"name": "app.example.com", "slug": "app-example-com"},
		},
		"port_maps": []interface{}{
			map[string]interface{}{
				"name":           "http-80-5000",
				"scheme":         "http",
				"container_port": 5000,
			},
		},
		"redirects": redirectValues,
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
				"ingress_class": ingressClass,
			},
		},
		"processes": map[string]interface{}{
			"web": map[string]interface{}{
				"web": web,
			},
		},
	}
}

func testRedirect(source string, destination string, code int) map[string]interface{} {
	return map[string]interface{}{
		"source":      source,
		"destination": destination,
		"code":        code,
		"permanent":   code == 301 || code == 308,
		"slug":        domainSlug(source),
	}
}

func docsOfKind(docs []map[string]interface{}, kind string) []map[string]interface{} {
	out := []map[string]interface{}{}
	for _, doc := range docs {
		if doc["kind"] == kind {
			out = append(out, doc)
		}
	}
	return out
}

func TestRedirectIngressTemplateNoRedirectsRendersNothing(t *testing.T) {
	for _, ingressClass := range []string{"nginx", "traefik"} {
		docs := renderRedirectIngressTemplate(t, testRedirectValues(ingressClass, true))
		if len(docs) != 0 {
			t.Fatalf("%s: expected no documents, got %d: %#v", ingressClass, len(docs), docs)
		}
	}
}

func TestRedirectIngressTemplateNginxRendersSnippetIngress(t *testing.T) {
	values := testRedirectValues("nginx", false,
		testRedirect("www.example.com", "app.example.com", 301),
		testRedirect("old.example.com", "app.example.com", 307),
	)

	docs := renderRedirectIngressTemplate(t, values)
	if len(docs) != 2 {
		t.Fatalf("expected 2 ingresses, got %d", len(docs))
	}

	doc := findDocByName(t, docs, "myapp-web-redirect-www-example-com")
	if got := ingressHost(t, doc); got != "www.example.com" {
		t.Fatalf("expected host www.example.com, got %q", got)
	}
	annotations := doc["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})
	snippet, _ := annotations["nginx.ingress.kubernetes.io/configuration-snippet"].(string)
	if strings.TrimSpace(snippet) != "return 301 http://app.example.com$request_uri;" {
		t.Fatalf("unexpected configuration-snippet %q", snippet)
	}
	if _, ok := doc["spec"].(map[string]interface{})["tls"]; ok {
		t.Fatalf("expected no tls block without tls enabled, got %#v", doc["spec"])
	}

	doc = findDocByName(t, docs, "myapp-web-redirect-old-example-com")
	annotations = doc["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})
	snippet, _ = annotations["nginx.ingress.kubernetes.io/configuration-snippet"].(string)
	if strings.TrimSpace(snippet) != "return 307 http://app.example.com$request_uri;" {
		t.Fatalf("unexpected configuration-snippet %q", snippet)
	}
}

func TestRedirectIngressTemplateNginxTLSUsesHTTPSAndAppSecret(t *testing.T) {
	values := testRedirectValues("nginx", true, testRedirect("www.example.com", "app.example.com", 302))

	docs := renderRedirectIngressTemplate(t, values)
	doc := findDocByName(t, docs, "myapp-web-redirect-www-example-com")
	annotations := doc["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})
	snippet, _ := annotations["nginx.ingress.kubernetes.io/configuration-snippet"].(string)
	if strings.TrimSpace(snippet) != "return 302 https://app.example.com$request_uri;" {
		t.Fatalf("unexpected configuration-snippet %q", snippet)
	}

	tls, ok := doc["spec"].(map[string]interface{})["tls"].([]interface{})
	if !ok || len(tls) != 1 {
		t.Fatalf("expected a single tls entry, got %#v", doc["spec"])
	}
	hosts := tls[0].(map[string]interface{})["hosts"].([]interface{})
	if len(hosts) != 1 || hosts[0] != "www.example.com" {
		t.Fatalf("expected tls hosts [www.example.com], got %#v", hosts)
	}
	if got := tls[0].(map[string]interface{})["secretName"]; got != "tls-myapp-web" {
		t.Fatalf("expected tls secretName tls-myapp-web, got %#v", got)
	}
}

func TestRedirectIngressTemplateTraefikRendersMiddlewareAndRoutes(t *testing.T) {
	values := testRedirectValues("traefik", true, testRedirect("www.example.com", "app.example.com", 301))

	docs := renderRedirectIngressTemplate(t, values)
	middlewares := docsOfKind(docs, "Middleware")
	if len(middlewares) != 1 {
		t.Fatalf("expected 1 middleware, got %d", len(middlewares))
	}
	redirectRegex := middlewares[0]["spec"].(map[string]interface{})["redirectRegex"].(map[string]interface{})
	if redirectRegex["regex"] != "^[^:]+://[^/]+(.*)" {
		t.Fatalf("unexpected regex %#v", redirectRegex["regex"])
	}
	if redirectRegex["replacement"] != "https://app.example.com${1}" {
		t.Fatalf("unexpected replacement %#v", redirectRegex["replacement"])
	}
	if redirectRegex["permanent"] != true {
		t.Fatalf("expected permanent redirect, got %#v", redirectRegex["permanent"])
	}

	routes := docsOfKind(docs, "IngressRoute")
	if len(routes) != 2 {
		t.Fatalf("expected web and websecure ingress routes, got %d", len(routes))
	}

	secure := findDocByName(t, routes, "myapp-web-redirect-www-example-com-websecure")
	spec := secure["spec"].(map[string]interface{})
	if got := spec["tls"].(map[string]interface{})["secretName"]; got != "tls-myapp-web" {
		t.Fatalf("expected tls secretName tls-myapp-web, got %#v", got)
	}
	route := spec["routes"].([]interface{})[0].(map[string]interface{})
	if route["match"] != "Host(`www.example.com`)" {
		t.Fatalf("unexpected match %#v", route["match"])
	}
	if names := middlewareNames(t, route); len(names) != 1 || names[0] != "myapp-web-redirect-www-example-com" {
		t.Fatalf("unexpected middlewares %#v", names)
	}

	plain := findDocByName(t, routes, "myapp-web-redirect-www-example-com")
	if _, ok := plain["spec"].(map[string]interface{})["tls"]; ok {
		t.Fatalf("expected no tls on web entrypoint route, got %#v", plain["spec"])
	}
}

func TestRedirectIngressTemplateTraefikTemporaryWithoutTLS(t *testing.T) {
	values := testRedirectValues("traefik", false, testRedirect("www.example.com", "app.example.com", 307))

	docs := renderRedirectIngressTemplate(t, values)
	if routes := docsOfKind(docs, "IngressRoute"); len(routes) != 1 {
		t.Fatalf("expected only a web entrypoint ingress route, got %d", len(routes))
	}
	middleware := docsOfKind(docs, "Middleware")[0]
	redirectRegex := middleware["spec"].(map[string]interface{})["redirectRegex"].(map[string]interface{})
	if redirectRegex["replacement"] != "http://app.example.com${1}" {
		t.Fatalf("unexpected replacement %#v", redirectRegex["replacement"])
	}
	if redirectRegex["permanent"] != false {
		t.Fatalf("expected temporary redirect, got %#v", redirectRegex["permanent"])
	}
}

func TestRedirectIngressTemplateOtherIngressClassRendersNothing(t *testing.T) {
	values := testRedirectValues("haproxy", true, testRedirect("www.example.com", "app.example.com", 301))

	if docs := renderRedirectIngressTemplate(t, values); len(docs) != 0 {
		t.Fatalf("expected no documents, got %d", len(docs))
	}
}

func TestCertificateTemplateIncludesRedirectSources(t *testing.T) {
	values := testCertificateValues("ClusterIssuer", "letsencrypt-prod", nil)
	web := values["processes"].(map[string]interface{})["web"].(map[string]interface{})["web"].(map[string]interface{})
	web["redirects"] = []interface{}{testRedirect("www.example.com", "app.example.com", 301)}

	docs := renderCertificateChart(t, values)
	certs := docs["certificate"]
	if len(certs) != 1 {
		t.Fatalf("expected 1 certificate, got %d", len(certs))
	}
	dnsNames := certs[0]["spec"].(map[string]interface{})["dnsNames"].([]interface{})
	if len(dnsNames) != 2 || dnsNames[0] != "app.example.com" || dnsNames[1] != "www.example.com" {
		t.Fatalf("expected dnsNames [app.example.com www.example.com], got %#v", dnsNames)
	}
}

func TestCertificateTemplateRendersWithOnlyRedirectSources(t *testing.T) {
	values := testCertificateValues("ClusterIssuer", "letsencrypt-prod", nil)
	web := values["processes"].(map[string]interface{})["web"].(map[string]interface{})["web"].(map[string]interface{})
	web["domains"] = []interface{}{}
	web["redirects"] = []interface{}{testRedirect("www.example.com", "app.example.com", 301)}

	docs := renderCertificateChart(t, values)
	certs := docs["certificate"]
	if len(certs) != 1 {
		t.Fatalf("expected 1 certificate, got %d", len(certs))
	}
	dnsNames := certs[0]["spec"].(map[string]interface{})["dnsNames"].([]interface{})
	if len(dnsNames) != 1 || dnsNames[0] != "www.example.com" {
		t.Fatalf("expected dnsNames [www.example.com], got %#v", dnsNames)
	}
}
