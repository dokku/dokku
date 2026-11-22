package nginxvhosts

import (
	"github.com/dokku/dokku/plugins/common"
)

// TriggerBackupAppExport exports the app's nginx properties into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExport("nginx", appName, scopeDir, "")
}

// TriggerBackupAppImport restores the app's nginx properties from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return common.BackupPropertyImport("nginx", appName, scopeDir)
}

// TriggerBackupGlobalExport exports the global nginx properties into the backup scope dir.
func TriggerBackupGlobalExport(scopeDir string) error {
	return common.BackupPropertyExport("nginx", "--global", scopeDir, "")
}

// TriggerBackupGlobalImport restores the global nginx properties from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return common.BackupPropertyImport("nginx", "--global", scopeDir)
}
