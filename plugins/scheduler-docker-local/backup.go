package schedulerdockerlocal

import (
	"github.com/dokku/dokku/plugins/common"
)

// TriggerBackupAppExport exports the app's scheduler-docker-local properties into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExport("scheduler-docker-local", appName, scopeDir, "")
}

// TriggerBackupAppImport restores the app's scheduler-docker-local properties from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return common.BackupPropertyImport("scheduler-docker-local", appName, scopeDir)
}

// TriggerBackupGlobalExport exports the global scheduler-docker-local properties into the backup scope dir.
func TriggerBackupGlobalExport(scopeDir string) error {
	return common.BackupPropertyExport("scheduler-docker-local", "--global", scopeDir, "")
}

// TriggerBackupGlobalImport restores the global scheduler-docker-local properties from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return common.BackupPropertyImport("scheduler-docker-local", "--global", scopeDir)
}
