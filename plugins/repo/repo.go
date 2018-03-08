package repo

import (
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
	"github.com/otiai10/copy"
)

var (
	// DefaultProperties is a map of all valid repo properties with corresponding default property values
	DefaultProperties = map[string]string{
		"container-copy-folder": "",
		"host-copy-folder":      "",
	}
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
			"container",
			"run", "--rm", dockerLabelArgs, dokkuGlobalRunArgs,
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

// ReportSingleApp is an internal function that displays the app report for one or more apps
func ReportSingleApp(appName, infoFlag string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	infoFlags := map[string]string{
		"--repo-container-copy-folder": common.PropertyGet("repo", appName, "container-copy-folder"),
		"--repo-host-copy-folder":      common.PropertyGet("repo", appName, "host-copy-folder"),
	}

	trimPrefix := false
	uppercaseFirstCharacter := true
	return common.ReportSingleApp("repo", appName, infoFlag, infoFlags, trimPrefix, uppercaseFirstCharacter)
}

func copyDirectory(sourceBasePath string, sourceFolder string, destinationPath string) error {
	if sourceFolder == "" {
		return nil
	}

	sourcePath := strings.Join([]string{sourceBasePath, sourceFolder}, "/")
	stat, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return copy.Copy(sourcePath, destinationPath)
	}

	return nil
}
