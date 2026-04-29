package scheduler_k3s

import (
	"encoding/base64"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGetConfigSecretReleaseName(t *testing.T) {
	got := GetConfigSecretReleaseName("myapp")
	want := "config-myapp"
	if got != want {
		t.Fatalf("GetConfigSecretReleaseName(\"myapp\") = %q, want %q", got, want)
	}
}

func TestGetConfigSecretName(t *testing.T) {
	got := GetConfigSecretName("myapp")
	want := "config-myapp"
	if got != want {
		t.Fatalf("GetConfigSecretName(\"myapp\") = %q, want %q", got, want)
	}
}

func TestConfigSecretValuesEncoding(t *testing.T) {
	values := &ConfigSecretValues{
		Global: ConfigSecretGlobalValues{
			AppName:   "myapp",
			Namespace: "default",
			Secrets: map[string]string{
				"DATABASE_URL": base64.StdEncoding.EncodeToString([]byte("postgres://localhost/db")),
				"REDIS_URL":    base64.StdEncoding.EncodeToString([]byte("redis://localhost:6379")),
			},
		},
	}

	data, err := yaml.Marshal(values)
	if err != nil {
		t.Fatalf("yaml.Marshal returned error: %v", err)
	}

	var roundTrip ConfigSecretValues
	if err := yaml.Unmarshal(data, &roundTrip); err != nil {
		t.Fatalf("yaml.Unmarshal returned error: %v", err)
	}

	if roundTrip.Global.AppName != "myapp" {
		t.Errorf("AppName = %q, want \"myapp\"", roundTrip.Global.AppName)
	}
	if roundTrip.Global.Namespace != "default" {
		t.Errorf("Namespace = %q, want \"default\"", roundTrip.Global.Namespace)
	}
	if len(roundTrip.Global.Secrets) != 2 {
		t.Errorf("Secrets length = %d, want 2", len(roundTrip.Global.Secrets))
	}

	decoded, err := base64.StdEncoding.DecodeString(roundTrip.Global.Secrets["DATABASE_URL"])
	if err != nil {
		t.Fatalf("base64 decode failed: %v", err)
	}
	if string(decoded) != "postgres://localhost/db" {
		t.Errorf("DATABASE_URL decoded = %q, want \"postgres://localhost/db\"", string(decoded))
	}
}
