package storage

import (
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
)

// stageDockerLocalEntry writes a docker-local entry at the default host
// path so the metadata commands have something to operate on.
func stageDockerLocalEntry(t *testing.T, name string) *Entry {
	t.Helper()
	entry := &Entry{
		Name:      name,
		Scheduler: SchedulerDockerLocal,
		HostPath:  filepath.Join(GetStorageDirectory(), name),
	}
	if err := SaveEntry(entry); err != nil {
		t.Fatalf("SaveEntry: %v", err)
	}
	return entry
}

func TestSetEntryMapKeyAddsAndOverwrites(t *testing.T) {
	RegisterTestingT(t)
	withTempLibRoot(t)
	stageDockerLocalEntry(t, "demo")

	Expect(setEntryMapKey(annotationsField, "demo", "first", "one")).To(Succeed())
	Expect(setEntryMapKey(annotationsField, "demo", "second", "two")).To(Succeed())

	loaded, err := LoadEntry("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(loaded.Annotations).To(Equal(map[string]string{"first": "one", "second": "two"}))

	Expect(setEntryMapKey(annotationsField, "demo", "first", "rewritten")).To(Succeed())
	loaded, err = LoadEntry("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(loaded.Annotations).To(Equal(map[string]string{"first": "rewritten", "second": "two"}))
}

// TestSetEntryMapKeyDeletesOneKey is the behavior the wholesale
// --annotation flag could never express: clearing one key without
// disturbing its siblings.
func TestSetEntryMapKeyDeletesOneKey(t *testing.T) {
	RegisterTestingT(t)
	withTempLibRoot(t)
	stageDockerLocalEntry(t, "demo")

	Expect(setEntryMapKey(annotationsField, "demo", "first", "one")).To(Succeed())
	Expect(setEntryMapKey(annotationsField, "demo", "second", "two")).To(Succeed())
	Expect(setEntryMapKey(annotationsField, "demo", "first", "")).To(Succeed())

	loaded, err := LoadEntry("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(loaded.Annotations).To(Equal(map[string]string{"second": "two"}))
}

// TestSetEntryMapKeyDropsEmptyMap keeps the omitempty tag honest: once the
// last key is gone the field should be absent from the JSON, not an empty
// object, and should still round-trip.
func TestSetEntryMapKeyDropsEmptyMap(t *testing.T) {
	RegisterTestingT(t)
	withTempLibRoot(t)
	stageDockerLocalEntry(t, "demo")

	Expect(setEntryMapKey(labelsField, "demo", "only", "value")).To(Succeed())
	Expect(setEntryMapKey(labelsField, "demo", "only", "")).To(Succeed())

	loaded, err := LoadEntry("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(loaded.Labels).To(BeEmpty())

	Expect(SaveEntry(loaded)).To(Succeed())
	reloaded, err := LoadEntry("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(reloaded.Labels).To(BeEmpty())
}

func TestSetEntryMapKeyKeepsAnnotationsAndLabelsSeparate(t *testing.T) {
	RegisterTestingT(t)
	withTempLibRoot(t)
	stageDockerLocalEntry(t, "demo")

	Expect(setEntryMapKey(annotationsField, "demo", "shared", "annotation")).To(Succeed())
	Expect(setEntryMapKey(labelsField, "demo", "shared", "label")).To(Succeed())

	loaded, err := LoadEntry("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(loaded.Annotations).To(Equal(map[string]string{"shared": "annotation"}))
	Expect(loaded.Labels).To(Equal(map[string]string{"shared": "label"}))
}

// TestSetEntryMapKeyPreservesSlashKeys covers Kubernetes-style keys, which
// are the common case for both annotations and labels.
func TestSetEntryMapKeyPreservesSlashKeys(t *testing.T) {
	RegisterTestingT(t)
	withTempLibRoot(t)
	stageDockerLocalEntry(t, "demo")

	key := "backup.velero.io/backup-volumes"
	Expect(setEntryMapKey(annotationsField, "demo", key, "demo")).To(Succeed())

	loaded, err := LoadEntry("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(loaded.Annotations).To(HaveKeyWithValue(key, "demo"))
}

func TestSetEntryMapKeyValidatesInputs(t *testing.T) {
	RegisterTestingT(t)
	withTempLibRoot(t)
	stageDockerLocalEntry(t, "demo")

	err := setEntryMapKey(annotationsField, "", "key", "value")
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("name is required"))

	err = setEntryMapKey(annotationsField, "demo", "", "value")
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("No annotation key specified"))

	err = setEntryMapKey(labelsField, "demo", "", "value")
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("No label key specified"))

	err = setEntryMapKey(annotationsField, "missing", "key", "value")
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("does not exist"))
}

func TestReportEntryMapValidatesFormatAndInfoFlag(t *testing.T) {
	RegisterTestingT(t)
	withTempLibRoot(t)
	stageDockerLocalEntry(t, "demo")

	err := reportEntryMap(annotationsField, "demo", "yaml", "")
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("Invalid format"))

	err = reportEntryMap(annotationsField, "demo", "json", "--storage-annotations.first")
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("cannot be specified when specifying an info flag"))

	// An info flag has no single answer across every entry, so it needs a name.
	err = reportEntryMap(annotationsField, "", "stdout", "--storage-annotations.first")
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("requires a storage entry name"))

	err = reportEntryMap(annotationsField, "missing", "stdout", "")
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("does not exist"))
}

func TestReportEntryMapRejectsUnknownInfoFlag(t *testing.T) {
	RegisterTestingT(t)
	withTempLibRoot(t)
	stageDockerLocalEntry(t, "demo")
	Expect(setEntryMapKey(annotationsField, "demo", "first", "one")).To(Succeed())

	err := reportEntryMap(annotationsField, "demo", "stdout", "--storage-annotations.nope")
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("Invalid flag passed, valid flags: --storage-annotations.first"))
}
