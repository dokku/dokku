package repo

import (
	"github.com/dokku/dokku/plugins/common"
)

// CommandGc runs 'git gc --aggressive' against the application's repo
func CommandGc(appName string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	return RepoGc(appName)
}

// CommandPurgeCache deletes the contents of the build cache stored in the repository
func CommandPurgeCache(appName string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	return PurgeCache(appName)
}
