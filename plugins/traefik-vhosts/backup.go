package traefikvhosts

import (
	"github.com/dokku/dokku/plugins/common"
)

// TriggerBackupAppExport exports the app's traefik properties into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExport("traefik", appName, scopeDir, "")
}

// TriggerBackupAppImport restores the app's traefik properties from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return common.BackupPropertyImport("traefik", appName, scopeDir)
}

// TriggerBackupGlobalExport exports the global traefik properties into the backup scope dir.
func TriggerBackupGlobalExport(scopeDir string) error {
	return common.BackupPropertyExport("traefik", "--global", scopeDir, "")
}

// TriggerBackupGlobalImport restores the global traefik properties from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return common.BackupPropertyImport("traefik", "--global", scopeDir)
}
