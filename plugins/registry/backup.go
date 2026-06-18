package registry

import (
	"github.com/dokku/dokku/plugins/common"
)

// TriggerBackupAppExport exports the app's registry properties into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExport("registry", appName, scopeDir, "")
}

// TriggerBackupAppImport restores the app's registry properties from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return common.BackupPropertyImport("registry", appName, scopeDir)
}

// TriggerBackupGlobalExport exports the global registry properties into the backup scope dir.
func TriggerBackupGlobalExport(scopeDir string) error {
	return common.BackupPropertyExport("registry", "--global", scopeDir, "")
}

// TriggerBackupGlobalImport restores the global registry properties from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return common.BackupPropertyImport("registry", "--global", scopeDir)
}
