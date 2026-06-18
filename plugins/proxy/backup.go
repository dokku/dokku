package proxy

import (
	"github.com/dokku/dokku/plugins/common"
)

// TriggerBackupAppExport exports the app's proxy properties into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExport("proxy", appName, scopeDir, "")
}

// TriggerBackupAppImport restores the app's proxy properties from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return common.BackupPropertyImport("proxy", appName, scopeDir)
}

// TriggerBackupGlobalExport exports the global proxy properties into the backup scope dir.
func TriggerBackupGlobalExport(scopeDir string) error {
	return common.BackupPropertyExport("proxy", "--global", scopeDir, "")
}

// TriggerBackupGlobalImport restores the global proxy properties from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return common.BackupPropertyImport("proxy", "--global", scopeDir)
}
