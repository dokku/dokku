package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	dockeroptions "github.com/dokku/dokku/plugins/docker-options"
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

	// Legacy colon form: keep writing into docker-options so existing
	// users see no behavior change.
	if strings.Contains(input.NameOrPath, ":") {
		if err := VerifyPaths(input.NameOrPath); err != nil {
			return err
		}
		if CheckIfPathExists(input.AppName, input.NameOrPath, MountPhases) {
			return errors.New("Mount path already exists.")
		}
		return dockeroptions.AddDockerOptionToPhases(input.AppName, MountPhases, fmt.Sprintf("-v %s", input.NameOrPath))
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
		if err := VerifyPaths(input.NameOrPath); err != nil {
			return err
		}
		if !CheckIfPathExists(input.AppName, input.NameOrPath, MountPhases) {
			return errors.New("Mount path does not exist.")
		}
		return dockeroptions.RemoveDockerOptionFromPhases(input.AppName, MountPhases, fmt.Sprintf("-v %s", input.NameOrPath))
	}

	return RemoveAttachment(input.AppName, input.NameOrPath, input.ContainerDir)
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
