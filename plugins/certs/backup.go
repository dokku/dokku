package certs

import (
	"os"
	"path/filepath"

	"github.com/dokku/dokku/plugins/common"
)

// backupCertsDir returns where an app's TLS files live within a backup scope dir.
func backupCertsDir(scopeDir string) string {
	return filepath.Join(common.BackupDataDir(scopeDir), "certs")
}

// TriggerBackupAppExport exports an app's TLS certificate and key into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	tlsPath := certTLSPath(appName)
	if !directoryHasEntries(tlsPath) {
		return nil
	}

	return common.Copy(tlsPath, backupCertsDir(scopeDir))
}

// TriggerBackupAppImport restores an app's TLS certificate and key from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	source := backupCertsDir(scopeDir)
	if !directoryHasEntries(source) {
		return nil
	}

	tlsPath := certTLSPath(appName)
	if err := common.Copy(source, tlsPath); err != nil {
		return err
	}

	return common.BackupChownTree(tlsPath)
}

// directoryHasEntries reports whether a path is a directory with at least one entry.
func directoryHasEntries(path string) bool {
	entries, err := os.ReadDir(path)
	return err == nil && len(entries) > 0
}
