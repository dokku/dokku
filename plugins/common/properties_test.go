package common

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

const (
	propertyTestPlugin = "test-plugin"
	propertyTestApp    = "--global"
	propertyTestName   = "chart-overrides.traefik"
)

func setupPropertyMapTest(t *testing.T) {
	t.Helper()
	RegisterTestingT(t)
	Expect(setupTests()).To(Succeed())
	t.Setenv("DOKKU_LIB_ROOT", t.TempDir())
	t.Setenv("DOKKU_SYSTEM_USER", "root")
	t.Setenv("DOKKU_SYSTEM_GROUP", "root")
	Expect(PropertySetup(propertyTestPlugin)).To(Succeed())
}

func TestPropertyMapWriteAndGetRoundTrip(t *testing.T) {
	setupPropertyMapTest(t)

	want := map[string]string{"installCRDs": "false", "version": "1.13.3"}
	Expect(PropertyMapWrite(propertyTestPlugin, propertyTestApp, propertyTestName, want)).To(Succeed())

	got, err := PropertyMapGet(propertyTestPlugin, propertyTestApp, propertyTestName)
	Expect(err).NotTo(HaveOccurred())
	Expect(got).To(Equal(want))
}

func TestPropertyMapHandlesSlashInKey(t *testing.T) {
	setupPropertyMapTest(t)

	key := "service.annotations.prometheus.io/scrape"
	Expect(PropertyMapSet(propertyTestPlugin, propertyTestApp, propertyTestName, key, "true")).To(Succeed())

	got, err := PropertyMapGet(propertyTestPlugin, propertyTestApp, propertyTestName)
	Expect(err).NotTo(HaveOccurred())
	Expect(got).To(HaveKeyWithValue(key, "true"))
}

func TestPropertyMapHandlesNewlineInValue(t *testing.T) {
	setupPropertyMapTest(t)

	value := "line one\nline two\nline three"
	Expect(PropertyMapSet(propertyTestPlugin, propertyTestApp, propertyTestName, "controller.config", value)).To(Succeed())

	got, err := PropertyMapGet(propertyTestPlugin, propertyTestApp, propertyTestName)
	Expect(err).NotTo(HaveOccurred())
	Expect(got).To(HaveKeyWithValue("controller.config", value))
}

func TestPropertyMapEmptyMap(t *testing.T) {
	setupPropertyMapTest(t)

	Expect(PropertyMapWrite(propertyTestPlugin, propertyTestApp, propertyTestName, map[string]string{})).To(Succeed())

	got, err := PropertyMapGet(propertyTestPlugin, propertyTestApp, propertyTestName)
	Expect(err).NotTo(HaveOccurred())
	Expect(got).To(BeEmpty())
}

func TestPropertyMapNilMapPersistsAsEmptyJSON(t *testing.T) {
	setupPropertyMapTest(t)

	Expect(PropertyMapWrite(propertyTestPlugin, propertyTestApp, propertyTestName, nil)).To(Succeed())

	b, err := os.ReadFile(getPropertyPath(propertyTestPlugin, propertyTestApp, propertyTestName))
	Expect(err).NotTo(HaveOccurred())
	Expect(string(b)).To(Equal("{}"))
}

func TestPropertyMapGetMissingReturnsEmpty(t *testing.T) {
	setupPropertyMapTest(t)

	got, err := PropertyMapGet(propertyTestPlugin, propertyTestApp, "chart-overrides.never-written")
	Expect(err).NotTo(HaveOccurred())
	Expect(got).To(BeEmpty())
}

func TestPropertyMapSetCreatesMissingFile(t *testing.T) {
	setupPropertyMapTest(t)

	Expect(PropertyExists(propertyTestPlugin, propertyTestApp, propertyTestName)).To(BeFalse())
	Expect(PropertyMapSet(propertyTestPlugin, propertyTestApp, propertyTestName, "version", "1.0")).To(Succeed())
	Expect(PropertyExists(propertyTestPlugin, propertyTestApp, propertyTestName)).To(BeTrue())

	got, err := PropertyMapGet(propertyTestPlugin, propertyTestApp, propertyTestName)
	Expect(err).NotTo(HaveOccurred())
	Expect(got).To(Equal(map[string]string{"version": "1.0"}))
}

func TestPropertyMapDeleteRemovesKeyOnly(t *testing.T) {
	setupPropertyMapTest(t)

	Expect(PropertyMapSet(propertyTestPlugin, propertyTestApp, propertyTestName, "version", "1.0")).To(Succeed())
	Expect(PropertyMapSet(propertyTestPlugin, propertyTestApp, propertyTestName, "installCRDs", "false")).To(Succeed())

	Expect(PropertyMapDelete(propertyTestPlugin, propertyTestApp, propertyTestName, "version")).To(Succeed())

	got, err := PropertyMapGet(propertyTestPlugin, propertyTestApp, propertyTestName)
	Expect(err).NotTo(HaveOccurred())
	Expect(got).To(Equal(map[string]string{"installCRDs": "false"}))
}

func TestPropertyMapDeleteMissingKeyIsNoOp(t *testing.T) {
	setupPropertyMapTest(t)

	Expect(PropertyMapSet(propertyTestPlugin, propertyTestApp, propertyTestName, "version", "1.0")).To(Succeed())

	Expect(PropertyMapDelete(propertyTestPlugin, propertyTestApp, propertyTestName, "missing")).To(Succeed())
	Expect(PropertyMapDelete(propertyTestPlugin, propertyTestApp, "chart-overrides.never-written", "anything")).To(Succeed())

	got, err := PropertyMapGet(propertyTestPlugin, propertyTestApp, propertyTestName)
	Expect(err).NotTo(HaveOccurred())
	Expect(got).To(Equal(map[string]string{"version": "1.0"}))
}

func TestPropertyMapLength(t *testing.T) {
	setupPropertyMapTest(t)

	length, err := PropertyMapLength(propertyTestPlugin, propertyTestApp, "chart-overrides.never-written")
	Expect(err).NotTo(HaveOccurred())
	Expect(length).To(Equal(0))

	Expect(PropertyMapSet(propertyTestPlugin, propertyTestApp, propertyTestName, "a", "1")).To(Succeed())
	Expect(PropertyMapSet(propertyTestPlugin, propertyTestApp, propertyTestName, "b", "2")).To(Succeed())

	length, err = PropertyMapLength(propertyTestPlugin, propertyTestApp, propertyTestName)
	Expect(err).NotTo(HaveOccurred())
	Expect(length).To(Equal(2))
}

func TestPropertyMapRewriteIsStable(t *testing.T) {
	setupPropertyMapTest(t)

	Expect(PropertyMapSet(propertyTestPlugin, propertyTestApp, propertyTestName, "version", "1.0")).To(Succeed())
	Expect(PropertyMapSet(propertyTestPlugin, propertyTestApp, propertyTestName, "version", "1.0")).To(Succeed())

	got, err := PropertyMapGet(propertyTestPlugin, propertyTestApp, propertyTestName)
	Expect(err).NotTo(HaveOccurred())
	Expect(got).To(Equal(map[string]string{"version": "1.0"}))
}

func TestPropertyMapFileMode(t *testing.T) {
	setupPropertyMapTest(t)

	Expect(PropertyMapSet(propertyTestPlugin, propertyTestApp, propertyTestName, "version", "1.0")).To(Succeed())

	fi, err := os.Stat(filepath.Clean(getPropertyPath(propertyTestPlugin, propertyTestApp, propertyTestName)))
	Expect(err).NotTo(HaveOccurred())
	Expect(fi.Mode().Perm()).To(Equal(os.FileMode(0600)))
}

func TestPropertyMapMalformedJSONReturnsError(t *testing.T) {
	setupPropertyMapTest(t)

	Expect(PropertyMapWrite(propertyTestPlugin, propertyTestApp, propertyTestName, map[string]string{"k": "v"})).To(Succeed())
	Expect(os.WriteFile(getPropertyPath(propertyTestPlugin, propertyTestApp, propertyTestName), []byte("not-json"), 0600)).To(Succeed())

	_, err := PropertyMapGet(propertyTestPlugin, propertyTestApp, propertyTestName)
	Expect(err).To(HaveOccurred())
}

func TestPropertyListAddConcurrent(t *testing.T) {
	setupPropertyMapTest(t)

	Expect(runConcurrently(50, func(i int) error {
		return PropertyListAdd(propertyTestPlugin, propertyTestApp, "list", fmt.Sprintf("value-%d", i), 0)
	})).To(Succeed())

	lines, err := PropertyListGet(propertyTestPlugin, propertyTestApp, "list")
	Expect(err).NotTo(HaveOccurred())
	Expect(lines).To(HaveLen(50))
	for i := 0; i < 50; i++ {
		Expect(lines).To(ContainElement(fmt.Sprintf("value-%d", i)))
	}
}

func TestPropertyListRemoveConcurrent(t *testing.T) {
	setupPropertyMapTest(t)

	values := []string{}
	for i := 0; i < 50; i++ {
		values = append(values, fmt.Sprintf("value-%d", i))
	}
	Expect(PropertyListWrite(propertyTestPlugin, propertyTestApp, "list", values)).To(Succeed())

	Expect(runConcurrently(len(values), func(i int) error {
		return PropertyListRemove(propertyTestPlugin, propertyTestApp, "list", values[i])
	})).To(Succeed())

	lines, err := PropertyListGet(propertyTestPlugin, propertyTestApp, "list")
	Expect(err).NotTo(HaveOccurred())
	Expect(lines).To(BeEmpty())
}

func TestPropertyListRemoveMissingDoesNotCreateDirectory(t *testing.T) {
	setupPropertyMapTest(t)

	Expect(PropertyListRemove(propertyTestPlugin, "missing-app", "list", "value")).NotTo(Succeed())
	Expect(PropertyListRemoveByPrefix(propertyTestPlugin, "missing-app", "list", "value")).NotTo(Succeed())
	Expect(PropertyMapDelete(propertyTestPlugin, "missing-app", "map", "key")).To(Succeed())
	Expect(DirectoryExists(getPluginAppPropertyPath(propertyTestPlugin, "missing-app"))).To(BeFalse())
}

func TestPropertyListRemoveNotFoundLeavesList(t *testing.T) {
	setupPropertyMapTest(t)

	Expect(PropertyListWrite(propertyTestPlugin, propertyTestApp, "list", []string{"a", "b"})).To(Succeed())
	Expect(PropertyListRemove(propertyTestPlugin, propertyTestApp, "list", "c")).To(MatchError("Property not found, nothing was removed"))

	lines, err := PropertyListGet(propertyTestPlugin, propertyTestApp, "list")
	Expect(err).NotTo(HaveOccurred())
	Expect(lines).To(Equal([]string{"a", "b"}))
}

func TestPropertyListUpdateErrorLeavesListUnchanged(t *testing.T) {
	setupPropertyMapTest(t)

	Expect(PropertyListWrite(propertyTestPlugin, propertyTestApp, "list", []string{"a", "b"})).To(Succeed())
	err := PropertyListUpdate(propertyTestPlugin, propertyTestApp, "list", func(lines []string) ([]string, error) {
		return nil, errors.New("boom")
	})
	Expect(err).To(MatchError("boom"))

	lines, err := PropertyListGet(propertyTestPlugin, propertyTestApp, "list")
	Expect(err).NotTo(HaveOccurred())
	Expect(lines).To(Equal([]string{"a", "b"}))
}

func TestPropertyMapSetConcurrent(t *testing.T) {
	setupPropertyMapTest(t)

	Expect(runConcurrently(50, func(i int) error {
		return PropertyMapSet(propertyTestPlugin, propertyTestApp, propertyTestName, fmt.Sprintf("key-%d", i), "value")
	})).To(Succeed())

	got, err := PropertyMapGet(propertyTestPlugin, propertyTestApp, propertyTestName)
	Expect(err).NotTo(HaveOccurred())
	Expect(got).To(HaveLen(50))
}

func TestPropertyLockSerializesReadModifyWrite(t *testing.T) {
	setupPropertyMapTest(t)

	Expect(runConcurrently(50, func(i int) error {
		return PropertyLock(propertyTestPlugin, propertyTestApp, func(p *LockedProperties) error {
			current, err := strconv.Atoi(PropertyGetDefault(propertyTestPlugin, propertyTestApp, "counter", "0"))
			if err != nil {
				return err
			}
			return p.Write("counter", strconv.Itoa(current+1))
		})
	})).To(Succeed())

	Expect(PropertyGet(propertyTestPlugin, propertyTestApp, "counter")).To(Equal("50"))
}

func TestLockDirectorySkipsDirectoriesHeldByParent(t *testing.T) {
	setupPropertyMapTest(t)

	dir := getPluginAppPropertyPath(propertyTestPlugin, propertyTestApp)
	Expect(makePluginAppPropertyPath(propertyTestPlugin, propertyTestApp)).To(Succeed())
	t.Setenv(LockedDirectoriesEnvVar, "/some/other/dir\n"+dir+"/")

	done := make(chan error, 1)
	go func() {
		done <- LockDirectory(dir, func() error {
			return PropertyWrite(propertyTestPlugin, propertyTestApp, "nested", "value")
		})
	}()

	select {
	case err := <-done:
		Expect(err).NotTo(HaveOccurred())
	case <-time.After(5 * time.Second):
		t.Fatal("nested lock on a directory listed in DOKKU_LOCKED_DIRECTORIES blocked")
	}
	Expect(PropertyGet(propertyTestPlugin, propertyTestApp, "nested")).To(Equal("value"))
}

func TestPropertyWritesLeaveNoTemporaryFiles(t *testing.T) {
	setupPropertyMapTest(t)

	Expect(PropertyWrite(propertyTestPlugin, propertyTestApp, "plain", "value")).To(Succeed())
	Expect(PropertyListWrite(propertyTestPlugin, propertyTestApp, "list", []string{"a"})).To(Succeed())
	Expect(PropertyMapSet(propertyTestPlugin, propertyTestApp, propertyTestName, "key", "value")).To(Succeed())

	entries, err := os.ReadDir(getPluginAppPropertyPath(propertyTestPlugin, propertyTestApp))
	Expect(err).NotTo(HaveOccurred())
	names := []string{}
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	Expect(names).To(ConsistOf("plain", "list", propertyTestName))
}

func TestPropertyGetAllIgnoresHiddenFiles(t *testing.T) {
	setupPropertyMapTest(t)

	Expect(PropertyWrite(propertyTestPlugin, propertyTestApp, "plain", "value")).To(Succeed())
	stray := filepath.Join(getPluginAppPropertyPath(propertyTestPlugin, propertyTestApp), ".plain.tmp-123")
	Expect(os.WriteFile(stray, []byte("partial"), 0600)).To(Succeed())

	all, err := PropertyGetAll(propertyTestPlugin, propertyTestApp)
	Expect(err).NotTo(HaveOccurred())
	Expect(all).To(Equal(map[string]string{"plain": "value"}))

	byPrefix, err := PropertyGetAllByPrefix(propertyTestPlugin, propertyTestApp, "")
	Expect(err).NotTo(HaveOccurred())
	Expect(byPrefix).To(Equal(map[string]string{"plain": "value"}))
}

// runConcurrently calls fn with 0 through n-1 from n goroutines started
// together, returning the joined errors of every call.
func runConcurrently(n int, fn func(i int) error) error {
	var wg sync.WaitGroup
	start := make(chan struct{})
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			errs[i] = fn(i)
		}(i)
	}
	close(start)
	wg.Wait()
	return errors.Join(errs...)
}
