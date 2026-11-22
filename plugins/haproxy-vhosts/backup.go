package haproxyvhosts

import (
	"github.com/dokku/dokku/plugins/common"
)

// TriggerBackupAppExport exports the app's haproxy properties into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExport("haproxy", appName, scopeDir, "")
}

// TriggerBackupAppImport restores the app's haproxy properties from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return common.BackupPropertyImport("haproxy", appName, scopeDir)
}

// TriggerBackupGlobalExport exports the global haproxy properties into the backup scope dir.
func TriggerBackupGlobalExport(scopeDir string) error {
	return common.BackupPropertyExport("haproxy", "--global", scopeDir, "")
}

// TriggerBackupGlobalImport restores the global haproxy properties from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return common.BackupPropertyImport("haproxy", "--global", scopeDir)
}
