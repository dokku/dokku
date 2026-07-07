package plugin

import (
	"testing"
)

func TestParsePlugnList(t *testing.T) {
	output := `plugn: dev
  00_dokku-standard    0.38.21 enabled    dokku core standard plugin
  smoke-test-plugin    0.2.0 disabled   a third party plugin
  bare-plugin          1.0.0 enabled`

	plugins := parsePlugnList(output)
	if len(plugins) != 3 {
		t.Fatalf("expected 3 plugins, got %d", len(plugins))
	}

	first := plugins[0]
	if first.Name != "00_dokku-standard" {
		t.Errorf("expected name 00_dokku-standard, got %q", first.Name)
	}
	if first.Version != "0.38.21" {
		t.Errorf("expected version 0.38.21, got %q", first.Version)
	}
	if !first.Enabled {
		t.Errorf("expected plugin to be enabled")
	}
	if first.Description != "dokku core standard plugin" {
		t.Errorf("expected multi-word description, got %q", first.Description)
	}

	if plugins[1].Enabled {
		t.Errorf("expected smoke-test-plugin to be disabled")
	}
	if plugins[1].Description != "a third party plugin" {
		t.Errorf("expected multi-word description, got %q", plugins[1].Description)
	}

	if plugins[2].Name != "bare-plugin" {
		t.Errorf("expected name bare-plugin, got %q", plugins[2].Name)
	}
	if plugins[2].Description != "" {
		t.Errorf("expected empty description, got %q", plugins[2].Description)
	}
}
