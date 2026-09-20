package storage

import (
	"encoding/json"
	"errors"
	"fmt"
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
    storage:annotations:set <name> <key> [<value>], Set or clear a single annotation on a storage entry
    storage:create <name> [<path>] [flags], Register a named storage entry
    storage:destroy <name> [--force] [--destroy-host-dir], Remove a named storage entry (must be unmounted from every app first)
    storage:ensure-directory [--chown option] <directory>, [DEPRECATED] use storage:create instead
    storage:exec <name> [-- <cmd>...], Run a command (or shell) in a temporary container that mounts the entry
    storage:info <name> [--format text|json], Show details for one storage entry
    storage:labels:report [<name>] [<flag>], Displays labels for one or more storage entries
    storage:labels:set <name> <key> [<value>], Set or clear a single label on a storage entry
    storage:list <app> [--format text|json], List bind mounts for an app's container(s) (legacy host:container view)
    storage:list-entries [--scheduler s] [--format text|json], List registered storage entries
    storage:migrate [<app>|--all], Re-run the legacy -v to attachment migration for an app
    storage:mount [--replace] <app> <host-dir:container-dir>... [flags], Create or replace bind mounts
    storage:report [<app>|--global] [<flag>], Displays a storage report for one or more apps
    storage:set <name> <property> [<value>], Update a storage entry in place
    storage:unmount [--all] <app> [<host-dir:container-dir>...] [flags], Remove one or all bind mounts
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

// ValidateChownOption reports whether a chown option is one the storage
// plugin understands, mirroring the set ResolveChownID maps. An empty value
// means no chown was requested. Unlike ResolveChownID it never shells out to
// docker for the user namespace offset, so it is safe to call as a pure
// validation step before anything has been written.
func ValidateChownOption(chownFlag string) error {
	switch chownFlag {
	case "", "herokuish", "heroku", "packeto", "paketo", "root", "false":
		return nil
	}

	if _, err := strconv.ParseUint(chownFlag, 10, 16); err != nil {
		return errors.New("Unsupported chown permissions")
	}

	return nil
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
//
// Replace and Specs drive the whole-set form, where the positional arguments
// are mount specs rather than a single entry and the remaining mount-time
// fields scope every spec in the call.
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
	Replace       bool
	Specs         []string
}

// CommandMount creates a new bind mount for an app. The second positional
// argument may be either a legacy host:container[:options] string (kept
// for back-compat on docker-local) or a registered storage entry name.
func CommandMount(input CommandMountInput) error {
	if err := common.VerifyAppName(input.AppName); err != nil {
		return err
	}

	if input.Replace {
		return replaceMounts(input)
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

	entry, err := LoadEntry(input.NameOrPath)
	if err != nil {
		return err
	}

	if entry.Scheduler != appScheduler {
		return fmt.Errorf("storage entry %q is scheduler=%s but cannot be mounted on a %s app; recreate it with --scheduler %s", entry.Name, entry.Scheduler, appScheduler, appScheduler)
	}

	if err := ValidateChownOption(input.VolumeChown); err != nil {
		return err
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

// mountSpec is one parsed positional argument of the storage:mount --replace
// form. It names an entry and the container path to bind it at, plus the
// mount-time fields the colon form can express on its own.
type mountSpec struct {
	EntryName     string
	ContainerPath string
	Readonly      bool
	VolumeOptions string
}

// parseMountSpec splits an "<entry>:<container-dir>[:<options>]" argument.
// ParseMountPath already implements that grammar - including hoisting the "ro"
// token out of the options list - so the only difference here is that the first
// field names a storage entry rather than a host path.
func parseMountSpec(spec string) (mountSpec, error) {
	parsed := ParseMountPath(spec)
	if parsed.HostPath == "" || parsed.ContainerPath == "" {
		return mountSpec{}, fmt.Errorf("Invalid mount specified: %s", spec)
	}

	return mountSpec{
		EntryName:     parsed.HostPath,
		ContainerPath: parsed.ContainerPath,
		Readonly:      parsed.Readonly,
		VolumeOptions: parsed.VolumeOptions,
	}, nil
}

// resolveMountSpecEntry maps a parsed spec onto the storage entry it names.
//
// The single-mount form tells a named entry from a colon-form host path by the
// presence of a colon, which every --replace spec has, so a leading "/" is the
// discriminator instead. Without that rule a mistyped entry name would match
// the docker volume grammar and silently register a legacy entry for a volume
// that does not exist.
//
// The returned bool reports whether the entry still needs registering, so the
// registry write can wait until every spec in the call has validated.
func resolveMountSpecEntry(appName string, appScheduler string, spec string, parsed mountSpec) (*Entry, bool, error) {
	if strings.HasPrefix(parsed.EntryName, "/") {
		if appScheduler != SchedulerDockerLocal {
			return nil, false, fmt.Errorf("colon-form mounts are only supported on docker-local apps; %s is a %s app. Create a named entry with 'dokku storage:create --scheduler %s' and mount it instead", appName, appScheduler, appScheduler)
		}

		if err := VerifyPaths(spec); err != nil {
			return nil, false, err
		}

		entry := LegacyMountToEntry(spec)
		return entry, !EntryExists(entry.Name), nil
	}

	if !EntryExists(parsed.EntryName) {
		return nil, false, fmt.Errorf("storage entry %q does not exist; create it first with `dokku storage:create`", parsed.EntryName)
	}

	entry, err := LoadEntry(parsed.EntryName)
	if err != nil {
		return nil, false, err
	}

	if entry.Scheduler != appScheduler {
		return nil, false, fmt.Errorf("storage entry %q is scheduler=%s but cannot be mounted on a %s app; recreate it with --scheduler %s", entry.Name, entry.Scheduler, appScheduler, appScheduler)
	}

	return entry, false, nil
}

// attachmentProcessType returns the process type an attachment is scoped to,
// treating an unset value as the default scope so attachments written before
// the field existed still sort into a scope.
func attachmentProcessType(attachment *Attachment) string {
	if attachment.ProcessType == "" {
		return DefaultProcessType
	}

	return attachment.ProcessType
}

// replaceMounts swaps one process type's entire mount set for the declared one
// in a single write, so a caller converging an app onto a declared set no
// longer issues one storage:mount or storage:unmount per difference and can no
// longer fail partway through holding a mixture of the two sets.
//
// Every spec is parsed and validated before anything is written, so a rejected
// spec leaves the stored set untouched. Attachments belonging to other process
// types are kept: an omitted --process-type scopes the replacement to
// _default_ the same way it scopes a single storage:mount.
func replaceMounts(input CommandMountInput) error {
	if input.ContainerDir != "" {
		return errors.New("The --container-dir flag cannot be used with --replace")
	}

	if len(input.Specs) == 0 {
		return errors.New("Must specify at least one mount, use storage:unmount --all to remove all mounts")
	}

	if err := ValidateChownOption(input.VolumeChown); err != nil {
		return err
	}

	phases := input.Phases
	if len(phases) == 0 {
		phases = []string{PhaseDeploy, PhaseRun}
	}

	processType := input.ProcessType
	if processType == "" {
		processType = DefaultProcessType
	}

	appScheduler := common.GetAppScheduler(input.AppName)
	declared := []*Attachment{}
	pending := []*Entry{}
	containerPaths := map[string]bool{}
	for _, spec := range input.Specs {
		parsed, err := parseMountSpec(spec)
		if err != nil {
			return err
		}

		entry, needsRegister, err := resolveMountSpecEntry(input.AppName, appScheduler, spec, parsed)
		if err != nil {
			return err
		}

		// Two entries bound at one container path inside one process type
		// would emit two -v flags for the same target, so the declared set
		// must name each container path at most once.
		if containerPaths[parsed.ContainerPath] {
			return fmt.Errorf("Container path %s is specified more than once", parsed.ContainerPath)
		}
		containerPaths[parsed.ContainerPath] = true

		attachment := &Attachment{
			EntryName:     entry.Name,
			ContainerPath: parsed.ContainerPath,
			Phases:        phases,
			ProcessType:   processType,
			Subpath:       input.Subpath,
			Readonly:      input.Readonly || parsed.Readonly,
			VolumeOptions: parsed.VolumeOptions,
			VolumeChown:   input.VolumeChown,
		}
		if err := attachment.Validate(); err != nil {
			return err
		}

		if needsRegister {
			pending = append(pending, entry)
		}
		declared = append(declared, attachment)
	}

	for _, entry := range pending {
		if err := SaveEntry(entry); err != nil {
			return err
		}
	}

	existing, err := LoadAttachments(input.AppName)
	if err != nil {
		return err
	}

	attachments := []*Attachment{}
	for _, attachment := range existing {
		if attachmentProcessType(attachment) != processType {
			attachments = append(attachments, attachment)
		}
	}
	attachments = append(attachments, declared...)

	if err := SaveAttachments(input.AppName, attachments); err != nil {
		return err
	}

	common.LogInfo1(fmt.Sprintf("Storage mounts for process type %s replaced on %s", processType, input.AppName))
	return nil
}

// CommandUnmountInput captures the arguments accepted by storage:unmount.
// Mounts holds one or more entry names or colon-form mount strings, each
// naming an attachment to remove.
//
// All and ProcessType drive the whole-set form, where no mount is named and
// every attachment in scope is removed at once.
type CommandUnmountInput struct {
	AppName      string
	Mounts       []string
	ContainerDir string
	All          bool
	ProcessType  string
}

// CommandUnmount removes an existing bind mount from an app.
func CommandUnmount(input CommandUnmountInput) error {
	if err := common.VerifyAppName(input.AppName); err != nil {
		return err
	}

	if input.All {
		return unmountAll(input)
	}

	if input.ProcessType != "" {
		return errors.New("The --process-type flag can only be used with --all")
	}

	if len(input.Mounts) == 0 {
		return errors.New("Must specify at least one mount, use storage:unmount --all to remove all mounts")
	}

	// --container-dir disambiguates a single entry mounted at several paths,
	// which has no meaning once more than one mount is named.
	if len(input.Mounts) > 1 && input.ContainerDir != "" {
		return errors.New("The --container-dir flag cannot be used with multiple mounts")
	}

	return unmountMounts(input)
}

// unmountTarget names the attachment one storage:unmount argument resolves to.
type unmountTarget struct {
	EntryName     string
	ContainerPath string
	Legacy        bool
}

// resolveUnmountTarget maps one storage:unmount argument onto the attachment it
// names. An argument without a colon is a registered entry name, disambiguated
// by --container-dir when the entry is mounted at more than one path. An
// argument with a colon names a registered entry when its first field is one,
// and otherwise falls back to the legacy host:container form, which keeps a
// host path or docker volume name resolving the way it always has.
func resolveUnmountTarget(mount string, containerDir string) (unmountTarget, error) {
	if !strings.Contains(mount, ":") {
		return unmountTarget{EntryName: mount, ContainerPath: containerDir}, nil
	}

	parsed := ParseMountPath(mount)
	if !strings.HasPrefix(parsed.HostPath, "/") && EntryExists(parsed.HostPath) {
		if parsed.ContainerPath == "" {
			return unmountTarget{}, fmt.Errorf("Invalid mount specified: %s", mount)
		}

		return unmountTarget{EntryName: parsed.HostPath, ContainerPath: parsed.ContainerPath}, nil
	}

	if err := VerifyPaths(mount); err != nil {
		return unmountTarget{}, err
	}

	if parsed.ContainerPath == "" {
		return unmountTarget{}, errors.New("Storage path must be two valid paths divided by colon.")
	}

	return unmountTarget{
		EntryName:     LegacyMountToEntry(mount).Name,
		ContainerPath: parsed.ContainerPath,
		Legacy:        true,
	}, nil
}

// unmountMounts removes every named mount in a single write. Each argument is
// resolved and matched against the app's attachments before anything is
// written, so an argument naming a mount the app does not have leaves the rest
// of the set in place rather than removing a prefix of it.
func unmountMounts(input CommandUnmountInput) error {
	attachments, err := LoadAttachments(input.AppName)
	if err != nil {
		return err
	}

	for _, mount := range input.Mounts {
		target, err := resolveUnmountTarget(mount, input.ContainerDir)
		if err != nil {
			return err
		}

		attachments, err = removeMatchingAttachments(attachments, input.AppName, target.EntryName, target.ContainerPath)
		if err != nil {
			// The legacy form historically said "Mount path does not
			// exist." - preserve that exact wording so existing automation
			// and the bats suite keep matching.
			if target.Legacy && strings.Contains(err.Error(), "is not mounted") {
				return errors.New("Mount path does not exist.")
			}
			return err
		}
	}

	return SaveAttachments(input.AppName, attachments)
}

// unmountAll removes every attachment on an app, or every attachment for one
// process type when --process-type is given. This is the empty case the
// --replace form refuses to treat as a request to remove everything, so that a
// generated spec list which expands to nothing cannot silently drop an app's
// mounts. A scope matching nothing is a no-op rather than an error, leaving the
// command idempotent.
func unmountAll(input CommandUnmountInput) error {
	if len(input.Mounts) > 0 {
		return errors.New("A mount cannot be specified with --all")
	}

	if input.ContainerDir != "" {
		return errors.New("The --container-dir flag cannot be used with --all")
	}

	if input.ProcessType == "" {
		if err := common.PropertyDelete(PluginName, input.AppName, AttachmentsProperty); err != nil {
			return err
		}

		common.LogInfo1(fmt.Sprintf("Removed all storage mounts on %s", input.AppName))
		return nil
	}

	attachments, err := LoadAttachments(input.AppName)
	if err != nil {
		return err
	}

	keep := []*Attachment{}
	for _, attachment := range attachments {
		if attachmentProcessType(attachment) != input.ProcessType {
			keep = append(keep, attachment)
		}
	}

	if err := SaveAttachments(input.AppName, keep); err != nil {
		return err
	}

	common.LogInfo1(fmt.Sprintf("Removed all storage mounts for process type %s on %s", input.ProcessType, input.AppName))
	return nil
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
