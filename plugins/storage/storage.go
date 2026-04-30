package storage

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	dockeroptions "github.com/dokku/dokku/plugins/docker-options"
)

// MountPhases are the phases where storage mounts are applied
var MountPhases = []string{"deploy", "run"}

// VerifyPaths validates the storage mount path format
func VerifyPaths(mountPath string) error {
	if strings.HasPrefix(mountPath, "/") {
		matched, err := regexp.MatchString(`^/.*:/`, mountPath)
		if err != nil {
			return err
		}
		if !matched {
			return errors.New("Storage path must be two valid paths divided by colon.")
		}
	} else {
		matched, err := regexp.MatchString(`^[a-zA-Z0-9][a-zA-Z0-9_.-]+:/`, mountPath)
		if err != nil {
			return err
		}
		if !matched {
			return errors.New("Volume name must be two characters or more. Volume name must not contain invalid characters. Storage path must be two valid paths divided by colon.")
		}
	}
	return nil
}

// CheckIfPathExists checks if a mount path exists in the specified phases
func CheckIfPathExists(appName string, mountPath string, phases []string) bool {
	for _, phase := range phases {
		options, err := dockeroptions.GetDockerOptionsForPhase(appName, phase)
		if err != nil {
			continue
		}
		for _, option := range options {
			if option == fmt.Sprintf("-v %s", mountPath) {
				return true
			}
		}
	}
	return false
}

// GetBindMounts returns the bind mounts for an app and phase, synthesized
// from the attachment store. The returned strings use the legacy colon
// form (host:container[:options]) so existing display paths keep
// working unchanged.
func GetBindMounts(appName string, phase string) ([]string, error) {
	entries, err := ListAppMountEntries(appName, phase)
	if err != nil {
		return nil, err
	}
	mounts := make([]string, 0, len(entries))
	for _, entry := range entries {
		mounts = append(mounts, formatStorageListEntry(entry))
	}
	return mounts, nil
}

// GetBindMountsForDisplay returns the bind mounts formatted for display
func GetBindMountsForDisplay(appName string, phase string) string {
	mounts, err := GetBindMounts(appName, phase)
	if err != nil {
		return ""
	}

	result := []string{}
	for _, mount := range mounts {
		result = append(result, fmt.Sprintf("-v %s", mount))
	}
	return strings.Join(result, " ")
}

// StorageListEntry represents a storage mount entry for JSON output.
type StorageListEntry struct {
	EntryName     string `json:"entry_name,omitempty"`
	HostPath      string `json:"host_path"`
	ContainerPath string `json:"container_path"`
	VolumeOptions string `json:"volume_options"`
}

// ListAppMountEntries returns one StorageListEntry per attachment on
// an app for the requested phase, joining each attachment with its
// referenced storage entry. Used by storage:list and the deprecated
// storage-list trigger.
func ListAppMountEntries(appName string, phase string) ([]StorageListEntry, error) {
	if phase == "" {
		phase = PhaseDeploy
	}
	attachments, err := AttachmentsForPhase(appName, phase)
	if err != nil {
		return nil, err
	}

	rows := make([]StorageListEntry, 0, len(attachments))
	for _, attachment := range attachments {
		entry, err := LoadEntry(attachment.EntryName)
		if err != nil {
			return nil, fmt.Errorf("attachment on %q references missing entry %q: %w", appName, attachment.EntryName, err)
		}

		host := entry.HostPath
		if host == "" {
			// k3s-only entries with no host path: surface the entry
			// name as the host token so the colon-form output is
			// well-formed and parseable.
			host = entry.Name
		}

		options := ""
		switch {
		case attachment.Readonly && attachment.VolumeOptions != "":
			options = "ro," + attachment.VolumeOptions
		case attachment.Readonly:
			options = "ro"
		case attachment.VolumeOptions != "":
			options = attachment.VolumeOptions
		}

		rows = append(rows, StorageListEntry{
			EntryName:     entry.Name,
			HostPath:      host,
			ContainerPath: attachment.ContainerPath,
			VolumeOptions: options,
		})
	}
	return rows, nil
}

// formatStorageListEntry renders a StorageListEntry into the legacy
// host:container[:options] colon form for textual output.
func formatStorageListEntry(entry StorageListEntry) string {
	if entry.VolumeOptions == "" {
		return fmt.Sprintf("%s:%s", entry.HostPath, entry.ContainerPath)
	}
	return fmt.Sprintf("%s:%s:%s", entry.HostPath, entry.ContainerPath, entry.VolumeOptions)
}

// ParseMountPath parses a mount path into its components
func ParseMountPath(mountPath string) StorageListEntry {
	parts := strings.SplitN(mountPath, ":", 3)
	entry := StorageListEntry{}

	if len(parts) >= 1 {
		entry.HostPath = parts[0]
	}
	if len(parts) >= 2 {
		entry.ContainerPath = parts[1]
	}
	if len(parts) >= 3 {
		entry.VolumeOptions = parts[2]
	}

	return entry
}

// GetStorageDirectory returns the storage directory path
func GetStorageDirectory() string {
	dokkuLibRoot := common.GetenvWithDefault("DOKKU_LIB_ROOT", "/var/lib/dokku")
	return fmt.Sprintf("%s/data/storage", dokkuLibRoot)
}

// ValidateDirectoryName validates a storage directory name
func ValidateDirectoryName(directory string) error {
	if directory == "" {
		return errors.New("Please specify a directory to create")
	}

	matched, err := regexp.MatchString(`^[A-Za-z0-9_-]+$`, directory)
	if err != nil {
		return err
	}
	if !matched {
		return errors.New("Directory can only contain the following set of characters: [A-Za-z0-9_-]")
	}
	return nil
}
