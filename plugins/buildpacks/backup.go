package buildpacks

import (
	"github.com/dokku/dokku/plugins/common"
)

// TriggerBackupAppExport exports the app's buildpacks properties into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExport("buildpacks", appName, scopeDir, "")
}

// TriggerBackupAppImport restores the app's buildpacks properties from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return common.BackupPropertyImport("buildpacks", appName, scopeDir)
}

// TriggerBackupGlobalExport exports the global buildpacks properties into the backup scope dir.
func TriggerBackupGlobalExport(scopeDir string) error {
	return common.BackupPropertyExport("buildpacks", "--global", scopeDir, "")
}

// TriggerBackupGlobalImport restores the global buildpacks properties from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return common.BackupPropertyImport("buildpacks", "--global", scopeDir)
}
