package backup

import (
	"path/filepath"
	"testing"
)

func TestManifestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ManifestFilename)

	manifest := NewManifest("2026-06-18T00:00:00Z", true)
	manifest.Apps["api"] = ManifestApp{Deployed: true, HasCode: true, Scheduler: "docker-local", Plugins: []string{"config", "ps"}}
	manifest.Services["postgres/db"] = ManifestService{LinkedApps: []string{"api"}}

	if err := WriteManifest(path, manifest); err != nil {
		t.Fatalf("WriteManifest: %v", err)
	}

	got, err := ReadManifest(path)
	if err != nil {
		t.Fatalf("ReadManifest: %v", err)
	}

	if got.SchemaVersion != ManifestSchemaVersion {
		t.Errorf("schema_version = %d, want %d", got.SchemaVersion, ManifestSchemaVersion)
	}
	if got.BackupVersion != BackupFormatVersion {
		t.Errorf("backup_version = %q, want %q", got.BackupVersion, BackupFormatVersion)
	}
	if !got.IncludeStorage {
		t.Errorf("include_storage = false, want true")
	}
	app, ok := got.Apps["api"]
	if !ok {
		t.Fatalf("app api missing from round-tripped manifest")
	}
	if !app.Deployed || !app.HasCode || app.Scheduler != "docker-local" {
		t.Errorf("app metadata not preserved: %+v", app)
	}
	if svc, ok := got.Services["postgres/db"]; !ok || len(svc.LinkedApps) != 1 || svc.LinkedApps[0] != "api" {
		t.Errorf("service metadata not preserved: %+v", got.Services)
	}
}

func TestManifestValidateRejectsNewerSchema(t *testing.T) {
	manifest := NewManifest("2026-06-18T00:00:00Z", false)
	manifest.SchemaVersion = ManifestSchemaVersion + 1
	if err := manifest.Validate(); err == nil {
		t.Errorf("Validate accepted a newer schema_version; want error")
	}
}

func TestManifestValidateRejectsMissingFields(t *testing.T) {
	cases := map[string]*Manifest{
		"missing backup_version": {SchemaVersion: 1},
		"missing schema_version": {BackupVersion: BackupFormatVersion},
	}
	for name, manifest := range cases {
		if err := manifest.Validate(); err == nil {
			t.Errorf("%s: Validate accepted; want error", name)
		}
	}
}

func TestManifestValidateAcceptsCurrent(t *testing.T) {
	if err := NewManifest("2026-06-18T00:00:00Z", false).Validate(); err != nil {
		t.Errorf("Validate rejected a current manifest: %v", err)
	}
}
