package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
)

// writeAttachmentsFile drops the JSON-encoded attachment list into the
// per-app property file directly, bypassing common.PropertyListWrite (which
// calls SetPermissions and chowns to dokku:dokku, which a developer machine
// or CI runner generally doesn't have).
func writeAttachmentsFile(t *testing.T, root string, appName string, attachments []*Attachment) {
	t.Helper()
	dir := filepath.Join(root, "config", "storage", appName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir attachment dir: %v", err)
	}
	path := filepath.Join(dir, AttachmentsProperty)
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create attachment file: %v", err)
	}
	defer f.Close()
	for _, a := range attachments {
		data, err := json.Marshal(a)
		if err != nil {
			t.Fatalf("marshal attachment: %v", err)
		}
		f.Write(append(data, '\n'))
	}
}

func TestAttachmentValidate(t *testing.T) {
	RegisterTestingT(t)

	good := &Attachment{
		EntryName:     "demo-data",
		ContainerPath: "/data",
		Phases:        []string{PhaseDeploy, PhaseRun},
	}
	Expect(good.Validate()).To(Succeed())

	noEntry := &Attachment{ContainerPath: "/data", Phases: []string{PhaseDeploy}}
	Expect(noEntry.Validate()).To(HaveOccurred())

	noPath := &Attachment{EntryName: "demo-data", Phases: []string{PhaseDeploy}}
	Expect(noPath.Validate()).To(HaveOccurred())

	relative := &Attachment{EntryName: "demo-data", ContainerPath: "data", Phases: []string{PhaseDeploy}}
	Expect(relative.Validate()).To(HaveOccurred())

	noPhases := &Attachment{EntryName: "demo-data", ContainerPath: "/data"}
	Expect(noPhases.Validate()).To(HaveOccurred())

	badPhase := &Attachment{EntryName: "demo-data", ContainerPath: "/data", Phases: []string{"build"}}
	Expect(badPhase.Validate()).To(HaveOccurred())
}

func TestListAppMountEntriesDockerLocal(t *testing.T) {
	RegisterTestingT(t)
	root := withTempLibRoot(t)

	Expect(SaveEntry(&Entry{
		Name:      "demo-data",
		Scheduler: SchedulerDockerLocal,
		HostPath:  "/var/lib/dokku/data/storage/demo-data",
	})).To(Succeed())

	writeAttachmentsFile(t, root, "demo", []*Attachment{
		{
			EntryName:     "demo-data",
			ContainerPath: "/data",
			Phases:        []string{PhaseDeploy, PhaseRun},
			Readonly:      true,
		},
	})

	rows, err := ListAppMountEntries("demo", PhaseDeploy)
	Expect(err).NotTo(HaveOccurred())
	Expect(rows).To(HaveLen(1))
	Expect(rows[0].EntryName).To(Equal("demo-data"))
	Expect(rows[0].HostPath).To(Equal("/var/lib/dokku/data/storage/demo-data"))
	Expect(rows[0].ContainerPath).To(Equal("/data"))
	Expect(rows[0].VolumeOptions).To(Equal("ro"))

	// Run phase shows it too because the attachment includes both phases.
	runRows, err := ListAppMountEntries("demo", PhaseRun)
	Expect(err).NotTo(HaveOccurred())
	Expect(runRows).To(HaveLen(1))

	// formatStorageListEntry produces the legacy colon form including options.
	Expect(formatStorageListEntry(rows[0])).To(Equal("/var/lib/dokku/data/storage/demo-data:/data:ro"))
}

func TestListAppMountEntriesK3sUsesEntryName(t *testing.T) {
	RegisterTestingT(t)
	root := withTempLibRoot(t)

	// k3s entry has no host path because the cluster provisions it.
	Expect(SaveEntry(&Entry{
		Name:         "demo-pvc",
		Scheduler:    SchedulerK3s,
		Size:         "2Gi",
		StorageClass: "longhorn",
		AccessMode:   "ReadWriteOnce",
	})).To(Succeed())

	writeAttachmentsFile(t, root, "demo", []*Attachment{
		{
			EntryName:     "demo-pvc",
			ContainerPath: "/data",
			Phases:        []string{PhaseDeploy},
			Subpath:       "uploads",
		},
	})

	rows, err := ListAppMountEntries("demo", PhaseDeploy)
	Expect(err).NotTo(HaveOccurred())
	Expect(rows).To(HaveLen(1))
	// HostPath falls back to the entry name so the colon form remains
	// well-formed for callers that only know about the legacy shape.
	Expect(rows[0].HostPath).To(Equal("demo-pvc"))
	Expect(rows[0].EntryName).To(Equal("demo-pvc"))
	Expect(rows[0].ContainerPath).To(Equal("/data"))
	Expect(formatStorageListEntry(rows[0])).To(Equal("demo-pvc:/data"))
}

func TestListAppMountEntriesPhaseFilter(t *testing.T) {
	RegisterTestingT(t)
	root := withTempLibRoot(t)

	Expect(SaveEntry(&Entry{
		Name:      "demo-deploy-only",
		Scheduler: SchedulerDockerLocal,
		HostPath:  "/srv/deploy",
	})).To(Succeed())
	Expect(SaveEntry(&Entry{
		Name:      "demo-run-only",
		Scheduler: SchedulerDockerLocal,
		HostPath:  "/srv/run",
	})).To(Succeed())

	writeAttachmentsFile(t, root, "demo", []*Attachment{
		{EntryName: "demo-deploy-only", ContainerPath: "/d", Phases: []string{PhaseDeploy}},
		{EntryName: "demo-run-only", ContainerPath: "/r", Phases: []string{PhaseRun}},
	})

	deployRows, err := ListAppMountEntries("demo", PhaseDeploy)
	Expect(err).NotTo(HaveOccurred())
	Expect(deployRows).To(HaveLen(1))
	Expect(deployRows[0].EntryName).To(Equal("demo-deploy-only"))

	runRows, err := ListAppMountEntries("demo", PhaseRun)
	Expect(err).NotTo(HaveOccurred())
	Expect(runRows).To(HaveLen(1))
	Expect(runRows[0].EntryName).To(Equal("demo-run-only"))
}
