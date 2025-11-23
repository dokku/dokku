package buildpacks

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	appjson "github.com/dokku/dokku/plugins/app-json"
	"github.com/dokku/dokku/plugins/common"
)

// getBuildpacks returns the buildpacks for a given app
// the order of application is:
// 1. .buildpacks: if it exists, use it
// 2. app.json: if it exists and has buildpacks, use them
// 3. buildpack properties: if they exist, use them
// 4. buildpack detection
func getBuildpacks(appName string) ([]string, error) {
	buildpacks, err := common.PropertyListGet("buildpacks", appName, "buildpacks")
	if err != nil {
		return buildpacks, err
	}

	appJSON, err := appjson.GetAppJSON(appName)
	if err != nil {
		return buildpacks, err
	}

	if len(appJSON.Buildpacks) == 0 {
		return buildpacks, nil
	}

	for _, b := range appJSON.Buildpacks {
		buildpacks = append(buildpacks, b.URL)
	}

	return buildpacks, nil
}

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

		if strings.HasPrefix(buildpack, "#") {
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

func checkoutBareGitRepo(sourceWorkDir string, branch string) (string, error) {
	tmpdir, err := os.MkdirTemp("", "bare-git-checkout")
	if err != nil {
		return "", fmt.Errorf("Unable to create temporary directory: %s", err.Error())
	}

	args := []string{
		"--git-dir=" + sourceWorkDir,
		"--work-tree=" + tmpdir,
		"checkout", "-f",
	}

	if branch != "" {
		args = append(args, branch)
	}

	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command: "git",
		Args:    args,
	})

	if err != nil {
		os.RemoveAll(tmpdir)
		return "", fmt.Errorf("Unable to checkout bare git repository: %s", result.StderrContents())
	}

	return tmpdir, nil
}
