package appjson

import (
	"github.com/dokku/dokku/plugins/common"
)

// TriggerBackupAppExport exports the app's app-json properties into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExport("app-json", appName, scopeDir, "")
}

// TriggerBackupAppImport restores the app's app-json properties from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return common.BackupPropertyImport("app-json", appName, scopeDir)
}

// TriggerBackupGlobalExport exports the global app-json properties into the backup scope dir.
func TriggerBackupGlobalExport(scopeDir string) error {
	return common.BackupPropertyExport("app-json", "--global", scopeDir, "")
}

// TriggerBackupGlobalImport restores the global app-json properties from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return common.BackupPropertyImport("app-json", "--global", scopeDir)
}
