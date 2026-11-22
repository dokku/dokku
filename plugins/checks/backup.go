package checks

import (
	"github.com/dokku/dokku/plugins/common"
)

// TriggerBackupAppExport exports the app's checks properties into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExport("checks", appName, scopeDir, "")
}

// TriggerBackupAppImport restores the app's checks properties from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return common.BackupPropertyImport("checks", appName, scopeDir)
}

// TriggerBackupGlobalExport exports the global checks properties into the backup scope dir.
func TriggerBackupGlobalExport(scopeDir string) error {
	return common.BackupPropertyExport("checks", "--global", scopeDir, "")
}

// TriggerBackupGlobalImport restores the global checks properties from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return common.BackupPropertyImport("checks", "--global", scopeDir)
}
