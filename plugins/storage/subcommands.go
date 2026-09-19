package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

const (
	helpHeader = `Usage: dokku storage[:COMMAND]

Manage mounted volumes

Additional commands:`

	helpContent = `
    storage:annotations:report [<name>] [<flag>], Displays annotations for one or more storage entries
    storage:annotations:set <name> <key> [<value>], Set or clear an annotation on a storage entry
    storage:create <name> [<path>] [flags], Register a named storage entry
    storage:destroy <name> [--force] [--destroy-host-dir], Remove a named storage entry (must be unmounted from every app first)
    storage:ensure-directory [--chown option] <directory>, [DEPRECATED] use storage:create instead
    storage:exec <name> [-- <cmd>...], Run a command (or shell) in a temporary container that mounts the entry
    storage:info <name> [--format text|json], Show details for one storage entry
    storage:labels:report [<name>] [<flag>], Displays labels for one or more storage entries
    storage:labels:set <name> <key> [<value>], Set or clear a label on a storage entry
    storage:list <app> [--format text|json], List bind mounts for app's container(s) (host:container)
    storage:list-entries [--scheduler s] [--format text|json], List registered storage entries
    storage:migrate [<app>|--all], Re-run the legacy -v to attachment migration for an app
    storage:mount <app> <host-dir:container-dir>, Create a new bind mount
    storage:mounts:clear <app>, Remove every storage attachment from an app
    storage:mounts:set <app> [FILE|-] [--replace], Replace the complete set of an app's storage attachments from a JSON array
    storage:report [<app>] [<flag>], Displays a storage report for one or more apps
    storage:set <name> <property> [<value>], Update a storage entry in place
    storage:unmount <app> <host-dir:container-dir>, Remove an existing bind mount
    storage:wait <name>, Wait for a storage entry's PVC to be bound (k3s)`
)

// CommandHelp displays help for the storage plugin
func CommandHelp() error {
	common.CommandUsage(helpHeader, helpContent)
	return nil
}

// CommandEnsureDirectory creates a persistent storage directory
func CommandEnsureDirectory(directory string, chownFlag string) error {
	if err := ValidateDirectoryName(directory); err != nil {
		return err
	}

	chownID, err := ResolveChownID(chownFlag)
	if err != nil {
		return err
	}

	storageDirectory := filepath.Join(GetStorageDirectory(), directory)
	common.LogInfo1(fmt.Sprintf("Ensuring %s exists", storageDirectory))

	if err := os.MkdirAll(storageDirectory, 0755); err != nil {
		return fmt.Errorf("Unable to create directory: %s", err.Error())
	}

	if chownID != "false" {
		common.LogVerboseQuiet(fmt.Sprintf("Setting directory ownership to %s:%s", chownID, chownID))

		if err := callStorageDirScript("chown-storage-dir", directory, chownID); err != nil {
			return fmt.Errorf("Unable to set directory ownership: %s", err.Error())
		}
	}

	common.LogVerboseQuiet("Directory ready for mounting")
	return nil
}

// ResolveChownID converts a chown flag value to a numeric UID
func ResolveChownID(chownFlag string) (string, error) {
	var chownID string

	switch chownFlag {
	case "herokuish":
		chownID = "32767"
	case "heroku":
		chownID = "1000"
	case "packeto":
		common.LogVerbose("Detected deprecated chown flag 'packeto'. Using 'paketo' instead. Please update your configuration.")
		chownID = "2000"
	case "paketo":
		chownID = "2000"
	case "root":
		chownID = "0"
	case "false":
		return "false", nil
	default:
		id, err := strconv.ParseUint(chownFlag, 10, 16)

		if err != nil {
			return "", errors.New("Unsupported chown permissions")
		} else {
			chownID = fmt.Sprint(id)
		}
	}

	userns, err := isUserNamespacesEnabled()
	if err != nil {
		return "", err
	}

	if userns && chownID != "false" {
		uid := 0
		fmt.Sscanf(chownID, "%d", &uid)
		uid += 165536
		chownID = fmt.Sprintf("%d", uid)
	}

	return chownID, nil
}

// isUserNamespacesEnabled checks if Docker user namespaces are enabled
func isUserNamespacesEnabled() (bool, error) {
	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command: common.DockerBin(),
		Args:    []string{"info", "-f", "{{range .SecurityOptions}}{{if eq . \"name=userns\"}}true{{end}}{{end}}"},
	})
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(result.StdoutContents()) == "true", nil
}

// CommandMountInput captures the optional flags accepted by storage:mount
// when the second argument is a named entry rather than a colon-form path.
type CommandMountInput struct {
	AppName       string
	NameOrPath    string
	ContainerDir  string
	Phases        []string
	ProcessType   string
	Subpath       string
	Readonly      bool
	VolumeOptions string
	VolumeChown   string
}

// CommandMount creates a new bind mount for an app. The second positional
// argument may be either a legacy host:container[:options] string (kept
// for back-compat on docker-local) or a registered storage entry name.
func CommandMount(input CommandMountInput) error {
	if err := common.VerifyAppName(input.AppName); err != nil {
		return err
	}

	appScheduler := common.GetAppScheduler(input.AppName)

	// Legacy colon form: synthesize a legacy-<hash> entry plus an
	// attachment so storage:list (now attachment-only) sees the mount.
	// The storage docker-args trigger emits the corresponding -v flag at
	// deploy time, so behavior at the docker-run boundary is unchanged.
	if strings.Contains(input.NameOrPath, ":") {
		if appScheduler != SchedulerDockerLocal {
			return fmt.Errorf("colon-form mounts are only supported on docker-local apps; %s is a %s app. Create a named entry with 'dokku storage:create --scheduler %s' and mount it instead", input.AppName, appScheduler, appScheduler)
		}
		return mountLegacyColon(input.AppName, input.NameOrPath)
	}

	// Named-entry form: persist as an attachment.
	if !EntryExists(input.NameOrPath) {
		return fmt.Errorf("storage entry %q does not exist; create it first with `dokku storage:create`", input.NameOrPath)
	}
	if input.ContainerDir == "" {
		return errors.New("--container-dir is required when mounting a named storage entry")
	}

	entry, err := lookupMountableEntry(input.NameOrPath)
	if err != nil {
		return err
	}

	if entry.Scheduler != appScheduler {
		return fmt.Errorf("storage entry %q is scheduler=%s but cannot be mounted on a %s app; recreate it with --scheduler %s", entry.Name, entry.Scheduler, appScheduler, appScheduler)
	}

	phases := input.Phases
	if len(phases) == 0 {
		phases = []string{PhaseDeploy, PhaseRun}
	}

	processType := input.ProcessType
	if processType == "" {
		processType = DefaultProcessType
	}

	attachment := &Attachment{
		EntryName:     entry.Name,
		ContainerPath: input.ContainerDir,
		Phases:        phases,
		ProcessType:   processType,
		Subpath:       input.Subpath,
		Readonly:      input.Readonly,
		VolumeOptions: input.VolumeOptions,
		VolumeChown:   input.VolumeChown,
	}

	// Idempotent contract: same (entry, container_dir, process_type)
	// tuple updates the mount-time fields in place rather than appending
	// a duplicate attachment. Lets declarative tooling change
	// --volume-options / --volume-chown / --phase / etc. without an
	// unmount-then-remount dance that briefly drops the volume.
	created, err := UpsertAttachment(input.AppName, attachment)
	if err != nil {
		return err
	}
	if created {
		common.LogInfo1(fmt.Sprintf("Storage entry %s mounted at %s on %s", entry.Name, input.ContainerDir, input.AppName))
	} else {
		common.LogInfo1(fmt.Sprintf("Storage entry %s mount at %s on %s updated", entry.Name, input.ContainerDir, input.AppName))
	}
	return nil
}

// CommandMountsSetInput captures the inputs accepted by
// storage:mounts:set. The desired attachment set is always read as one JSON
// array, either from Filename or, when Filename is empty or "-", from stdin.
type CommandMountsSetInput struct {
	AppName  string
	Filename string
}

// CommandMountsSet replaces an app's entire attachment set with the declared
// set in one call. Every attachment is validated before the write, so a
// rejected item leaves the previously stored mounts untouched.
func CommandMountsSet(input CommandMountsSetInput) error {
	if err := common.VerifyAppName(input.AppName); err != nil {
		return err
	}

	reader, closeReader, err := mountsSetReader(input.Filename)
	if err != nil {
		return err
	}
	defer closeReader()

	attachments, err := decodeAttachments(reader)
	if err != nil {
		return err
	}

	appScheduler := common.GetAppScheduler(input.AppName)
	for i, attachment := range attachments {
		if err := validateAttachmentSetItem(appScheduler, attachment); err != nil {
			return fmt.Errorf("attachment %d: %s", i, err.Error())
		}
	}

	if err := SaveAttachments(input.AppName, attachments); err != nil {
		return err
	}

	common.LogInfo1(fmt.Sprintf("Storage attachments for %s replaced with %d mount(s)", input.AppName, len(attachments)))
	return nil
}

// CommandMountsClear removes every attachment from an app. Registry entries
// and the data they point at are left alone, and clearing an app that has no
// attachments succeeds.
func CommandMountsClear(appName string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if err := common.PropertyDelete(PluginName, appName, AttachmentsProperty); err != nil {
		return err
	}

	common.LogInfo1(fmt.Sprintf("Storage attachments for %s cleared", appName))
	return nil
}

// mountsSetReader returns a reader for the desired attachment set along with a
// closer. An empty filename or exactly "-" reads stdin; anything else is a file
// path, including a name that merely starts with a dash.
func mountsSetReader(filename string) (io.Reader, func(), error) {
	if filename == "" || filename == "-" {
		return os.Stdin, func() {}, nil
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, func() {}, err
	}

	return file, func() { file.Close() }, nil
}

// decodeAttachments reads one JSON array of attachments. The input must hold
// a single array and nothing else, and must contain at least one attachment;
// removing everything is what storage:mounts:clear is for. Unknown keys are
// ignored so that output from commands whose JSON carries extra fields, such
// as storage:list --format json, can be fed back in.
func decodeAttachments(reader io.Reader) ([]*Attachment, error) {
	decoder := json.NewDecoder(reader)

	// Decode via the raw form so an omitted "phases" can be told apart from an
	// explicit empty list: the first takes the defaults storage:mount applies,
	// the second is a declaration of nothing and is rejected.
	var raw []json.RawMessage
	err := decoder.Decode(&raw)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, errors.New("Must specify at least one mount via stdin, use storage:mounts:clear to remove all mounts")
		}

		return nil, fmt.Errorf("Unable to parse mounts: %s", err.Error())
	}

	if raw == nil || len(raw) == 0 {
		return nil, errors.New("Must specify at least one mount, use storage:mounts:clear to remove all mounts")
	}

	var extra any
	err = decoder.Decode(&extra)
	if err == nil {
		return nil, errors.New("Unable to parse mounts: unexpected content after the JSON array")
	}
	if !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("Unable to parse mounts: %s", err.Error())
	}

	attachments := make([]*Attachment, 0, len(raw))
	seen := map[attachmentIdentity]struct{}{}
	for i, item := range raw {
		var element any
		if err := json.Unmarshal(item, &element); err != nil {
			return nil, fmt.Errorf("attachment %d: %s", i, err.Error())
		}

		if element == nil {
			return nil, fmt.Errorf("attachment %d: entry is null", i)
		}

		var fields map[string]json.RawMessage
		if err := json.Unmarshal(item, &fields); err != nil {
			return nil, fmt.Errorf("attachment %d: expected an object with entry_name and container_path", i)
		}

		if _, ok := fields["phases"]; ok {
			var phases []string
			if err := json.Unmarshal(fields["phases"], &phases); err != nil {
				return nil, fmt.Errorf("attachment %d: phases must be a list of phase names", i)
			}

			// A JSON null decodes to a nil slice and means "unspecified", so it
			// takes the same defaults as an omitted key. Only an explicit empty
			// list is a declaration of nothing and rejected.
			if phases != nil && len(phases) == 0 {
				return nil, fmt.Errorf("attachment %d: %s", i, attachmentNoPhasesError)
			}
		}

		attachment := &Attachment{}
		if err := json.Unmarshal(item, attachment); err != nil {
			return nil, fmt.Errorf("attachment %d: %s", i, err.Error())
		}

		normalizeAttachment(attachment)

		identity := attachmentIdentity{
			entryName:     attachment.EntryName,
			containerPath: attachment.ContainerPath,
			processType:   attachment.ProcessType,
		}
		if _, ok := seen[identity]; ok {
			return nil, fmt.Errorf("attachment %d: storage entry %q is declared more than once at %q for process type %q", i, attachment.EntryName, attachment.ContainerPath, attachment.ProcessType)
		}
		seen[identity] = struct{}{}

		attachments = append(attachments, attachment)
	}

	return attachments, nil
}

// attachmentNoPhasesError is the exact wording Attachment.Validate uses for a
// missing phase, so mount:set and the single-mount commands agree.
const attachmentNoPhasesError = "attachment must specify at least one phase"

// attachmentIdentity is the tuple AddAttachment and UpsertAttachment treat as
// one attachment. Two entries sharing it are the same declaration written
// twice, so the whole-set form refuses to guess which one wins.
type attachmentIdentity struct {
	containerPath string
	entryName     string
	processType   string
}

// normalizeAttachment applies the same defaults storage:mount applies to a
// single named entry, so a declared set and an equivalent series of mount
// calls produce identical stored attachments.
func normalizeAttachment(attachment *Attachment) {
	if len(attachment.Phases) == 0 {
		attachment.Phases = []string{PhaseDeploy, PhaseRun}
	}

	if attachment.ProcessType == "" {
		attachment.ProcessType = DefaultProcessType
	}
}

// validateAttachmentSetItem checks one declared attachment against the app it
// is being mounted on: the entry name must be usable, the structural fields
// must be valid, the referenced entry must exist, and its scheduler must match
// the app's. Entry names are not path-validated - the registry is
// authoritative, which is what makes existing legacy-<hash> entries mountable.
func validateAttachmentSetItem(appScheduler string, attachment *Attachment) error {
	if err := attachment.Validate(); err != nil {
		return err
	}

	if err := ValidateEntryName(attachment.EntryName, true); err != nil {
		return err
	}

	entry, err := lookupMountableEntry(attachment.EntryName)
	if err != nil {
		return err
	}

	if entry.Scheduler != appScheduler {
		return fmt.Errorf("storage entry %q is scheduler=%s but cannot be mounted on a %s app; recreate it with --scheduler %s", entry.Name, entry.Scheduler, appScheduler, appScheduler)
	}

	return nil
}

// lookupMountableEntry loads a registered storage entry, reporting a missing
// entry with the same actionable hint storage:mount gives.
func lookupMountableEntry(name string) (*Entry, error) {
	if !EntryExists(name) {
		return nil, fmt.Errorf("storage entry %q does not exist; create it first with `dokku storage:create`", name)
	}

	return LoadEntry(name)
}

// CommandUnmountInput captures the flags accepted by storage:unmount when
// the second argument is a named entry.
type CommandUnmountInput struct {
	AppName      string
	NameOrPath   string
	ContainerDir string
}

// CommandUnmount removes an existing bind mount from an app.
func CommandUnmount(input CommandUnmountInput) error {
	if err := common.VerifyAppName(input.AppName); err != nil {
		return err
	}

	if strings.Contains(input.NameOrPath, ":") {
		return unmountLegacyColon(input.AppName, input.NameOrPath)
	}

	return RemoveAttachment(input.AppName, input.NameOrPath, input.ContainerDir)
}

// mountLegacyColon translates a `<host>:<container>[:options]` mount
// string into a synthesized legacy-<hash> entry plus an attachment.
// Idempotent: re-running the same mount errors with the existing
// "already mounted" message via AddAttachment's duplicate check.
func mountLegacyColon(appName string, mountPath string) error {
	if err := VerifyPaths(mountPath); err != nil {
		return err
	}
	parsed := ParseMountPath(mountPath)
	if parsed.ContainerPath == "" {
		return errors.New("Storage path must be two valid paths divided by colon.")
	}

	entry := LegacyMountToEntry(mountPath)
	if !EntryExists(entry.Name) {
		if err := SaveEntry(entry); err != nil {
			return err
		}
	}

	attachment := &Attachment{
		EntryName:     entry.Name,
		ContainerPath: parsed.ContainerPath,
		Phases:        []string{PhaseDeploy, PhaseRun},
		ProcessType:   DefaultProcessType,
		Readonly:      parsed.Readonly,
		VolumeOptions: parsed.VolumeOptions,
	}

	if err := AddAttachment(appName, attachment); err != nil {
		// AddAttachment's duplicate error mentions the entry name, but
		// the legacy form historically said "Mount path already
		// exists." - preserve that exact wording so existing automation
		// and the bats suite keep matching.
		if strings.Contains(err.Error(), "is already mounted at") {
			return errors.New("Mount path already exists.")
		}
		return err
	}
	return nil
}

// unmountLegacyColon is the inverse of mountLegacyColon. The legacy
// mount string identifies an entry+container-path tuple deterministically
// via LegacyMountToEntry, so we can route to RemoveAttachment.
func unmountLegacyColon(appName string, mountPath string) error {
	if err := VerifyPaths(mountPath); err != nil {
		return err
	}
	parsed := ParseMountPath(mountPath)
	if parsed.ContainerPath == "" {
		return errors.New("Storage path must be two valid paths divided by colon.")
	}

	entry := LegacyMountToEntry(mountPath)
	if err := RemoveAttachment(appName, entry.Name, parsed.ContainerPath); err != nil {
		// Match the legacy wording for "not currently mounted".
		if strings.Contains(err.Error(), "is not mounted") {
			return errors.New("Mount path does not exist.")
		}
		return err
	}
	return nil
}

// CommandList lists all bind mounts for an app. Reads attachments
// directly from the storage plugin's own state rather than going through
// the deprecated `storage-list` plugn trigger.
func CommandList(appName string, format string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if format != "text" && format != "json" {
		return errors.New("Invalid --format value specified")
	}

	rows, err := ListAppMountEntries(appName, PhaseDeploy)
	if err != nil {
		return err
	}

	if format == "json" {
		output, err := json.Marshal(rows)
		if err != nil {
			return err
		}
		fmt.Println(string(output))
		return nil
	}

	common.LogInfo1Quiet(fmt.Sprintf("%s volume bind-mounts:", appName))
	for _, row := range rows {
		line := formatStorageListEntry(row)
		if os.Getenv("DOKKU_QUIET_OUTPUT") != "" {
			fmt.Println(line)
		} else {
			common.LogVerbose(line)
		}
	}
	return nil
}

// CommandReport displays a storage report for one or more apps
func CommandReport(appName string, format string, infoFlag string) error {
	if appName == "--global" {
		// Global storage report: list every registered entry and its
		// attachment count; falls back to the legacy per-app loop when
		// no entries exist so existing automation keeps working.
		reportFormat := format
		if reportFormat == "stdout" {
			reportFormat = "text"
		}
		return CommandReportGlobal(reportFormat)
	}

	if appName == "" {
		apps, err := common.DokkuApps()
		if err != nil {
			if errors.Is(err, common.NoAppsExist) {
				common.LogWarn(err.Error())
				return nil
			}
			return err
		}
		for _, app := range apps {
			if err := ReportSingleApp(app, format, infoFlag); err != nil {
				return err
			}
		}
		return nil
	}

	return ReportSingleApp(appName, format, infoFlag)
}
