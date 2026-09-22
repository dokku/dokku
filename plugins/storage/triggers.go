package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// TriggerInstall sets up the storage plugin on installation and runs
// the bulk legacy-mount migration once per upgrade.
func TriggerInstall() error {
	if err := common.PropertySetup(PluginName); err != nil {
		return fmt.Errorf("Unable to install the storage plugin: %s", err.Error())
	}

	storageDir := GetStorageDirectory()

	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return fmt.Errorf("Unable to create storage directory: %s", err.Error())
	}

	if err := common.SetPermissions(common.SetPermissionInput{
		Filename: storageDir,
		Mode:     0755,
	}); err != nil {
		return fmt.Errorf("Unable to set storage directory permissions: %s", err.Error())
	}

	if err := EnsureEntriesDirectory(); err != nil {
		return err
	}

	// The install trigger runs as root; the directories it just made are
	// root-owned, but the dokku user is the one that runs storage:create
	// and writes entry files into them. migrationFlagDir() stays in the
	// list for one release cycle so the upgrade-cycle conversion in
	// MigrateLegacyMounts can read pre-existing flag files.
	// TODO(post-deprecation): drop migrationFlagDir() from this list.
	for _, dir := range []string{RegistryDirectory(), EntriesDirectory(), migrationFlagDir()} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("Unable to create %s: %s", dir, err.Error())
		}
		if err := common.SetPermissions(common.SetPermissionInput{
			Filename: dir,
			Mode:     0755,
		}); err != nil {
			return fmt.Errorf("Unable to set permissions on %s: %s", dir, err.Error())
		}
	}

	if err := repairRegistryOwnership(); err != nil {
		return fmt.Errorf("Unable to repair storage registry ownership: %s", err.Error())
	}

	if err := MigrateLegacyMounts(); err != nil {
		return fmt.Errorf("storage migration failed: %w", err)
	}

	distro := detectDistro()
	if distro == "" {
		return nil
	}

	sudoersFile := "/etc/sudoers.d/dokku-storage"
	content := ""
	for _, script := range StorageDirScripts() {
		content += fmt.Sprintf("%%dokku ALL=(ALL) NOPASSWD:%s *\n", script)
	}
	content += "Defaults env_keep += \"DOKKU_LIB_ROOT\"\n"

	if err := os.WriteFile(sudoersFile, []byte(content), 0440); err != nil {
		return fmt.Errorf("Unable to write sudoers file: %s", err.Error())
	}

	return nil
}

// repairRegistryOwnership rewrites ownership on every regular file under
// the registry tree so the dokku user can read entry / migration-flag
// files. Earlier 0.38 builds wrote those files as root because SaveEntry
// and touchMigrationFlag bypassed the permission helpers; on an upgrade
// the per-app migration flag short-circuits MigrateLegacyMounts so a
// plain SaveEntry fix would leave the broken files in place. The walk is
// idempotent: calling SetPermissions on a file already owned by the
// target user produces a no-op chown.
func repairRegistryOwnership() error {
	return filepath.Walk(RegistryDirectory(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.IsDir() {
			return nil
		}
		return common.SetPermissions(common.SetPermissionInput{
			Filename: path,
			Mode:     info.Mode().Perm(),
		})
	})
}

// TriggerPostDelete removes the attachment store for an app that's being
// destroyed. Entries are global and survive app deletion. The whole
// per-app property folder goes away rather than being rewritten as
// empty, matching what other plugins do on app deletion.
func TriggerPostDelete(appName string) error {
	if appName == "" {
		return nil
	}
	return common.PropertyDestroy(PluginName, appName)
}

// TriggerPostAppCloneSetup copies attachments from the source app to the
// cloned app. Entries are global so no entry-level work is needed.
func TriggerPostAppCloneSetup(oldName string, newName string) error {
	if oldName == "" || newName == "" {
		return nil
	}
	if !common.PropertyExists(PluginName, oldName, AttachmentsProperty) {
		return nil
	}
	attachments, err := LoadAttachments(oldName)
	if err != nil {
		return err
	}
	return SaveAttachments(newName, attachments)
}

// TriggerPostAppRenameSetup moves attachments from the old name to the
// new name and removes the old per-app property folder.
func TriggerPostAppRenameSetup(oldName string, newName string) error {
	if oldName == "" || newName == "" {
		return nil
	}
	if !common.PropertyExists(PluginName, oldName, AttachmentsProperty) {
		return nil
	}
	attachments, err := LoadAttachments(oldName)
	if err != nil {
		return err
	}
	if err := SaveAttachments(newName, attachments); err != nil {
		return err
	}
	return common.PropertyDestroy(PluginName, oldName)
}

// detectDistro returns the Linux distribution name
func detectDistro() string {
	if runtime.GOOS != "linux" {
		return ""
	}

	distro := os.Getenv("DOKKU_DISTRO")
	if distro != "" {
		return distro
	}

	if common.FileExists("/etc/debian_version") {
		return "debian"
	}
	if common.FileExists("/etc/arch-release") {
		return "arch"
	}

	return ""
}

// TriggerStorageList outputs storage mounts for an app.
//
// Deprecated: the storage-list plugn trigger is retained for back-compat
// with external plugins that may still call it, but in-process callers
// should use storage.ListAppMountEntries directly. A deprecation warning
// is emitted on every invocation.
func TriggerStorageList(appName string, phase string, format string) error {
	common.LogWarn("Deprecated: please use the 'storage-app-mounts' plugn trigger or the storage Go package directly instead of 'storage-list'")

	rows, err := ListAppMountEntries(appName, phase)
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
	for _, row := range rows {
		fmt.Println(formatStorageListEntry(row))
	}
	return nil
}

// AppMountPair pairs an Attachment with the Entry it references; consumed
// by scheduler plugins via the storage-app-mounts trigger.
type AppMountPair struct {
	Entry      *Entry      `json:"entry"`
	Attachment *Attachment `json:"attachment"`
}

// TriggerStorageAppMounts emits the (entry, attachment) pairs an app has
// for the given phase, in JSON. Schedulers consume this at deploy time.
func TriggerStorageAppMounts(appName string, phase string) error {
	if phase == "" {
		phase = PhaseDeploy
	}
	attachments, err := AttachmentsForPhase(appName, phase)
	if err != nil {
		return err
	}

	pairs := []AppMountPair{}
	for _, attachment := range attachments {
		entry, err := LoadEntry(attachment.EntryName)
		if err != nil {
			return fmt.Errorf("attachment on %q references missing entry %q: %w", appName, attachment.EntryName, err)
		}
		pairs = append(pairs, AppMountPair{Entry: entry, Attachment: attachment})
	}

	output, err := json.Marshal(pairs)
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

// TriggerDockerArgs emits `-v` flags for the docker-local attachments in the
// requested phase that apply to the given process type. Plugn concatenates
// this with docker-options' equivalent trigger output, so docker-local apps
// continue to receive their bind mounts through the standard pipeline.
//
// An empty processType names a container that belongs to no process - the
// app.json deploy-task path invokes the trigger without one - and so receives
// the default scope. Unlike docker-options, which emits its default scope from
// a separate docker-args-deploy handler, this is storage's only deploy-time
// emitter: returning early on an empty process type would drop every mount.
func TriggerDockerArgs(appName string, phase string, processType string) error {
	flags, err := dockerVFlagsForProcess(appName, phase, processType)
	if err != nil {
		return err
	}

	for _, flag := range flags {
		fmt.Printf(" %s", flag)
	}
	return nil
}

// dockerVFlagsForProcess returns the `-v` flags a single process type's
// container receives for the given phase.
//
// Every attachment in the phase is resolved and scheduler-checked, not just
// the ones in scope, so an entry created for the wrong scheduler still fails
// the deploy no matter which process type happens to be starting.
//
// Apps on another scheduler emit nothing rather than failing: the docker-args
// triggers are invoked for every app, whatever its scheduler, because the
// app.json deploy-task path builds its ephemeral container with docker
// directly. A k3s app's volumes are PersistentVolumeClaims mounted by its own
// scheduler and there is nothing to bind-mount here.
func dockerVFlagsForProcess(appName string, phase string, processType string) ([]string, error) {
	if appSchedulerFor(appName) != SchedulerDockerLocal {
		return nil, nil
	}

	attachments, err := AttachmentsForPhase(appName, phase)
	if err != nil {
		return nil, err
	}

	flags := []string{}
	for _, attachment := range attachments {
		entry, err := LoadEntry(attachment.EntryName)
		if err != nil {
			return nil, fmt.Errorf("attachment on %q references missing entry %q: %w", appName, attachment.EntryName, err)
		}
		if entry.Scheduler != SchedulerDockerLocal {
			return nil, fmt.Errorf("storage entry %q is scheduler=%s but is mounted on a docker-local app; recreate it with --scheduler docker-local", entry.Name, entry.Scheduler)
		}
		if !attachment.AppliesToProcessType(processType) {
			continue
		}
		flag := buildDockerVFlag(entry, attachment)
		if flag == "" {
			continue
		}
		flags = append(flags, flag)
	}
	return flags, nil
}

// appSchedulerFor resolves the scheduler an app deploys with. It is a variable
// so tests can exercise the non-docker-local path without a plugn install.
var appSchedulerFor = common.GetAppScheduler

// buildDockerVFlag formats the Docker -v argument for a docker-local
// attachment.
func buildDockerVFlag(entry *Entry, attachment *Attachment) string {
	host := entry.HostPath
	container := attachment.ContainerPath
	if host == "" || container == "" {
		return ""
	}
	options := []string{}
	if attachment.Readonly {
		options = append(options, "ro")
	}
	if attachment.VolumeOptions != "" {
		options = append(options, attachment.VolumeOptions)
	}
	flag := fmt.Sprintf("-v %s:%s", host, container)
	if len(options) > 0 {
		flag = fmt.Sprintf("%s:%s", flag, strings.Join(options, ","))
	}
	return flag
}
