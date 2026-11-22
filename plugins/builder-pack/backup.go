package builderpack

import (
	"github.com/dokku/dokku/plugins/common"
)

// backupExcludedProperties are transient properties that must not be restored.
var backupExcludedProperties = []string{"clear-cache"}

// TriggerBackupAppExport exports the app's builder-pack properties into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExportFiltered("builder-pack", appName, scopeDir, "", backupExcludedProperties)
}

// TriggerBackupAppImport restores the app's builder-pack properties from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return common.BackupPropertyImport("builder-pack", appName, scopeDir)
}

// TriggerBackupGlobalExport exports the global builder-pack properties into the backup scope dir.
func TriggerBackupGlobalExport(scopeDir string) error {
	return common.BackupPropertyExportFiltered("builder-pack", "--global", scopeDir, "", backupExcludedProperties)
}

// TriggerBackupGlobalImport restores the global builder-pack properties from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return common.BackupPropertyImport("builder-pack", "--global", scopeDir)
}
