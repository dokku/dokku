package config

import (
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/dokku/dokku/plugins/common"
)

// setupMigrateEnv points the dokku env at temporary directories and tells the
// permission helpers to chown files to the current user (a no-op) so the test
// works without root. The package-level paths in config_test.go are captured at
// init against the real dokku directories, so these tests must not use them.
func setupMigrateEnv(t *testing.T) (dokkuRoot string, libRoot string) {
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

	return dokkuRoot, libRoot
}

func writeLegacyAppEnv(t *testing.T, dokkuRoot, appName, contents string) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dokkuRoot, appName), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	path := filepath.Join(dokkuRoot, appName, "ENV")
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}

func expectEnvValue(t *testing.T, appName, key, want string) {
	t.Helper()
	got, ok := Get(appName, key)
	if !ok {
		t.Fatalf("expected %s to be set for %s", key, appName)
	}
	if got != want {
		t.Errorf("%s for %s = %q, want %q", key, appName, got, want)
	}
}

func TestMigrateEnvFiles_DrainsAndRemovesAppFile(t *testing.T) {
	dokkuRoot, _ := setupMigrateEnv(t)

	legacy := writeLegacyAppEnv(t, dokkuRoot, "alpha", "export DOKKU_CHECKS_SKIPPED=worker\nexport MY_VAR=value\n")

	if err := MigrateEnvFiles(); err != nil {
		t.Fatalf("MigrateEnvFiles: %v", err)
	}

	expectEnvValue(t, "alpha", "DOKKU_CHECKS_SKIPPED", "worker")
	expectEnvValue(t, "alpha", "MY_VAR", "value")

	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Errorf("expected %s to be removed, got err=%v", legacy, err)
	}
	if common.PropertyGet("config", "alpha", envMigratedProperty) != "true" {
		t.Errorf("expected %s property to be set for alpha", envMigratedProperty)
	}
}

func TestMigrateEnvFiles_DrainsAndRemovesGlobalFile(t *testing.T) {
	dokkuRoot, _ := setupMigrateEnv(t)

	legacy := filepath.Join(dokkuRoot, "ENV")
	if err := os.WriteFile(legacy, []byte("export DOKKU_WAIT_TO_RETIRE=30\n"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := MigrateEnvFiles(); err != nil {
		t.Fatalf("MigrateEnvFiles: %v", err)
	}

	expectEnvValue(t, "--global", "DOKKU_WAIT_TO_RETIRE", "30")

	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Errorf("expected %s to be removed, got err=%v", legacy, err)
	}
	if common.PropertyGet("config", "--global", envMigratedProperty) != "true" {
		t.Errorf("expected %s property to be set globally", envMigratedProperty)
	}
}

// TestMigrateEnvFiles_ImportsHandEditedFile covers a legacy file that turns up
// after the migration was already recorded, which means it was written by hand
// rather than through `dokku config:*`. Its contents are merged in rather than
// discarded.
func TestMigrateEnvFiles_ImportsHandEditedFile(t *testing.T) {
	dokkuRoot, _ := setupMigrateEnv(t)

	writeLegacyAppEnv(t, dokkuRoot, "alpha", "export FIRST=one\n")
	if err := MigrateEnvFiles(); err != nil {
		t.Fatalf("first MigrateEnvFiles: %v", err)
	}

	legacy := writeLegacyAppEnv(t, dokkuRoot, "alpha", "export SECOND=two\n")
	if err := MigrateEnvFiles(); err != nil {
		t.Fatalf("second MigrateEnvFiles: %v", err)
	}

	expectEnvValue(t, "alpha", "FIRST", "one")
	expectEnvValue(t, "alpha", "SECOND", "two")
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Errorf("expected %s to be removed, got err=%v", legacy, err)
	}
}

// TestMigrateEnvFiles_LegacyValueWins documents the merge precedence the
// hand-edit path depends on: the legacy file overwrites the value already held
// at the config path.
func TestMigrateEnvFiles_LegacyValueWins(t *testing.T) {
	dokkuRoot, _ := setupMigrateEnv(t)

	if err := os.MkdirAll(filepath.Join(dokkuRoot, "alpha"), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := common.PropertySetupApp("config", "alpha"); err != nil {
		t.Fatalf("PropertySetupApp: %v", err)
	}
	if err := setupAppConfigDir("alpha"); err != nil {
		t.Fatalf("setupAppConfigDir: %v", err)
	}
	if err := SetMany("alpha", map[string]string{"SHARED": "new"}, false, false); err != nil {
		t.Fatalf("SetMany: %v", err)
	}

	writeLegacyAppEnv(t, dokkuRoot, "alpha", "export SHARED=legacy\n")
	if err := MigrateEnvFiles(); err != nil {
		t.Fatalf("MigrateEnvFiles: %v", err)
	}

	expectEnvValue(t, "alpha", "SHARED", "legacy")
}

func TestMigrateEnvFiles_MarksAppsWithoutLegacyFile(t *testing.T) {
	dokkuRoot, _ := setupMigrateEnv(t)

	if err := os.MkdirAll(filepath.Join(dokkuRoot, "alpha"), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	if err := MigrateEnvFiles(); err != nil {
		t.Fatalf("MigrateEnvFiles: %v", err)
	}

	if common.PropertyGet("config", "alpha", envMigratedProperty) != "true" {
		t.Errorf("expected %s property to be set for an app with no legacy file", envMigratedProperty)
	}
}

// TestMigrateEnvFiles_KeepsLegacyFileWhenWriteFails verifies the legacy file
// survives a failure to write the merged environment, so nothing is lost.
func TestMigrateEnvFiles_KeepsLegacyFileWhenWriteFails(t *testing.T) {
	dokkuRoot, libRoot := setupMigrateEnv(t)

	legacy := writeLegacyAppEnv(t, dokkuRoot, "alpha", "export MY_VAR=value\n")

	// Occupy the app's config directory path with a regular file so
	// MkdirAll cannot create it, even when the tests run as root.
	if err := os.MkdirAll(filepath.Join(libRoot, "config"), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(libRoot, "config", "alpha"), []byte(""), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := MigrateEnvFiles(); err == nil {
		t.Fatalf("expected MigrateEnvFiles to fail")
	}

	if _, err := os.Stat(legacy); err != nil {
		t.Errorf("expected %s to survive a failed migration, got err=%v", legacy, err)
	}
	if common.PropertyGet("config", "alpha", envMigratedProperty) == "true" {
		t.Errorf("did not expect %s property to be set after a failed migration", envMigratedProperty)
	}
}
