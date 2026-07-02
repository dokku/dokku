package builder

import (
	"github.com/dokku/dokku/plugins/common"
)

// TriggerBackupAppExport exports the app's builder properties into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExport("builder", appName, scopeDir, "")
}

// TriggerBackupAppImport restores the app's builder properties from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return common.BackupPropertyImport("builder", appName, scopeDir)
}

// TriggerBackupGlobalExport exports the global builder properties into the backup scope dir.
func TriggerBackupGlobalExport(scopeDir string) error {
	return common.BackupPropertyExport("builder", "--global", scopeDir, "")
}

// TriggerBackupGlobalImport restores the global builder properties from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return common.BackupPropertyImport("builder", "--global", scopeDir)
}
