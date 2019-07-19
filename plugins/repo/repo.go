package repo

import (
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
)

// PurgeCache deletes the contents of the build cache stored in the repository
func PurgeCache(appName string) error {
	err := common.VerifyAppName(appName)
	if err != nil {
		return err
	}

	cacheDir := strings.Join([]string{common.MustGetEnv("DOKKU_ROOT"), appName, "cache"}, "/")
	cacheHostDir := strings.Join([]string{common.MustGetEnv("DOKKU_HOST_ROOT"), appName, "cache"}, "/")
	dokkuGlobalRunArgs := common.MustGetEnv("DOKKU_GLOBAL_RUN_ARGS")
	image := config.GetWithDefault(appName, "DOKKU_IMAGE", os.Getenv("DOKKU_IMAGE"))
	if info, _ := os.Stat(cacheDir); info != nil && info.IsDir() {
		dockerLabelArgs := fmt.Sprintf("--label=com.dokku.app-name=%s", appName)
		purgeCacheCmd := common.NewShellCmd(strings.Join([]string{
			common.DockerBin(),
			"run --rm", dockerLabelArgs, dokkuGlobalRunArgs,
			"-v", strings.Join([]string{cacheHostDir, ":/cache"}, ""), image,
			`find /cache -depth -mindepth 1 -maxdepth 1 -exec rm -Rf {} ;`}, " "))
		purgeCacheCmd.Execute()
		err := os.MkdirAll(cacheDir, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}
