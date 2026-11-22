package openrestyvhosts

import (
	"github.com/dokku/dokku/plugins/common"
)

// TriggerBackupAppExport exports the app's openresty properties into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExport("openresty", appName, scopeDir, "")
}

// TriggerBackupAppImport restores the app's openresty properties from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return common.BackupPropertyImport("openresty", appName, scopeDir)
}

// TriggerBackupGlobalExport exports the global openresty properties into the backup scope dir.
func TriggerBackupGlobalExport(scopeDir string) error {
	return common.BackupPropertyExport("openresty", "--global", scopeDir, "")
}

// TriggerBackupGlobalImport restores the global openresty properties from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return common.BackupPropertyImport("openresty", "--global", scopeDir)
}
