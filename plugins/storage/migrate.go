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

// migrationFlagDir contains per-app flag files written after a successful
// per-app migration; presence means the app's legacy -v lines have been
// drained into the attachment store. Lives under the registry directory
// so it doesn't collide with the property-store path
// `config/storage/<appName>`.
func migrationFlagDir() string {
	return filepath.Join(RegistryDirectory(), "migrations")
}

func migrationFlagFile(appName string) string {
	return filepath.Join(migrationFlagDir(), appName)
}

// MigrateApp re-runs the legacy migration for a single app, ignoring
// the per-app flag file so an operator can force-rescan an app whose
// docker-options state changed after the bulk install-time pass.
// Used by `dokku storage:migrate <app>`.
func MigrateApp(appName string) error {
	if err := os.MkdirAll(migrationFlagDir(), 0755); err != nil {
		return fmt.Errorf("unable to create migration flag dir: %w", err)
	}
	_ = os.Remove(migrationFlagFile(appName))
	if err := migrateApp(appName); err != nil {
		return fmt.Errorf("storage migration failed for app %q: %w", appName, err)
	}
	return nil
}

// MigrateLegacyMounts walks every app and converts its legacy `-v`
// docker-options entries into named storage entries plus attachments.
// Idempotent: per-app flag files short-circuit re-runs.
func MigrateLegacyMounts() error {
	if err := os.MkdirAll(migrationFlagDir(), 0755); err != nil {
		return fmt.Errorf("unable to create migration flag dir: %w", err)
	}

	apps, err := common.DokkuApps()
	if err != nil {
		if errors.Is(err, common.NoAppsExist) {
			return nil
		}
		return err
	}

	for _, app := range apps {
		if _, err := os.Stat(migrationFlagFile(app)); err == nil {
			continue
		}
		if err := migrateApp(app); err != nil {
			return fmt.Errorf("storage migration failed for app %q: %w", app, err)
		}
	}
	return nil
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

	for _, mount := range mounts {
		if err := migrateMount(appName, mount, phaseMap[mount]); err != nil {
			return err
		}
	}

	return touchMigrationFlag(appName)
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

func touchMigrationFlag(appName string) error {
	f, err := os.Create(migrationFlagFile(appName))
	if err != nil {
		return err
	}
	return f.Close()
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
