package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	dockeroptions "github.com/dokku/dokku/plugins/docker-options"
)

// MigratedProperty is the per-app marker recording that legacy
// docker-options `-v` lines have been drained into named storage
// entries plus attachments. Written via the property store so it is
// visible to property-store backup/restore tooling, replacing the old
// filesystem flag file at `data/storage-registry/migrations/<app>`.
const MigratedProperty = "legacy-mounts-migrated"

// MigratedPropertyVersion is the value MigratedProperty carries once an app
// has been through the current migration. Earlier releases wrote "true" and
// drained only the default process-type scope, so an app still holding that
// value is rescanned once to pick up its process-scoped `-v` lines.
const MigratedPropertyVersion = "2"

// migrationFlagDir / migrationFlagFile remain for one release cycle so
// the upgrade-cycle helper convertLegacyMigrationFlag can drain any
// leftover flag files into the new property.
// TODO(post-deprecation): remove both functions and the
// migrationFlagDir entry in triggers.go's install directory list.
func migrationFlagDir() string {
	return filepath.Join(RegistryDirectory(), "migrations")
}

func migrationFlagFile(appName string) string {
	return filepath.Join(migrationFlagDir(), appName)
}

// MigrateApp re-runs the legacy migration for a single app so an
// operator can force-rescan an app whose docker-options state changed
// after the bulk install-time pass. Used by `dokku storage:migrate
// <app>`. The marker property is intentionally left in place; migrateApp
// only writes it when something was actually drained, so a re-run on
// an already-drained app with no new `-v` lines preserves the existing
// "had legacy state, drained" signal.
func MigrateApp(appName string) error {
	if err := convertLegacyMigrationFlag(appName); err != nil {
		return err
	}
	if err := migrateApp(appName, true); err != nil {
		return fmt.Errorf("storage migration failed for app %q: %w", appName, err)
	}
	return nil
}

// MigrateLegacyMounts walks every app and converts its legacy `-v`
// docker-options entries into named storage entries plus attachments.
// Idempotent: the per-app legacy-mounts-migrated property short-circuits
// re-runs.
func MigrateLegacyMounts() error {
	apps, err := common.DokkuApps()
	if err != nil {
		if errors.Is(err, common.NoAppsExist) {
			return nil
		}
		return err
	}

	for _, app := range apps {
		if err := convertLegacyMigrationFlag(app); err != nil {
			return err
		}
		if err := migrateApp(app, false); err != nil {
			return fmt.Errorf("storage migration failed for app %q: %w", app, err)
		}
	}
	return nil
}

// convertLegacyMigrationFlag drains the legacy filesystem flag file
// (data/storage-registry/migrations/<app>) into the new
// legacy-mounts-migrated property and removes the file. Runs as part
// of the per-app loop so installs upgrading from the previous release
// surface the marker in the property store without re-running the
// drain logic. No-op when the flag file is absent.
// TODO(post-deprecation): remove this helper and its callers.
func convertLegacyMigrationFlag(appName string) error {
	path := migrationFlagFile(appName)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if err := common.PropertyWrite(PluginName, appName, MigratedProperty, "true"); err != nil {
		return err
	}
	return os.Remove(path)
}

// migrateApp performs the per-app migration across every process-type scope
// docker-options holds options for. The default scope is migrated first: the
// per-scope drain skips a named-scope mount that the default scope already
// covers, which only works once the default scope's attachments exist.
//
// force bypasses the version gate, which is what `dokku storage:migrate <app>`
// wants - it exists so an operator can rescan an app whose docker-options
// state changed after the bulk install-time pass.
func migrateApp(appName string, force bool) error {
	migrated := common.PropertyGet(PluginName, appName, MigratedProperty)
	if !force && migrated == MigratedPropertyVersion {
		return nil
	}

	processTypes := []string{DefaultProcessType}
	named, err := dockeroptions.ListProcessTypesWithOptions(appName)
	if err != nil {
		return err
	}
	processTypes = append(processTypes, named...)

	drained := false
	for _, processType := range processTypes {
		scopeDrained, err := migrateProcessScope(appName, processType)
		if err != nil {
			return err
		}
		drained = drained || scopeDrained
	}

	// An app that never had legacy state keeps no marker at all, so
	// `storage:report` (and future tooling) can still tell it apart from one
	// that did and was drained. An app already carrying an older marker keeps
	// one, advanced to the current version so it stops being rescanned.
	if !drained && migrated == "" {
		return nil
	}

	return common.PropertyWrite(PluginName, appName, MigratedProperty, MigratedPropertyVersion)
}

// migrateProcessScope drains one process-type scope's legacy `-v` lines into
// attachments. Reports whether anything was drained.
func migrateProcessScope(appName string, processType string) (bool, error) {
	deployLines, err := dockeroptions.GetDockerOptionsForProcessPhase(appName, processType, PhaseDeploy)
	if err != nil {
		return false, err
	}
	runLines, err := dockeroptions.GetDockerOptionsForProcessPhase(appName, processType, PhaseRun)
	if err != nil {
		return false, err
	}

	deployMounts := filterMountLines(deployLines)
	runMounts := filterMountLines(runLines)

	// Group identical mount strings across phases.
	phaseMap := map[string][]string{}
	for _, mount := range deployMounts {
		phaseMap[mount] = appendUnique(phaseMap[mount], PhaseDeploy)
	}
	for _, mount := range runMounts {
		phaseMap[mount] = appendUnique(phaseMap[mount], PhaseRun)
	}

	mounts := make([]string, 0, len(phaseMap))
	for mount := range phaseMap {
		mounts = append(mounts, mount)
	}
	sort.Strings(mounts)

	if len(mounts) == 0 {
		return false, nil
	}

	for _, mount := range mounts {
		if err := migrateMount(appName, processType, mount, phaseMap[mount]); err != nil {
			return false, err
		}
	}

	return true, nil
}

func migrateMount(appName string, processType string, mount string, phases []string) error {
	parsed := ParseMountPath(mount)
	entry := LegacyMountToEntry(mount)

	if EntryExists(entry.Name) {
		existing, err := LoadEntry(entry.Name)
		if err != nil {
			return err
		}
		if existing.HostPath != entry.HostPath || existing.Scheduler != entry.Scheduler {
			return fmt.Errorf("legacy entry %q already exists with conflicting fields", entry.Name)
		}
	} else {
		if err := SaveEntry(entry); err != nil {
			return err
		}
	}

	containerPath := parsed.ContainerPath
	if containerPath == "" {
		// Defensive: legacy mounts must have a container path; skip if not.
		return nil
	}

	attachment := &Attachment{
		EntryName:     entry.Name,
		ContainerPath: containerPath,
		Phases:        phases,
		ProcessType:   processType,
		Readonly:      parsed.Readonly,
		VolumeOptions: parsed.VolumeOptions,
	}

	existing, err := LoadAttachments(appName)
	if err != nil {
		return err
	}
	// A named scope whose mount the default scope already covers needs no
	// attachment of its own - the default one reaches that process anyway, and
	// recording both would manufacture two -v flags for one container path.
	// The docker-options line is still drained so it stops being emitted twice.
	if !attachmentExists(existing, attachment) && !coveredByDefaultScope(existing, attachment) {
		existing = append(existing, attachment)
		if err := SaveAttachments(appName, existing); err != nil {
			return err
		}
	}

	mountedPhases := []string{}
	for _, phase := range phases {
		mountedPhases = append(mountedPhases, phase)
	}
	if err := dockeroptions.RemoveDockerOptionFromProcessPhases(appName, []string{processType}, mountedPhases, fmt.Sprintf("-v %s", mount)); err != nil {
		return err
	}
	return nil
}

// coveredByDefaultScope reports whether a named-scope candidate binds the same
// entry at the same container path as an attachment already in the default
// scope. Always false for a default-scope candidate, which is its own cover.
func coveredByDefaultScope(attachments []*Attachment, candidate *Attachment) bool {
	if candidate.EffectiveProcessType() == DefaultProcessType {
		return false
	}

	for _, existing := range attachments {
		if existing.EffectiveProcessType() != DefaultProcessType {
			continue
		}
		if existing.EntryName == candidate.EntryName && existing.ContainerPath == candidate.ContainerPath {
			return true
		}
	}
	return false
}

func filterMountLines(lines []string) []string {
	out := []string{}
	for _, line := range lines {
		if strings.HasPrefix(line, "-v ") {
			out = append(out, strings.TrimPrefix(line, "-v "))
		}
	}
	return out
}

func appendUnique(slice []string, value string) []string {
	for _, existing := range slice {
		if existing == value {
			return slice
		}
	}
	return append(slice, value)
}

func attachmentExists(attachments []*Attachment, candidate *Attachment) bool {
	for _, existing := range attachments {
		if existing.EntryName == candidate.EntryName &&
			existing.ContainerPath == candidate.ContainerPath &&
			existing.EffectiveProcessType() == candidate.EffectiveProcessType() {
			return true
		}
	}
	return false
}
