package builderherokuish

import (
	"github.com/dokku/dokku/plugins/common"
)

// TriggerBackupAppExport exports the app's builder-herokuish properties into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExport("builder-herokuish", appName, scopeDir, "")
}

// TriggerBackupAppImport restores the app's builder-herokuish properties from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return common.BackupPropertyImport("builder-herokuish", appName, scopeDir)
}

// TriggerBackupGlobalExport exports the global builder-herokuish properties into the backup scope dir.
func TriggerBackupGlobalExport(scopeDir string) error {
	return common.BackupPropertyExport("builder-herokuish", "--global", scopeDir, "")
}

// TriggerBackupGlobalImport restores the global builder-herokuish properties from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return common.BackupPropertyImport("builder-herokuish", "--global", scopeDir)
}
