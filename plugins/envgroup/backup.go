package envgroup

import (
	"fmt"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
)

// groupTaskBody is the dokku_envgroup task body emitted into a global backup
type groupTaskBody struct {
	Group  string            `yaml:"group"`
	Config map[string]string `yaml:"config"`
}

// groupRecipe is the read shape for a global envgroup backup slice
type groupRecipe []struct {
	Tasks []struct {
		Group *groupTaskBody `yaml:"dokku_envgroup"`
	} `yaml:"tasks"`
}

// membershipTaskBody is the dokku_envgroup_app task body emitted into an app backup
type membershipTaskBody struct {
	App    string   `yaml:"app"`
	Groups []string `yaml:"groups"`
}

// membershipRecipe is the read shape for an app envgroup backup slice
type membershipRecipe []struct {
	Tasks []struct {
		Membership *membershipTaskBody `yaml:"dokku_envgroup_app"`
	} `yaml:"tasks"`
}

// TriggerBackupGlobalExport writes every group and its environment into the backup scope dir
func TriggerBackupGlobalExport(scopeDir string) error {
	groups, err := ListGroups()
	if err != nil {
		return err
	}

	if len(groups) == 0 {
		return nil
	}

	tasks := make([]common.BackupTask, 0, len(groups))
	for _, groupName := range groups {
		env, err := LoadGroupEnv(groupName)
		if err != nil {
			return err
		}

		tasks = append(tasks, common.BackupTask{"dokku_envgroup": groupTaskBody{
			Group:  groupName,
			Config: env.Map(),
		}})
	}

	recipe := common.BackupRecipe{{
		Name:  "envgroup",
		Tasks: tasks,
	}}
	return common.BackupWriteRecipe(common.BackupRecipePath(pluginName, scopeDir), recipe)
}

// TriggerBackupGlobalImport recreates every group in the backup scope dir,
// replacing the environment of a group that already exists
func TriggerBackupGlobalImport(scopeDir string) error {
	path := common.BackupRecipePath(pluginName, scopeDir)
	if !common.FileExists(path) {
		return nil
	}

	var recipe groupRecipe
	if err := common.BackupReadRecipeFile(path, &recipe); err != nil {
		return err
	}

	for _, play := range recipe {
		for _, task := range play.Tasks {
			if task.Group == nil {
				continue
			}

			if err := restoreGroup(task.Group.Group, task.Group.Config); err != nil {
				return err
			}
		}
	}

	return nil
}

// TriggerBackupAppExport writes an app's group memberships into the backup scope dir
func TriggerBackupAppExport(appName string, scopeDir string) error {
	groups, err := AppGroups(appName)
	if err != nil {
		return err
	}

	if len(groups) == 0 {
		return nil
	}

	recipe := common.BackupRecipe{{
		Name: "envgroup",
		Tasks: []common.BackupTask{{"dokku_envgroup_app": membershipTaskBody{
			App:    appName,
			Groups: groups,
		}}},
	}}
	return common.BackupWriteRecipe(common.BackupRecipePath(pluginName, scopeDir), recipe)
}

// TriggerBackupAppImport restores an app's group memberships from the backup
// scope dir, skipping any group that does not exist on this host
func TriggerBackupAppImport(appName string, scopeDir string) error {
	path := common.BackupRecipePath(pluginName, scopeDir)
	if !common.FileExists(path) {
		return nil
	}

	var recipe membershipRecipe
	if err := common.BackupReadRecipeFile(path, &recipe); err != nil {
		return err
	}

	for _, play := range recipe {
		for _, task := range play.Tasks {
			if task.Membership == nil {
				continue
			}

			groups := []string{}
			for _, groupName := range task.Membership.Groups {
				if !GroupExists(groupName) {
					common.LogWarn(fmt.Sprintf("Skipping env group %s for app %s, the group does not exist", groupName, appName))
					continue
				}
				if !containsString(groups, groupName) {
					groups = append(groups, groupName)
				}
			}

			if err := common.PropertyListWrite(pluginName, appName, groupsProperty, groups); err != nil {
				return err
			}
		}
	}

	return nil
}

// restoreGroup creates a group if needed and replaces its environment
func restoreGroup(groupName string, entries map[string]string) error {
	if err := validateGroupName(groupName); err != nil {
		return err
	}

	if !GroupExists(groupName) {
		if err := createGroup(groupName); err != nil {
			return fmt.Errorf("Unable to create env group %s: %s", groupName, err.Error())
		}
	}

	_, err := config.UpdateEnvFile(groupName, groupFile(groupName), func(env *config.Env) bool {
		env.Clear()
		for k, v := range entries {
			if config.ValidateKey(k) != nil {
				common.LogWarn(fmt.Sprintf("Skipping invalid key %s for env group %s", k, groupName))
				continue
			}
			env.Set(k, v)
		}
		return true
	})
	return err
}
