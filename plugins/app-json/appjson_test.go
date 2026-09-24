package appjson

import (
	"strings"
	"testing"
)

func TestValidateHealthchecksAcceptsWellFormedValues(t *testing.T) {
	tests := []struct {
		name        string
		healthcheck Healthcheck
	}{
		{name: "empty", healthcheck: Healthcheck{}},
		{name: "root path", healthcheck: Healthcheck{Path: "/"}},
		{name: "nested path", healthcheck: Healthcheck{Path: "/health/ready"}},
		{name: "query string", healthcheck: Healthcheck{Path: "/health?full=1&verbose=true"}},
		{name: "percent encoded", healthcheck: Healthcheck{Path: "/health%20check"}},
		{name: "http scheme", healthcheck: Healthcheck{Path: "/", Scheme: "http"}},
		{name: "https scheme", healthcheck: Healthcheck{Path: "/", Scheme: "https"}},
		{name: "uppercase scheme", healthcheck: Healthcheck{Path: "/", Scheme: "HTTPS"}},
		{name: "command check", healthcheck: Healthcheck{Command: []string{"/bin/sh", "-c", "echo ok"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appJSON := AppJSON{Healthchecks: map[string][]Healthcheck{"web": {tt.healthcheck}}}
			if err := validateHealthchecks(appJSON); err != nil {
				t.Errorf("validateHealthchecks() returned unexpected error: %v", err)
			}
		})
	}
}

func TestValidateHealthchecksRejectsMalformedValues(t *testing.T) {
	tests := []struct {
		name        string
		healthcheck Healthcheck
		wantField   string
	}{
		{name: "relative path", healthcheck: Healthcheck{Path: "health"}, wantField: "path"},
		{name: "space", healthcheck: Healthcheck{Path: "/health check"}, wantField: "path"},
		{name: "tab", healthcheck: Healthcheck{Path: "/health\tcheck"}, wantField: "path"},
		{name: "newline", healthcheck: Healthcheck{Path: "/health\ncheck"}, wantField: "path"},
		{name: "control character", healthcheck: Healthcheck{Path: "/health\x00"}, wantField: "path"},
		{name: "double quote", healthcheck: Healthcheck{Path: `/health"`}, wantField: "path"},
		{name: "single quote", healthcheck: Healthcheck{Path: `/health'`}, wantField: "path"},
		{name: "backslash", healthcheck: Healthcheck{Path: `/health\ check`}, wantField: "path"},
		{name: "unknown scheme", healthcheck: Healthcheck{Path: "/", Scheme: "tcp"}, wantField: "scheme"},
		{name: "scheme with whitespace", healthcheck: Healthcheck{Path: "/", Scheme: "http s"}, wantField: "scheme"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appJSON := AppJSON{Healthchecks: map[string][]Healthcheck{"web": {tt.healthcheck}}}
			err := validateHealthchecks(appJSON)
			if err == nil {
				t.Fatalf("validateHealthchecks() returned nil, want error for %s", tt.wantField)
			}

			if !strings.Contains(err.Error(), tt.wantField) {
				t.Errorf("validateHealthchecks() error = %q, want it to mention %q", err.Error(), tt.wantField)
			}
		})
	}
}

func TestValidateHealthchecksNamesTheOffendingCheck(t *testing.T) {
	appJSON := AppJSON{Healthchecks: map[string][]Healthcheck{
		"web": {
			{Name: "ok", Path: "/ready"},
			{Path: "/health check"},
		},
	}}

	err := validateHealthchecks(appJSON)
	if err == nil {
		t.Fatal("validateHealthchecks() returned nil, want error")
	}

	for _, want := range []string{"#2", "web"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("validateHealthchecks() error = %q, want it to mention %q", err.Error(), want)
		}
	}
}
