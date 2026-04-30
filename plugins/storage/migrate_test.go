package storage

import (
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"testing"

	"github.com/dokku/dokku/plugins/common"
	dockeroptions "github.com/dokku/dokku/plugins/docker-options"
)

// setupMigrationEnv points DOKKU_LIB_ROOT and DOKKU_ROOT at temp dirs and
// tells the permission helpers to chown to the current user (a no-op),
// so the test works without root and without the dokku system user.
// Mirrors the helper in docker-options/migration_test.go.
func setupMigrationEnv(t *testing.T) (libRoot string, dokkuRoot string) {
	t.Helper()
	libRoot = t.TempDir()
	dokkuRoot = t.TempDir()

	t.Setenv("DOKKU_LIB_ROOT", libRoot)
	t.Setenv("DOKKU_ROOT", dokkuRoot)
	t.Setenv("PLUGIN_PATH", filepath.Join(libRoot, "plugins"))

	current, err := user.Current()
	if err != nil {
		t.Fatalf("user.Current: %v", err)
	}
	group, err := user.LookupGroupId(current.Gid)
	if err != nil {
		t.Fatalf("user.LookupGroupId: %v", err)
	}
	t.Setenv("DOKKU_SYSTEM_USER", current.Username)
	t.Setenv("DOKKU_SYSTEM_GROUP", group.Name)

	// MigrateLegacyMounts creates the flag dir before iterating apps;
	// migrateApp (the per-app helper) assumes it already exists. Pre-
	// create it so tests that call migrateApp directly don't trip.
	if err := os.MkdirAll(migrationFlagDir(), 0755); err != nil {
		t.Fatalf("mkdir migration flag dir: %v", err)
	}

	return libRoot, dokkuRoot
}

// stageApp registers the named app under DOKKU_ROOT so DokkuApps can
// see it, and seeds its docker-options property files for the requested
// phases. Each map entry is a phase ("deploy", "run", "build") whose
// value is the full property-list contents (one option per line).
func stageApp(t *testing.T, dokkuRoot string, app string, options map[string][]string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dokkuRoot, app), 0755); err != nil {
		t.Fatalf("mkdir app: %v", err)
	}
	for phase, lines := range options {
		if err := dockeroptions.AddDockerOptionToPhases(app, []string{phase}, ""); err != nil {
			t.Fatalf("seed property file: %v", err)
		}
		// Replace whatever AddDockerOptionToPhases left with the precise
		// list the test wants, since it appends rather than overwriting.
		key := "_default_." + phase
		if err := common.PropertyListWrite("docker-options", app, key, lines); err != nil {
			t.Fatalf("PropertyListWrite: %v", err)
		}
	}
}

func phaseOptions(t *testing.T, app string, phase string) []string {
	t.Helper()
	got, err := dockeroptions.GetDockerOptionsForPhase(app, phase)
	if err != nil {
		t.Fatalf("GetDockerOptionsForPhase %s/%s: %v", app, phase, err)
	}
	return got
}

func TestMigrateAppSinglePhase(t *testing.T) {
	_, dokkuRoot := setupMigrationEnv(t)
	stageApp(t, dokkuRoot, "alpha", map[string][]string{
		"deploy": {"-v /var/log:/log", "--restart=on-failure:5"},
	})

	if err := migrateApp("alpha"); err != nil {
		t.Fatalf("migrateApp: %v", err)
	}

	entry := LegacyMountToEntry("/var/log:/log")
	if !EntryExists(entry.Name) {
		t.Fatalf("expected entry %s on disk", entry.Name)
	}
	loaded, err := LoadEntry(entry.Name)
	if err != nil {
		t.Fatalf("LoadEntry: %v", err)
	}
	if loaded.HostPath != "/var/log" {
		t.Errorf("entry HostPath = %s, want /var/log", loaded.HostPath)
	}
	if loaded.Scheduler != SchedulerDockerLocal {
		t.Errorf("entry Scheduler = %s, want %s", loaded.Scheduler, SchedulerDockerLocal)
	}

	attachments, err := LoadAttachments("alpha")
	if err != nil {
		t.Fatalf("LoadAttachments: %v", err)
	}
	if len(attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(attachments))
	}
	att := attachments[0]
	if att.EntryName != entry.Name || att.ContainerPath != "/log" {
		t.Errorf("attachment = %+v, want entry=%s container=/log", att, entry.Name)
	}
	if len(att.Phases) != 1 || att.Phases[0] != PhaseDeploy {
		t.Errorf("attachment phases = %v, want [deploy]", att.Phases)
	}

	// docker-options drained the -v line but kept everything else.
	deployAfter := phaseOptions(t, "alpha", "deploy")
	if !equalSorted(deployAfter, []string{"--restart=on-failure:5"}) {
		t.Errorf("deploy options after migration = %v, want [--restart=on-failure:5]", deployAfter)
	}

	// Migration flag file is in place.
	if _, err := os.Stat(migrationFlagFile("alpha")); err != nil {
		t.Errorf("migration flag missing: %v", err)
	}
}

func TestMigrateAppCrossPhaseGroupsIntoOneAttachment(t *testing.T) {
	_, dokkuRoot := setupMigrationEnv(t)
	// Same -v in both phases groups into one attachment with phases = [deploy, run].
	stageApp(t, dokkuRoot, "alpha", map[string][]string{
		"deploy": {"-v /data:/d"},
		"run":    {"-v /data:/d"},
	})

	if err := migrateApp("alpha"); err != nil {
		t.Fatalf("migrateApp: %v", err)
	}

	attachments, err := LoadAttachments("alpha")
	if err != nil {
		t.Fatalf("LoadAttachments: %v", err)
	}
	if len(attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(attachments))
	}
	got := attachments[0].Phases
	sort.Strings(got)
	if !equalSorted(got, []string{PhaseDeploy, PhaseRun}) {
		t.Errorf("phases = %v, want [deploy run]", got)
	}

	// Drain affected both phases.
	if got := phaseOptions(t, "alpha", "deploy"); len(got) != 0 {
		t.Errorf("deploy options after migration = %v, want []", got)
	}
	if got := phaseOptions(t, "alpha", "run"); len(got) != 0 {
		t.Errorf("run options after migration = %v, want []", got)
	}
}

func TestMigrateAppPreservesVolumeOptions(t *testing.T) {
	_, dokkuRoot := setupMigrationEnv(t)
	stageApp(t, dokkuRoot, "alpha", map[string][]string{
		"deploy": {"-v /ro:/r:ro", "-v /label:/l:Z"},
	})

	if err := migrateApp("alpha"); err != nil {
		t.Fatalf("migrateApp: %v", err)
	}

	atts, err := LoadAttachments("alpha")
	if err != nil {
		t.Fatalf("LoadAttachments: %v", err)
	}
	byPath := map[string]*Attachment{}
	for _, a := range atts {
		byPath[a.ContainerPath] = a
	}

	if a := byPath["/r"]; a == nil {
		t.Fatalf("missing /r attachment")
	} else if !a.Readonly || a.VolumeOptions != "" {
		t.Errorf("/r attachment = readonly=%v opts=%q, want readonly=true opts=\"\"", a.Readonly, a.VolumeOptions)
	}

	if a := byPath["/l"]; a == nil {
		t.Fatalf("missing /l attachment")
	} else if a.Readonly || a.VolumeOptions != "Z" {
		t.Errorf("/l attachment = readonly=%v opts=%q, want readonly=false opts=\"Z\"", a.Readonly, a.VolumeOptions)
	}
}

func TestMigrateAppIdempotent(t *testing.T) {
	_, dokkuRoot := setupMigrationEnv(t)
	stageApp(t, dokkuRoot, "alpha", map[string][]string{
		"deploy": {"-v /var/log:/log"},
	})

	if err := migrateApp("alpha"); err != nil {
		t.Fatalf("first migrateApp: %v", err)
	}

	// Re-stage the same -v line by hand to simulate someone re-adding
	// the legacy form via docker-options:add after migration. The
	// migration is gated by the per-app flag file at the
	// MigrateLegacyMounts level, but migrateApp itself should still be
	// safe to call directly: a second run produces no duplicates.
	if err := common.PropertyListWrite("docker-options", "alpha", "_default_.deploy", []string{"-v /var/log:/log"}); err != nil {
		t.Fatalf("re-stage: %v", err)
	}
	if err := migrateApp("alpha"); err != nil {
		t.Fatalf("second migrateApp: %v", err)
	}

	atts, err := LoadAttachments("alpha")
	if err != nil {
		t.Fatalf("LoadAttachments: %v", err)
	}
	if len(atts) != 1 {
		t.Errorf("expected 1 attachment after rerun, got %d", len(atts))
	}
}

func TestMigrateLegacyMountsBulkAndFlagFastPath(t *testing.T) {
	_, dokkuRoot := setupMigrationEnv(t)
	stageApp(t, dokkuRoot, "alpha", map[string][]string{
		"deploy": {"-v /alpha-data:/data"},
	})
	stageApp(t, dokkuRoot, "beta", map[string][]string{
		"deploy": {"-v /beta-data:/data"},
		"run":    {"-v /beta-data:/data"},
	})

	if err := MigrateLegacyMounts(); err != nil {
		t.Fatalf("MigrateLegacyMounts: %v", err)
	}

	for _, app := range []string{"alpha", "beta"} {
		if _, err := os.Stat(migrationFlagFile(app)); err != nil {
			t.Errorf("flag file missing for %s: %v", app, err)
		}
		atts, err := LoadAttachments(app)
		if err != nil {
			t.Fatalf("LoadAttachments %s: %v", app, err)
		}
		if len(atts) != 1 {
			t.Errorf("%s: expected 1 attachment, got %d", app, len(atts))
		}
	}

	// Re-stage a -v on alpha and rerun MigrateLegacyMounts. The flag
	// file short-circuits per-app migration, so the new line stays in
	// docker-options.
	if err := common.PropertyListWrite("docker-options", "alpha", "_default_.deploy", []string{"-v /sneaky:/x"}); err != nil {
		t.Fatalf("re-stage alpha: %v", err)
	}
	if err := MigrateLegacyMounts(); err != nil {
		t.Fatalf("MigrateLegacyMounts (rerun): %v", err)
	}
	got := phaseOptions(t, "alpha", "deploy")
	if !equalSorted(got, []string{"-v /sneaky:/x"}) {
		t.Errorf("expected sneaky -v line untouched after second pass, got %v", got)
	}
	atts, err := LoadAttachments("alpha")
	if err != nil {
		t.Fatalf("LoadAttachments alpha rerun: %v", err)
	}
	if len(atts) != 1 {
		t.Errorf("expected 1 attachment after rerun, got %d", len(atts))
	}
}

func TestMigrateAppRefusesEntryNameCollision(t *testing.T) {
	libRoot, dokkuRoot := setupMigrationEnv(t)
	_ = libRoot

	// Plant a colliding entry at the legacy-<hash> name with conflicting
	// fields so migration's "fields match" check trips.
	collide := LegacyMountToEntry("/var/log:/log")
	collide.HostPath = "/something/else"
	if err := SaveEntry(collide); err != nil {
		t.Fatalf("seed colliding entry: %v", err)
	}

	stageApp(t, dokkuRoot, "alpha", map[string][]string{
		"deploy": {"-v /var/log:/log"},
	})

	if err := migrateApp("alpha"); err == nil {
		t.Fatalf("expected migrateApp to error on entry name collision")
	}

	// Drain did not happen.
	if got := phaseOptions(t, "alpha", "deploy"); len(got) != 1 || got[0] != "-v /var/log:/log" {
		t.Errorf("expected legacy line preserved on collision, got %v", got)
	}
}

func equalSorted(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	ax := append([]string(nil), a...)
	bx := append([]string(nil), b...)
	sort.Strings(ax)
	sort.Strings(bx)
	for i := range ax {
		if ax[i] != bx[i] {
			return false
		}
	}
	return true
}
