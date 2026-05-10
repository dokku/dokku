package nginxvhosts

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gliderlabs/sigil"
	_ "github.com/gliderlabs/sigil/builtin"
)

func templatePath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(file), "templates", "nginx.conf.sigil")
}

func defaultVars() map[string]interface{} {
	return map[string]interface{}{
		"APP":                          "app",
		"DOKKU_ROOT":                   "/home/dokku",
		"DOKKU_LIB_ROOT":               "/var/lib/dokku",
		"NOSSL_SERVER_NAME":            "app.example.com",
		"SSL_SERVER_NAME":              "",
		"SSL_INUSE":                    "",
		"APP_SSL_PATH":                 "/home/dokku/app/tls",
		"DOKKU_APP_WEB_LISTENERS":      "127.0.0.1:5000",
		"DOKKU_APP_WEB_LISTENER_HOST":  "app.web.docker",
		"PROXY_PORT_MAP":               "http:80:5000",
		"PROXY_UPSTREAM_PORTS":         "5000",
		"PROXY_PORT":                   "80",
		"PROXY_SSL_PORT":               "443",
		"PROXY_KEEPALIVE":              "",
		"NGINX_BIND_ADDRESS_IP4":       "",
		"NGINX_BIND_ADDRESS_IP6":       "::",
		"NGINX_DNS_RESOLVER":           "127.0.0.1:1053",
		"NGINX_DNS_ZONE":               "docker",
		"NGINX_ACCESS_LOG_PATH":        "/var/log/nginx/app-access.log",
		"NGINX_ACCESS_LOG_FORMAT":      "",
		"NGINX_ERROR_LOG_PATH":         "/var/log/nginx/app-error.log",
		"NGINX_UNDERSCORE_IN_HEADERS":  "off",
		"CLIENT_BODY_TIMEOUT":          "60s",
		"CLIENT_HEADER_TIMEOUT":        "60s",
		"CLIENT_MAX_BODY_SIZE":         "1m",
		"KEEPALIVE_TIMEOUT":            "75s",
		"LINGERING_TIMEOUT":            "5s",
		"SEND_TIMEOUT":                 "60s",
		"PROXY_CONNECT_TIMEOUT":        "60s",
		"PROXY_READ_TIMEOUT":           "60s",
		"PROXY_SEND_TIMEOUT":           "60s",
		"PROXY_BUFFER_SIZE":            "4k",
		"PROXY_BUFFERING":              "on",
		"PROXY_BUFFERS":                "8 4k",
		"PROXY_BUSY_BUFFERS_SIZE":      "8k",
		"PROXY_X_FORWARDED_FOR":        "$remote_addr",
		"PROXY_X_FORWARDED_PORT":       "$server_port",
		"PROXY_X_FORWARDED_PROTO":      "$scheme",
		"PROXY_X_FORWARDED_SSL":        "",
		"HTTP2_DIRECTIVE_SUPPORTED":    "true",
	}
}

func renderTemplate(t *testing.T, vars map[string]interface{}) string {
	t.Helper()
	data, err := os.ReadFile(templatePath(t))
	if err != nil {
		t.Fatalf("read template: %v", err)
	}
	buf, err := sigil.Execute(data, vars, "nginx.conf.sigil")
	if err != nil {
		t.Fatalf("sigil.Execute: %v", err)
	}
	return buf.String()
}

func mustContain(t *testing.T, out, needle string) {
	t.Helper()
	if !strings.Contains(out, needle) {
		t.Errorf("expected output to contain %q\n--- output ---\n%s", needle, out)
	}
}

func mustNotContain(t *testing.T, out, needle string) {
	t.Helper()
	if strings.Contains(out, needle) {
		t.Errorf("expected output NOT to contain %q\n--- output ---\n%s", needle, out)
	}
}

func TestTemplate_HTTPOnlyBasicProxy(t *testing.T) {
	out := renderTemplate(t, defaultVars())
	mustContain(t, out, "listen      [::]:80;")
	mustContain(t, out, "resolver 127.0.0.1:1053 valid=10s ipv6=off;")
	mustContain(t, out, `set $dokku_upstream "app.web.docker:5000";`)
	mustContain(t, out, "proxy_pass http://$dokku_upstream;")
	mustNotContain(t, out, "upstream app-5000 {")
	mustContain(t, out, "error_page 500 501 502 503")
	mustContain(t, out, "server_name app.example.com;")
	mustNotContain(t, out, "ssl_certificate")
	mustNotContain(t, out, "return 301 https")
	mustNotContain(t, out, "http2_push_preload")
}

func TestTemplate_HTTPOnlyBasicProxyWithResolverOff(t *testing.T) {
	v := defaultVars()
	v["NGINX_DNS_RESOLVER"] = "off"
	out := renderTemplate(t, v)
	mustContain(t, out, "listen      [::]:80;")
	mustContain(t, out, "proxy_pass  http://app-5000;")
	mustContain(t, out, "upstream app-5000 {")
	mustNotContain(t, out, "resolver 127.0.0.1:1053")
	mustNotContain(t, out, "$dokku_upstream")
}

func TestTemplate_HTTPSEmitsPushPreloadWhenSupported(t *testing.T) {
	v := defaultVars()
	v["PROXY_PORT_MAP"] = "https:443:5000"
	v["SSL_INUSE"] = "true"
	v["SSL_SERVER_NAME"] = "app.example.com"
	v["HTTP2_PUSH_SUPPORTED"] = "true"
	out := renderTemplate(t, v)
	mustContain(t, out, "http2_push_preload on;")
}

func TestTemplate_HTTPSOmitsPushPreloadWhenUnsupported(t *testing.T) {
	v := defaultVars()
	v["PROXY_PORT_MAP"] = "https:443:5000"
	v["SSL_INUSE"] = "true"
	v["SSL_SERVER_NAME"] = "app.example.com"
	v["HTTP2_PUSH_SUPPORTED"] = "false"
	out := renderTemplate(t, v)
	mustNotContain(t, out, "http2_push_preload")
}

func TestTemplate_HTTPRedirectsToHTTPSWhenSSLInUse(t *testing.T) {
	v := defaultVars()
	v["PROXY_PORT_MAP"] = "http:80:5000 https:443:5000"
	v["SSL_INUSE"] = "true"
	v["SSL_SERVER_NAME"] = "app.example.com"
	v["NGINX_DNS_RESOLVER"] = "off"
	out := renderTemplate(t, v)

	port80, _, ok := strings.Cut(out, "listen      [::]:443")
	if !ok {
		t.Fatalf("expected an https server block in output:\n%s", out)
	}
	mustContain(t, port80, "return 301 https://$host:443$request_uri;")
	mustContain(t, port80, "include /home/dokku/app/nginx.conf.d/*.conf;")
	mustNotContain(t, port80, "proxy_pass  http://app-5000;")

	mustContain(t, out, "ssl_certificate           /home/dokku/app/tls/server.crt;")
	mustContain(t, out, "proxy_pass  http://app-5000;")
}

func TestTemplate_HTTPSWithHTTP2Directive(t *testing.T) {
	v := defaultVars()
	v["PROXY_PORT_MAP"] = "https:443:5000"
	v["SSL_INUSE"] = "true"
	v["SSL_SERVER_NAME"] = "app.example.com"
	v["HTTP2_DIRECTIVE_SUPPORTED"] = "true"
	out := renderTemplate(t, v)
	mustContain(t, out, "listen      [::]:443 ssl;")
	mustContain(t, out, "http2 on;")
	mustNotContain(t, out, "listen      [::]:443 ssl http2;")
}

func TestTemplate_HTTPSWithHTTP2Parameter(t *testing.T) {
	v := defaultVars()
	v["PROXY_PORT_MAP"] = "https:443:5000"
	v["SSL_INUSE"] = "true"
	v["SSL_SERVER_NAME"] = "app.example.com"
	v["HTTP2_DIRECTIVE_SUPPORTED"] = "false"
	out := renderTemplate(t, v)
	mustContain(t, out, "listen      [::]:443 ssl http2;")
	mustNotContain(t, out, "http2 on;")
}

func TestTemplate_NoWebListenersReturns502(t *testing.T) {
	v := defaultVars()
	v["DOKKU_APP_WEB_LISTENERS"] = ""
	out := renderTemplate(t, v)
	mustContain(t, out, "return 502;")
	mustNotContain(t, out, "proxy_pass http://$dokku_upstream;")
	mustNotContain(t, out, "proxy_pass  http://app-5000;")
}

func TestTemplate_GRPCNoSSL(t *testing.T) {
	v := defaultVars()
	v["PROXY_PORT_MAP"] = "grpc:50051:50051"
	v["PROXY_UPSTREAM_PORTS"] = "50051"
	v["NGINX_DNS_RESOLVER"] = "off"
	out := renderTemplate(t, v)
	mustContain(t, out, "grpc_pass  grpc://app-50051;")
	mustContain(t, out, "http2")
	mustNotContain(t, out, "ssl_certificate")
}

func TestTemplate_GRPCNoSSLWithResolver(t *testing.T) {
	v := defaultVars()
	v["PROXY_PORT_MAP"] = "grpc:50051:50051"
	v["PROXY_UPSTREAM_PORTS"] = "50051"
	out := renderTemplate(t, v)
	mustContain(t, out, "resolver 127.0.0.1:1053 valid=10s ipv6=off;")
	mustContain(t, out, `set $dokku_upstream "app.web.docker:50051";`)
	mustContain(t, out, "grpc_pass grpc://$dokku_upstream;")
}

func TestTemplate_GRPCS(t *testing.T) {
	v := defaultVars()
	v["PROXY_PORT_MAP"] = "grpcs:443:50051"
	v["PROXY_UPSTREAM_PORTS"] = "50051"
	v["SSL_INUSE"] = "true"
	v["SSL_SERVER_NAME"] = "app.example.com"
	v["NGINX_DNS_RESOLVER"] = "off"
	out := renderTemplate(t, v)
	mustContain(t, out, "ssl_certificate           /home/dokku/app/tls/server.crt;")
	mustContain(t, out, "grpc_pass  grpc://app-50051;")
}

func TestTemplate_GRPCSkippedWithoutListeners(t *testing.T) {
	v := defaultVars()
	v["PROXY_PORT_MAP"] = "grpc:50051:50051"
	v["PROXY_UPSTREAM_PORTS"] = "50051"
	v["DOKKU_APP_WEB_LISTENERS"] = ""
	out := renderTemplate(t, v)
	mustNotContain(t, out, "grpc_pass")
	mustNotContain(t, out, "server {")
}

func TestTemplate_BindAddressIPv4(t *testing.T) {
	v := defaultVars()
	v["NGINX_BIND_ADDRESS_IP4"] = "127.0.0.1"
	out := renderTemplate(t, v)
	mustContain(t, out, "listen      127.0.0.1:80;")

	v2 := defaultVars()
	out2 := renderTemplate(t, v2)
	mustNotContain(t, out2, "listen      127.0.0.1")
	mustNotContain(t, out2, ":80;\n  listen      :80;")
}

func TestTemplate_UpstreamBlock(t *testing.T) {
	v := defaultVars()
	v["PROXY_PORT_MAP"] = "http:80:5000 http:8080:5001"
	v["PROXY_UPSTREAM_PORTS"] = "5000 5001"
	v["DOKKU_APP_WEB_LISTENERS"] = "10.0.0.1:5000 10.0.0.2:5000"
	v["NGINX_DNS_RESOLVER"] = "off"
	out := renderTemplate(t, v)
	mustContain(t, out, "upstream app-5000 {")
	mustContain(t, out, "upstream app-5001 {")
	mustContain(t, out, "server 10.0.0.1:5000;")
	mustContain(t, out, "server 10.0.0.2:5000;")
	mustContain(t, out, "server 10.0.0.1:5001;")
	mustContain(t, out, "server 10.0.0.2:5001;")
}

func TestTemplate_UpstreamBlockOmittedInResolverMode(t *testing.T) {
	v := defaultVars()
	v["PROXY_PORT_MAP"] = "http:80:5000 http:8080:5001"
	v["PROXY_UPSTREAM_PORTS"] = "5000 5001"
	v["DOKKU_APP_WEB_LISTENERS"] = "10.0.0.1:5000 10.0.0.2:5000"
	out := renderTemplate(t, v)
	mustNotContain(t, out, "upstream app-5000 {")
	mustNotContain(t, out, "upstream app-5001 {")
	mustNotContain(t, out, "server 10.0.0.1:5000;")
}

func TestTemplate_UpstreamWithKeepalive(t *testing.T) {
	v := defaultVars()
	v["PROXY_KEEPALIVE"] = "16"
	v["NGINX_DNS_RESOLVER"] = "off"
	out := renderTemplate(t, v)
	mustContain(t, out, "keepalive 16;")

	v2 := defaultVars()
	v2["NGINX_DNS_RESOLVER"] = "off"
	out2 := renderTemplate(t, v2)
	mustNotContain(t, out2, "keepalive 16;")
}

func TestTemplate_AccessLogFormat(t *testing.T) {
	v := defaultVars()
	v["NGINX_ACCESS_LOG_FORMAT"] = "json"
	v["NGINX_ACCESS_LOG_PATH"] = "/var/log/nginx/app-access.log"
	out := renderTemplate(t, v)
	mustContain(t, out, "access_log  /var/log/nginx/app-access.log json;")

	v2 := defaultVars()
	v2["NGINX_ACCESS_LOG_FORMAT"] = "json"
	v2["NGINX_ACCESS_LOG_PATH"] = "off"
	out2 := renderTemplate(t, v2)
	mustContain(t, out2, "access_log  off;")
	mustNotContain(t, out2, "access_log  off json;")
}

func TestTemplate_XForwardedSSL(t *testing.T) {
	v := defaultVars()
	v["PROXY_X_FORWARDED_SSL"] = "on"
	out := renderTemplate(t, v)
	mustContain(t, out, "proxy_set_header X-Forwarded-Ssl on;")

	v2 := defaultVars()
	out2 := renderTemplate(t, v2)
	mustNotContain(t, out2, "X-Forwarded-Ssl")
}

func TestTemplate_NginxConfDIncludeAlways(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(map[string]interface{})
	}{
		{"http", func(v map[string]interface{}) {}},
		{"http_redirect", func(v map[string]interface{}) {
			v["PROXY_PORT_MAP"] = "http:80:5000 https:443:5000"
			v["SSL_INUSE"] = "true"
			v["SSL_SERVER_NAME"] = "app.example.com"
		}},
		{"grpc", func(v map[string]interface{}) {
			v["PROXY_PORT_MAP"] = "grpc:50051:50051"
			v["PROXY_UPSTREAM_PORTS"] = "50051"
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := defaultVars()
			tc.mutate(v)
			out := renderTemplate(t, v)
			mustContain(t, out, "include /home/dokku/app/nginx.conf.d/*.conf;")
		})
	}
}
