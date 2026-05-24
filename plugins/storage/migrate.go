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
	if err := migrateApp(appName); err != nil {
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
		if common.PropertyExists(PluginName, app, MigratedProperty) {
			continue
		}
		if err := migrateApp(app); err != nil {
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

// migrateApp performs the per-app migration. Order is intentional: write
// the entry, write the attachment, drain the docker-options line. Any
// step failing aborts before drain, so the original behavior is
// preserved on partial failure.
func migrateApp(appName string) error {
	deployLines, err := dockeroptions.GetDockerOptionsForPhase(appName, PhaseDeploy)
	if err != nil {
		return err
	}
	runLines, err := dockeroptions.GetDockerOptionsForPhase(appName, PhaseRun)
	if err != nil {
		return err
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
		// No legacy `-v` lines for this app. Leave the marker unset so
		// `storage:report` (and future tooling) distinguishes apps that
		// never had legacy state from apps that did and were drained.
		return nil
	}

	for _, mount := range mounts {
		if err := migrateMount(appName, mount, phaseMap[mount]); err != nil {
			return err
		}
	}

	return common.PropertyWrite(PluginName, appName, MigratedProperty, "true")
}

func migrateMount(appName string, mount string, phases []string) error {
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
		ProcessType:   DefaultProcessType,
	}
	if parsed.VolumeOptions != "" {
		if parsed.VolumeOptions == "ro" {
			attachment.Readonly = true
		} else {
			attachment.VolumeOptions = parsed.VolumeOptions
		}
	}

	existing, err := LoadAttachments(appName)
	if err != nil {
		return err
	}
	if !attachmentExists(existing, attachment) {
		existing = append(existing, attachment)
		if err := SaveAttachments(appName, existing); err != nil {
			return err
		}
	}

	mountedPhases := []string{}
	for _, phase := range phases {
		mountedPhases = append(mountedPhases, phase)
	}
	if err := dockeroptions.RemoveDockerOptionFromPhases(appName, mountedPhases, fmt.Sprintf("-v %s", mount)); err != nil {
		return err
	}
	return nil
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
			existing.ProcessType == candidate.ProcessType {
			return true
		}
	}
	return false
}
