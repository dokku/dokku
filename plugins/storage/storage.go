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

// GetBindMounts returns the bind mounts for an app and phase
func GetBindMounts(appName string, phase string) ([]string, error) {
	mounts := []string{}
	options, err := dockeroptions.GetDockerOptionsForPhase(appName, phase)
	if err != nil {
		return mounts, err
	}

	for _, option := range options {
		if strings.HasPrefix(option, "-v ") {
			mount := strings.TrimPrefix(option, "-v ")
			mounts = append(mounts, mount)
		}
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

// StorageListEntry represents a storage mount entry for JSON output
type StorageListEntry struct {
	HostPath      string `json:"host_path"`
	ContainerPath string `json:"container_path"`
	VolumeOptions string `json:"volume_options"`
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
