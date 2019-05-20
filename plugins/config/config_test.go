package config

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/dokku/dokku/plugins/common"

	. "github.com/onsi/gomega"
)

var (
	testAppName      = "test-app-1"
	dokkuRoot        = common.MustGetEnv("DOKKU_ROOT")
	testAppDir       = strings.Join([]string{dokkuRoot, testAppName}, "/")
	globalConfigFile = strings.Join([]string{dokkuRoot, "ENV"}, "/")
)

func setupTestApp() (err error) {
	Expect(os.MkdirAll(testAppDir, 0766)).To(Succeed())
	b := []byte("export testKey=TESTING\n")
	if err = ioutil.WriteFile(strings.Join([]string{testAppDir, "/ENV"}, ""), b, 0644); err != nil {
		return
	}

	b = []byte("export testKey=GLOBAL_TESTING\nexport globalKey=GLOBAL_VALUE")
	if err = ioutil.WriteFile(globalConfigFile, b, 0644); err != nil {
		return
	}
	return
}

func teardownTestApp() {
	os.RemoveAll(testAppDir)
}

func TestConfigGetWithDefault(t *testing.T) {
	RegisterTestingT(t)
	Expect(setupTestApp()).To(Succeed())
	Expect(GetWithDefault(testAppName, "unknownKey", "UNKNOWN")).To(Equal("UNKNOWN"))
	Expect(GetWithDefault(testAppName, "testKey", "testKey")).To(Equal("TESTING"))
	Expect(GetWithDefault(testAppName+"-nonexistent", "testKey", "default")).To(Equal("default"))
	teardownTestApp()
}

func TestConfigGet(t *testing.T) {
	RegisterTestingT(t)
	Expect(setupTestApp()).To(Succeed())
	defer teardownTestApp()

	expectValue(testAppName, "testKey", "TESTING")
	expectValue("", "testKey", "GLOBAL_TESTING")

	expectNoValue(testAppName, "testKey2")
	expectNoValue("", "testKey2")
}

func TestConfigSetMany(t *testing.T) {
	RegisterTestingT(t)
	Expect(setupTestApp()).To(Succeed())
	defer teardownTestApp()

	expectValue(testAppName, "testKey", "TESTING")

	vals := map[string]string{"testKey": "updated", "testKey2": "new"}
	Expect(SetMany(testAppName, vals, false)).To(Succeed())
	expectValue(testAppName, "testKey", "updated")
	expectValue(testAppName, "testKey2", "new")

	vals = map[string]string{"testKey": "updated_global", "testKey2": "new_global"}
	Expect(SetMany("", vals, false)).To(Succeed())
	expectValue("", "testKey", "updated_global")
	expectValue("", "testKey2", "new_global")
	expectValue("", "globalKey", "GLOBAL_VALUE")
	expectValue(testAppName, "testKey", "updated")
	expectValue(testAppName, "testKey2", "new")

	Expect(SetMany(testAppName+"does_not_exist", vals, false)).ToNot(Succeed())
}

func TestConfigUnsetAll(t *testing.T) {
	RegisterTestingT(t)
	Expect(setupTestApp()).To(Succeed())
	defer teardownTestApp()

	expectValue(testAppName, "testKey", "TESTING")
	expectValue("", "testKey", "GLOBAL_TESTING")

	Expect(UnsetAll(testAppName, false)).To(Succeed())
	expectNoValue(testAppName, "testKey")
	expectNoValue(testAppName, "noKey")
	expectNoValue(testAppName, "globalKey")

	Expect(UnsetAll(testAppName+"does-not-exist", false)).ToNot(Succeed())
}

func TestConfigUnsetMany(t *testing.T) {
	RegisterTestingT(t)
	Expect(setupTestApp()).To(Succeed())
	defer teardownTestApp()

	expectValue(testAppName, "testKey", "TESTING")
	expectValue("", "testKey", "GLOBAL_TESTING")

	keys := []string{"testKey", "noKey"}
	Expect(UnsetMany(testAppName, keys, false)).To(Succeed())
	expectNoValue(testAppName, "testKey")
	expectValue("", "testKey", "GLOBAL_TESTING")

	Expect(UnsetMany(testAppName, keys, false)).To(Succeed())
	expectNoValue(testAppName, "testKey")
	expectNoValue(testAppName, "globalKey")

	Expect(UnsetMany(testAppName+"does-not-exist", keys, false)).ToNot(Succeed())
}

func TestEnvironmentLoading(t *testing.T) {
	RegisterTestingT(t)
	Expect(setupTestApp()).To(Succeed())
	defer teardownTestApp()

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
	env.Write()

	expectValue(testAppName, "testKey", "TESTING-updated")
	expectValue(testAppName, "testKey2", "TESTING-'\n'-updated")
	expectValue("", "testKey", "GLOBAL_TESTING")
	Expect(err).To(Succeed())
}

func TestInvalidKeys(t *testing.T) {
	RegisterTestingT(t)
	Expect(setupTestApp()).To(Succeed())
	defer teardownTestApp()

	invalidKeys := []string{"0invalidKey", "invalid:key", "invalid=Key", "!invalidKey"}
	for _, key := range invalidKeys {
		Expect(SetMany(testAppName, map[string]string{key: "value"}, false)).NotTo(Succeed())
		Expect(UnsetMany(testAppName, []string{key}, false)).NotTo(Succeed())
		value, ok := Get(testAppName, key)
		Expect(ok).To(Equal(false))
		value = GetWithDefault(testAppName, key, "default")
		Expect(value).To(Equal("default"))
	}
}

func TestInvalidEnvOnDisk(t *testing.T) {
	RegisterTestingT(t)
	Expect(setupTestApp()).To(Succeed())
	defer teardownTestApp()

	appConfigFile := strings.Join([]string{testAppDir, "/ENV"}, "")
	b := []byte("export --invalid-key=TESTING\nexport valid_key=value\n")
	if err := ioutil.WriteFile(appConfigFile, b, 0644); err != nil {
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
	content, err := ioutil.ReadFile(appConfigFile)
	Expect(err).NotTo(HaveOccurred())
	Expect(strings.Contains(string(content), "--invalid-key")).To(BeFalse())

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
