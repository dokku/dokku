package buildpacks

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

func rewriteBuildpacksFile(sourceWorkDir string) error {
	buildpacksPath := filepath.Join(sourceWorkDir, ".buildpacks")
	if !common.FileExists(buildpacksPath) {
		return nil
	}

	buildpacks, err := common.FileToSlice(buildpacksPath)
	if err != nil {
		return err
	}

	for i, buildpack := range buildpacks {
		if buildpack == "" {
			continue
		}

		buildpack, err = validBuildpackURL(buildpack)
		if err != nil {
			return fmt.Errorf("Unable to parse .buildpacks file, line %d: %s", i, err)
		}

		buildpacks[i] = buildpack
	}

	return common.WriteSliceToFile(common.WriteSliceToFileInput{
		Filename: buildpacksPath,
		Lines:    buildpacks,
		Mode:     os.FileMode(0600),
	})
}

func validBuildpackURL(buildpack string) (string, error) {
	if buildpack == "" {
		return buildpack, errors.New("Must specify a buildpack url or reference")
	}

	reHerokuValue := regexp.MustCompile(`(?m)^([\w-]+\/[\w-]+)$`)
	if found := reHerokuValue.Find([]byte(buildpack)); found != nil {
		parts := strings.SplitN(buildpack, "/", 2)
		if parts[0] == "heroku-community" {
			parts[0] = "heroku"
		}
		return fmt.Sprintf("https://github.com/%s/heroku-buildpack-%s.git", parts[0], parts[1]), nil
	}

	reString := regexp.MustCompile(`(?m)^(http|https|git)(:\/\/|@)([^\/:]+)[\/:]([^\/:]+)\/(.+)(.git(#derp)?)?$`)
	if found := reString.Find([]byte(buildpack)); found != nil {
		return buildpack, nil
	}

	return buildpack, fmt.Errorf("Invalid buildpack specified: %v", buildpack)
}

func isBareGitRepo(directory string) bool {
	headFile := filepath.Join(directory, "HEAD")
	objectsDir := filepath.Join(directory, "objects")
	if common.FileExists(headFile) && common.DirectoryExists(objectsDir) {
		return true
	}
	return false
}

func checkoutBareGitRepo(sourceWorkDir string) (string, error) {
	tmpdir, err := os.MkdirTemp("", "bare-git-checkout")
	if err != nil {
		return "", fmt.Errorf("Unable to create temporary directory: %s", err.Error())
	}

	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command: "git",
		Args:    []string{"--git-dir=" + sourceWorkDir, "--work-tree=" + tmpdir, "checkout", "-f"},
	})

	if err != nil {
		os.RemoveAll(tmpdir)
		return "", fmt.Errorf("Unable to checkout bare git repository: %s", result.StderrContents())
	}

	return tmpdir, nil
}
