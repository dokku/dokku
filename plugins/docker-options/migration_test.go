package dockeroptions

import (
	"os"
	"os/user"
	"path/filepath"
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

	for _, app := range []string{"alpha", "beta"} {
		for _, phase := range []string{"BUILD", "DEPLOY", "RUN"} {
			legacy := filepath.Join(dokkuRoot, app, "DOCKER_OPTIONS_"+phase)
			migrated := legacy + ".migrated"
			if app == "alpha" && (phase == "BUILD" || phase == "DEPLOY") || app == "beta" && phase == "DEPLOY" {
				if _, err := os.Stat(migrated); err != nil {
					t.Errorf("expected %s to exist: %v", migrated, err)
				}
				if _, err := os.Stat(legacy); !os.IsNotExist(err) {
					t.Errorf("expected %s to be gone, got err=%v", legacy, err)
				}
			} else {
				if _, err := os.Stat(legacy); !os.IsNotExist(err) {
					t.Errorf("did not expect %s to exist, got err=%v", legacy, err)
				}
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
