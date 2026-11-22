package logs

import (
	"github.com/dokku/dokku/plugins/common"
)

// TriggerBackupAppExport exports the app's logs properties into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExport("logs", appName, scopeDir, "")
}

// TriggerBackupAppImport restores the app's logs properties from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return common.BackupPropertyImport("logs", appName, scopeDir)
}

// TriggerBackupGlobalExport exports the global logs properties into the backup scope dir.
func TriggerBackupGlobalExport(scopeDir string) error {
	return common.BackupPropertyExport("logs", "--global", scopeDir, "")
}

// TriggerBackupGlobalImport restores the global logs properties from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return common.BackupPropertyImport("logs", "--global", scopeDir)
}
