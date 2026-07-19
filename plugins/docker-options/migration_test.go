package dockeroptions

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dokku/dokku/plugins/common"
)

// setupMigrationEnv points the dokku env at temporary directories and tells the
// permission helpers to chown files to the current user (a no-op) so the test
// works without root.
func setupMigrationEnv(t *testing.T) (dokkuRoot string) {
	t.Helper()

	libRoot := t.TempDir()
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

	return dokkuRoot
}

func writeLegacyDockerOptionsFile(t *testing.T, dokkuRoot, app, phase, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dokkuRoot, app), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	path := filepath.Join(dokkuRoot, app, "DOCKER_OPTIONS_"+phase)
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func TestMigrateLegacyDockerOptionsFiles_MigratesAndIsIdempotent(t *testing.T) {
	dokkuRoot := setupMigrationEnv(t)

	writeLegacyDockerOptionsFile(t, dokkuRoot, "alpha", "DEPLOY", "-v /var/log:/log\n# a comment\n\n--restart=on-failure:5\n")
	writeLegacyDockerOptionsFile(t, dokkuRoot, "alpha", "BUILD", "--build-arg FOO=bar\n")
	writeLegacyDockerOptionsFile(t, dokkuRoot, "beta", "DEPLOY", "-p 8080:5000\n")

	if err := migrateLegacyDockerOptionsFiles(); err != nil {
		t.Fatalf("first migration: %v", err)
	}

	deploy, err := common.PropertyListGet("docker-options", "alpha", "_default_.deploy")
	if err != nil {
		t.Fatalf("PropertyListGet alpha deploy: %v", err)
	}
	wantDeploy := []string{"-v /var/log:/log", "--restart=on-failure:5"}
	if !equalStrings(deploy, wantDeploy) {
		t.Errorf("alpha deploy = %v, want %v", deploy, wantDeploy)
	}

	build, err := common.PropertyListGet("docker-options", "alpha", "_default_.build")
	if err != nil {
		t.Fatalf("PropertyListGet alpha build: %v", err)
	}
	wantBuild := []string{"--build-arg FOO=bar"}
	if !equalStrings(build, wantBuild) {
		t.Errorf("alpha build = %v, want %v", build, wantBuild)
	}

	betaDeploy, err := common.PropertyListGet("docker-options", "beta", "_default_.deploy")
	if err != nil {
		t.Fatalf("PropertyListGet beta deploy: %v", err)
	}
	wantBetaDeploy := []string{"-p 8080:5000"}
	if !equalStrings(betaDeploy, wantBetaDeploy) {
		t.Errorf("beta deploy = %v, want %v", betaDeploy, wantBetaDeploy)
	}

	if !common.PropertyExists("docker-options", "--global", "migrated-from-files") {
		t.Errorf("migrated-from-files marker not set after first migration")
	}

	migratedPhases := map[string]map[string]bool{
		"alpha": {"build": true, "deploy": true},
		"beta":  {"deploy": true},
	}
	for _, app := range []string{"alpha", "beta"} {
		for _, phase := range []string{"BUILD", "DEPLOY", "RUN"} {
			legacy := filepath.Join(dokkuRoot, app, "DOCKER_OPTIONS_"+phase)
			migrated := legacy + ".migrated"
			if _, err := os.Stat(legacy); !os.IsNotExist(err) {
				t.Errorf("expected %s to be gone, got err=%v", legacy, err)
			}
			if _, err := os.Stat(migrated); !os.IsNotExist(err) {
				t.Errorf("expected %s to be gone (no more rename), got err=%v", migrated, err)
			}
			lowerPhase := strings.ToLower(phase)
			exists := common.PropertyExists("docker-options", app, "migrated-"+lowerPhase)
			if migratedPhases[app][lowerPhase] {
				if !exists {
					t.Errorf("expected migrated-%s property for %s to be set", lowerPhase, app)
				}
			} else if exists {
				t.Errorf("did not expect migrated-%s property for %s to be set", lowerPhase, app)
			}
		}
	}

	sneakyPath := filepath.Join(dokkuRoot, "alpha", "DOCKER_OPTIONS_DEPLOY")
	sneakyContents := "-v /tmp/sneaky:/sneaky\n"
	if err := os.WriteFile(sneakyPath, []byte(sneakyContents), 0644); err != nil {
		t.Fatalf("re-create legacy file: %v", err)
	}

	if err := migrateLegacyDockerOptionsFiles(); err != nil {
		t.Fatalf("second migration: %v", err)
	}

	deployAfter, err := common.PropertyListGet("docker-options", "alpha", "_default_.deploy")
	if err != nil {
		t.Fatalf("PropertyListGet alpha deploy (post re-run): %v", err)
	}
	if !equalStrings(deployAfter, wantDeploy) {
		t.Errorf("alpha deploy after rerun = %v, want %v (idempotency violated)", deployAfter, wantDeploy)
	}

	gotSneaky, err := os.ReadFile(sneakyPath)
	if err != nil {
		t.Fatalf("expected sneaky legacy file untouched, got err=%v", err)
	}
	if string(gotSneaky) != sneakyContents {
		t.Errorf("sneaky legacy file modified: %q", gotSneaky)
	}
}

// TestMigrateLegacyDockerOptionsFiles_ConvertsLegacyMigratedMarker
// covers the upgrade-cycle path where a previous-release `.migrated`
// sentinel exists on disk. The conversion writes the per-phase
// property and removes the file.
func TestMigrateLegacyDockerOptionsFiles_ConvertsLegacyMigratedMarker(t *testing.T) {
	dokkuRoot := setupMigrationEnv(t)

	if err := os.MkdirAll(filepath.Join(dokkuRoot, "alpha"), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	sentinel := filepath.Join(dokkuRoot, "alpha", "DOCKER_OPTIONS_DEPLOY.migrated")
	if err := os.WriteFile(sentinel, []byte{}, 0644); err != nil {
		t.Fatalf("WriteFile sentinel: %v", err)
	}

	if err := migrateLegacyDockerOptionsFiles(); err != nil {
		t.Fatalf("migrateLegacyDockerOptionsFiles: %v", err)
	}

	if !common.PropertyExists("docker-options", "alpha", "migrated-deploy") {
		t.Errorf("expected migrated-deploy property to be set")
	}
	if _, err := os.Stat(sentinel); !os.IsNotExist(err) {
		t.Errorf("expected sentinel file gone, got err=%v", err)
	}
}

// TestMigrateLegacyDockerOptionsFiles_ConvertsMigratedMarkerEvenWhenGloballyMigrated
// is the regression test for the upgrade-cycle ordering fix: users on
// the previous release have `migrated-from-files` ALREADY set AND
// `.migrated` sentinels on disk. The conversion pass must run before
// the global short-circuit so these sentinels still get drained.
func TestMigrateLegacyDockerOptionsFiles_ConvertsMigratedMarkerEvenWhenGloballyMigrated(t *testing.T) {
	dokkuRoot := setupMigrationEnv(t)

	if err := os.MkdirAll(filepath.Join(dokkuRoot, "alpha"), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	sentinel := filepath.Join(dokkuRoot, "alpha", "DOCKER_OPTIONS_DEPLOY.migrated")
	if err := os.WriteFile(sentinel, []byte{}, 0644); err != nil {
		t.Fatalf("WriteFile sentinel: %v", err)
	}
	if err := common.PropertyWrite("docker-options", "--global", "migrated-from-files", "true"); err != nil {
		t.Fatalf("seed global marker: %v", err)
	}

	if err := migrateLegacyDockerOptionsFiles(); err != nil {
		t.Fatalf("migrateLegacyDockerOptionsFiles: %v", err)
	}

	if !common.PropertyExists("docker-options", "alpha", "migrated-deploy") {
		t.Errorf("expected migrated-deploy property to be set despite global short-circuit")
	}
	if _, err := os.Stat(sentinel); !os.IsNotExist(err) {
		t.Errorf("expected sentinel file gone, got err=%v", err)
	}
}

// TestMigrateLegacyDockerOptionsFiles_SkipsEmptyContentFiles verifies
// that a legacy file containing only comments/whitespace gets removed
// but does NOT receive a per-phase property write - the property is
// only set when actual content was drained.
func TestMigrateLegacyDockerOptionsFiles_SkipsEmptyContentFiles(t *testing.T) {
	dokkuRoot := setupMigrationEnv(t)

	writeLegacyDockerOptionsFile(t, dokkuRoot, "alpha", "DEPLOY", "# only a comment\n\n   \n")

	if err := migrateLegacyDockerOptionsFiles(); err != nil {
		t.Fatalf("migrateLegacyDockerOptionsFiles: %v", err)
	}

	if common.PropertyExists("docker-options", "alpha", "migrated-deploy") {
		t.Errorf("did not expect migrated-deploy property to be set for empty-content file")
	}
	legacy := filepath.Join(dokkuRoot, "alpha", "DOCKER_OPTIONS_DEPLOY")
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Errorf("expected empty-content legacy file removed, got err=%v", err)
	}
	if common.PropertyExists("docker-options", "alpha", "_default_.deploy") {
		t.Errorf("did not expect _default_.deploy property list for empty-content file")
	}
}

func TestRepairTraefikLabelBackticks(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    string
		changed bool
	}{
		{"traefik rule space form", "--label 'traefik.http.routers.web.rule=Host(\\`app.example.com\\`)'", "--label 'traefik.http.routers.web.rule=Host(`app.example.com`)'", true},
		{"traefik rule whole-quoted equals form", "'--label=traefik.http.routers.web.rule=Host(\\`x\\`)'", "'--label=traefik.http.routers.web.rule=Host(`x`)'", true},
		{"non-traefik label left untouched", "--label 'some.key=Host(\\`x\\`)'", "--label 'some.key=Host(\\`x\\`)'", false},
		{"already valid traefik label", "--label 'traefik.rule=Host(`x`)'", "--label 'traefik.rule=Host(`x`)'", false},
		{"non-label option left untouched", "-v /tmp", "-v /tmp", false},
		{"traefik label without backticks", "--label 'traefik.enable=true'", "--label 'traefik.enable=true'", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, changed := repairTraefikLabelBackticks(tc.in)
			if got != tc.want || changed != tc.changed {
				t.Errorf("repairTraefikLabelBackticks(%q) = (%q, %v), want (%q, %v)", tc.in, got, changed, tc.want, tc.changed)
			}
		})
	}
}

func TestMigrateTraefikLabelBackticks(t *testing.T) {
	dokkuRoot := setupMigrationEnv(t)
	if err := os.MkdirAll(filepath.Join(dokkuRoot, "alpha"), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	broken := "--label 'traefik.http.routers.web.rule=Host(\\`app.example.com\\`)'"
	if err := common.PropertyListWrite("docker-options", "alpha", "_default_.deploy", []string{broken}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	if err := migrateTraefikLabelBackticks(); err != nil {
		t.Fatalf("migrateTraefikLabelBackticks: %v", err)
	}

	got, err := GetDockerOptionsForProcessPhase("alpha", "_default_", "deploy")
	if err != nil {
		t.Fatalf("GetDockerOptionsForProcessPhase: %v", err)
	}
	want := "--label 'traefik.http.routers.web.rule=Host(`app.example.com`)'"
	if !equalStrings(got, []string{want}) {
		t.Errorf("got %q, want [%q]", got, []string{want})
	}
	if common.PropertyGet("docker-options", "--global", traefikLabelMigrationKey) != "true" {
		t.Errorf("expected %s guard to be set", traefikLabelMigrationKey)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
