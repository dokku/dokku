package ps

import (
	"github.com/dokku/dokku/plugins/common"
)

// psBackupExcludes lists transient properties that must not be restored.
var psBackupExcludes = []string{"scale.old"}

// TriggerBackupAppExport exports the app's ps properties into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExportFiltered("ps", appName, scopeDir, "", psBackupExcludes)
}

// TriggerBackupAppImport restores the app's ps properties from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return common.BackupPropertyImport("ps", appName, scopeDir)
}

// TriggerBackupGlobalExport exports the global ps properties into the backup scope dir.
func TriggerBackupGlobalExport(scopeDir string) error {
	return common.BackupPropertyExportFiltered("ps", "--global", scopeDir, "", psBackupExcludes)
}

// TriggerBackupGlobalImport restores the global ps properties from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return common.BackupPropertyImport("ps", "--global", scopeDir)
}
