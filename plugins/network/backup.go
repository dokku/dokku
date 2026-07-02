package network

import (
	"github.com/dokku/dokku/plugins/common"
)

// TriggerBackupAppExport exports the app's network properties into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExport("network", appName, scopeDir, "")
}

// TriggerBackupAppImport restores the app's network properties from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return common.BackupPropertyImport("network", appName, scopeDir)
}

// TriggerBackupGlobalExport exports the global network properties into the backup scope dir.
func TriggerBackupGlobalExport(scopeDir string) error {
	return common.BackupPropertyExport("network", "--global", scopeDir, "")
}

// TriggerBackupGlobalImport restores the global network properties from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return common.BackupPropertyImport("network", "--global", scopeDir)
}
