package backup

import (
	"encoding/json"
	"fmt"
	"os"
)

// Manifest is the top-level inventory of a backup. It is the source of truth
// for import: it drives ordering, scope resolution, and the redeploy decision.
type Manifest struct {
	// SchemaVersion is the manifest schema version (see ManifestSchemaVersion).
	SchemaVersion int `json:"schema_version"`

	// BackupVersion identifies the tarball format (see BackupFormatVersion).
	BackupVersion string `json:"backup_version"`

	// DokkuVersion is the dokku version that produced the backup.
	DokkuVersion string `json:"dokku_version"`

	// CreatedAt is the RFC3339 UTC timestamp of when the backup was created.
	CreatedAt string `json:"created_at"`

	// Hostname is the host that produced the backup (informational).
	Hostname string `json:"hostname,omitempty"`

	// IncludeStorage records whether persistent storage volume data was bundled.
	IncludeStorage bool `json:"include_storage"`

	// Apps maps an app name to its metadata.
	Apps map[string]ManifestApp `json:"apps,omitempty"`

	// Services maps a "type/name" key to its metadata.
	Services map[string]ManifestService `json:"services,omitempty"`
}

// ManifestApp is the per-app metadata used to drive an ordered import.
type ManifestApp struct {
	// Deployed records whether the app was deployed at export time.
	Deployed bool `json:"deployed"`

	// HasCode records whether a git bundle was captured for the app.
	HasCode bool `json:"has_code"`

	// Scheduler is the app's scheduler at export time (informational).
	Scheduler string `json:"scheduler,omitempty"`

	// Plugins lists the plugins that contributed a slice for the app.
	Plugins []string `json:"plugins,omitempty"`

	// DataFiles lists the raw data files bundled for the app, relative to the
	// app scope directory.
	DataFiles []string `json:"data_files,omitempty"`
}

// ManifestService is the per-service metadata used during import.
type ManifestService struct {
	// LinkedApps lists the apps linked to the service at export time.
	LinkedApps []string `json:"linked_apps,omitempty"`
}

// NewManifest returns a manifest stamped with the current format, dokku
// version, host, and the supplied creation timestamp.
func NewManifest(createdAt string, includeStorage bool) *Manifest {
	return &Manifest{
		SchemaVersion:  ManifestSchemaVersion,
		BackupVersion:  BackupFormatVersion,
		DokkuVersion:   dokkuVersion(),
		CreatedAt:      createdAt,
		Hostname:       hostname(),
		IncludeStorage: includeStorage,
		Apps:           map[string]ManifestApp{},
		Services:       map[string]ManifestService{},
	}
}

// WriteManifest marshals a manifest to path with 0600 permissions.
func WriteManifest(path string, manifest *Manifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal manifest: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("unable to write manifest: %w", err)
	}
	return nil
}

// ReadManifest reads and unmarshals a manifest from path.
func ReadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("unable to parse manifest: %w", err)
	}
	return &manifest, nil
}

// Validate ensures the manifest is well-formed and can be imported by this
// version of dokku. It rejects a schema_version newer than supported.
func (m *Manifest) Validate() error {
	if m.BackupVersion == "" {
		return fmt.Errorf("backup is missing a backup_version; it may be corrupt")
	}
	if m.SchemaVersion == 0 {
		return fmt.Errorf("backup is missing a schema_version; it may be corrupt")
	}
	if m.SchemaVersion > ManifestSchemaVersion {
		return fmt.Errorf("backup schema_version %d is newer than supported version %d; upgrade dokku to import this backup", m.SchemaVersion, ManifestSchemaVersion)
	}
	return nil
}
