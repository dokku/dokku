package storage

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestParseMountSpec(t *testing.T) {
	RegisterTestingT(t)

	parsed, err := parseMountSpec("demo-data:/app/storage")
	Expect(err).NotTo(HaveOccurred())
	Expect(parsed.EntryName).To(Equal("demo-data"))
	Expect(parsed.ContainerPath).To(Equal("/app/storage"))
	Expect(parsed.Readonly).To(BeFalse())
	Expect(parsed.VolumeOptions).To(BeEmpty())

	parsed, err = parseMountSpec("demo-data:/app/storage:ro,noexec,nosuid")
	Expect(err).NotTo(HaveOccurred())
	Expect(parsed.Readonly).To(BeTrue())
	Expect(parsed.VolumeOptions).To(Equal("noexec,nosuid"))

	for _, spec := range []string{"demo-data", ":/app/storage", ""} {
		_, err := parseMountSpec(spec)
		Expect(err).To(HaveOccurred(), "expected %q to be rejected", spec)
		Expect(err.Error()).To(ContainSubstring("Invalid mount specified"))
	}
}

func TestValidateChownOption(t *testing.T) {
	RegisterTestingT(t)

	for _, chownFlag := range []string{"", "herokuish", "heroku", "paketo", "packeto", "root", "false", "0", "1000", "65535"} {
		Expect(ValidateChownOption(chownFlag)).To(Succeed(), "expected %q to be accepted", chownFlag)
	}

	for _, chownFlag := range []string{"bogus", "-1", "65536", "1.5", "herokuish "} {
		err := ValidateChownOption(chownFlag)
		Expect(err).To(HaveOccurred(), "expected %q to be rejected", chownFlag)
		Expect(err.Error()).To(ContainSubstring("Unsupported chown permissions"))
	}
}

func TestCommandMountRejectsInvalidVolumeChown(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	stageDockerLocalEntry(t, "demo-data")

	err := CommandMount(CommandMountInput{
		AppName:      "demo",
		NameOrPath:   "demo-data",
		ContainerDir: "/app/storage",
		VolumeChown:  "bogus",
	})
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("Unsupported chown permissions"))

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(BeEmpty())
}

func TestCommandMountReplaceReplacesTheScope(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	stageDockerLocalEntry(t, "demo-data")
	stageDockerLocalEntry(t, "demo-cache")
	stageDockerLocalEntry(t, "demo-logs")

	Expect(CommandMount(CommandMountInput{
		AppName:      "demo",
		NameOrPath:   "demo-data",
		ContainerDir: "/app/storage",
	})).To(Succeed())

	Expect(CommandMount(CommandMountInput{
		AppName:     "demo",
		Replace:     true,
		Specs:       []string{"demo-cache:/cache", "demo-logs:/logs"},
		ProcessType: DefaultProcessType,
	})).To(Succeed())

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(HaveLen(2))
	Expect(attachments[0].EntryName).To(Equal("demo-cache"))
	Expect(attachments[0].ContainerPath).To(Equal("/cache"))
	Expect(attachments[0].Phases).To(Equal([]string{PhaseDeploy, PhaseRun}))
	Expect(attachments[0].ProcessType).To(Equal(DefaultProcessType))
	Expect(attachments[1].EntryName).To(Equal("demo-logs"))
}

func TestCommandMountReplaceLeavesOtherProcessTypesUntouched(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	stageDockerLocalEntry(t, "demo-data")
	stageDockerLocalEntry(t, "demo-cache")
	stageDockerLocalEntry(t, "demo-logs")

	Expect(CommandMount(CommandMountInput{
		AppName:      "demo",
		NameOrPath:   "demo-data",
		ContainerDir: "/app/storage",
	})).To(Succeed())

	Expect(CommandMount(CommandMountInput{
		AppName:      "demo",
		NameOrPath:   "demo-cache",
		ContainerDir: "/cache",
		ProcessType:  "web",
	})).To(Succeed())

	Expect(CommandMount(CommandMountInput{
		AppName: "demo",
		Replace: true,
		Specs:   []string{"demo-logs:/logs"},
	})).To(Succeed())

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(HaveLen(2))
	Expect(attachments[0].EntryName).To(Equal("demo-cache"))
	Expect(attachments[0].ProcessType).To(Equal("web"))
	Expect(attachments[1].EntryName).To(Equal("demo-logs"))
	Expect(attachments[1].ProcessType).To(Equal(DefaultProcessType))
}

func TestCommandMountReplaceAppliesScopeFlags(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	stageDockerLocalEntry(t, "demo-data")
	stageDockerLocalEntry(t, "demo-cache")

	Expect(CommandMount(CommandMountInput{
		AppName:     "demo",
		Replace:     true,
		Specs:       []string{"demo-data:/app/storage:Z", "demo-cache:/cache:ro,noexec"},
		Phases:      []string{PhaseDeploy},
		ProcessType: "web",
		Subpath:     "uploads",
		VolumeChown: "herokuish",
	})).To(Succeed())

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(HaveLen(2))
	for _, attachment := range attachments {
		Expect(attachment.Phases).To(Equal([]string{PhaseDeploy}))
		Expect(attachment.ProcessType).To(Equal("web"))
		Expect(attachment.Subpath).To(Equal("uploads"))
		Expect(attachment.VolumeChown).To(Equal("herokuish"))
	}

	Expect(attachments[0].Readonly).To(BeFalse())
	Expect(attachments[0].VolumeOptions).To(Equal("Z"))
	Expect(attachments[1].Readonly).To(BeTrue())
	Expect(attachments[1].VolumeOptions).To(Equal("noexec"))
}

func TestCommandMountReplaceReadonlyFlagAppliesToEverySpec(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	stageDockerLocalEntry(t, "demo-data")
	stageDockerLocalEntry(t, "demo-cache")

	Expect(CommandMount(CommandMountInput{
		AppName:  "demo",
		Replace:  true,
		Specs:    []string{"demo-data:/app/storage", "demo-cache:/cache"},
		Readonly: true,
	})).To(Succeed())

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(HaveLen(2))
	Expect(attachments[0].Readonly).To(BeTrue())
	Expect(attachments[1].Readonly).To(BeTrue())
}

func TestCommandMountReplaceIsIdempotent(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	stageDockerLocalEntry(t, "demo-data")
	stageDockerLocalEntry(t, "demo-cache")

	input := CommandMountInput{
		AppName: "demo",
		Replace: true,
		Specs:   []string{"demo-data:/app/storage", "demo-cache:/cache"},
	}

	Expect(CommandMount(input)).To(Succeed())
	first, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())

	Expect(CommandMount(input)).To(Succeed())
	second, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())

	Expect(second).To(Equal(first))
}

func TestCommandMountReplaceRejectsEmptySpecs(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	stageDockerLocalEntry(t, "demo-data")

	Expect(CommandMount(CommandMountInput{
		AppName:      "demo",
		NameOrPath:   "demo-data",
		ContainerDir: "/app/storage",
	})).To(Succeed())

	err := CommandMount(CommandMountInput{AppName: "demo", Replace: true})
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(Equal("Must specify at least one mount, use storage:unmount --all to remove all mounts"))

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(HaveLen(1))
}

func TestCommandMountReplaceRejectsContainerDirFlag(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	stageDockerLocalEntry(t, "demo-data")

	err := CommandMount(CommandMountInput{
		AppName:      "demo",
		Replace:      true,
		Specs:        []string{"demo-data:/app/storage"},
		ContainerDir: "/app/storage",
	})
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(Equal("The --container-dir flag cannot be used with --replace"))
}

func TestCommandMountReplaceRejectsDuplicateContainerPath(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	stageDockerLocalEntry(t, "demo-data")
	stageDockerLocalEntry(t, "demo-cache")

	err := CommandMount(CommandMountInput{
		AppName: "demo",
		Replace: true,
		Specs:   []string{"demo-data:/app/storage", "demo-cache:/app/storage"},
	})
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(Equal("Container path /app/storage is specified more than once"))

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(BeEmpty())
}

func TestCommandMountReplaceLeavesStoreUntouchedOnInvalidSpec(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	stageDockerLocalEntry(t, "demo-data")
	stageDockerLocalEntry(t, "demo-cache")

	Expect(CommandMount(CommandMountInput{
		AppName:      "demo",
		NameOrPath:   "demo-data",
		ContainerDir: "/app/storage",
	})).To(Succeed())

	err := CommandMount(CommandMountInput{
		AppName: "demo",
		Replace: true,
		Specs:   []string{"demo-cache:/cache", "demo-missing:/missing"},
	})
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring(`storage entry "demo-missing" does not exist`))

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(HaveLen(1))
	Expect(attachments[0].EntryName).To(Equal("demo-data"))
}

func TestCommandMountReplaceRejectsRelativeContainerPath(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	stageDockerLocalEntry(t, "demo-data")

	err := CommandMount(CommandMountInput{
		AppName: "demo",
		Replace: true,
		Specs:   []string{"demo-data:app/storage"},
	})
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("must be absolute"))

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(BeEmpty())
}

func TestCommandMountReplaceRejectsSchedulerMismatch(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")

	Expect(SaveEntry(&Entry{
		Name:      "demo-pvc",
		Scheduler: SchedulerK3s,
	})).To(Succeed())

	err := CommandMount(CommandMountInput{
		AppName: "demo",
		Replace: true,
		Specs:   []string{"demo-pvc:/data"},
	})
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(Equal(`storage entry "demo-pvc" is scheduler=k3s but cannot be mounted on a docker-local app; recreate it with --scheduler docker-local`))
}

func TestCommandMountReplaceRejectsUnregisteredEntry(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")

	err := CommandMount(CommandMountInput{
		AppName: "demo",
		Replace: true,
		Specs:   []string{"demo-data:/app/storage"},
	})
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring(`storage entry "demo-data" does not exist`))
}

func TestCommandMountReplaceAcceptsLegacyHostPath(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")

	Expect(CommandMount(CommandMountInput{
		AppName: "demo",
		Replace: true,
		Specs:   []string{"/var/lib/dokku/data/storage/shared:/app/shared"},
	})).To(Succeed())

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(HaveLen(1))
	Expect(attachments[0].ContainerPath).To(Equal("/app/shared"))

	expected := LegacyMountToEntry("/var/lib/dokku/data/storage/shared:/app/shared")
	Expect(attachments[0].EntryName).To(Equal(expected.Name))
	Expect(EntryExists(expected.Name)).To(BeTrue())
}

func TestCommandUnmountRemovesMultipleMounts(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	stageDockerLocalEntry(t, "demo-data")
	stageDockerLocalEntry(t, "demo-cache")
	stageDockerLocalEntry(t, "demo-logs")

	Expect(CommandMount(CommandMountInput{
		AppName: "demo",
		Replace: true,
		Specs:   []string{"demo-data:/app/storage", "demo-cache:/cache", "demo-logs:/logs"},
	})).To(Succeed())

	Expect(CommandUnmount(CommandUnmountInput{
		AppName: "demo",
		Mounts:  []string{"demo-data", "demo-logs"},
	})).To(Succeed())

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(HaveLen(1))
	Expect(attachments[0].EntryName).To(Equal("demo-cache"))
}

func TestCommandUnmountAcceptsEntryColonContainerPath(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	stageDockerLocalEntry(t, "demo-data")

	Expect(CommandMount(CommandMountInput{
		AppName: "demo",
		Replace: true,
		Specs:   []string{"demo-data:/app/storage", "demo-data:/app/uploads"},
	})).To(Succeed())

	Expect(CommandUnmount(CommandUnmountInput{
		AppName: "demo",
		Mounts:  []string{"demo-data:/app/uploads"},
	})).To(Succeed())

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(HaveLen(1))
	Expect(attachments[0].ContainerPath).To(Equal("/app/storage"))
}

func TestCommandUnmountLeavesStoreUntouchedOnUnknownMount(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	stageDockerLocalEntry(t, "demo-data")
	stageDockerLocalEntry(t, "demo-cache")

	Expect(CommandMount(CommandMountInput{
		AppName: "demo",
		Replace: true,
		Specs:   []string{"demo-data:/app/storage", "demo-cache:/cache"},
	})).To(Succeed())

	err := CommandUnmount(CommandUnmountInput{
		AppName: "demo",
		Mounts:  []string{"demo-data", "demo-missing"},
	})
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring(`storage entry "demo-missing" is not mounted`))

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(HaveLen(2))
}

func TestCommandUnmountRejectsContainerDirWithMultipleMounts(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")

	err := CommandUnmount(CommandUnmountInput{
		AppName:      "demo",
		Mounts:       []string{"demo-data", "demo-cache"},
		ContainerDir: "/app/storage",
	})
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(Equal("The --container-dir flag cannot be used with multiple mounts"))
}

func TestCommandUnmountRequiresAMount(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")

	err := CommandUnmount(CommandUnmountInput{AppName: "demo"})
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(Equal("Must specify at least one mount, use storage:unmount --all to remove all mounts"))
}

func TestCommandUnmountRejectsProcessTypeWithoutAll(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")

	err := CommandUnmount(CommandUnmountInput{
		AppName:     "demo",
		Mounts:      []string{"demo-data"},
		ProcessType: "web",
	})
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(Equal("The --process-type flag can only be used with --all"))
}

func TestCommandUnmountAllRemovesEverything(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	stageDockerLocalEntry(t, "demo-data")
	stageDockerLocalEntry(t, "demo-cache")

	Expect(CommandMount(CommandMountInput{
		AppName: "demo",
		Replace: true,
		Specs:   []string{"demo-data:/app/storage", "demo-cache:/cache"},
	})).To(Succeed())

	Expect(CommandUnmount(CommandUnmountInput{AppName: "demo", All: true})).To(Succeed())

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(BeEmpty())

	// Idempotent: clearing an app with no mounts is not an error.
	Expect(CommandUnmount(CommandUnmountInput{AppName: "demo", All: true})).To(Succeed())
}

func TestCommandUnmountAllScopedToProcessType(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	stageDockerLocalEntry(t, "demo-data")
	stageDockerLocalEntry(t, "demo-cache")

	Expect(CommandMount(CommandMountInput{
		AppName: "demo",
		Replace: true,
		Specs:   []string{"demo-data:/app/storage"},
	})).To(Succeed())

	Expect(CommandMount(CommandMountInput{
		AppName:     "demo",
		Replace:     true,
		Specs:       []string{"demo-cache:/cache"},
		ProcessType: "web",
	})).To(Succeed())

	Expect(CommandUnmount(CommandUnmountInput{
		AppName:     "demo",
		All:         true,
		ProcessType: "web",
	})).To(Succeed())

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(HaveLen(1))
	Expect(attachments[0].EntryName).To(Equal("demo-data"))

	// A scope that matches nothing is a no-op rather than an error.
	Expect(CommandUnmount(CommandUnmountInput{
		AppName:     "demo",
		All:         true,
		ProcessType: "worker",
	})).To(Succeed())

	attachments, err = LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(HaveLen(1))
}

func TestCommandUnmountAllRejectsNamedMount(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")

	err := CommandUnmount(CommandUnmountInput{
		AppName: "demo",
		All:     true,
		Mounts:  []string{"demo-data"},
	})
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(Equal("A mount cannot be specified with --all"))
}
