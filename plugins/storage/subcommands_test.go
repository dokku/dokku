package storage

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
)

func TestResolveChownIDRejectsOutOfBoundsValues(t *testing.T) {
	RegisterTestingT(t)

	for _, chownFlag := range []string{"-1", "65536", "100000", "231072", "abc", "1.5", ""} {
		_, err := ResolveChownID(chownFlag)
		Expect(err).To(HaveOccurred(), "expected %q to be rejected", chownFlag)
		Expect(err.Error()).To(ContainSubstring("Unsupported chown permissions"))
	}
}

func setupTestApp(t *testing.T, appName string) {
	t.Helper()
	libRoot := withTempLibRoot(t)
	dokkuRoot := t.TempDir()
	t.Setenv("DOKKU_ROOT", dokkuRoot)
	t.Setenv("PLUGIN_PATH", filepath.Join(libRoot, "plugins"))
	if err := os.MkdirAll(filepath.Join(dokkuRoot, appName), 0755); err != nil {
		t.Fatalf("mkdir app root: %v", err)
	}
}

func TestCommandMountRejectsSchedulerMismatch(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")

	Expect(SaveEntry(&Entry{
		Name:      "demo-pvc",
		Scheduler: SchedulerK3s,
	})).To(Succeed())

	err := CommandMount(CommandMountInput{
		AppName:      "demo",
		NameOrPath:   "demo-pvc",
		ContainerDir: "/data",
	})
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(Equal(`storage entry "demo-pvc" is scheduler=k3s but cannot be mounted on a docker-local app; recreate it with --scheduler docker-local`))

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(BeEmpty())
}

func TestCommandMountAcceptsMatchingScheduler(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")

	Expect(SaveEntry(&Entry{
		Name:      "demo-data",
		Scheduler: SchedulerDockerLocal,
		HostPath:  "/data",
	})).To(Succeed())

	err := CommandMount(CommandMountInput{
		AppName:      "demo",
		NameOrPath:   "demo-data",
		ContainerDir: "/data",
	})
	Expect(err).NotTo(HaveOccurred())

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(HaveLen(1))
	Expect(attachments[0].EntryName).To(Equal("demo-data"))
	Expect(attachments[0].ContainerPath).To(Equal("/data"))
}
