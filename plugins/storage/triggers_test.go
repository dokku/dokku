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

func TestTriggerDockerArgsRejectsSchedulerMismatch(t *testing.T) {
	RegisterTestingT(t)
	root := withTempLibRoot(t)

	Expect(SaveEntry(&Entry{
		Name:      "demo-pvc",
		Scheduler: SchedulerK3s,
	})).To(Succeed())

	writeAttachmentsFile(t, root, "demo", []*Attachment{
		{
			EntryName:     "demo-pvc",
			ContainerPath: "/data",
			Phases:        []string{PhaseDeploy},
		},
	})

	err := TriggerDockerArgs("demo", PhaseDeploy, "web")
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(Equal(`storage entry "demo-pvc" is scheduler=k3s but is mounted on a docker-local app; recreate it with --scheduler docker-local`))
}

// TestTriggerDockerArgsRejectsSchedulerMismatchOutOfScope pins that the
// scheduler check runs over every attachment in the phase, not just the ones
// the process type being deployed would actually mount. An entry created for
// the wrong scheduler has to fail the deploy whichever process starts first.
func TestTriggerDockerArgsRejectsSchedulerMismatchOutOfScope(t *testing.T) {
	RegisterTestingT(t)
	root := withTempLibRoot(t)

	Expect(SaveEntry(&Entry{
		Name:      "demo-pvc",
		Scheduler: SchedulerK3s,
	})).To(Succeed())

	writeAttachmentsFile(t, root, "demo", []*Attachment{
		{
			EntryName:     "demo-pvc",
			ContainerPath: "/data",
			Phases:        []string{PhaseDeploy},
			ProcessType:   "worker",
		},
	})

	_, err := dockerVFlagsForProcess("demo", PhaseDeploy, "web")
	Expect(err).To(HaveOccurred())
}

// TestDockerVFlagsForProcessScopes is the docker-local half of #9059. An
// attachment scoped to a named process type reaches that process only; the
// default scope reaches every process, including containers that belong to no
// process at all - the app.json deploy-task path passes an empty process type.
func TestDockerVFlagsForProcessScopes(t *testing.T) {
	RegisterTestingT(t)
	root := withTempLibRoot(t)

	for _, name := range []string{"shared", "web-only", "legacy"} {
		Expect(SaveEntry(&Entry{
			Name:      name,
			Scheduler: SchedulerDockerLocal,
			HostPath:  "/host/" + name,
		})).To(Succeed())
	}

	writeAttachmentsFile(t, root, "demo", []*Attachment{
		{
			EntryName:     "shared",
			ContainerPath: "/shared",
			Phases:        []string{PhaseDeploy},
			ProcessType:   DefaultProcessType,
		},
		{
			EntryName:     "web-only",
			ContainerPath: "/web",
			Phases:        []string{PhaseDeploy},
			ProcessType:   "web",
		},
		{
			// Written before the process-type field existed.
			EntryName:     "legacy",
			ContainerPath: "/legacy",
			Phases:        []string{PhaseDeploy},
		},
	})

	for _, testCase := range []struct {
		processType string
		expected    []string
	}{
		{processType: "web", expected: []string{"-v /host/legacy:/legacy", "-v /host/shared:/shared", "-v /host/web-only:/web"}},
		{processType: "worker", expected: []string{"-v /host/legacy:/legacy", "-v /host/shared:/shared"}},
		{processType: "", expected: []string{"-v /host/legacy:/legacy", "-v /host/shared:/shared"}},
		{processType: DefaultProcessType, expected: []string{"-v /host/legacy:/legacy", "-v /host/shared:/shared"}},
	} {
		flags, err := dockerVFlagsForProcess("demo", PhaseDeploy, testCase.processType)
		Expect(err).NotTo(HaveOccurred())
		Expect(flags).To(Equal(testCase.expected), "process type %q", testCase.processType)
	}
}
