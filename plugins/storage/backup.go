package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// managedStorageRoot returns the root under which dokku-managed storage
// directories live. Only entry host paths under this root are bundled when
// --include-storage is requested.
func managedStorageRoot() string {
	return filepath.Join(common.GetenvWithDefault("DOKKU_LIB_ROOT", "/var/lib/dokku"), "data", "storage")
}

// writeEntryFile serializes a storage entry into a backup directory.
func writeEntryFile(dir string, entry *Entry) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, entry.Name+".json"), data, 0600)
}

// restoreEntryFiles reads entry JSON files from a backup directory and saves any
// that do not already exist, so mounts that reference them resolve on import.
func restoreEntryFiles(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	for _, file := range entries {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return err
		}
		var entry Entry
		if err := json.Unmarshal(data, &entry); err != nil {
			return err
		}
		if _, err := LoadEntry(entry.Name); err == nil {
			continue
		}
		if err := SaveEntry(&entry); err != nil {
			return err
		}
	}
	return nil
}

// TriggerBackupAppExport exports an app's storage mounts and the entries they
// reference. When DOKKU_BACKUP_INCLUDE_STORAGE is set, the contents of each
// dokku-managed entry directory are bundled as well.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	if err := common.BackupPropertyExport("storage", appName, scopeDir, ""); err != nil {
		return err
	}

	attachments, err := LoadAttachments(appName)
	if err != nil || len(attachments) == 0 {
		return err
	}

	includeData := os.Getenv("DOKKU_BACKUP_INCLUDE_STORAGE") == "1"
	root := managedStorageRoot()
	entriesDir := filepath.Join(scopeDir, "data", "storage-entries")
	dataDir := filepath.Join(scopeDir, "data", "storage-data")
	seen := map[string]bool{}

	for _, attachment := range attachments {
		if seen[attachment.EntryName] {
			continue
		}
		seen[attachment.EntryName] = true

		entry, err := LoadEntry(attachment.EntryName)
		if err != nil {
			continue
		}
		if err := writeEntryFile(entriesDir, entry); err != nil {
			return err
		}

		if !includeData || entry.HostPath == "" {
			continue
		}
		if entry.HostPath != root && !strings.HasPrefix(entry.HostPath, root+string(os.PathSeparator)) {
			common.LogWarn(fmt.Sprintf("skipping non-managed storage path for %s: %s", appName, entry.HostPath))
			continue
		}
		if !common.DirectoryExists(entry.HostPath) {
			continue
		}
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return err
		}
		archive := filepath.Join(dataDir, entry.Name+".tar.gz")
		if _, err := common.CallExecCommand(common.ExecCommandInput{
			Command: "tar",
			Args:    []string{"-czf", archive, "-C", entry.HostPath, "."},
		}); err != nil {
			return err
		}
	}
	return nil
}

// TriggerBackupAppImport restores an app's storage mounts, recreates any
// referenced entries that are missing, and extracts bundled entry data.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	if err := common.BackupPropertyImport("storage", appName, scopeDir); err != nil {
		return err
	}

	entriesDir := filepath.Join(scopeDir, "data", "storage-entries")
	if err := restoreEntryFiles(entriesDir); err != nil {
		return err
	}

	dataDir := filepath.Join(scopeDir, "data", "storage-data")
	if !common.DirectoryExists(dataDir) {
		return nil
	}

	attachments, err := LoadAttachments(appName)
	if err != nil {
		return err
	}
	for _, attachment := range attachments {
		archive := filepath.Join(dataDir, attachment.EntryName+".tar.gz")
		if !common.FileExists(archive) {
			continue
		}
		entry, err := LoadEntry(attachment.EntryName)
		if err != nil || entry.HostPath == "" {
			continue
		}
		if err := os.MkdirAll(entry.HostPath, 0755); err != nil {
			return err
		}
		if _, err := common.CallExecCommand(common.ExecCommandInput{
			Command: "tar",
			Args:    []string{"-xzf", archive, "-C", entry.HostPath},
		}); err != nil {
			return err
		}
		if err := common.SetPermissions(common.SetPermissionInput{Filename: entry.HostPath, Mode: 0755}); err != nil {
			return err
		}
	}
	return nil
}

// TriggerBackupGlobalExport exports every storage entry into the global backup
// scope, so a full backup can recreate the entire storage registry.
func TriggerBackupGlobalExport(scopeDir string) error {
	entries, err := ListEntries()
	if err != nil || len(entries) == 0 {
		return err
	}
	entriesDir := filepath.Join(scopeDir, "data", "storage-entries")
	for _, entry := range entries {
		if err := writeEntryFile(entriesDir, entry); err != nil {
			return err
		}
	}
	return nil
}

// TriggerBackupGlobalImport restores storage entries from the global backup scope.
func TriggerBackupGlobalImport(scopeDir string) error {
	return restoreEntryFiles(filepath.Join(scopeDir, "data", "storage-entries"))
}
