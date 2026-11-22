package buildernixpacks

import (
	"github.com/dokku/dokku/plugins/common"
)

// TriggerBackupAppExport exports the app's builder-nixpacks properties into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExport("builder-nixpacks", appName, scopeDir, "")
}

// TriggerBackupAppImport restores the app's builder-nixpacks properties from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return common.BackupPropertyImport("builder-nixpacks", appName, scopeDir)
}

// TriggerBackupGlobalExport exports the global builder-nixpacks properties into the backup scope dir.
func TriggerBackupGlobalExport(scopeDir string) error {
	return common.BackupPropertyExport("builder-nixpacks", "--global", scopeDir, "")
}

// TriggerBackupGlobalImport restores the global builder-nixpacks properties from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return common.BackupPropertyImport("builder-nixpacks", "--global", scopeDir)
}
