package builderlambda

import (
	"github.com/dokku/dokku/plugins/common"
)

// TriggerBackupAppExport exports the app's builder-lambda properties into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExport("builder-lambda", appName, scopeDir, "")
}

// TriggerBackupAppImport restores the app's builder-lambda properties from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return common.BackupPropertyImport("builder-lambda", appName, scopeDir)
}

// TriggerBackupGlobalExport exports the global builder-lambda properties into the backup scope dir.
func TriggerBackupGlobalExport(scopeDir string) error {
	return common.BackupPropertyExport("builder-lambda", "--global", scopeDir, "")
}

// TriggerBackupGlobalImport restores the global builder-lambda properties from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return common.BackupPropertyImport("builder-lambda", "--global", scopeDir)
}
