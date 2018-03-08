package repo

import (
	"errors"
	"strings"

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

// CommandReport displays a repo report for one or more apps
func CommandReport(appName string, infoFlag string) error {
	if strings.HasPrefix(appName, "--") {
		infoFlag = appName
		appName = ""
	}

	if len(appName) == 0 {
		apps, err := common.DokkuApps()
		if err != nil {
			return err
		}
		for _, appName := range apps {
			if err := ReportSingleApp(appName, infoFlag); err != nil {
				return err
			}
		}
		return nil
	}

	return ReportSingleApp(appName, infoFlag)
}

// CommandSet set or clear a repo property for an app
func CommandSet(appName string, property string, value string) error {
	common.CommandPropertySet("repo", appName, property, value, DefaultProperties)
	return nil
}
