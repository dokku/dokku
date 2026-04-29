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

// CommandCreateInput captures the flags accepted by storage:create.
type CommandCreateInput struct {
	Name           string
	Path           string
	Scheduler      string
	Size           string
	AccessMode     string
	StorageClass   string
	Namespace      string
	Chown          string
	ReclaimPolicy  string
	Annotations    map[string]string
	Labels         map[string]string
}

// CommandCreate registers a new storage entry.
func CommandCreate(input CommandCreateInput) error {
	if err := ValidateEntryName(input.Name, false); err != nil {
		return err
	}

	scheduler := input.Scheduler
	if scheduler == "" {
		scheduler = SchedulerDockerLocal
	}

	hostPath := input.Path
	if scheduler == SchedulerDockerLocal && hostPath == "" {
		hostPath = filepath.Join(GetStorageDirectory(), input.Name)
	}

	entry := &Entry{
		Name:          input.Name,
		Scheduler:     scheduler,
		HostPath:      hostPath,
		Size:          input.Size,
		AccessMode:    input.AccessMode,
		StorageClass:  input.StorageClass,
		Namespace:     input.Namespace,
		Chown:         input.Chown,
		ReclaimPolicy: input.ReclaimPolicy,
		Annotations:   input.Annotations,
		Labels:        input.Labels,
		SchemaVersion: SchemaVersion,
	}

	if err := entry.Validate(); err != nil {
		return err
	}

	if EntryExists(entry.Name) {
		existing, err := LoadEntry(entry.Name)
		if err != nil {
			return err
		}
		if existing.Scheduler != entry.Scheduler {
			return fmt.Errorf("storage entry %q already exists with scheduler %q; refusing to redefine", entry.Name, existing.Scheduler)
		}
		common.LogInfo1Quiet(fmt.Sprintf("Storage entry %s already exists, leaving in place", entry.Name))
	}

	if scheduler == SchedulerDockerLocal {
		if err := ensureDockerLocalPath(entry); err != nil {
			return err
		}
	}

	if err := SaveEntry(entry); err != nil {
		return err
	}

	if scheduler == SchedulerK3s {
		if err := callSchedulerCreateTrigger(entry); err != nil {
			// Roll back the on-disk entry so the disk and the cluster stay in sync.
			_ = DeleteEntry(entry.Name)
			return fmt.Errorf("scheduler refused storage entry %q: %w", entry.Name, err)
		}
	}

	common.LogInfo1(fmt.Sprintf("Storage entry %s created", entry.Name))
	return nil
}

// CommandDestroy removes a registered storage entry. Refuses to remove
// an entry that any app still has attached.
func CommandDestroy(name string) error {
	if name == "" {
		return errors.New("storage entry name is required")
	}
	if !EntryExists(name) {
		return fmt.Errorf("storage entry %q does not exist", name)
	}

	using, err := AppsUsingEntry(name)
	if err != nil {
		return err
	}
	if len(using) > 0 {
		return fmt.Errorf("storage entry %q is still mounted by app(s): %s", name, strings.Join(using, ", "))
	}

	entry, err := LoadEntry(name)
	if err != nil {
		return err
	}

	if entry.Scheduler == SchedulerK3s {
		if err := callSchedulerDestroyTrigger(entry); err != nil {
			return fmt.Errorf("scheduler refused to remove storage entry %q: %w", name, err)
		}
	}

	if err := DeleteEntry(name); err != nil {
		return err
	}
	common.LogInfo1(fmt.Sprintf("Storage entry %s destroyed", name))
	return nil
}

// CommandInfo prints a single entry's details, optionally as JSON.
func CommandInfo(name string, format string) error {
	if name == "" {
		return errors.New("storage entry name is required")
	}
	if !EntryExists(name) {
		return fmt.Errorf("storage entry %q does not exist", name)
	}

	entry, err := LoadEntry(name)
	if err != nil {
		return err
	}

	if format == "json" {
		data, err := json.MarshalIndent(entry, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	common.LogInfo1Quiet(fmt.Sprintf("Storage entry %s", entry.Name))
	common.LogVerbose(fmt.Sprintf("Scheduler:        %s", entry.Scheduler))
	if entry.HostPath != "" {
		common.LogVerbose(fmt.Sprintf("Host path:        %s", entry.HostPath))
	}
	if entry.Size != "" {
		common.LogVerbose(fmt.Sprintf("Size:             %s", entry.Size))
	}
	if entry.AccessMode != "" {
		common.LogVerbose(fmt.Sprintf("Access mode:      %s", entry.AccessMode))
	}
	if entry.StorageClass != "" {
		common.LogVerbose(fmt.Sprintf("Storage class:    %s", entry.StorageClass))
	}
	if entry.Namespace != "" {
		common.LogVerbose(fmt.Sprintf("Namespace:        %s", entry.Namespace))
	}
	if entry.Chown != "" {
		common.LogVerbose(fmt.Sprintf("Chown:            %s", entry.Chown))
	}
	if entry.ReclaimPolicy != "" {
		common.LogVerbose(fmt.Sprintf("Reclaim policy:   %s", entry.ReclaimPolicy))
	}
	return nil
}

// CommandListEntries lists registered entries (distinct from the legacy
// per-app storage:list which lists colon-form mounts).
func CommandListEntries(scheduler string, format string) error {
	entries, err := ListEntries()
	if err != nil {
		return err
	}

	filtered := []*Entry{}
	for _, entry := range entries {
		if scheduler != "" && entry.Scheduler != scheduler {
			continue
		}
		filtered = append(filtered, entry)
	}

	if format == "json" {
		data, err := json.MarshalIndent(filtered, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	if len(filtered) == 0 {
		common.LogInfo1Quiet("No storage entries registered")
		return nil
	}
	common.LogInfo1Quiet("Storage entries:")
	for _, entry := range filtered {
		common.LogVerbose(fmt.Sprintf("%s\t%s\t%s", entry.Name, entry.Scheduler, entry.HostPath))
	}
	return nil
}

// ensureDockerLocalPath creates the host directory referenced by a
// docker-local entry if it doesn't already exist. Idempotent: a
// pre-existing directory is left in place.
func ensureDockerLocalPath(entry *Entry) error {
	info, err := os.Stat(entry.HostPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("unable to stat %s: %w", entry.HostPath, err)
	}
	if err == nil && !info.IsDir() {
		return fmt.Errorf("storage entry path %s exists but is not a directory", entry.HostPath)
	}
	if err != nil {
		if mkdirErr := os.MkdirAll(entry.HostPath, 0755); mkdirErr != nil {
			return fmt.Errorf("unable to create %s: %w", entry.HostPath, mkdirErr)
		}
		common.LogVerbose(fmt.Sprintf("Created %s", entry.HostPath))
	}

	if entry.Chown != "" && entry.Chown != "false" {
		chownID, err := resolveChownID(entry.Chown)
		if err != nil {
			return err
		}
		if chownID != "false" {
			pluginPath := common.MustGetEnv("PLUGIN_AVAILABLE_PATH")
			chownScript := filepath.Join(pluginPath, "storage", "bin", "chown-storage-dir")
			result, err := common.CallExecCommand(common.ExecCommandInput{
				Command: "sudo",
				Args:    []string{chownScript, entry.HostPath, chownID},
			})
			if err != nil {
				return fmt.Errorf("unable to chown %s: %w", entry.HostPath, err)
			}
			if result.ExitCode != 0 {
				return fmt.Errorf("unable to chown %s: %s", entry.HostPath, result.StderrContents())
			}
		}
	}
	return nil
}

// callSchedulerCreateTrigger asks the scheduler plugin (k3s) to provision
// the underlying PVC/PV. The scheduler is responsible for any
// cluster-level validation (storage class existence, etc.).
func callSchedulerCreateTrigger(entry *Entry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	results, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "storage-create",
		Args:        []string{entry.Name},
		StreamStdio: true,
		Stdin:       strings.NewReader(string(data)),
	})
	if err != nil {
		return err
	}
	if results.ExitCode != 0 {
		return fmt.Errorf("storage-create trigger exited with %d: %s", results.ExitCode, results.StderrContents())
	}
	return nil
}

// callSchedulerDestroyTrigger asks the scheduler plugin to release the
// underlying PVC/PV.
func callSchedulerDestroyTrigger(entry *Entry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	results, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "storage-destroy",
		Args:        []string{entry.Name},
		StreamStdio: true,
		Stdin:       strings.NewReader(string(data)),
	})
	if err != nil {
		return err
	}
	if results.ExitCode != 0 {
		return fmt.Errorf("storage-destroy trigger exited with %d: %s", results.ExitCode, results.StderrContents())
	}
	return nil
}
