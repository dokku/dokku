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

	// A trailing colon is the no-options case, not an empty option.
	parsed, err = parseMountSpec("demo-data:/app/storage:")
	Expect(err).NotTo(HaveOccurred())
	Expect(parsed.VolumeOptions).To(BeEmpty())
	Expect(parsed.Phases).To(Equal([]string{PhaseDeploy, PhaseRun}))

	for _, spec := range []string{"demo-data", ":/app/storage", ""} {
		_, err := parseMountSpec(spec)
		Expect(err).To(HaveOccurred(), "expected %q to be rejected", spec)
		Expect(err.Error()).To(ContainSubstring("Invalid mount specified"))
	}
}

// TestParseMountSpecTokens covers the key=value tokens the option list carries
// on top of the bare docker options it has always passed through. Order within
// the list is not significant, so the cases deliberately mix the two.
func TestParseMountSpecTokens(t *testing.T) {
	RegisterTestingT(t)

	for _, testCase := range []struct {
		spec     string
		expected mountSpec
	}{
		{
			spec:     "demo-data:/app/storage",
			expected: mountSpec{Phases: []string{PhaseDeploy, PhaseRun}},
		},
		{
			spec:     "demo-data:/app/storage:rw",
			expected: mountSpec{Phases: []string{PhaseDeploy, PhaseRun}},
		},
		{
			spec:     "demo-data:/app/storage:volume-subpath=uploads",
			expected: mountSpec{Phases: []string{PhaseDeploy, PhaseRun}, Subpath: "uploads"},
		},
		{
			spec:     "demo-data:/app/storage:volume-chown=herokuish",
			expected: mountSpec{Phases: []string{PhaseDeploy, PhaseRun}, VolumeChown: "herokuish"},
		},
		{
			spec:     "demo-data:/app/storage:phase=deploy",
			expected: mountSpec{Phases: []string{PhaseDeploy}},
		},
		{
			spec:     "demo-data:/app/storage:phase=deploy,phase=run",
			expected: mountSpec{Phases: []string{PhaseDeploy, PhaseRun}},
		},
		{
			spec: "demo-data:/app/storage:ro,Z,volume-subpath=uploads,volume-chown=herokuish",
			expected: mountSpec{
				Phases:        []string{PhaseDeploy, PhaseRun},
				Subpath:       "uploads",
				Readonly:      true,
				VolumeOptions: "Z",
				VolumeChown:   "herokuish",
			},
		},
		{
			// The same fields written in a different order, with the
			// remaining bare options keeping the order they were given in.
			spec: "demo-data:/app/storage:volume-chown=herokuish,noexec,phase=run,nosuid,ro,volume-subpath=uploads",
			expected: mountSpec{
				Phases:        []string{PhaseRun},
				Subpath:       "uploads",
				Readonly:      true,
				VolumeOptions: "noexec,nosuid",
				VolumeChown:   "herokuish",
			},
		},
	} {
		parsed, err := parseMountSpec(testCase.spec)
		Expect(err).NotTo(HaveOccurred(), "expected %q to be accepted", testCase.spec)

		expected := testCase.expected
		expected.EntryName = "demo-data"
		expected.ContainerPath = "/app/storage"
		Expect(parsed).To(Equal(expected), "parsing %q", testCase.spec)
	}
}

// TestParseMountSpecNormalizesPhaseOrder pins phases to a canonical order, so
// two specs naming the same phases store identical JSON and render identically
// in storage:report rather than looking like a change to a declarative caller.
func TestParseMountSpecNormalizesPhaseOrder(t *testing.T) {
	RegisterTestingT(t)

	for _, spec := range []string{
		"demo-data:/app/storage:phase=run,phase=deploy",
		"demo-data:/app/storage:phase=deploy,phase=run",
	} {
		parsed, err := parseMountSpec(spec)
		Expect(err).NotTo(HaveOccurred())
		Expect(parsed.Phases).To(Equal([]string{PhaseDeploy, PhaseRun}), "parsing %q", spec)
	}
}

// TestParseMountSpecAllowsSeparatorsInValues documents that only a comma ends
// a token value: the spec splits on the first two colons and a token on its
// first equals sign, so both characters survive inside a value.
func TestParseMountSpecAllowsSeparatorsInValues(t *testing.T) {
	RegisterTestingT(t)

	parsed, err := parseMountSpec("demo-data:/app/storage:volume-subpath=a:b")
	Expect(err).NotTo(HaveOccurred())
	Expect(parsed.Subpath).To(Equal("a:b"))

	parsed, err = parseMountSpec("demo-data:/app/storage:volume-subpath=a=b")
	Expect(err).NotTo(HaveOccurred())
	Expect(parsed.Subpath).To(Equal("a=b"))
}

func TestParseMountSpecRejectsInvalidTokens(t *testing.T) {
	RegisterTestingT(t)

	for _, testCase := range []struct {
		spec     string
		contains string
	}{
		{"demo-data:/app/storage:volume-subpth=uploads", `specifies unknown key "volume-subpth"`},
		{"demo-data:/app/storage:volume-subpath=", "has an empty value for volume-subpath"},
		{"demo-data:/app/storage:volume-chown=", "has an empty value for volume-chown"},
		{"demo-data:/app/storage:phase=", "has an empty value for phase"},
		{"demo-data:/app/storage:ro,,Z", "has an empty mount option"},
		{"demo-data:/app/storage:volume-subpath=a,volume-subpath=b", "specifies volume-subpath more than once"},
		{"demo-data:/app/storage:volume-chown=root,volume-chown=false", "specifies volume-chown more than once"},
		{"demo-data:/app/storage:phase=run,phase=run", "specifies phase run more than once"},
		{"demo-data:/app/storage:ro,rw", "sets both ro and rw"},
		{"demo-data:/app/storage:phase=build", `specifies unsupported phase "build"`},
		{"demo-data:/app/storage:volume-chown=bogus", "Unsupported chown permissions"},
	} {
		_, err := parseMountSpec(testCase.spec)
		Expect(err).To(HaveOccurred(), "expected %q to be rejected", testCase.spec)
		Expect(err.Error()).To(ContainSubstring(testCase.contains), "rejecting %q", testCase.spec)
		// Every rejection names the spec it came from, since a --replace
		// call can carry any number of them.
		Expect(err.Error()).To(ContainSubstring(testCase.spec), "rejecting %q", testCase.spec)
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

// TestCommandMountReplaceVariesFieldsPerSpec is the point of the token
// grammar: every mount-time field except the process type is declared by the
// spec that carries it, so one call can give two mounts different values.
func TestCommandMountReplaceVariesFieldsPerSpec(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	stageDockerLocalEntry(t, "demo-data")
	stageDockerLocalEntry(t, "demo-cache")

	Expect(CommandMount(CommandMountInput{
		AppName: "demo",
		Replace: true,
		Specs: []string{
			"demo-data:/app/storage:Z,volume-subpath=uploads,volume-chown=herokuish",
			"demo-cache:/cache:ro,noexec,phase=deploy",
		},
		ProcessType: "web",
	})).To(Succeed())

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(HaveLen(2))

	// --process-type is the one field that still scopes the whole call.
	Expect(attachments[0].ProcessType).To(Equal("web"))
	Expect(attachments[1].ProcessType).To(Equal("web"))

	Expect(attachments[0].Phases).To(Equal([]string{PhaseDeploy, PhaseRun}))
	Expect(attachments[0].Subpath).To(Equal("uploads"))
	Expect(attachments[0].VolumeChown).To(Equal("herokuish"))
	Expect(attachments[0].VolumeOptions).To(Equal("Z"))
	Expect(attachments[0].Readonly).To(BeFalse())

	Expect(attachments[1].Phases).To(Equal([]string{PhaseDeploy}))
	Expect(attachments[1].Subpath).To(BeEmpty())
	Expect(attachments[1].VolumeChown).To(BeEmpty())
	Expect(attachments[1].VolumeOptions).To(Equal("noexec"))
	Expect(attachments[1].Readonly).To(BeTrue())
}

// TestCommandMountReplaceReadonlyIsPerSpec covers the ro/rw pair, including
// that rw is consumed into the attachment rather than left in VolumeOptions,
// where it would be handed to docker as a mount option.
func TestCommandMountReplaceReadonlyIsPerSpec(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	stageDockerLocalEntry(t, "demo-data")
	stageDockerLocalEntry(t, "demo-cache")
	stageDockerLocalEntry(t, "demo-logs")

	Expect(CommandMount(CommandMountInput{
		AppName: "demo",
		Replace: true,
		Specs:   []string{"demo-data:/app/storage:ro", "demo-cache:/cache:rw", "demo-logs:/logs"},
	})).To(Succeed())

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(HaveLen(3))
	Expect(attachments[0].Readonly).To(BeTrue())
	Expect(attachments[1].Readonly).To(BeFalse())
	Expect(attachments[1].VolumeOptions).To(BeEmpty())
	Expect(attachments[2].Readonly).To(BeFalse())
}

// TestCommandMountReplaceDefaultsPhasesWithoutAToken pins the default a spec
// without a phase token gets, which is the same pair a mount has always had.
func TestCommandMountReplaceDefaultsPhasesWithoutAToken(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	stageDockerLocalEntry(t, "demo-data")

	Expect(CommandMount(CommandMountInput{
		AppName: "demo",
		Replace: true,
		Specs:   []string{"demo-data:/app/storage:Z"},
	})).To(Succeed())

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(HaveLen(1))
	Expect(attachments[0].Phases).To(Equal([]string{PhaseDeploy, PhaseRun}))
}

// TestCommandMountReplaceRejectsMountTimeFlags covers the other half of "one
// spelling per field": a flag that a token now expresses is refused rather
// than silently acting as a default. The --volume-options row is the case that
// was previously accepted and then dropped on the floor.
func TestCommandMountReplaceRejectsMountTimeFlags(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	stageDockerLocalEntry(t, "demo-data")

	for _, testCase := range []struct {
		name     string
		input    CommandMountInput
		contains string
	}{
		{"--phase", CommandMountInput{Phases: []string{PhaseDeploy}}, "The --phase flag cannot be used with --replace; set phase=deploy in the mount spec instead"},
		{"--volume-subpath", CommandMountInput{Subpath: "uploads"}, "The --volume-subpath flag cannot be used with --replace; set volume-subpath=<path> in the mount spec instead"},
		{"--volume-readonly", CommandMountInput{Readonly: true}, "The --volume-readonly flag cannot be used with --replace; set ro in the mount spec instead"},
		{"--volume-chown", CommandMountInput{VolumeChown: "herokuish"}, "The --volume-chown flag cannot be used with --replace; set volume-chown=<option> in the mount spec instead"},
		{"--volume-options", CommandMountInput{VolumeOptions: "Z"}, "The --volume-options flag cannot be used with --replace; set the option as a bare token in the mount spec instead"},
	} {
		input := testCase.input
		input.AppName = "demo"
		input.Replace = true
		input.Specs = []string{"demo-data:/app/storage"}

		err := CommandMount(input)
		Expect(err).To(HaveOccurred(), "expected %s to be rejected", testCase.name)
		Expect(err.Error()).To(Equal(testCase.contains), "rejecting %s", testCase.name)

		attachments, err := LoadAttachments("demo")
		Expect(err).NotTo(HaveOccurred())
		Expect(attachments).To(BeEmpty(), "rejecting %s wrote an attachment", testCase.name)
	}
}

// TestCommandMountReplaceValidatesTokensPerSpec keeps the all-or-nothing
// contract over the failures the token grammar adds: a bad token in a later
// spec leaves an earlier one unwritten and the stored set as it was.
func TestCommandMountReplaceValidatesTokensPerSpec(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	stageDockerLocalEntry(t, "demo-data")
	stageDockerLocalEntry(t, "demo-cache")

	Expect(CommandMount(CommandMountInput{
		AppName:      "demo",
		NameOrPath:   "demo-data",
		ContainerDir: "/app/storage",
	})).To(Succeed())

	for _, testCase := range []struct {
		spec     string
		contains string
	}{
		{"demo-cache:/cache:volume-chown=bogus", "Unsupported chown permissions"},
		{"demo-cache:/cache:phase=build", `specifies unsupported phase "build"`},
		{"demo-cache:/cache:volume-subpth=uploads", `specifies unknown key "volume-subpth"`},
	} {
		err := CommandMount(CommandMountInput{
			AppName: "demo",
			Replace: true,
			Specs:   []string{"demo-cache:/cache2", testCase.spec},
		})
		Expect(err).To(HaveOccurred(), "expected %q to be rejected", testCase.spec)
		Expect(err.Error()).To(ContainSubstring(testCase.contains))
		Expect(err.Error()).To(ContainSubstring(testCase.spec))

		attachments, err := LoadAttachments("demo")
		Expect(err).NotTo(HaveOccurred())
		Expect(attachments).To(HaveLen(1), "rejecting %q changed the stored set", testCase.spec)
		Expect(attachments[0].EntryName).To(Equal("demo-data"))
	}
}

// TestCommandMountReplaceRejectsPathHeldByAnotherScope covers the scopes a
// --replace call does not replace. The declared set is already unique on
// container path within its own scope, but a path held by a different process
// type would still collide at deploy time.
func TestCommandMountReplaceRejectsPathHeldByAnotherScope(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	stageDockerLocalEntry(t, "demo-data")
	stageDockerLocalEntry(t, "demo-cache")

	Expect(CommandMount(CommandMountInput{
		AppName:      "demo",
		NameOrPath:   "demo-data",
		ContainerDir: "/app/storage",
		ProcessType:  "web",
	})).To(Succeed())

	err := CommandMount(CommandMountInput{
		AppName: "demo",
		Replace: true,
		Specs:   []string{"demo-cache:/app/storage"},
	})
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("Container path /app/storage on app demo is already mounted by storage entry demo-data for process type web"))

	// The rejected call leaves the stored set exactly as it was.
	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(HaveLen(1))
	Expect(attachments[0].EntryName).To(Equal("demo-data"))
	Expect(attachments[0].ProcessType).To(Equal("web"))
}

// TestCommandMountReplaceAllowsPathInAReplacedScope is the other side of it:
// the scope being replaced is dropped before the check, so re-declaring a path
// that scope already held is the ordinary idempotent case, not a conflict.
func TestCommandMountReplaceAllowsPathInAReplacedScope(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	stageDockerLocalEntry(t, "demo-data")
	stageDockerLocalEntry(t, "demo-cache")

	Expect(CommandMount(CommandMountInput{
		AppName:      "demo",
		NameOrPath:   "demo-data",
		ContainerDir: "/app/storage",
	})).To(Succeed())

	Expect(CommandMount(CommandMountInput{
		AppName: "demo",
		Replace: true,
		Specs:   []string{"demo-cache:/app/storage"},
	})).To(Succeed())

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(HaveLen(1))
	Expect(attachments[0].EntryName).To(Equal("demo-cache"))
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

// TestCommandMountReplaceAcceptsLegacyHostPathWithTokens guards the entry
// hash: LegacyMountToEntry only ever sees the first colon field, so a token in
// the third one must not move a host path onto a different legacy-<hash>
// entry than the same path without it.
func TestCommandMountReplaceAcceptsLegacyHostPathWithTokens(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")

	Expect(CommandMount(CommandMountInput{
		AppName: "demo",
		Replace: true,
		Specs:   []string{"/var/lib/dokku/data/storage/shared:/app/shared:ro,volume-subpath=uploads"},
	})).To(Succeed())

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(HaveLen(1))
	Expect(attachments[0].Subpath).To(Equal("uploads"))
	Expect(attachments[0].Readonly).To(BeTrue())

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

// TestCommandUnmountIgnoresMountSpecTokens keeps a mount spec pasteable into
// an unmount call. An unmount target is identified by entry and container path
// alone, and the third field has always been ignored here - including "ro" -
// so the tokens that field now also carries are ignored the same way rather
// than being rejected only because they are newer.
func TestCommandUnmountIgnoresMountSpecTokens(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	stageDockerLocalEntry(t, "demo-data")

	Expect(CommandMount(CommandMountInput{
		AppName: "demo",
		Replace: true,
		Specs:   []string{"demo-data:/app/storage", "demo-data:/app/uploads:ro,volume-subpath=uploads"},
	})).To(Succeed())

	Expect(CommandUnmount(CommandUnmountInput{
		AppName: "demo",
		Mounts:  []string{"demo-data:/app/uploads:ro,volume-subpath=uploads"},
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
