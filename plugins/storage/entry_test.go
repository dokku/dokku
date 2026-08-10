package storage

import (
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"

	. "github.com/onsi/gomega"
)

// withTempLibRoot points DOKKU_LIB_ROOT at a temp dir and overrides the
// permission helpers' target user/group to the current process user so
// SetPermissions is a no-op chown rather than a hard failure on dev and
// CI machines that lack a dokku system user.
func withTempLibRoot(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("DOKKU_LIB_ROOT", dir)

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

	withMode := &Entry{Name: "foo", Scheduler: SchedulerDockerLocal, HostPath: "/var/lib/dokku/data/storage/foo", Mode: "0777"}
	Expect(withMode.Validate()).To(Succeed())

	withBadMode := &Entry{Name: "foo", Scheduler: SchedulerDockerLocal, HostPath: "/var/lib/dokku/data/storage/foo", Mode: "0999"}
	Expect(withBadMode.Validate()).To(HaveOccurred())
}

// TestEntryValidateDockerLocalReclaimPolicy covers the reclaim policy now
// that it governs whether storage:destroy removes the host directory on
// docker-local, not just the k3s PV.
func TestEntryValidateDockerLocalReclaimPolicy(t *testing.T) {
	RegisterTestingT(t)

	defaultPath := "/var/lib/dokku/data/storage/foo"

	retain := &Entry{Name: "foo", Scheduler: SchedulerDockerLocal, HostPath: defaultPath, ReclaimPolicy: ReclaimPolicyRetain}
	Expect(retain.Validate()).To(Succeed())

	del := &Entry{Name: "foo", Scheduler: SchedulerDockerLocal, HostPath: defaultPath, ReclaimPolicy: ReclaimPolicyDelete}
	Expect(del.Validate()).To(Succeed())

	bad := &Entry{Name: "foo", Scheduler: SchedulerDockerLocal, HostPath: defaultPath, ReclaimPolicy: "Recycle"}
	err := bad.Validate()
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("reclaim policy"))

	// Delete can only be honored where the sudo helper is allowed to
	// operate, so a custom host path is refused up front rather than
	// silently ignored at destroy time.
	custom := &Entry{Name: "foo", Scheduler: SchedulerDockerLocal, HostPath: "/mnt/custom", ReclaimPolicy: ReclaimPolicyDelete}
	err = custom.Validate()
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("default host path"))

	// Retain on a custom path stays legal; nothing gets removed either way.
	customRetain := &Entry{Name: "foo", Scheduler: SchedulerDockerLocal, HostPath: "/mnt/custom", ReclaimPolicy: ReclaimPolicyRetain}
	Expect(customRetain.Validate()).To(Succeed())
}

func TestNormalizeDirectoryMode(t *testing.T) {
	RegisterTestingT(t)

	accepted := map[string]string{
		"":     "",
		"755":  "0755",
		"777":  "0777",
		"0755": "0755",
		"0777": "0777",
		"2775": "2775",
		"1777": "1777",
		"0000": "0000",
	}
	for input, expected := range accepted {
		normalized, err := NormalizeDirectoryMode(input)
		Expect(err).NotTo(HaveOccurred(), "expected %q to be accepted", input)
		Expect(normalized).To(Equal(expected), "unexpected normalization of %q", input)
	}

	for _, input := range []string{"8", "88", "888", "0888", "07555", "0x1ff", "u+rwx", "-1", " 755 ", "rwx"} {
		_, err := NormalizeDirectoryMode(input)
		Expect(err).To(HaveOccurred(), "expected %q to be rejected", input)
		Expect(err.Error()).To(ContainSubstring("Unsupported directory mode"))
	}
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

	withMode := &Entry{Name: "foo", Scheduler: SchedulerK3s, Size: "2Gi", StorageClass: "longhorn", Mode: "0777"}
	err = withMode.Validate()
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("--mode"))
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

	// mode is docker-local only, so it round-trips through its own entry.
	local := &Entry{
		Name:          "demo-local",
		Scheduler:     SchedulerDockerLocal,
		HostPath:      filepath.Join(GetStorageDirectory(), "demo-local"),
		Chown:         "herokuish",
		Mode:          "0777",
		ReclaimPolicy: ReclaimPolicyDelete,
	}
	Expect(SaveEntry(local)).To(Succeed())

	loadedLocal, err := LoadEntry("demo-local")
	Expect(err).NotTo(HaveOccurred())
	Expect(loadedLocal.Mode).To(Equal("0777"))
	Expect(loadedLocal.Chown).To(Equal("herokuish"))
	Expect(loadedLocal.ReclaimPolicy).To(Equal(ReclaimPolicyDelete))

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

// TestSaveEntryChownsToSystemUser is a regression test for #8557: the
// install trigger runs as root, so SaveEntry must chown the file to the
// dokku user before ps:rebuild can read it.
func TestSaveEntryChownsToSystemUser(t *testing.T) {
	RegisterTestingT(t)
	withTempLibRoot(t)

	Expect(SaveEntry(&Entry{
		Name:      "demo-data",
		Scheduler: SchedulerDockerLocal,
		HostPath:  "/data/demo",
	})).To(Succeed())

	info, err := os.Stat(entryPath("demo-data"))
	Expect(err).NotTo(HaveOccurred())
	stat, ok := info.Sys().(*syscall.Stat_t)
	Expect(ok).To(BeTrue())

	current, err := user.Current()
	Expect(err).NotTo(HaveOccurred())
	Expect(strconv.Itoa(int(stat.Uid))).To(Equal(current.Uid))
}
