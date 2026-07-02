package builds

import (
	"github.com/dokku/dokku/plugins/common"
)

// TriggerBackupAppExport exports the app's builds properties into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExport("builds", appName, scopeDir, "")
}

// TriggerBackupAppImport restores the app's builds properties from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return common.BackupPropertyImport("builds", appName, scopeDir)
}

// TriggerBackupGlobalExport exports the global builds properties into the backup scope dir.
func TriggerBackupGlobalExport(scopeDir string) error {
	return common.BackupPropertyExport("builds", "--global", scopeDir, "")
}

// TriggerBackupGlobalImport restores the global builds properties from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return common.BackupPropertyImport("builds", "--global", scopeDir)
}
