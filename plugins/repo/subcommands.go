package repo

import (
	"errors"

	"github.com/dokku/dokku/plugins/common"
)

// CommandGc runs 'git gc --aggressive' against the application's repo
func CommandGc(appName string) error {
	if appName == "" {
		return errors.New("Please specify an app to run the command on")
	}
	err := common.VerifyAppName(appName)
	if err != nil {
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
	if appName == "" {
		common.LogFail("Please specify an app to run the command on")
	}

	return PurgeCache(appName)
}
