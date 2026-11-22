package apps

import (
	"github.com/dokku/dokku/plugins/common"
)

// TriggerBackupAppExport exports the app's apps properties (deploy source,
// created-at) into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExport("apps", appName, scopeDir, "")
}

// TriggerBackupAppImport restores the app's apps properties from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return common.BackupPropertyImport("apps", appName, scopeDir)
}

// TriggerBackupGlobalExport exports the global apps properties (for example
// disable-autocreation) into the backup scope dir.
func TriggerBackupGlobalExport(scopeDir string) error {
	return common.BackupPropertyExport("apps", "--global", scopeDir, "")
}

// TriggerBackupGlobalImport restores the global apps properties from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return common.BackupPropertyImport("apps", "--global", scopeDir)
}
