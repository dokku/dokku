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
	// and writes entry / migration-flag files into them.
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

	if err := MigrateLegacyMounts(); err != nil {
		return fmt.Errorf("storage migration failed: %w", err)
	}

	distro := detectDistro()
	if distro == "" {
		return nil
	}

	pluginPath := common.MustGetEnv("PLUGIN_AVAILABLE_PATH")
	chownScript := filepath.Join(pluginPath, "storage", "bin", "chown-storage-dir")

	sudoersFile := "/etc/sudoers.d/dokku-storage"
	content := fmt.Sprintf("%%dokku ALL=(ALL) NOPASSWD:%s *\n", chownScript)
	content += "Defaults env_keep += \"DOKKU_LIB_ROOT\"\n"

	if err := os.WriteFile(sudoersFile, []byte(content), 0440); err != nil {
		return fmt.Errorf("Unable to write sudoers file: %s", err.Error())
	}

	return nil
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
	common.LogWarn("the storage-list plugn trigger is deprecated; use the storage-app-mounts trigger or the storage Go package directly")

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

// TriggerDockerArgs emits `-v` flags for each docker-local attachment in
// the requested phase. Plugn concatenates this with docker-options'
// equivalent trigger output, so docker-local apps continue to receive
// their bind mounts through the standard pipeline.
func TriggerDockerArgs(appName string, phase string) error {
	attachments, err := AttachmentsForPhase(appName, phase)
	if err != nil {
		return err
	}

	for _, attachment := range attachments {
		entry, err := LoadEntry(attachment.EntryName)
		if err != nil {
			return fmt.Errorf("attachment on %q references missing entry %q: %w", appName, attachment.EntryName, err)
		}
		if entry.Scheduler != SchedulerDockerLocal {
			continue
		}
		flag := buildDockerVFlag(entry, attachment)
		if flag == "" {
			continue
		}
		fmt.Printf(" %s", flag)
	}
	return nil
}

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
