package repo

import (
	"fmt"

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

	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command:     "git",
		Args:        []string{"gc", "--aggressive"},
		Env:         cmdEnv,
		StreamStdio: true,
	})
	if err != nil {
		return fmt.Errorf("Unable to run git gc: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Unable to run git gc: %s", result.StderrContents())
	}

	return nil
}

// CommandPurgeCache deletes the contents of the build cache stored in the repository
func CommandPurgeCache(appName string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	return PurgeCache(appName)
}
