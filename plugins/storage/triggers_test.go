package storage

import (
	"os"
	"os/user"
	"strconv"
	"syscall"
	"testing"

	. "github.com/onsi/gomega"
)

// TestRepairRegistryOwnership covers the install-time repair that exists
// to fix #8557 on systems that already ran the buggy 0.38.0 install. The
// repair must rewrite ownership without rewriting the file mode, and
// must tolerate a missing registry tree (clean install).
func TestRepairRegistryOwnership(t *testing.T) {
	RegisterTestingT(t)
	withTempLibRoot(t)

	Expect(EnsureEntriesDirectory()).To(Succeed())
	path := entryPath("legacy-deadbeef")
	Expect(os.WriteFile(path, []byte(`{"name":"legacy-deadbeef"}`), 0640)).To(Succeed())

	Expect(repairRegistryOwnership()).To(Succeed())

	info, err := os.Stat(path)
	Expect(err).NotTo(HaveOccurred())
	Expect(info.Mode().Perm()).To(Equal(os.FileMode(0640)))

	stat, ok := info.Sys().(*syscall.Stat_t)
	Expect(ok).To(BeTrue())

	current, err := user.Current()
	Expect(err).NotTo(HaveOccurred())
	Expect(strconv.Itoa(int(stat.Uid))).To(Equal(current.Uid))

	Expect(os.RemoveAll(RegistryDirectory())).To(Succeed())
	Expect(repairRegistryOwnership()).To(Succeed())
}

// TestBuildDockerVFlagVolumeOptions covers the rendering rules in
// buildDockerVFlag for the various combinations of Readonly and
// VolumeOptions. This is the docker-args path that the running
// container actually receives.
func TestBuildDockerVFlagVolumeOptions(t *testing.T) {
	RegisterTestingT(t)

	entry := &Entry{HostPath: "/host"}

	plain := buildDockerVFlag(entry, &Attachment{ContainerPath: "/container"})
	Expect(plain).To(Equal("-v /host:/container"))

	opts := buildDockerVFlag(entry, &Attachment{ContainerPath: "/container", VolumeOptions: "Z"})
	Expect(opts).To(Equal("-v /host:/container:Z"))

	ro := buildDockerVFlag(entry, &Attachment{ContainerPath: "/container", Readonly: true})
	Expect(ro).To(Equal("-v /host:/container:ro"))

	roOpts := buildDockerVFlag(entry, &Attachment{ContainerPath: "/container", Readonly: true, VolumeOptions: "noexec,nosuid"})
	Expect(roOpts).To(Equal("-v /host:/container:ro,noexec,nosuid"))
}
