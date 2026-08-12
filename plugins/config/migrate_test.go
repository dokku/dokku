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

// expectPreservedKey asserts the stale ENV file moved aside at path still holds
// the given key, so the values it was not allowed to import remain recoverable.
func expectPreservedKey(t *testing.T, path, key, want string) {
	t.Helper()

	preserved, err := loadFromFile("preserved", path)
	if err != nil {
		t.Fatalf("loadFromFile(%s): %v", path, err)
	}
	got, ok := preserved.Get(key)
	if !ok {
		t.Fatalf("expected %s to hold %s", path, key)
	}
	if got != want {
		t.Errorf("%s in %s = %q, want %q", key, path, got, want)
	}
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

// TestMigrateEnvFiles_PreservesStaleFileWithoutImporting covers a legacy file
// that is still there once the migration has been recorded. Nothing it holds is
// imported, and it is moved aside so the values remain recoverable.
func TestMigrateEnvFiles_PreservesStaleFileWithoutImporting(t *testing.T) {
	dokkuRoot, _ := setupMigrateEnv(t)

	writeLegacyAppEnv(t, dokkuRoot, "alpha", "export FIRST=one\n")
	if err := MigrateEnvFiles(); err != nil {
		t.Fatalf("first MigrateEnvFiles: %v", err)
	}

	legacy := writeLegacyAppEnv(t, dokkuRoot, "alpha", "export FIRST=one\nexport SECOND=two\n")
	if err := MigrateEnvFiles(); err != nil {
		t.Fatalf("second MigrateEnvFiles: %v", err)
	}

	expectEnvValue(t, "alpha", "FIRST", "one")
	if _, ok := Get("alpha", "SECOND"); ok {
		t.Errorf("did not expect SECOND to be imported from a stale legacy file")
	}

	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Errorf("expected %s to be cleared out, got err=%v", legacy, err)
	}
	expectPreservedKey(t, legacy+".migrated", "SECOND", "two")
}

// TestMigrateEnvFiles_KeepsConfigSetAfterMigration covers the 0.38.26 regression
// reported in #8929: releases 0.38.0 through 0.38.25 recorded the migration and
// left the legacy file in place, so draining it a second time replayed the
// environment as it stood at that upgrade over every config:set made since.
func TestMigrateEnvFiles_KeepsConfigSetAfterMigration(t *testing.T) {
	dokkuRoot, _ := setupMigrateEnv(t)

	writeLegacyAppEnv(t, dokkuRoot, "alpha", "export DATABASE_URL=value-A\n")
	if err := MigrateEnvFiles(); err != nil {
		t.Fatalf("first MigrateEnvFiles: %v", err)
	}

	if err := SetMany("alpha", map[string]string{"DATABASE_URL": "value-B"}, false, false); err != nil {
		t.Fatalf("SetMany: %v", err)
	}

	legacy := writeLegacyAppEnv(t, dokkuRoot, "alpha", "export DATABASE_URL=value-A\n")
	if err := MigrateEnvFiles(); err != nil {
		t.Fatalf("second MigrateEnvFiles: %v", err)
	}

	expectEnvValue(t, "alpha", "DATABASE_URL", "value-B")
	expectPreservedKey(t, legacy+".migrated", "DATABASE_URL", "value-A")
}

// TestMigrateEnvFiles_DoesNotResurrectUnsetKeys covers the other half of #8929: a
// key unset after the migration was recorded came back when the legacy file was
// drained again, which for a rotated secret meant putting it back into service.
func TestMigrateEnvFiles_DoesNotResurrectUnsetKeys(t *testing.T) {
	dokkuRoot, _ := setupMigrateEnv(t)

	writeLegacyAppEnv(t, dokkuRoot, "alpha", "export DOKKU_PROXY_PORT=80\n")
	if err := MigrateEnvFiles(); err != nil {
		t.Fatalf("first MigrateEnvFiles: %v", err)
	}

	if err := UnsetMany("alpha", []string{"DOKKU_PROXY_PORT"}, false); err != nil {
		t.Fatalf("UnsetMany: %v", err)
	}

	writeLegacyAppEnv(t, dokkuRoot, "alpha", "export DOKKU_PROXY_PORT=80\n")
	if err := MigrateEnvFiles(); err != nil {
		t.Fatalf("second MigrateEnvFiles: %v", err)
	}

	if _, ok := Get("alpha", "DOKKU_PROXY_PORT"); ok {
		t.Errorf("did not expect DOKKU_PROXY_PORT to be resurrected from a stale legacy file")
	}
}

// TestMigrateEnvFiles_RemovesStaleFileMatchingCurrentConfig covers the common
// upgrade: the leftover file still agrees with the current config, so there is
// nothing to preserve and no reason to say anything about it.
func TestMigrateEnvFiles_RemovesStaleFileMatchingCurrentConfig(t *testing.T) {
	dokkuRoot, _ := setupMigrateEnv(t)

	writeLegacyAppEnv(t, dokkuRoot, "alpha", "export MY_VAR=value\n")
	if err := MigrateEnvFiles(); err != nil {
		t.Fatalf("first MigrateEnvFiles: %v", err)
	}

	legacy := writeLegacyAppEnv(t, dokkuRoot, "alpha", "export MY_VAR=value\n")
	if err := MigrateEnvFiles(); err != nil {
		t.Fatalf("second MigrateEnvFiles: %v", err)
	}

	expectEnvValue(t, "alpha", "MY_VAR", "value")
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Errorf("expected %s to be removed, got err=%v", legacy, err)
	}
	if _, err := os.Stat(legacy + ".migrated"); !os.IsNotExist(err) {
		t.Errorf("did not expect %s.migrated to be left behind, got err=%v", legacy, err)
	}
}

// TestMigrateEnvFiles_PreservesStaleGlobalFile covers the global file, which
// releases 0.38.0 through 0.38.25 never removed at all, so every host upgraded
// through that range still has one.
func TestMigrateEnvFiles_PreservesStaleGlobalFile(t *testing.T) {
	dokkuRoot, _ := setupMigrateEnv(t)

	legacy := filepath.Join(dokkuRoot, "ENV")
	if err := os.WriteFile(legacy, []byte("export DOKKU_WAIT_TO_RETIRE=30\n"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := MigrateEnvFiles(); err != nil {
		t.Fatalf("first MigrateEnvFiles: %v", err)
	}

	if err := SetMany("--global", map[string]string{"DOKKU_WAIT_TO_RETIRE": "60"}, false, false); err != nil {
		t.Fatalf("SetMany: %v", err)
	}

	if err := os.WriteFile(legacy, []byte("export DOKKU_WAIT_TO_RETIRE=30\n"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := MigrateEnvFiles(); err != nil {
		t.Fatalf("second MigrateEnvFiles: %v", err)
	}

	expectEnvValue(t, "--global", "DOKKU_WAIT_TO_RETIRE", "60")
	expectPreservedKey(t, legacy+".migrated", "DOKKU_WAIT_TO_RETIRE", "30")
}

// TestMigrateEnvFiles_LegacyValueWins documents the precedence of the one drain
// that does import: before the migration is recorded the legacy file is the
// source of truth, so it overwrites the value already held at the config path.
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
