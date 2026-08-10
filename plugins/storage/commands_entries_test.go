package storage

import (
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
)

func TestEnsureDockerLocalPathCreatesDefaultPath(t *testing.T) {
	RegisterTestingT(t)
	withTempLibRoot(t)

	hostPath := filepath.Join(GetStorageDirectory(), "demo")
	entry := &Entry{Name: "demo", Scheduler: SchedulerDockerLocal, HostPath: hostPath}

	Expect(ensureDockerLocalPath(entry)).To(Succeed())
	Expect(hostPath).To(BeADirectory())

	// Idempotent: a second run leaves the existing directory in place.
	Expect(ensureDockerLocalPath(entry)).To(Succeed())
	Expect(hostPath).To(BeADirectory())
}

// TestEnsureDockerLocalPathSkipsNamedVolumes guards the docker named-volume
// case: the host path is a token the docker engine resolves, not a path on
// disk, so stat'ing and creating it would produce a stray directory in
// whatever working directory the command happened to run from.
func TestEnsureDockerLocalPathSkipsNamedVolumes(t *testing.T) {
	RegisterTestingT(t)
	withTempLibRoot(t)

	cwd := t.TempDir()
	t.Chdir(cwd)

	entry := &Entry{Name: "demo", Scheduler: SchedulerDockerLocal, HostPath: "myvolume"}
	Expect(ensureDockerLocalPath(entry)).To(Succeed())
	Expect(filepath.Join(cwd, "myvolume")).NotTo(BeADirectory())
}

func TestEnsureDockerLocalPathRefusesModeOnCustomPath(t *testing.T) {
	RegisterTestingT(t)
	withTempLibRoot(t)

	entry := &Entry{
		Name:      "demo",
		Scheduler: SchedulerDockerLocal,
		HostPath:  filepath.Join(t.TempDir(), "custom"),
		Mode:      "0777",
	}

	err := ensureDockerLocalPath(entry)
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("--mode is only supported when the storage entry uses the default host path"))
	Expect(entry.HostPath).NotTo(BeADirectory())
}

func TestEnsureDockerLocalPathRefusesChownOnCustomPath(t *testing.T) {
	RegisterTestingT(t)
	withTempLibRoot(t)

	entry := &Entry{
		Name:      "demo",
		Scheduler: SchedulerDockerLocal,
		HostPath:  filepath.Join(t.TempDir(), "custom"),
		Chown:     "herokuish",
	}

	err := ensureDockerLocalPath(entry)
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("--chown is only supported when the storage entry uses the default host path"))
}

// TestEnsureDockerLocalPathAllowsChownFalseOnCustomPath documents that the
// escape hatch the refusal message points at actually works.
func TestEnsureDockerLocalPathAllowsChownFalseOnCustomPath(t *testing.T) {
	RegisterTestingT(t)
	withTempLibRoot(t)

	hostPath := filepath.Join(t.TempDir(), "custom")
	entry := &Entry{Name: "demo", Scheduler: SchedulerDockerLocal, HostPath: hostPath, Chown: "false"}

	Expect(ensureDockerLocalPath(entry)).To(Succeed())
	Expect(hostPath).To(BeADirectory())
}

func TestApplyPropertyChangeScalars(t *testing.T) {
	RegisterTestingT(t)

	entry := &Entry{Name: "demo", Scheduler: SchedulerDockerLocal}

	Expect(applyPropertyChange(entry, PropertyChange{Property: "chown", Value: "herokuish"})).To(Succeed())
	Expect(entry.Chown).To(Equal("herokuish"))

	Expect(applyPropertyChange(entry, PropertyChange{Property: "namespace", Value: "dokku"})).To(Succeed())
	Expect(entry.Namespace).To(Equal("dokku"))

	Expect(applyPropertyChange(entry, PropertyChange{Property: "reclaim-policy", Value: ReclaimPolicyDelete})).To(Succeed())
	Expect(entry.ReclaimPolicy).To(Equal(ReclaimPolicyDelete))

	Expect(applyPropertyChange(entry, PropertyChange{Property: "size", Value: "2Gi"})).To(Succeed())
	Expect(entry.Size).To(Equal("2Gi"))

	// mode is canonicalized on the way in
	Expect(applyPropertyChange(entry, PropertyChange{Property: "mode", Value: "770"})).To(Succeed())
	Expect(entry.Mode).To(Equal("0770"))
}

// TestApplyPropertyChangeUnsets is the hole this shape exists to close: an
// empty value clears the field rather than being indistinguishable from an
// omitted flag.
func TestApplyPropertyChangeUnsets(t *testing.T) {
	RegisterTestingT(t)

	entry := &Entry{
		Name:          "demo",
		Scheduler:     SchedulerDockerLocal,
		Chown:         "herokuish",
		Mode:          "0770",
		Namespace:     "dokku",
		ReclaimPolicy: ReclaimPolicyDelete,
		Size:          "2Gi",
	}

	for _, property := range []string{"chown", "mode", "namespace", "reclaim-policy", "size"} {
		Expect(applyPropertyChange(entry, PropertyChange{Property: property})).To(Succeed(), "unsetting %q", property)
	}

	Expect(entry.Chown).To(BeEmpty())
	Expect(entry.Mode).To(BeEmpty())
	Expect(entry.Namespace).To(BeEmpty())
	Expect(entry.ReclaimPolicy).To(BeEmpty())
	Expect(entry.Size).To(BeEmpty())
}

func TestApplyPropertyChangeRejectsInvalidInput(t *testing.T) {
	RegisterTestingT(t)

	entry := &Entry{Name: "demo", Scheduler: SchedulerDockerLocal}

	err := applyPropertyChange(entry, PropertyChange{Property: "bogus", Value: "x"})
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("Invalid property specified, valid properties include: access-mode, chown, mode, namespace, reclaim-policy, size, storage-class-name"))

	err = applyPropertyChange(entry, PropertyChange{Property: "", Value: "x"})
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("No property specified"))

	err = applyPropertyChange(entry, PropertyChange{Property: "mode", Value: "0888"})
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("Unsupported directory mode"))
}

// TestApplyPropertyChangeRefusesInPlaceSwaps covers both a differing value
// and an empty one, since clearing a bound PVC's access-mode or storage
// class is equally a change Kubernetes cannot apply.
func TestApplyPropertyChangeRefusesInPlaceSwaps(t *testing.T) {
	RegisterTestingT(t)

	entry := &Entry{
		Name:         "demo",
		Scheduler:    SchedulerK3s,
		Size:         "2Gi",
		AccessMode:   "ReadWriteOnce",
		StorageClass: "longhorn",
	}

	err := applyPropertyChange(entry, PropertyChange{Property: "access-mode", Value: "ReadWriteMany"})
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("cannot change access-mode in place"))

	err = applyPropertyChange(entry, PropertyChange{Property: "access-mode"})
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("cannot change access-mode in place"))

	err = applyPropertyChange(entry, PropertyChange{Property: "storage-class-name", Value: "other"})
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("cannot change storage-class-name in place"))

	err = applyPropertyChange(entry, PropertyChange{Property: "storage-class-name"})
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("cannot change storage-class-name in place"))

	// Re-stating the current value is a no-op, not a change.
	Expect(applyPropertyChange(entry, PropertyChange{Property: "access-mode", Value: "ReadWriteOnce"})).To(Succeed())
	Expect(applyPropertyChange(entry, PropertyChange{Property: "storage-class-name", Value: "longhorn"})).To(Succeed())
}

func TestCommandSetPersistsAndValidates(t *testing.T) {
	RegisterTestingT(t)
	withTempLibRoot(t)
	stageDockerLocalEntry(t, "demo")

	Expect(CommandSet(CommandSetInput{
		Name:    "demo",
		Changes: []PropertyChange{{Property: "namespace", Value: "dokku"}},
	})).To(Succeed())

	loaded, err := LoadEntry("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(loaded.Namespace).To(Equal("dokku"))

	// An invalid change leaves the stored entry untouched.
	err = CommandSet(CommandSetInput{
		Name:    "demo",
		Changes: []PropertyChange{{Property: "bogus", Value: "x"}},
	})
	Expect(err).To(HaveOccurred())

	loaded, err = LoadEntry("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(loaded.Namespace).To(Equal("dokku"))

	err = CommandSet(CommandSetInput{
		Name:    "missing",
		Changes: []PropertyChange{{Property: "namespace", Value: "dokku"}},
	})
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("does not exist"))
}
