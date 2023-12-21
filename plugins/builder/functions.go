package builder

import (
	"bytes"
	"errors"
	"os"
	"path"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

func listImagesByImageRepo(imageRepo string) ([]string, error) {
	command := []string{
		common.DockerBin(),
		"image",
		"ls",
		"--quiet",
		imageRepo,
	}

	var stderr bytes.Buffer
	listCmd := common.NewShellCmd(strings.Join(command, " "))
	listCmd.ShowOutput = false
	listCmd.Command.Stderr = &stderr
	b, err := listCmd.Output()

	if err != nil {
		return []string{}, errors.New(strings.TrimSpace(stderr.String()))
	}

	output := strings.Split(strings.TrimSpace(string(b[:])), "\n")
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
