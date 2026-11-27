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
	stubGroupEnv(t, map[string]string{})
	stubManagedKeys(t, "")
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
	Expect(CommandSet(testAppName, vals, false, true, false, false)).To(Succeed())
	expectValue(testAppName, "testKey", "updated")
	expectValue(testAppName, "testKey2", "new")

	vals = []string{"testKey=updated_global", "testKey2=new_global"}
	Expect(CommandSet("", vals, true, true, false, false)).To(Succeed())
	expectValue("", "testKey", "updated_global")
	expectValue("", "testKey2", "new_global")
	expectValue("", "globalKey", "GLOBAL_VALUE")
	expectValue(testAppName, "testKey", "updated")
	expectValue(testAppName, "testKey2", "new")

	Expect(CommandSet(testAppName+"does_not_exist", vals, false, true, false, false)).ToNot(Succeed())
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

func TestReplaceEnvFileEntries(t *testing.T) {
	RegisterTestingT(t)
	_, libRoot := setupIsolatedEnv(t)

	filename := filepath.Join(libRoot, "config", "replace-app", "ENV")
	Expect(os.MkdirAll(filepath.Dir(filename), 0755)).To(Succeed())
	original := []byte("export A=1\nexport B=2\nexport C=3\n")
	Expect(os.WriteFile(filename, original, 0600)).To(Succeed())

	changed, removed, err := ReplaceEnvFileEntries("replace-app", filename, map[string]string{"A": "1", "B": "2", "C": "3"}, nil)
	Expect(err).To(Succeed())
	Expect(changed).To(BeFalse())
	Expect(removed).To(BeEmpty())
	contents, err := os.ReadFile(filename)
	Expect(err).To(Succeed())
	Expect(contents).To(Equal(original))

	changed, removed, err = ReplaceEnvFileEntries("replace-app", filename, map[string]string{"A": "1", "B": "20", "C": "3"}, nil)
	Expect(err).To(Succeed())
	Expect(changed).To(BeTrue())
	Expect(removed).To(BeEmpty())

	changed, removed, err = ReplaceEnvFileEntries("replace-app", filename, map[string]string{"B": "20"}, nil)
	Expect(err).To(Succeed())
	Expect(changed).To(BeTrue())
	Expect(removed).To(Equal([]string{"A", "C"}))
	env, err := LoadEnvFile("replace-app", filename)
	Expect(err).To(Succeed())
	Expect(env.Map()).To(Equal(map[string]string{"B": "20"}))

	_, _, err = ReplaceEnvFileEntries("replace-app", filename, map[string]string{"1BAD": "x"}, nil)
	Expect(err).To(MatchError("Invalid key name: '1BAD'"))
	env, _ = LoadEnvFile("replace-app", filename)
	Expect(env.Map()).To(Equal(map[string]string{"B": "20"}))
}

func TestReplaceEnvFileEntriesPreserve(t *testing.T) {
	RegisterTestingT(t)
	_, libRoot := setupIsolatedEnv(t)

	filename := filepath.Join(libRoot, "config", "replace-app", "ENV")
	Expect(os.MkdirAll(filepath.Dir(filename), 0755)).To(Succeed())
	Expect(os.WriteFile(filename, []byte("export GIT_REV=abc\nexport NO_VHOST=1\nexport OLD=x\n"), 0600)).To(Succeed())

	preserve := []string{"GIT_REV", "NO_VHOST", "MISSING"}
	changed, removed, err := ReplaceEnvFileEntries("replace-app", filename, map[string]string{"NEW": "y"}, preserve)
	Expect(err).To(Succeed())
	Expect(changed).To(BeTrue())
	Expect(removed).To(Equal([]string{"OLD"}))
	env, err := LoadEnvFile("replace-app", filename)
	Expect(err).To(Succeed())
	Expect(env.Map()).To(Equal(map[string]string{"GIT_REV": "abc", "NO_VHOST": "1", "NEW": "y"}))

	changed, _, err = ReplaceEnvFileEntries("replace-app", filename, map[string]string{"NEW": "y"}, preserve)
	Expect(err).To(Succeed())
	Expect(changed).To(BeFalse())

	changed, removed, err = ReplaceEnvFileEntries("replace-app", filename, map[string]string{"NEW": "y", "GIT_REV": "def"}, preserve)
	Expect(err).To(Succeed())
	Expect(changed).To(BeTrue())
	Expect(removed).To(BeEmpty())
	env, _ = LoadEnvFile("replace-app", filename)
	Expect(env.Map()).To(Equal(map[string]string{"GIT_REV": "def", "NO_VHOST": "1", "NEW": "y"}))

	changed, removed, err = ReplaceEnvFileEntries("replace-app", filename, map[string]string{"NEW": "y"}, nil)
	Expect(err).To(Succeed())
	Expect(changed).To(BeTrue())
	Expect(removed).To(Equal([]string{"GIT_REV", "NO_VHOST"}))
}

// stubManagedKeys replaces the config-managed-keys trigger with fixed output
// so tests can run without plugn
func stubManagedKeys(t *testing.T, output string) {
	t.Helper()

	original := managedKeysLoader
	managedKeysLoader = func(appName string) (string, error) {
		return output, nil
	}
	t.Cleanup(func() {
		managedKeysLoader = original
	})
}

func TestManagedKeys(t *testing.T) {
	RegisterTestingT(t)
	stubManagedKeys(t, "GIT_REV\n\n  NO_VHOST  \nGIT_REV\nnot-a-key\n")

	keys, err := ManagedKeys(testAppName)
	Expect(err).To(Succeed())
	Expect(keys).To(Equal([]string{"GIT_REV", "NO_VHOST"}))

	keys, err = ManagedKeys("--global")
	Expect(err).To(Succeed())
	Expect(keys).To(BeEmpty())
}

func TestConfigSetReplace(t *testing.T) {
	RegisterTestingT(t)
	testAppDir := setupTestApp(t)
	Expect(os.WriteFile(filepath.Join(testAppDir, "ENV"), []byte("export GIT_REV=abc\nexport OLD=x\nexport testKey=TESTING\n"), 0600)).To(Succeed())
	stubManagedKeys(t, "GIT_REV\n")

	Expect(CommandSet(testAppName, []string{"testKey=NEW"}, false, true, false, true)).To(Succeed())
	expectValue(testAppName, "testKey", "NEW")
	expectValue(testAppName, "GIT_REV", "abc")
	expectNoValue(testAppName, "OLD")

	Expect(CommandSet(testAppName, []string{}, false, true, false, true)).To(MatchError("At least one env pair must be given"))
}
