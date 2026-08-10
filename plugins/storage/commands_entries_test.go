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
