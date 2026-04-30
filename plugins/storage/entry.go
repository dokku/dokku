package storage

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// SchemaVersion is the on-disk schema version emitted by this plugin.
// Bumped when the format of Entry / Attachment changes incompatibly.
const SchemaVersion = 1

// SchedulerDockerLocal is the default scheduler value for storage entries.
const SchedulerDockerLocal = "docker-local"

// SchedulerK3s is the k3s scheduler value for storage entries.
const SchedulerK3s = "k3s"

// LegacyEntryPrefix is reserved for migration-synthesized entries; users
// cannot create entries whose name starts with this prefix.
const LegacyEntryPrefix = "legacy-"

// MaxEntryNameLength caps the storage entry name so that the helm release
// name "storage-<name>" stays within Helm's practical 53-char ceiling.
const MaxEntryNameLength = 45

// ReclaimPolicyRetain keeps the underlying PV after the entry is destroyed.
const ReclaimPolicyRetain = "Retain"

// ReclaimPolicyDelete deletes the underlying PV after the entry is destroyed.
const ReclaimPolicyDelete = "Delete"

var (
	// dns1123LabelRegexp matches a Kubernetes DNS-1123 label, the format
	// required for resource names like PVCs and Helm releases.
	dns1123LabelRegexp = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

	// dockerNamedVolumeRegexp matches a Docker named-volume token. Same
	// rule docker uses to disambiguate volume names from bind paths.
	dockerNamedVolumeRegexp = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]+$`)

	// supportedSchedulers lists the values accepted for --scheduler.
	supportedSchedulers = map[string]bool{
		SchedulerDockerLocal: true,
		SchedulerK3s:         true,
	}

	// supportedAccessModes lists the values accepted for --access-mode on k3s.
	supportedAccessModes = map[string]bool{
		"ReadWriteOnce":    true,
		"ReadOnlyMany":     true,
		"ReadWriteMany":    true,
		"ReadWriteOncePod": true,
	}
)

// Entry is the source of truth for a storage volume. One file per entry
// lives at $DOKKU_LIB_ROOT/config/storage/entries/<name>.json.
type Entry struct {
	Name           string            `json:"name"`
	Scheduler      string            `json:"scheduler"`
	HostPath       string            `json:"host_path,omitempty"`
	Size           string            `json:"size,omitempty"`
	AccessMode     string            `json:"access_mode,omitempty"`
	StorageClass   string            `json:"storage_class,omitempty"`
	Namespace      string            `json:"namespace,omitempty"`
	Chown          string            `json:"chown,omitempty"`
	ReclaimPolicy  string            `json:"reclaim_policy,omitempty"`
	Annotations    map[string]string `json:"annotations,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
	SchemaVersion  int               `json:"schema_version"`
}

// RegistryDirectory returns the parent directory for storage-plugin
// state that lives outside the property store. Kept under data/ rather
// than config/ so it doesn't collide with the property-list paths
// `config/storage/<appName>/<property>` that the rest of the plugin
// relies on for per-app attachments.
func RegistryDirectory() string {
	root := common.GetenvWithDefault("DOKKU_LIB_ROOT", "/var/lib/dokku")
	return filepath.Join(root, "data", "storage-registry")
}

// EntriesDirectory returns the directory holding per-entry JSON files.
func EntriesDirectory() string {
	return filepath.Join(RegistryDirectory(), "entries")
}

// entryPath returns the on-disk path for a named entry.
func entryPath(name string) string {
	return filepath.Join(EntriesDirectory(), name+".json")
}

// EnsureEntriesDirectory makes sure the entries directory exists with
// permissive enough mode that the dokku user can read it.
func EnsureEntriesDirectory() error {
	dir := EntriesDirectory()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("unable to create storage entries directory: %w", err)
	}
	return nil
}

// EntryExists reports whether an entry with the given name is registered.
func EntryExists(name string) bool {
	_, err := os.Stat(entryPath(name))
	return err == nil
}

// LoadEntry reads a registered entry by name.
func LoadEntry(name string) (*Entry, error) {
	data, err := os.ReadFile(entryPath(name))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("storage entry %q does not exist", name)
		}
		return nil, fmt.Errorf("unable to read storage entry %q: %w", name, err)
	}

	entry := &Entry{}
	if err := json.Unmarshal(data, entry); err != nil {
		return nil, fmt.Errorf("unable to parse storage entry %q: %w", name, err)
	}
	return entry, nil
}

// SaveEntry persists an entry to its JSON file.
func SaveEntry(entry *Entry) error {
	if err := EnsureEntriesDirectory(); err != nil {
		return err
	}

	if entry.SchemaVersion == 0 {
		entry.SchemaVersion = SchemaVersion
	}

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to encode storage entry %q: %w", entry.Name, err)
	}

	path := entryPath(entry.Name)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("unable to write storage entry %q: %w", entry.Name, err)
	}
	return nil
}

// DeleteEntry removes the on-disk record for an entry.
func DeleteEntry(name string) error {
	if err := os.Remove(entryPath(name)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("unable to remove storage entry %q: %w", name, err)
	}
	return nil
}

// ListEntries returns every registered entry sorted by name.
func ListEntries() ([]*Entry, error) {
	dir := EntriesDirectory()
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to list storage entries: %w", err)
	}

	entries := []*Entry{}
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() || !strings.HasSuffix(dirEntry.Name(), ".json") {
			continue
		}
		name := strings.TrimSuffix(dirEntry.Name(), ".json")
		entry, err := LoadEntry(name)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	return entries, nil
}

// ValidateEntryName checks the user-supplied name against length and DNS
// label rules. Returns nil for legacy names if allowLegacyPrefix is true,
// which is only set during migration.
func ValidateEntryName(name string, allowLegacyPrefix bool) error {
	if name == "" {
		return errors.New("storage entry name is required")
	}
	if len(name) > MaxEntryNameLength {
		return fmt.Errorf("storage entry name %q is too long (max %d characters); helm release names cap our budget", name, MaxEntryNameLength)
	}
	if !dns1123LabelRegexp.MatchString(name) {
		return fmt.Errorf("storage entry name %q must be a DNS-1123 label (lowercase letters, digits, and dashes; cannot start or end with a dash)", name)
	}
	if !allowLegacyPrefix && strings.HasPrefix(name, LegacyEntryPrefix) {
		return fmt.Errorf("storage entry name %q must not start with reserved prefix %q", name, LegacyEntryPrefix)
	}
	return nil
}

// Validate checks an Entry's fields against the cross-field rules for its
// scheduler. It is reused by both storage:create and the legacy migration.
func (e *Entry) Validate() error {
	if e == nil {
		return errors.New("entry is nil")
	}
	if err := ValidateEntryName(e.Name, strings.HasPrefix(e.Name, LegacyEntryPrefix)); err != nil {
		return err
	}

	if !supportedSchedulers[e.Scheduler] {
		return fmt.Errorf("storage entry %q has unsupported scheduler %q (supported: docker-local, k3s)", e.Name, e.Scheduler)
	}

	switch e.Scheduler {
	case SchedulerDockerLocal:
		if e.HostPath == "" {
			return fmt.Errorf("storage entry %q (docker-local) is missing host_path", e.Name)
		}
		if !filepath.IsAbs(e.HostPath) && !dockerNamedVolumeRegexp.MatchString(e.HostPath) {
			return fmt.Errorf("storage entry %q host_path must be an absolute path or a docker named volume token, got %q", e.Name, e.HostPath)
		}
		if e.Size != "" {
			return fmt.Errorf("storage entry %q (docker-local) does not accept --size", e.Name)
		}
		if e.StorageClass != "" {
			return fmt.Errorf("storage entry %q (docker-local) does not accept --storage-class-name", e.Name)
		}
		if e.AccessMode != "" {
			return fmt.Errorf("storage entry %q (docker-local) does not accept --access-mode", e.Name)
		}
	case SchedulerK3s:
		if e.Size == "" {
			return fmt.Errorf("storage entry %q (k3s) requires --size", e.Name)
		}
		if e.StorageClass != "" && e.HostPath != "" {
			return fmt.Errorf("storage entry %q (k3s) cannot set both --storage-class-name and a host path; the cluster provisions class-backed volumes", e.Name)
		}
		if e.AccessMode != "" && !supportedAccessModes[e.AccessMode] {
			return fmt.Errorf("storage entry %q has unsupported access mode %q", e.Name, e.AccessMode)
		}
		if e.HostPath != "" && !filepath.IsAbs(e.HostPath) {
			return fmt.Errorf("storage entry %q host_path must be absolute, got %q", e.Name, e.HostPath)
		}
		if e.ReclaimPolicy != "" && e.ReclaimPolicy != ReclaimPolicyRetain && e.ReclaimPolicy != ReclaimPolicyDelete {
			return fmt.Errorf("storage entry %q has unsupported reclaim policy %q", e.Name, e.ReclaimPolicy)
		}
	}

	return nil
}

// LegacyMountToEntry synthesizes a deterministic Entry for a legacy
// "host:container[:options]" mount string. The synthesized name is
// stable across runs so re-running the migration is idempotent and so
// two apps mounting the same host path converge on a single entry.
func LegacyMountToEntry(mountPath string) *Entry {
	parsed := ParseMountPath(mountPath)
	host := parsed.HostPath

	hashInput := host
	if !strings.HasPrefix(host, "/") {
		hashInput = "vol:" + host
	}
	sum := sha1.Sum([]byte(hashInput))
	name := LegacyEntryPrefix + hex.EncodeToString(sum[:])[:10]

	hostPath := host
	if !strings.HasPrefix(host, "/") {
		// Named docker volumes pass through unchanged on docker-local;
		// the docker engine resolves them. Store as-is so the args
		// shim can emit the original token.
		hostPath = host
	}

	return &Entry{
		Name:          name,
		Scheduler:     SchedulerDockerLocal,
		HostPath:      hostPath,
		SchemaVersion: SchemaVersion,
	}
}
