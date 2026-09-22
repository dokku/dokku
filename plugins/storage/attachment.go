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

// EffectiveProcessType returns the process type an attachment is scoped to,
// treating an unset value as the default scope so attachments written before
// the field existed still sort into a scope.
func (a *Attachment) EffectiveProcessType() string {
	if a == nil || a.ProcessType == "" {
		return DefaultProcessType
	}

	return a.ProcessType
}

// AppliesToProcessType reports whether an attachment is mounted into the
// container being built for the given process type. The default scope applies
// to every process; a named scope applies only to itself. An empty processType
// names a container that belongs to no process at all - a `dokku run` one-off,
// an app.json deploy task, a k3s cron job - and so sees the default scope only.
func (a *Attachment) AppliesToProcessType(processType string) bool {
	if processType == "" {
		processType = DefaultProcessType
	}

	scope := a.EffectiveProcessType()
	return scope == DefaultProcessType || scope == processType
}

// scopesOverlap reports whether two process-type scopes can both apply to one
// container. The default scope overlaps every named scope; two named scopes
// overlap only when they are equal.
func scopesOverlap(first string, second string) bool {
	return first == DefaultProcessType || second == DefaultProcessType || first == second
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

// ensureContainerPathFree rejects a candidate whose container path is already
// claimed by an attachment in an overlapping scope. Two attachments bound at
// one container path in overlapping scopes emit two -v flags for the same
// target, leaving which volume the process actually sees up to Docker.
func ensureContainerPathFree(appName string, attachments []*Attachment, candidate *Attachment) error {
	for _, existing := range attachments {
		if existing.ContainerPath != candidate.ContainerPath {
			continue
		}
		if !scopesOverlap(existing.EffectiveProcessType(), candidate.EffectiveProcessType()) {
			continue
		}

		return fmt.Errorf("Container path %s on app %s is already mounted by storage entry %s for process type %s",
			candidate.ContainerPath, appName, existing.EntryName, existing.EffectiveProcessType())
	}

	return nil
}

// AddAttachment appends an attachment, rejecting duplicates of the same
// (entry_name, container_path, process_type) tuple as well as container paths
// already claimed by an attachment in an overlapping scope.
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
			existing.EffectiveProcessType() == attachment.EffectiveProcessType() {
			return fmt.Errorf("storage entry %q is already mounted at %q for process type %q on app %q",
				attachment.EntryName, attachment.ContainerPath, attachment.EffectiveProcessType(), appName)
		}
	}

	if err := ensureContainerPathFree(appName, attachments, attachment); err != nil {
		return err
	}

	attachments = append(attachments, attachment)
	return SaveAttachments(appName, attachments)
}

// UpsertAttachment inserts an attachment, or updates the mount-time fields
// on an existing one matching the same (entry_name, container_path,
// process_type) tuple. Returns created=true when a new attachment was
// appended, created=false when an existing one was rewritten.
//
// Mount-time fields (Phases, Subpath, Readonly, VolumeOptions, VolumeChown)
// are overwritten wholesale, not merged: passing an empty Subpath or a
// false Readonly clears any previously-set value, mirroring the operator's
// "set the flags as I just typed them" intent and the same idempotent
// contract storage:create offers for entries.
//
// A new attachment is rejected when its container path is already claimed by
// an attachment in an overlapping scope.
func UpsertAttachment(appName string, attachment *Attachment) (bool, error) {
	if err := attachment.Validate(); err != nil {
		return false, err
	}

	attachments, err := LoadAttachments(appName)
	if err != nil {
		return false, err
	}

	for _, existing := range attachments {
		if existing.EntryName == attachment.EntryName &&
			existing.ContainerPath == attachment.ContainerPath &&
			existing.EffectiveProcessType() == attachment.EffectiveProcessType() {
			existing.Phases = attachment.Phases
			existing.Subpath = attachment.Subpath
			existing.Readonly = attachment.Readonly
			existing.VolumeOptions = attachment.VolumeOptions
			existing.VolumeChown = attachment.VolumeChown
			return false, SaveAttachments(appName, attachments)
		}
	}

	// Only the append branch needs the overlap check: rewriting an existing
	// tuple in place would otherwise trip over the attachment it is updating.
	if err := ensureContainerPathFree(appName, attachments, attachment); err != nil {
		return false, err
	}

	attachments = append(attachments, attachment)
	return true, SaveAttachments(appName, attachments)
}

// RemoveAttachment removes the attachment matching the given entry and
// optional container path. If containerPath is empty and there is more
// than one match, returns an error.
func RemoveAttachment(appName string, entryName string, containerPath string) error {
	attachments, err := LoadAttachments(appName)
	if err != nil {
		return err
	}

	keep, err := removeMatchingAttachments(attachments, appName, entryName, containerPath)
	if err != nil {
		return err
	}

	return SaveAttachments(appName, keep)
}

// removeMatchingAttachments returns the given list with the attachments
// matching the entry and optional container path dropped. Working against an
// in-memory list lets storage:unmount resolve several mounts before any of
// them is written, so one unresolvable mount removes none of them.
func removeMatchingAttachments(attachments []*Attachment, appName string, entryName string, containerPath string) ([]*Attachment, error) {
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
			return nil, fmt.Errorf("storage entry %q is not mounted on app %q", entryName, appName)
		}
		return nil, fmt.Errorf("storage entry %q is not mounted at %q on app %q", entryName, containerPath, appName)
	}
	if len(matches) > 1 && containerPath == "" {
		paths := []string{}
		for _, attachment := range matches {
			paths = append(paths, attachment.ContainerPath)
		}
		sort.Strings(paths)
		return nil, fmt.Errorf("storage entry %q is mounted at multiple paths on app %q (%s); pass --container-dir to disambiguate",
			entryName, appName, strings.Join(paths, ", "))
	}

	return keep, nil
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
