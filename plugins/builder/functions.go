package builder

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

func listImagesByImageRepo(imageRepo string) ([]string, error) {
	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command: common.DockerBin(),
		Args:    []string{"image", "ls", "--quiet", imageRepo},
	})
	if err != nil {
		return []string{}, fmt.Errorf("Unable to list images: %w", err)
	}
	if result.ExitCode != 0 {
		return []string{}, fmt.Errorf("Unable to list images: %s", result.StderrContents())
	}

	output := strings.Split(result.StdoutContents(), "\n")
	return output, nil
}

func removeAllContents(basePath string) error {
	dir, err := os.ReadDir(basePath)
	if err != nil {
		return err
	}

	for _, d := range dir {
		os.RemoveAll(path.Join([]string{basePath, d.Name()}...))
	}

	return nil
}
