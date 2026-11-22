package ports

import (
	"github.com/dokku/dokku/plugins/common"
)

// portsBackupExcludes lists derived properties that must not be restored.
var portsBackupExcludes = []string{"map-detected"}

// TriggerBackupAppExport exports the app's ports properties into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExportFiltered("ports", appName, scopeDir, "", portsBackupExcludes)
}

// TriggerBackupAppImport restores the app's ports properties from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return common.BackupPropertyImport("ports", appName, scopeDir)
}
