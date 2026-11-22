package scheduler_k3s

import (
	"github.com/dokku/dokku/plugins/common"
)

// TriggerBackupAppExport exports the app's scheduler-k3s properties into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExport("scheduler-k3s", appName, scopeDir, "")
}

// TriggerBackupAppImport restores the app's scheduler-k3s properties from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return common.BackupPropertyImport("scheduler-k3s", appName, scopeDir)
}

// TriggerBackupGlobalExport exports the global scheduler-k3s properties into the backup scope dir.
func TriggerBackupGlobalExport(scopeDir string) error {
	return common.BackupPropertyExport("scheduler-k3s", "--global", scopeDir, "")
}

// TriggerBackupGlobalImport restores the global scheduler-k3s properties from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return common.BackupPropertyImport("scheduler-k3s", "--global", scopeDir)
}
