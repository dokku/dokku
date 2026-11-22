package builderrailpack

import (
	"github.com/dokku/dokku/plugins/common"
)

// TriggerBackupAppExport exports the app's builder-railpack properties into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExport("builder-railpack", appName, scopeDir, "")
}

// TriggerBackupAppImport restores the app's builder-railpack properties from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return common.BackupPropertyImport("builder-railpack", appName, scopeDir)
}

// TriggerBackupGlobalExport exports the global builder-railpack properties into the backup scope dir.
func TriggerBackupGlobalExport(scopeDir string) error {
	return common.BackupPropertyExport("builder-railpack", "--global", scopeDir, "")
}

// TriggerBackupGlobalImport restores the global builder-railpack properties from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return common.BackupPropertyImport("builder-railpack", "--global", scopeDir)
}
