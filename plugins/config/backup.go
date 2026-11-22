package config

import (
	"github.com/dokku/dokku/plugins/common"
)

// configTaskBody is the docket dokku_config task body emitted into a backup.
type configTaskBody struct {
	App     string            `yaml:"app,omitempty"`
	Global  bool              `yaml:"global,omitempty"`
	Restart bool              `yaml:"restart"`
	Config  map[string]string `yaml:"config"`
}

// configRecipe is the read shape for a config backup slice.
type configRecipe []struct {
	Tasks []struct {
		Config *configTaskBody `yaml:"dokku_config"`
	} `yaml:"tasks"`
}

// backupExport writes the app or global environment as a dokku_config slice.
func backupExport(appName string, scopeDir string, global bool) error {
	var env *Env
	var err error
	if global {
		env, err = LoadGlobalEnv()
	} else {
		env, err = LoadAppEnv(appName)
	}
	if err != nil {
		return err
	}

	values := env.Map()
	if len(values) == 0 {
		return nil
	}

	body := configTaskBody{Restart: false, Config: values}
	if global {
		body.Global = true
	} else {
		body.App = appName
	}

	recipe := common.BackupRecipe{{
		Name:  "config",
		Tasks: []common.BackupTask{{"dokku_config": body}},
	}}
	return common.BackupWriteRecipe(common.BackupRecipePath("config", scopeDir), recipe)
}

// backupImport reapplies a config slice via SetMany, replacing existing values
// so the reapply is idempotent.
func backupImport(appName string, scopeDir string, global bool) error {
	path := common.BackupRecipePath("config", scopeDir)
	if !common.FileExists(path) {
		return nil
	}

	var recipe configRecipe
	if err := common.BackupReadRecipeFile(path, &recipe); err != nil {
		return err
	}

	target := appName
	if global {
		target = "--global"
	}

	for _, play := range recipe {
		for _, task := range play.Tasks {
			if task.Config == nil {
				continue
			}
			if err := SetMany(target, task.Config.Config, true, false); err != nil {
				return err
			}
		}
	}
	return nil
}

// TriggerBackupAppExport exports an app's environment into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return backupExport(appName, scopeDir, false)
}

// TriggerBackupAppImport restores an app's environment from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return backupImport(appName, scopeDir, false)
}

// TriggerBackupGlobalExport exports the global environment into the backup scope dir.
func TriggerBackupGlobalExport(scopeDir string) error {
	return backupExport("", scopeDir, true)
}

// TriggerBackupGlobalImport restores the global environment from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return backupImport("", scopeDir, true)
}
