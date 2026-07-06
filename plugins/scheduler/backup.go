package scheduler

import (
	"github.com/dokku/dokku/plugins/common"
)

// TriggerBackupAppExport exports the app's scheduler properties into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExport("scheduler", appName, scopeDir, "")
}

// TriggerBackupAppImport restores the app's scheduler properties from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return common.BackupPropertyImport("scheduler", appName, scopeDir)
}

// TriggerBackupGlobalExport exports the global scheduler properties into the backup scope dir.
func TriggerBackupGlobalExport(scopeDir string) error {
	return common.BackupPropertyExport("scheduler", "--global", scopeDir, "")
}

// TriggerBackupGlobalImport restores the global scheduler properties from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return common.BackupPropertyImport("scheduler", "--global", scopeDir)
}
