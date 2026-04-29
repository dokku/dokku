package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// AttachmentsProperty is the property-list key used to store attachments
// for an app. All process types live under one list and carry their own
// ProcessType field; we filter on read.
const AttachmentsProperty = "mounts"

// PluginName is the name we register with the property system.
const PluginName = "storage"

// PhaseDeploy is the standard deploy phase identifier.
const PhaseDeploy = "deploy"

// PhaseRun is the standard run phase identifier.
const PhaseRun = "run"

// DefaultProcessType is the wildcard process type that applies to every
// process. Mirrors docker-options' DefaultProcessType.
const DefaultProcessType = "_default_"

// Attachment is the source of truth for *how* an app uses a storage entry.
// One Attachment binds one entry into one container path on one app.
type Attachment struct {
	EntryName     string   `json:"entry_name"`
	ContainerPath string   `json:"container_path"`
	Phases        []string `json:"phases"`
	ProcessType   string   `json:"process_type,omitempty"`
	Subpath       string   `json:"subpath,omitempty"`
	Readonly      bool     `json:"readonly,omitempty"`
	VolumeOptions string   `json:"volume_options,omitempty"`
	VolumeChown   string   `json:"volume_chown,omitempty"`
}

// Validate checks an Attachment's fields against structural rules.
func (a *Attachment) Validate() error {
	if a == nil {
		return errors.New("attachment is nil")
	}
	if a.EntryName == "" {
		return errors.New("attachment is missing entry name")
	}
	if a.ContainerPath == "" {
		return errors.New("attachment is missing container path")
	}
	if !strings.HasPrefix(a.ContainerPath, "/") {
		return fmt.Errorf("attachment container path %q must be absolute", a.ContainerPath)
	}
	if len(a.Phases) == 0 {
		return errors.New("attachment must specify at least one phase")
	}
	for _, phase := range a.Phases {
		if phase != PhaseDeploy && phase != PhaseRun {
			return fmt.Errorf("attachment phase %q is not supported (must be %q or %q)", phase, PhaseDeploy, PhaseRun)
		}
	}
	return nil
}

// LoadAttachments returns every attachment registered against an app.
func LoadAttachments(appName string) ([]*Attachment, error) {
	lines, err := common.PropertyListGet(PluginName, appName, AttachmentsProperty)
	if err != nil {
		return nil, fmt.Errorf("unable to read storage attachments for %q: %w", appName, err)
	}

	attachments := []*Attachment{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		attachment := &Attachment{}
		if err := json.Unmarshal([]byte(line), attachment); err != nil {
			return nil, fmt.Errorf("unable to parse storage attachment for %q: %w", appName, err)
		}
		attachments = append(attachments, attachment)
	}
	return attachments, nil
}

// SaveAttachments overwrites the entire attachment list for an app.
func SaveAttachments(appName string, attachments []*Attachment) error {
	lines := make([]string, 0, len(attachments))
	for _, attachment := range attachments {
		data, err := json.Marshal(attachment)
		if err != nil {
			return fmt.Errorf("unable to encode storage attachment for %q: %w", appName, err)
		}
		lines = append(lines, string(data))
	}
	return common.PropertyListWrite(PluginName, appName, AttachmentsProperty, lines)
}

// AddAttachment appends an attachment, rejecting duplicates of the same
// (entry_name, container_path, process_type) tuple.
func AddAttachment(appName string, attachment *Attachment) error {
	if err := attachment.Validate(); err != nil {
		return err
	}

	attachments, err := LoadAttachments(appName)
	if err != nil {
		return err
	}

	for _, existing := range attachments {
		if existing.EntryName == attachment.EntryName &&
			existing.ContainerPath == attachment.ContainerPath &&
			existing.ProcessType == attachment.ProcessType {
			return fmt.Errorf("storage entry %q is already mounted at %q for process type %q on app %q",
				attachment.EntryName, attachment.ContainerPath, attachment.ProcessType, appName)
		}
	}

	attachments = append(attachments, attachment)
	return SaveAttachments(appName, attachments)
}

// RemoveAttachment removes the attachment matching the given entry and
// optional container path. If containerPath is empty and there is more
// than one match, returns an error.
func RemoveAttachment(appName string, entryName string, containerPath string) error {
	attachments, err := LoadAttachments(appName)
	if err != nil {
		return err
	}

	matches := []*Attachment{}
	keep := []*Attachment{}
	for _, attachment := range attachments {
		if attachment.EntryName == entryName && (containerPath == "" || attachment.ContainerPath == containerPath) {
			matches = append(matches, attachment)
			continue
		}
		keep = append(keep, attachment)
	}

	if len(matches) == 0 {
		if containerPath == "" {
			return fmt.Errorf("storage entry %q is not mounted on app %q", entryName, appName)
		}
		return fmt.Errorf("storage entry %q is not mounted at %q on app %q", entryName, containerPath, appName)
	}
	if len(matches) > 1 && containerPath == "" {
		paths := []string{}
		for _, attachment := range matches {
			paths = append(paths, attachment.ContainerPath)
		}
		sort.Strings(paths)
		return fmt.Errorf("storage entry %q is mounted at multiple paths on app %q (%s); pass --container-dir to disambiguate",
			entryName, appName, strings.Join(paths, ", "))
	}

	return SaveAttachments(appName, keep)
}

// AttachmentsForPhase returns the subset of an app's attachments that
// apply to the given phase, sorted for stable output.
func AttachmentsForPhase(appName string, phase string) ([]*Attachment, error) {
	attachments, err := LoadAttachments(appName)
	if err != nil {
		return nil, err
	}

	filtered := []*Attachment{}
	for _, attachment := range attachments {
		for _, p := range attachment.Phases {
			if p == phase {
				filtered = append(filtered, attachment)
				break
			}
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].EntryName != filtered[j].EntryName {
			return filtered[i].EntryName < filtered[j].EntryName
		}
		return filtered[i].ContainerPath < filtered[j].ContainerPath
	})
	return filtered, nil
}

// AppsUsingEntry returns the list of app names that have at least one
// attachment referencing the given entry name. Used by storage:destroy
// to refuse removing an entry that's still mounted.
func AppsUsingEntry(entryName string) ([]string, error) {
	apps, err := common.DokkuApps()
	if err != nil {
		if errors.Is(err, common.NoAppsExist) {
			return nil, nil
		}
		return nil, err
	}

	using := []string{}
	for _, app := range apps {
		attachments, err := LoadAttachments(app)
		if err != nil {
			return nil, err
		}
		for _, attachment := range attachments {
			if attachment.EntryName == entryName {
				using = append(using, app)
				break
			}
		}
	}
	sort.Strings(using)
	return using, nil
}
