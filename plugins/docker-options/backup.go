package dockeroptions

import (
	"github.com/dokku/dokku/plugins/common"
)

// TriggerBackupAppExport exports the app's docker-options into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExport("docker-options", appName, scopeDir, "")
}

// TriggerBackupAppImport restores the app's docker-options from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return common.BackupPropertyImport("docker-options", appName, scopeDir)
}
