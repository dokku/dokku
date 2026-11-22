package buildpacks

import (
	"github.com/dokku/dokku/plugins/common"
)

// TriggerBackupAppExport exports the app's buildpacks properties into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return common.BackupPropertyExport("buildpacks", appName, scopeDir, "")
}

// TriggerBackupAppImport restores the app's buildpacks properties from the
// backup scope dir. A stack value from an archive taken before stacks moved to
// the builder plugins is migrated to the matching builder.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	if err := common.BackupPropertyImport("buildpacks", appName, scopeDir); err != nil {
		return err
	}
	return migrateStackPropertyForApp(appName)
}

// TriggerBackupGlobalExport exports the global buildpacks properties into the backup scope dir.
func TriggerBackupGlobalExport(scopeDir string) error {
	return common.BackupPropertyExport("buildpacks", "--global", scopeDir, "")
}

// TriggerBackupGlobalImport restores the global buildpacks properties from the
// backup scope dir, migrating a legacy stack value to the matching builder.
func TriggerBackupGlobalImport(scopeDir string) error {
	if err := common.BackupPropertyImport("buildpacks", "--global", scopeDir); err != nil {
		return err
	}
	return migrateStackPropertyForApp("--global")
}
