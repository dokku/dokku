package common

import (
	"os"
	"path/filepath"
	"testing"

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
