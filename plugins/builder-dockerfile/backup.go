package builderdockerfile

import (
	"github.com/dokku/dokku/plugins/common"
)

// TriggerBackupAppExport exports the app's builder-dockerfile properties into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExport("builder-dockerfile", appName, scopeDir, "")
}

// TriggerBackupAppImport restores the app's builder-dockerfile properties from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return common.BackupPropertyImport("builder-dockerfile", appName, scopeDir)
}

// TriggerBackupGlobalExport exports the global builder-dockerfile properties into the backup scope dir.
func TriggerBackupGlobalExport(scopeDir string) error {
	return common.BackupPropertyExport("builder-dockerfile", "--global", scopeDir, "")
}

// TriggerBackupGlobalImport restores the global builder-dockerfile properties from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return common.BackupPropertyImport("builder-dockerfile", "--global", scopeDir)
}
