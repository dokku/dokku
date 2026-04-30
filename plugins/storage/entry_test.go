package storage

import (
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

func withTempLibRoot(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("DOKKU_LIB_ROOT", dir)
	return dir
}

func TestValidateEntryName(t *testing.T) {
	RegisterTestingT(t)

	Expect(ValidateEntryName("foo", false)).To(Succeed())
	Expect(ValidateEntryName("foo-bar", false)).To(Succeed())
	Expect(ValidateEntryName("a", false)).To(Succeed())
	Expect(ValidateEntryName("foo123", false)).To(Succeed())

	err := ValidateEntryName("", false)
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("required"))

	err = ValidateEntryName("Foo", false)
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("DNS-1123"))

	err = ValidateEntryName("foo_bar", false)
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("DNS-1123"))

	err = ValidateEntryName("-foo", false)
	Expect(err).To(HaveOccurred())

	err = ValidateEntryName("foo-", false)
	Expect(err).To(HaveOccurred())

	long := strings.Repeat("a", MaxEntryNameLength+1)
	err = ValidateEntryName(long, false)
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("too long"))

	maxOK := strings.Repeat("a", MaxEntryNameLength)
	Expect(ValidateEntryName(maxOK, false)).To(Succeed())

	err = ValidateEntryName("legacy-foo", false)
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("reserved"))

	Expect(ValidateEntryName("legacy-foo", true)).To(Succeed())
}

func TestEntryValidateDockerLocal(t *testing.T) {
	RegisterTestingT(t)

	good := &Entry{Name: "foo", Scheduler: SchedulerDockerLocal, HostPath: "/var/lib/dokku/data/storage/foo"}
	Expect(good.Validate()).To(Succeed())

	noPath := &Entry{Name: "foo", Scheduler: SchedulerDockerLocal}
	Expect(noPath.Validate()).To(HaveOccurred())

	// A leading-slash-less but DNS-1123-ish token is treated as a docker
	// named volume and accepted.
	namedVolume := &Entry{Name: "foo", Scheduler: SchedulerDockerLocal, HostPath: "myvolume"}
	Expect(namedVolume.Validate()).To(Succeed())

	// Slash-containing relative paths still fail, as do tokens with bad chars.
	relativePath := &Entry{Name: "foo", Scheduler: SchedulerDockerLocal, HostPath: "relative/path"}
	Expect(relativePath.Validate()).To(HaveOccurred())

	badToken := &Entry{Name: "foo", Scheduler: SchedulerDockerLocal, HostPath: "with spaces"}
	Expect(badToken.Validate()).To(HaveOccurred())

	withSize := &Entry{Name: "foo", Scheduler: SchedulerDockerLocal, HostPath: "/data", Size: "2Gi"}
	Expect(withSize.Validate()).To(HaveOccurred())

	withClass := &Entry{Name: "foo", Scheduler: SchedulerDockerLocal, HostPath: "/data", StorageClass: "longhorn"}
	Expect(withClass.Validate()).To(HaveOccurred())
}

func TestEntryValidateK3s(t *testing.T) {
	RegisterTestingT(t)

	dynamic := &Entry{Name: "foo", Scheduler: SchedulerK3s, Size: "2Gi", StorageClass: "longhorn", AccessMode: "ReadWriteOnce"}
	Expect(dynamic.Validate()).To(Succeed())

	hostPath := &Entry{Name: "foo", Scheduler: SchedulerK3s, Size: "2Gi", HostPath: "/data/foo"}
	Expect(hostPath.Validate()).To(Succeed())

	noSize := &Entry{Name: "foo", Scheduler: SchedulerK3s, StorageClass: "longhorn"}
	Expect(noSize.Validate()).To(HaveOccurred())

	bothPathAndClass := &Entry{Name: "foo", Scheduler: SchedulerK3s, Size: "2Gi", HostPath: "/data", StorageClass: "longhorn"}
	err := bothPathAndClass.Validate()
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("storage-class-name"))

	badAccessMode := &Entry{Name: "foo", Scheduler: SchedulerK3s, Size: "2Gi", StorageClass: "longhorn", AccessMode: "Bogus"}
	Expect(badAccessMode.Validate()).To(HaveOccurred())

	badReclaim := &Entry{Name: "foo", Scheduler: SchedulerK3s, Size: "2Gi", StorageClass: "longhorn", ReclaimPolicy: "Recycle"}
	Expect(badReclaim.Validate()).To(HaveOccurred())
}

func TestEntryValidateScheduler(t *testing.T) {
	RegisterTestingT(t)

	bad := &Entry{Name: "foo", Scheduler: "nomad"}
	err := bad.Validate()
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("scheduler"))
}

func TestEntryRoundTrip(t *testing.T) {
	RegisterTestingT(t)
	root := withTempLibRoot(t)

	original := &Entry{
		Name:         "demo-data",
		Scheduler:    SchedulerK3s,
		Size:         "2Gi",
		StorageClass: "longhorn",
		AccessMode:   "ReadWriteOnce",
		Namespace:    "dokku",
		Annotations:  map[string]string{"backup.velero.io/backup-volumes": "demo-data"},
		Labels:       map[string]string{"app.kubernetes.io/managed-by": "dokku"},
	}

	Expect(SaveEntry(original)).To(Succeed())
	Expect(EntryExists("demo-data")).To(BeTrue())

	loaded, err := LoadEntry("demo-data")
	Expect(err).NotTo(HaveOccurred())
	Expect(loaded.Name).To(Equal(original.Name))
	Expect(loaded.Scheduler).To(Equal(original.Scheduler))
	Expect(loaded.Size).To(Equal(original.Size))
	Expect(loaded.StorageClass).To(Equal(original.StorageClass))
	Expect(loaded.AccessMode).To(Equal(original.AccessMode))
	Expect(loaded.Namespace).To(Equal(original.Namespace))
	Expect(loaded.Annotations).To(Equal(original.Annotations))
	Expect(loaded.Labels).To(Equal(original.Labels))
	Expect(loaded.SchemaVersion).To(Equal(SchemaVersion))

	expectedPath := filepath.Join(root, "data", "storage-registry", "entries", "demo-data.json")
	Expect(expectedPath).To(BeARegularFile())

	Expect(DeleteEntry("demo-data")).To(Succeed())
	Expect(EntryExists("demo-data")).To(BeFalse())
}

func TestListEntries(t *testing.T) {
	RegisterTestingT(t)
	withTempLibRoot(t)

	entries, err := ListEntries()
	Expect(err).NotTo(HaveOccurred())
	Expect(entries).To(BeEmpty())

	Expect(SaveEntry(&Entry{Name: "b", Scheduler: SchedulerDockerLocal, HostPath: "/b"})).To(Succeed())
	Expect(SaveEntry(&Entry{Name: "a", Scheduler: SchedulerDockerLocal, HostPath: "/a"})).To(Succeed())

	entries, err = ListEntries()
	Expect(err).NotTo(HaveOccurred())
	Expect(entries).To(HaveLen(2))
	Expect(entries[0].Name).To(Equal("a"))
	Expect(entries[1].Name).To(Equal("b"))
}

func TestLegacyMountToEntry(t *testing.T) {
	RegisterTestingT(t)

	a := LegacyMountToEntry("/host/path:/container/path")
	b := LegacyMountToEntry("/host/path:/container/different")
	c := LegacyMountToEntry("/host/path:/container/path:ro")

	// Same host path produces the same entry name regardless of container
	// path or options - container path lives on the attachment, not the
	// entry, so apps mounting the same source converge.
	Expect(a.Name).To(Equal(b.Name))
	Expect(a.Name).To(Equal(c.Name))
	Expect(a.Name).To(HavePrefix(LegacyEntryPrefix))
	Expect(len(a.Name)).To(Equal(len(LegacyEntryPrefix) + 10))

	// Different host paths produce different names.
	other := LegacyMountToEntry("/different/path:/container/path")
	Expect(other.Name).NotTo(Equal(a.Name))

	// Named docker volumes are distinguished from absolute paths.
	named := LegacyMountToEntry("my_volume:/container/path")
	Expect(named.Name).NotTo(Equal(a.Name))
	Expect(named.HostPath).To(Equal("my_volume"))

	// The synthesized entry is a valid docker-local entry that passes
	// Validate (other than the legacy- prefix, which Validate accepts).
	abs := LegacyMountToEntry("/host/path:/container/path")
	Expect(abs.Scheduler).To(Equal(SchedulerDockerLocal))
	Expect(abs.HostPath).To(Equal("/host/path"))
	Expect(abs.Validate()).To(Succeed())
}
