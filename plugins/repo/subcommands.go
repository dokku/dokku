package repo

import (
	"github.com/dokku/dokku/plugins/common"
)

// CommandGc runs 'git gc --aggressive' against the application's repo
func CommandGc(appName string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	appRoot := common.AppRoot(appName)
	cmdEnv := map[string]string{
		"GIT_DIR": appRoot,
	}
	gitGcCmd := common.NewShellCmd("git gc --aggressive")
	gitGcCmd.Env = cmdEnv
	gitGcCmd.Execute()
	return nil
}

// CommandPurgeCache deletes the contents of the build cache stored in the repository
func CommandPurgeCache(appName string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	return PurgeCache(appName)
}
