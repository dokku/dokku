package resource

import (
	"github.com/dokku/dokku/plugins/common"
)

// TriggerBackupAppExport exports the app's resource properties into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExport("resource", appName, scopeDir, "")
}

// TriggerBackupAppImport restores the app's resource properties from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return common.BackupPropertyImport("resource", appName, scopeDir)
}

// TriggerBackupGlobalExport exports the global resource properties into the backup scope dir.
func TriggerBackupGlobalExport(scopeDir string) error {
	return common.BackupPropertyExport("resource", "--global", scopeDir, "")
}

// TriggerBackupGlobalImport restores the global resource properties from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return common.BackupPropertyImport("resource", "--global", scopeDir)
}
