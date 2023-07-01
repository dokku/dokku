package builder

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

func listImagesByAppLabel(appName string) ([]string, error) {
	command := []string{
		common.DockerBin(),
		"image",
		"list",
		"--quiet",
		"--filter",
		fmt.Sprintf("label=com.dokku.app-name=%v", appName),
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

func listImagesByImageRepo(imageRepo string) ([]string, error) {
	command := []string{
		common.DockerBin(),
		"image",
		"list",
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
	dir, err := ioutil.ReadDir(basePath)
	if err != nil {
		return err
	}

	for _, d := range dir {
		os.RemoveAll(path.Join([]string{basePath, d.Name()}...))
	}

	return nil
}
