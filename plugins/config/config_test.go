package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

const testAppName = "test-app-1"

// setupTestApp points the dokku environment at temporary directories and seeds
// the app and global config the tests read, returning the app's config
// directory. Nothing is written to a real dokku installation and no dokku user
// needs to exist, which matters now that a failure to apply ownership is
// reported rather than discarded. The directories are removed with the test.
func setupTestApp(t *testing.T) (testAppDir string) {
	t.Helper()

	dokkuRoot, libRoot := setupIsolatedEnv(t)
	t.Setenv("PLUGIN_ENABLED_PATH", filepath.Join(libRoot, "plugins", "enabled"))
	t.Setenv("PLUGIN_CORE_AVAILABLE_PATH", filepath.Join(libRoot, "core-plugins", "available"))

	Expect(os.MkdirAll(filepath.Join(dokkuRoot, testAppName), 0755)).To(Succeed())

	testAppDir = filepath.Join(libRoot, "config", testAppName)
	Expect(os.MkdirAll(testAppDir, 0755)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(testAppDir, "ENV"), []byte("export testKey=TESTING\n"), 0600)).To(Succeed())

	globalConfigFile := filepath.Join(libRoot, "config", "--global", "ENV")
	Expect(os.MkdirAll(filepath.Dir(globalConfigFile), 0755)).To(Succeed())
	Expect(os.WriteFile(globalConfigFile, []byte("export testKey=GLOBAL_TESTING\nexport globalKey=GLOBAL_VALUE"), 0600)).To(Succeed())

	return testAppDir
}

func TestConfigGetWithDefault(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t)
	Expect(GetWithDefault(testAppName, "unknownKey", "UNKNOWN")).To(Equal("UNKNOWN"))
	Expect(GetWithDefault(testAppName, "testKey", "testKey")).To(Equal("TESTING"))
	Expect(GetWithDefault(testAppName+"-nonexistent", "testKey", "default")).To(Equal("default"))
}

func TestConfigGet(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t)

	expectValue(testAppName, "testKey", "TESTING")
	expectValue("", "testKey", "GLOBAL_TESTING")

	expectNoValue(testAppName, "testKey2")
	expectNoValue("", "testKey2")
}

func TestConfigSetMany(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t)

	expectValue(testAppName, "testKey", "TESTING")

	vals := []string{"testKey=updated", "testKey2=new"}
	Expect(CommandSet(testAppName, vals, false, true, false)).To(Succeed())
	expectValue(testAppName, "testKey", "updated")
	expectValue(testAppName, "testKey2", "new")

	vals = []string{"testKey=updated_global", "testKey2=new_global"}
	Expect(CommandSet("", vals, true, true, false)).To(Succeed())
	expectValue("", "testKey", "updated_global")
	expectValue("", "testKey2", "new_global")
	expectValue("", "globalKey", "GLOBAL_VALUE")
	expectValue(testAppName, "testKey", "updated")
	expectValue(testAppName, "testKey2", "new")

	Expect(CommandSet(testAppName+"does_not_exist", vals, false, true, false)).ToNot(Succeed())
}

func TestConfigUnsetAll(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t)

	expectValue(testAppName, "testKey", "TESTING")
	expectValue("", "testKey", "GLOBAL_TESTING")

	Expect(CommandClear(testAppName, false, true)).To(Succeed())
	expectNoValue(testAppName, "testKey")
	expectNoValue(testAppName, "noKey")
	expectNoValue(testAppName, "globalKey")

	Expect(CommandClear(testAppName+"does-not-exist", false, true)).ToNot(Succeed())
}

func TestConfigUnsetMany(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t)

	expectValue(testAppName, "testKey", "TESTING")
	expectValue("", "testKey", "GLOBAL_TESTING")

	keys := []string{"testKey", "noKey"}
	Expect(CommandUnset(testAppName, keys, false, true)).To(Succeed())
	expectNoValue(testAppName, "testKey")
	expectValue("", "testKey", "GLOBAL_TESTING")

	Expect(CommandUnset(testAppName, keys, false, true)).To(Succeed())
	expectNoValue(testAppName, "testKey")
	expectNoValue(testAppName, "globalKey")

	Expect(CommandUnset(testAppName+"does-not-exist", keys, false, true)).ToNot(Succeed())
}

func TestConfigImport(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t)

	tempFile, err := os.CreateTemp("", "test-config-import-*.env")
	Expect(err).To(Succeed())
	defer os.Remove(tempFile.Name())
	content := `
testKey=TESTING-updated1
testKey2=TESTING-updated2
`
	_, err = tempFile.WriteString(content)
	Expect(err).To(Succeed())
	tempFile.Close()

	Expect(CommandImport(testAppName, false, false, true, "envfile", tempFile.Name())).To(Succeed())
	expectValue(testAppName, "testKey", "TESTING-updated1")
	expectValue(testAppName, "testKey2", "TESTING-updated2")

	env, err := LoadAppEnv(testAppName)
	Expect(err).To(Succeed())
	env.Set("testKey", "TESTING-original1")
	env.Set("testKey2", "TESTING-original2")
	env.Set("testKey3", "TESTING-original3")
	Expect(env.Write()).To(Succeed())

	Expect(CommandImport(testAppName, false, false, true, "envfile", tempFile.Name())).To(Succeed())
	expectValue(testAppName, "testKey", "TESTING-updated1")
	expectValue(testAppName, "testKey2", "TESTING-updated2")
	expectValue(testAppName, "testKey3", "TESTING-original3")

	Expect(CommandImport(testAppName, false, true, true, "envfile", tempFile.Name())).To(Succeed())
	expectValue(testAppName, "testKey", "TESTING-updated1")
	expectValue(testAppName, "testKey2", "TESTING-updated2")
	expectNoValue(testAppName, "testKey3")
}

func TestConfigImportJSON(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t)

	tempFile, err := os.CreateTemp("", "test-config-import-*.json")
	Expect(err).To(Succeed())
	defer os.Remove(tempFile.Name())

	content := `{"testKey": "TESTING-updated1", "testKey2": "TESTING-updated2"}`
	_, err = tempFile.WriteString(content)
	Expect(err).To(Succeed())
	tempFile.Close()

	Expect(CommandImport(testAppName, false, false, true, "json", tempFile.Name())).To(Succeed())
	expectValue(testAppName, "testKey", "TESTING-updated1")
	expectValue(testAppName, "testKey2", "TESTING-updated2")

	env, err := LoadAppEnv(testAppName)
	Expect(err).To(Succeed())
	env.Set("testKey", "TESTING-original1")
	env.Set("testKey2", "TESTING-original2")
	env.Set("testKey3", "TESTING-original3")
	Expect(env.Write()).To(Succeed())

	Expect(CommandImport(testAppName, false, false, true, "json", tempFile.Name())).To(Succeed())
	expectValue(testAppName, "testKey", "TESTING-updated1")
	expectValue(testAppName, "testKey2", "TESTING-updated2")
	expectValue(testAppName, "testKey3", "TESTING-original3")

	Expect(CommandImport(testAppName, false, true, true, "json", tempFile.Name())).To(Succeed())
	expectValue(testAppName, "testKey", "TESTING-updated1")
	expectValue(testAppName, "testKey2", "TESTING-updated2")
	expectNoValue(testAppName, "testKey3")
}

func TestEnvironmentLoading(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t)

	env, err := LoadMergedAppEnv(testAppName)
	Expect(err).To(Succeed())
	v, _ := env.Get("testKey")
	Expect(v).To(Equal("TESTING"))
	v, _ = env.Get("globalKey")
	Expect(v).To(Equal("GLOBAL_VALUE"))
	Expect(env.Write()).ToNot(Succeed())

	env, err = LoadAppEnv(testAppName)
	env.Set("testKey", "TESTING-updated")
	env.Set("testKey2", "TESTING-'\n'-updated")
	Expect(env.Write()).To(Succeed())

	expectValue(testAppName, "testKey", "TESTING-updated")
	expectValue(testAppName, "testKey2", "TESTING-'\n'-updated")
	expectValue("", "testKey", "GLOBAL_TESTING")
	Expect(err).To(Succeed())
}

func TestInvalidKeys(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t)

	invalidKeys := []string{"0invalidKey", "invalid:key", "invalid=Key", "!invalidKey"}
	for _, key := range invalidKeys {
		Expect(SetMany(testAppName, map[string]string{key: "value"}, false, false)).NotTo(Succeed())
		Expect(UnsetMany(testAppName, []string{key}, false)).NotTo(Succeed())
		value, ok := Get(testAppName, key)
		Expect(ok).To(Equal(false))
		Expect(value).To(Equal(""))
		value2 := GetWithDefault(testAppName, key, "default")
		Expect(value2).To(Equal("default"))
	}
}

func TestInvalidEnvOnDisk(t *testing.T) {
	RegisterTestingT(t)
	testAppDir := setupTestApp(t)

	appConfigFile := filepath.Join(testAppDir, "ENV")
	b := []byte("export --invalid-key=TESTING\nexport valid_key=value\n")
	if err := os.WriteFile(appConfigFile, b, 0644); err != nil {
		return
	}

	env, err := LoadAppEnv(testAppName)
	Expect(err).NotTo(HaveOccurred())
	_, ok := env.Get("--invalid-key")
	Expect(ok).To(Equal(false))
	value, ok := env.Get("valid_key")
	Expect(ok).To(Equal(true))
	Expect(value).To(Equal("value"))

	//LoadAppEnv eliminates it from the file
	content, err := os.ReadFile(appConfigFile)
	Expect(err).NotTo(HaveOccurred())
	Expect(strings.Contains(string(content), "--invalid-key")).To(BeFalse())

}

// TestConfigSetManyReportsWriteFailure covers the reported bug: a write that
// never lands used to print "Setting config vars", fire the update trigger and
// exit successfully. The app's config directory path is occupied by a regular
// file so the write fails with ENOTDIR, which root cannot bypass; the read
// treats the same path as an empty environment and succeeds.
func TestConfigSetManyReportsWriteFailure(t *testing.T) {
	RegisterTestingT(t)
	_, libRoot := setupIsolatedEnv(t)

	Expect(os.MkdirAll(filepath.Join(libRoot, "config"), 0755)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(libRoot, "config", "alpha"), []byte(""), 0600)).To(Succeed())

	Expect(SetMany("alpha", map[string]string{"testKey": "value"}, false, false)).ToNot(Succeed())
	expectNoValue("alpha", "testKey")
}

// TestConfigUnsetManyReportsWriteFailure and its UnsetAll counterpart cover the
// same bug for the removal paths. No filesystem state lets the read load a
// non-empty environment and still fail the write for root, because both go
// through one path resolved once at load, so the permission step inside Write
// is made to fail instead. That step runs against the staged temporary file, so
// the environment on disk is left untouched.
func TestConfigUnsetManyReportsWriteFailure(t *testing.T) {
	RegisterTestingT(t)
	setupIsolatedEnv(t)
	setupIsolatedApp(t, "alpha")

	t.Setenv("DOKKU_SYSTEM_USER", "dokku-no-such-user")
	Expect(UnsetMany("alpha", []string{"testKey"}, false)).ToNot(Succeed())
	expectValue("alpha", "testKey", "TESTING")
}

func TestConfigUnsetAllReportsWriteFailure(t *testing.T) {
	RegisterTestingT(t)
	setupIsolatedEnv(t)
	setupIsolatedApp(t, "alpha")

	t.Setenv("DOKKU_SYSTEM_USER", "dokku-no-such-user")
	Expect(UnsetAll("alpha", false)).ToNot(Succeed())
	expectValue("alpha", "testKey", "TESTING")
}

// TestTriggerConfigUnsetReportsFailure covers the config-unset plugin trigger,
// which discarded the error it was handed and always exited successfully. An
// invalid key fails validation before anything is read or written.
func TestTriggerConfigUnsetReportsFailure(t *testing.T) {
	RegisterTestingT(t)
	setupIsolatedEnv(t)

	Expect(TriggerConfigUnset("alpha", "invalid-key", false)).ToNot(Succeed())
}

// setupIsolatedApp creates the config directory for an app inside an isolated
// environment and seeds it with a value to remove
func setupIsolatedApp(t *testing.T, appName string) {
	t.Helper()

	Expect(setupAppConfigDir(appName)).To(Succeed())
	Expect(SetMany(appName, map[string]string{"testKey": "TESTING"}, false, false)).To(Succeed())
}

func expectValue(appName string, key string, expected string) {
	v, ok := Get(appName, key)
	Expect(ok).To(Equal(true))
	Expect(v).To(Equal(expected))
}

func expectNoValue(appName string, key string) {
	_, ok := Get(appName, key)
	Expect(ok).To(Equal(false))
}
