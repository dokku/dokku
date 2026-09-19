package storage

import (
	"bytes"
	"errors"
	"github.com/dokku/dokku/plugins/common"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

// attachmentsPath returns the property file storage:mounts:set writes for an
// app under the current DOKKU_LIB_ROOT.
func attachmentsPath(t *testing.T, appName string) string {
	t.Helper()
	return filepath.Join(common.MustGetEnv("DOKKU_LIB_ROOT"), "config", PluginName, appName, AttachmentsProperty)
}

// mountsSet runs storage:mounts:set against the desired set that
// captureStdin has already installed on os.Stdin.
func mountsSet(t *testing.T, appName string) error {
	t.Helper()
	return CommandMountsSet(CommandMountsSetInput{
		AppName:  appName,
		Filename: "-",
	})
}

// captureStdin swaps os.Stdin for the duration of fn so a stdin read in
// CommandMountsSet sees the supplied input.
func captureStdin(t *testing.T, input string, fn func() error) error {
	t.Helper()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}

	original := os.Stdin
	os.Stdin = reader
	defer func() {
		os.Stdin = original
		reader.Close()
	}()

	go func() {
		io.WriteString(writer, input)
		writer.Close()
	}()

	return fn()
}

func seedDockerLocalEntry(t *testing.T, appName string) {
	t.Helper()
	Expect(SaveEntry(&Entry{
		Name:      appName + "-data",
		Scheduler: SchedulerDockerLocal,
		HostPath:  "/data",
	})).To(Succeed())
}

func TestCommandMountsSetReplacesCompleteSet(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	seedDockerLocalEntry(t, "demo")

	// An attachment that the declared set omits must disappear, and the
	// retained one must have its mount-time fields overwritten wholesale.
	writeAttachmentsFile(t, common.MustGetEnv("DOKKU_LIB_ROOT"), "demo", []*Attachment{
		{EntryName: "demo-data", ContainerPath: "/data", Phases: []string{PhaseDeploy, PhaseRun}, ProcessType: DefaultProcessType},
		{EntryName: "demo-old", ContainerPath: "/old", Phases: []string{PhaseRun}, ProcessType: "worker"},
	})

	desired := `[
		{"entry_name":"demo-data","container_path":"/data","volume_options":"Z","readonly":true},
		{"entry_name":"demo-data","container_path":"/logs","phases":["run"],"process_type":"worker","subpath":"exports","volume_chown":"1000"}
	]`

	Expect(captureStdin(t, desired, func() error {
		return mountsSet(t, "demo")
	})).To(Succeed())

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(HaveLen(2))

	Expect(attachments[0].EntryName).To(Equal("demo-data"))
	Expect(attachments[0].ContainerPath).To(Equal("/data"))
	Expect(attachments[0].VolumeOptions).To(Equal("Z"))
	Expect(attachments[0].Readonly).To(BeTrue())
	Expect(attachments[0].Phases).To(Equal([]string{PhaseDeploy, PhaseRun}))
	Expect(attachments[0].ProcessType).To(Equal(DefaultProcessType))

	Expect(attachments[1].ContainerPath).To(Equal("/logs"))
	Expect(attachments[1].ProcessType).To(Equal("worker"))
	Expect(attachments[1].Phases).To(Equal([]string{PhaseRun}))
	Expect(attachments[1].Subpath).To(Equal("exports"))
	Expect(attachments[1].VolumeChown).To(Equal("1000"))

	// The omitted attachment is gone; input order is preserved.
	for _, attachment := range attachments {
		Expect(attachment.ContainerPath).NotTo(Equal("/old"))
	}
}

func TestCommandMountsSetIsIdempotent(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	seedDockerLocalEntry(t, "demo")

	desired := `[{"entry_name":"demo-data","container_path":"/data"}]`
	Expect(captureStdin(t, desired, func() error {
		return mountsSet(t, "demo")
	})).To(Succeed())

	first, err := os.ReadFile(attachmentsPath(t, "demo"))
	Expect(err).NotTo(HaveOccurred())

	Expect(captureStdin(t, desired, func() error {
		return mountsSet(t, "demo")
	})).To(Succeed())

	second, err := os.ReadFile(attachmentsPath(t, "demo"))
	Expect(err).NotTo(HaveOccurred())

	Expect(second).To(Equal(first))
}

func TestCommandMountsSetReadsFromFileArgument(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	seedDockerLocalEntry(t, "demo")

	path := filepath.Join(t.TempDir(), "mounts.json")
	Expect(os.WriteFile(path, []byte(`[{"entry_name":"demo-data","container_path":"/data"}]`), 0644)).To(Succeed())

	Expect(CommandMountsSet(CommandMountsSetInput{AppName: "demo", Filename: path})).To(Succeed())

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(HaveLen(1))
	Expect(attachmentsPath(t, "demo")).To(BeAnExistingFile())
}

func TestCommandMountsSetAcceptsListJSONWithExtraKeys(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	seedDockerLocalEntry(t, "demo")

	// storage:list --format json carries host_path and omits the fields this
	// representation defaults. Unknown keys are ignored rather than rejected.
	listJSON := `[{"entry_name":"demo-data","host_path":"/data","container_path":"/data","readonly":true,"volume_options":"Z"}]`
	Expect(captureStdin(t, listJSON, func() error {
		return mountsSet(t, "demo")
	})).To(Succeed())

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(HaveLen(1))
	Expect(attachments[0].Readonly).To(BeTrue())
	Expect(attachments[0].VolumeOptions).To(Equal("Z"))
	Expect(attachments[0].Phases).To(Equal([]string{PhaseDeploy, PhaseRun}))
	Expect(attachments[0].ProcessType).To(Equal(DefaultProcessType))
	Expect(attachmentsPath(t, "demo")).To(BeAnExistingFile())
}

func TestCommandMountsSetRejectsInvalidInput(t *testing.T) {
	cases := []struct {
		name     string
		desired  string
		contains string
	}{
		{"empty stream", "", "Must specify at least one mount via stdin"},
		{"whitespace only", "   \n", "Must specify at least one mount via stdin"},
		{"empty array", "[]", "Must specify at least one mount, use storage:mounts:clear to remove all mounts"},
		{"null root", "null", "Must specify at least one mount, use storage:mounts:clear to remove all mounts"},
		{"null element", `[null]`, "entry is null"},
		{"malformed json", `[{"entry_name":`, "Unable to parse mounts"},
		{"object root", `{"entry_name":"demo-data","container_path":"/data"}`, "Unable to parse mounts"},
		{"trailing document", `[{"entry_name":"demo-data","container_path":"/data"}] []`, "unexpected content after the JSON array"},
		{"trailing garbage", `[{"entry_name":"demo-data","container_path":"/data"}] x`, "Unable to parse mounts"},
		{"explicit empty phases", `[{"entry_name":"demo-data","container_path":"/data","phases":[]}]`, "must specify at least one phase"},
		{"unknown phase", `[{"entry_name":"demo-data","container_path":"/data","phases":["release"]}]`, `phase "release" is not supported`},
		{"relative container path", `[{"entry_name":"demo-data","container_path":"data"}]`, "must be absolute"},
		{"missing entry name", `[{"container_path":"/data"}]`, "missing entry name"},
		{"invalid entry name", `[{"entry_name":"Demo_Data","container_path":"/data"}]`, "must be a DNS-1123 label"},
		{"missing entry", `[{"entry_name":"demo-absent","container_path":"/data"}]`, "does not exist; create it first with `dokku storage:create`"},
		{"duplicate tuple", `[{"entry_name":"demo-data","container_path":"/data"},{"entry_name":"demo-data","container_path":"/data"}]`, "is declared more than once"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			RegisterTestingT(t)
			setupTestApp(t, "demo")
			seedDockerLocalEntry(t, "demo")

			err := captureStdin(t, testCase.desired, func() error {
				return mountsSet(t, "demo")
			})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(testCase.contains))
		})
	}
}

func TestCommandMountsSetRejectsDuplicateDifferingOnlyInMountFields(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	seedDockerLocalEntry(t, "demo")

	// Same identity tuple, different mount-time fields: the whole-set form
	// refuses to guess which declaration wins.
	desired := `[
		{"entry_name":"demo-data","container_path":"/data","readonly":true},
		{"entry_name":"demo-data","container_path":"/data","volume_options":"Z"}
	]`

	err := captureStdin(t, desired, func() error {
		return mountsSet(t, "demo")
	})

	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("is declared more than once"))
}

func TestCommandMountsSetAllowsTupleDistinguishedByProcessType(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	seedDockerLocalEntry(t, "demo")

	desired := `[
		{"entry_name":"demo-data","container_path":"/data"},
		{"entry_name":"demo-data","container_path":"/data","process_type":"worker"}
	]`

	Expect(captureStdin(t, desired, func() error {
		return mountsSet(t, "demo")
	})).To(Succeed())

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(HaveLen(2))
	Expect(attachments[0].ProcessType).To(Equal(DefaultProcessType))
	Expect(attachments[1].ProcessType).To(Equal("worker"))
}

func TestCommandMountsSetRejectsSchedulerMismatch(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")

	Expect(SaveEntry(&Entry{
		Name:      "demo-pvc",
		Scheduler: SchedulerK3s,
	})).To(Succeed())

	desired := `[{"entry_name":"demo-pvc","container_path":"/data"}]`
	err := captureStdin(t, desired, func() error {
		return mountsSet(t, "demo")
	})

	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(Equal(`attachment 0: storage entry "demo-pvc" is scheduler=k3s but cannot be mounted on a docker-local app; recreate it with --scheduler docker-local`))
}

func TestCommandMountsSetPreservesExistingFileWhenALaterItemIsInvalid(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	seedDockerLocalEntry(t, "demo")

	writeAttachmentsFile(t, common.MustGetEnv("DOKKU_LIB_ROOT"), "demo", []*Attachment{
		{EntryName: "demo-data", ContainerPath: "/keep", Phases: []string{PhaseDeploy, PhaseRun}, ProcessType: DefaultProcessType},
	})

	before, err := os.ReadFile(attachmentsPath(t, "demo"))
	Expect(err).NotTo(HaveOccurred())
	Expect(before).NotTo(BeEmpty())

	desired := `[
		{"entry_name":"demo-data","container_path":"/valid"},
		{"entry_name":"demo-absent","container_path":"/invalid"}
	]`

	Expect(captureStdin(t, desired, func() error {
		return mountsSet(t, "demo")
	})).To(HaveOccurred())

	after, err := os.ReadFile(attachmentsPath(t, "demo"))
	Expect(err).NotTo(HaveOccurred())
	Expect(after).To(Equal(before))

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(HaveLen(1))
	Expect(attachments[0].ContainerPath).To(Equal("/keep"))
}

func TestCommandMountsSetFailureDoesNotCreateState(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	seedDockerLocalEntry(t, "demo")

	desired := `[{"entry_name":"demo-data","container_path":"/data"},{"entry_name":"demo-absent","container_path":"/absent"}]`
	Expect(captureStdin(t, desired, func() error {
		return mountsSet(t, "demo")
	})).To(HaveOccurred())

	Expect(attachmentsPath(t, "demo")).NotTo(BeAnExistingFile())

	// Validation reads the registry but never writes to it.
	Expect(EntryExists("demo-absent")).To(BeFalse())
}

func TestCommandMountsSetRejectsUnknownApp(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")

	desired := `[{"entry_name":"demo-data","container_path":"/data"}]`
	if err := captureStdin(t, desired, func() error {
		return mountsSet(t, "absent")
	}); err == nil {
		t.Fatal("expected an error for an app that does not exist")
	}
}

func TestCommandMountsClearRemovesEveryAttachment(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	seedDockerLocalEntry(t, "demo")

	writeAttachmentsFile(t, common.MustGetEnv("DOKKU_LIB_ROOT"), "demo", []*Attachment{
		{EntryName: "demo-data", ContainerPath: "/data", Phases: []string{PhaseDeploy, PhaseRun}, ProcessType: DefaultProcessType},
		{EntryName: "demo-data", ContainerPath: "/logs", Phases: []string{PhaseRun}, ProcessType: "worker"},
	})

	Expect(CommandMountsClear("demo")).To(Succeed())

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(BeEmpty())

	// The registry entry and its host data are untouched: clearing attachments
	// is not entry destruction.
	Expect(EntryExists("demo-data")).To(BeTrue())

	// Clearing again is a no-op rather than an error.
	Expect(CommandMountsClear("demo")).To(Succeed())
}

func TestCommandMountsClearDoesNotAffectOtherApps(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")

	writeAttachmentsFile(t, common.MustGetEnv("DOKKU_LIB_ROOT"), "other", []*Attachment{
		{EntryName: "other-data", ContainerPath: "/other", Phases: []string{PhaseDeploy, PhaseRun}, ProcessType: DefaultProcessType},
	})

	Expect(CommandMountsClear("demo")).To(Succeed())

	attachments, err := LoadAttachments("other")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(HaveLen(1))
	Expect(attachments[0].EntryName).To(Equal("other-data"))
}

func TestCommandMountsClearRejectsUnknownApp(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")

	Expect(CommandMountsClear("absent")).To(HaveOccurred())
}

func TestDecodeAttachmentsRequiresEOFAfterArray(t *testing.T) {
	RegisterTestingT(t)

	_, err := decodeAttachments(strings.NewReader(`[{"entry_name":"demo-data","container_path":"/data"}] trailing`))
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("Unable to parse mounts"))

	_, err = decodeAttachments(strings.NewReader(`[{"entry_name":"demo-data","container_path":"/data"}]`))
	Expect(err).NotTo(HaveOccurred())
}

func TestDecodeAttachmentsAppliesDefaults(t *testing.T) {
	RegisterTestingT(t)

	attachments, err := decodeAttachments(bytes.NewReader([]byte(`[{"entry_name":"demo-data","container_path":"/data"}]`)))
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(HaveLen(1))
	Expect(attachments[0].Phases).To(Equal([]string{PhaseDeploy, PhaseRun}))
	Expect(attachments[0].ProcessType).To(Equal(DefaultProcessType))
}

// A JSON null phases value decodes to a nil slice and means "unspecified", so
// it must take the same defaults as an omitted key. Regression: this was
// rejected because a nil slice also has length zero.
func TestDecodeAttachmentsTreatsNullPhasesAsUnspecified(t *testing.T) {
	RegisterTestingT(t)

	for _, desired := range []string{
		`[{"entry_name":"demo-data","container_path":"/data","phases":null}]`,
		`[{"entry_name":"demo-data","container_path":"/data"}]`,
	} {
		attachments, err := decodeAttachments(strings.NewReader(desired))
		Expect(err).NotTo(HaveOccurred(), "input %s", desired)
		Expect(attachments).To(HaveLen(1))
		Expect(attachments[0].Phases).To(Equal([]string{PhaseDeploy, PhaseRun}), "input %s", desired)
	}

	// An explicit empty list is still a declaration of nothing.
	_, err := decodeAttachments(strings.NewReader(`[{"entry_name":"demo-data","container_path":"/data","phases":[]}]`))
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("must specify at least one phase"))
}

func TestCommandMountsSetTreatsNullPhasesAsUnspecified(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t, "demo")
	seedDockerLocalEntry(t, "demo")

	desired := `[{"entry_name":"demo-data","container_path":"/data","phases":null}]`
	Expect(captureStdin(t, desired, func() error {
		return mountsSet(t, "demo")
	})).To(Succeed())

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(HaveLen(1))
	Expect(attachments[0].Phases).To(Equal([]string{PhaseDeploy, PhaseRun}))
}

func TestMountsSetReaderUsesStdinForDash(t *testing.T) {
	RegisterTestingT(t)

	for _, filename := range []string{"", "-"} {
		reader, closer, err := mountsSetReader(filename)
		Expect(err).NotTo(HaveOccurred())
		Expect(reader).To(Equal(io.Reader(os.Stdin)), "expected %q to read stdin", filename)
		closer()
	}

	_, _, err := mountsSetReader(filepath.Join(t.TempDir(), "absent.json"))
	Expect(err).To(HaveOccurred())
	Expect(errors.Is(err, os.ErrNotExist)).To(BeTrue())
}

// Regression: a filename merely starting with a dash is a file path, not a
// request for stdin. Treating it as stdin would silently apply different input
// than the operator named. The name must genuinely start with a dash, so the
// test chdirs rather than prefixing a directory to it.
func TestMountsSetReaderTreatsDashPrefixedNameAsFile(t *testing.T) {
	RegisterTestingT(t)

	dir := t.TempDir()
	name := "-mounts.json"
	Expect(os.WriteFile(filepath.Join(dir, name), []byte(`[{"entry_name":"demo-data","container_path":"/data"}]`), 0644)).To(Succeed())

	t.Chdir(dir)

	reader, closer, err := mountsSetReader(name)
	Expect(err).NotTo(HaveOccurred())
	defer closer()

	Expect(io.Reader(reader)).NotTo(Equal(io.Reader(os.Stdin)))

	attachments, err := decodeAttachments(reader)
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(HaveLen(1))
	Expect(attachments[0].EntryName).To(Equal("demo-data"))
}
