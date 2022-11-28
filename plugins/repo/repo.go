package repo

import (
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// PurgeCache deletes the contents of the build cache stored in the repository
func PurgeCache(appName string) error {
	cacheDir := strings.Join([]string{common.AppRoot(appName), "cache"}, "/")
	if info, _ := os.Stat(cacheDir); info != nil && info.IsDir() {
		purgeCacheCmd := common.NewShellCmd(strings.Join([]string{
			common.DockerBin(),
			"volume",
			"rm", "-f", "cache-$APP"}, " "))
		purgeCacheCmd.Execute()
		err := os.MkdirAll(cacheDir, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}
