package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
)

// stubGroupEnv replaces the env group layer loader for the duration of a test
// so merged environments can be loaded without plugn
func stubGroupEnv(t *testing.T, groups map[string]string) {
	t.Helper()

	original := groupEnvLoader
	groupEnvLoader = func(appName string) (map[string]string, error) {
		return groups, nil
	}
	t.Cleanup(func() {
		groupEnvLoader = original
	})
}

func TestMergedEnvGroupPrecedence(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t)
	stubGroupEnv(t, map[string]string{
		"testKey":   "GROUP_TESTING",
		"globalKey": "GROUP_VALUE",
		"groupKey":  "GROUP_ONLY",
	})

	env, err := LoadMergedAppEnv(testAppName)
	Expect(err).To(Succeed())

	v, _ := env.Get("testKey")
	Expect(v).To(Equal("TESTING"))
	v, _ = env.Get("globalKey")
	Expect(v).To(Equal("GROUP_VALUE"))
	v, _ = env.Get("groupKey")
	Expect(v).To(Equal("GROUP_ONLY"))
	Expect(env.Write()).ToNot(Succeed())

	appEnv, err := LoadAppEnv(testAppName)
	Expect(err).To(Succeed())
	_, ok := appEnv.Get("groupKey")
	Expect(ok).To(BeFalse())
}

func TestMergedEnvGroupError(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t)

	original := groupEnvLoader
	groupEnvLoader = func(appName string) (map[string]string, error) {
		return nil, errors.New("group failure")
	}
	t.Cleanup(func() {
		groupEnvLoader = original
	})

	_, err := LoadMergedAppEnv(testAppName)
	Expect(err).To(MatchError("group failure"))
}

func TestParseEnvPairs(t *testing.T) {
	RegisterTestingT(t)

	pairs, err := ParseEnvPairs([]string{"A=1", "B=two=2", "C="}, false)
	Expect(err).To(Succeed())
	Expect(pairs).To(Equal(map[string]string{"A": "1", "B": "two=2", "C": ""}))

	pairs, err = ParseEnvPairs([]string{"A=aGVsbG8="}, true)
	Expect(err).To(Succeed())
	Expect(pairs).To(Equal(map[string]string{"A": "hello"}))

	_, err = ParseEnvPairs([]string{"A"}, false)
	Expect(err).To(MatchError("Invalid env pair: A"))

	_, err = ParseEnvPairs([]string{"A=!!"}, true)
	Expect(err).To(HaveOccurred())
}

func TestParseExportFormat(t *testing.T) {
	RegisterTestingT(t)

	format, err := ParseExportFormat("json")
	Expect(err).To(Succeed())
	Expect(format).To(Equal(ExportFormatJSON))

	format, err = ParseExportFormat("shell")
	Expect(err).To(Succeed())
	Expect(format).To(Equal(ExportFormatShell))

	_, err = ParseExportFormat("yaml")
	Expect(err).To(MatchError("Unknown export format: yaml"))
}

func TestEnvFileHelpers(t *testing.T) {
	RegisterTestingT(t)
	_, libRoot := setupIsolatedEnv(t)

	filename := filepath.Join(libRoot, "data", "envgroup", "dev", "ENV")
	Expect(os.MkdirAll(filepath.Dir(filename), 0755)).To(Succeed())

	changed, err := SetEnvFileEntries("dev", filename, map[string]string{"A": "1", "B": "2"}, false)
	Expect(err).To(Succeed())
	Expect(changed).To(BeTrue())

	env, err := LoadEnvFile("dev", filename)
	Expect(err).To(Succeed())
	Expect(env.Map()).To(Equal(map[string]string{"A": "1", "B": "2"}))

	changed, err = SetEnvFileEntries("dev", filename, map[string]string{"C": "3"}, true)
	Expect(err).To(Succeed())
	Expect(changed).To(BeTrue())
	env, _ = LoadEnvFile("dev", filename)
	Expect(env.Map()).To(Equal(map[string]string{"C": "3"}))

	_, err = SetEnvFileEntries("dev", filename, map[string]string{"1BAD": "x"}, false)
	Expect(err).To(MatchError("Invalid key name: '1BAD'"))

	changed, err = UnsetEnvFileKeys("dev", filename, []string{"MISSING"})
	Expect(err).To(Succeed())
	Expect(changed).To(BeFalse())

	changed, err = UnsetEnvFileKeys("dev", filename, []string{"C"})
	Expect(err).To(Succeed())
	Expect(changed).To(BeTrue())

	changed, err = ClearEnvFile("dev", filename)
	Expect(err).To(Succeed())
	Expect(changed).To(BeFalse())

	changed, err = UpdateEnvFile("dev", filename, func(env *Env) bool {
		env.Set("D", "4")
		return true
	})
	Expect(err).To(Succeed())
	Expect(changed).To(BeTrue())

	changed, err = ClearEnvFile("dev", filename)
	Expect(err).To(Succeed())
	Expect(changed).To(BeTrue())
	env, _ = LoadEnvFile("dev", filename)
	Expect(env.Len()).To(Equal(0))
}

func TestNewEnvFromMap(t *testing.T) {
	RegisterTestingT(t)

	source := map[string]string{"A": "1"}
	env := NewEnvFromMap("dev", source)
	env.Set("B", "2")
	Expect(source).To(Equal(map[string]string{"A": "1"}))
	Expect(env.Keys()).To(Equal([]string{"A", "B"}))
	Expect(env.Write()).ToNot(Succeed())
}
