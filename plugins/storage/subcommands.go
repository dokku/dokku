package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

const (
	helpHeader = `Usage: dokku storage[:COMMAND]

Manage mounted volumes

Additional commands:`

	helpContent = `
    storage:create <name> [<path>] [flags], Register a named storage entry
    storage:destroy <name>, Remove a named storage entry (must be unmounted from every app first)
    storage:ensure-directory [--chown option] <directory>, [DEPRECATED] use storage:create instead
    storage:exec <name> [-- <cmd>...], Run a command (or shell) in a temporary container that mounts the entry
    storage:info <name> [--format text|json], Show details for one storage entry
    storage:list <app> [--format text|json], List bind mounts for app's container(s) (host:container)
    storage:list-entries [--scheduler s] [--format text|json], List registered storage entries
    storage:migrate [<app>|--all], Re-run the legacy -v to attachment migration for an app
    storage:mount <app> <host-dir:container-dir>, Create a new bind mount
    storage:report [<app>] [<flag>], Displays a storage report for one or more apps
    storage:set <name> [flags], Update a storage entry in place
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

		pluginPath := common.MustGetEnv("PLUGIN_AVAILABLE_PATH")
		chownScript := filepath.Join(pluginPath, "storage", "bin", "chown-storage-dir")

		result, err := common.CallExecCommand(common.ExecCommandInput{
			Command: "sudo",
			Args:    []string{chownScript, directory, chownID},
		})
		if err != nil {
			return fmt.Errorf("Unable to set directory ownership: %s", err.Error())
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("Unable to set directory ownership: %s", result.StderrContents())
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
		return "", errors.New("Unsupported chown permissions")
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

	// Legacy colon form: synthesize a legacy-<hash> entry plus an
	// attachment so storage:list (now attachment-only) sees the mount.
	// The storage docker-args trigger emits the corresponding -v flag at
	// deploy time, so behavior at the docker-run boundary is unchanged.
	if strings.Contains(input.NameOrPath, ":") {
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

	return AddAttachment(input.AppName, attachment)
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
	}
	switch parsed.VolumeOptions {
	case "":
	case "ro":
		attachment.Readonly = true
	default:
		attachment.VolumeOptions = parsed.VolumeOptions
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
