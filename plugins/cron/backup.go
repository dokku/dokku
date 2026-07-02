package cron

import (
	"github.com/dokku/dokku/plugins/common"
)

// TriggerBackupAppExport exports the app's cron properties into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExport("cron", appName, scopeDir, "")
}

// TriggerBackupAppImport restores the app's cron properties from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return common.BackupPropertyImport("cron", appName, scopeDir)
}

// TriggerBackupGlobalExport exports the global cron properties into the backup scope dir.
func TriggerBackupGlobalExport(scopeDir string) error {
	return common.BackupPropertyExport("cron", "--global", scopeDir, "")
}

// TriggerBackupGlobalImport restores the global cron properties from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return common.BackupPropertyImport("cron", "--global", scopeDir)
}
