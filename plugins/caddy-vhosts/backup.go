package caddyvhosts

import (
	"github.com/dokku/dokku/plugins/common"
)

// TriggerBackupAppExport exports the app's caddy properties into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExport("caddy", appName, scopeDir, "")
}

// TriggerBackupAppImport restores the app's caddy properties from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return common.BackupPropertyImport("caddy", appName, scopeDir)
}

// TriggerBackupGlobalExport exports the global caddy properties into the backup scope dir.
func TriggerBackupGlobalExport(scopeDir string) error {
	return common.BackupPropertyExport("caddy", "--global", scopeDir, "")
}

// TriggerBackupGlobalImport restores the global caddy properties from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return common.BackupPropertyImport("caddy", "--global", scopeDir)
}
