package repo

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dokku/dokku/plugins/common"
)

// PurgeCacheFailed wraps error to allow returning the correct exit code
type PurgeCacheFailed struct {
	exitCode int
}

// ExitCode returns an exit code to use in case this error bubbles
// up into an os.Exit() call
func (err *PurgeCacheFailed) ExitCode() int {
	return err.exitCode
}

// Error returns a standard non-existent app error
func (err *PurgeCacheFailed) Error() string {
	return fmt.Sprintf("failed to purge cache, exit code %d", err.exitCode)
}

// PurgeCache deletes the contents of the build cache stored in the repository
func PurgeCache(appName string) error {
	containerIDs, _ := common.DockerFilterContainers([]string{
		fmt.Sprintf("label=com.dokku.app-name=%v", appName),
		"label=com.dokku.image-stage=build",
	})
	if len(containerIDs) > 0 {
		common.DockerRemoveContainers(containerIDs)
	}
	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command:      common.DockerBin(),
		Args:         []string{"volume", "rm", "-f", fmt.Sprintf("cache-%s", appName)},
		StreamStderr: true,
	})
	if err != nil {
		return fmt.Errorf("Unable to remove cache volume: %w", err)
	}
	if result.ExitCode != 0 {
		return &PurgeCacheFailed{result.ExitCode}
	}

	return nil
}

func RepoGc(appName string) error {
	heads := map[string][]byte{}
	headsDir := filepath.Join(common.AppRoot(appName), "refs", "heads")
	if common.DirectoryExists(headsDir) {
		headFiles, err := os.ReadDir(headsDir)
		if err != nil {
			return fmt.Errorf("Unable to read heads directory: %w", err)
		}

		for _, head := range headFiles {
			if head.IsDir() {
				continue
			}
			headContents, err := os.ReadFile(filepath.Join(headsDir, head.Name()))
			if err != nil {
				return fmt.Errorf("Unable to read head file: %w", err)
			}
			heads[head.Name()] = headContents
		}
	}

	defer func() {
		for head, contents := range heads {
			if err := os.WriteFile(filepath.Join(headsDir, head), contents, 0644); err != nil {
				common.LogWarn(fmt.Sprintf("Unable to write head file: %s\n", err))
			}
		}
	}()
	appRoot := common.AppRoot(appName)
	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command: "git",
		Args:    []string{"gc", "--aggressive"},
		Env: map[string]string{
			"GIT_DIR": appRoot,
		},
		StreamStderr: true,
	})
	if err != nil {
		return fmt.Errorf("Unable to run git gc: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Unable to run git gc: %s", result.StderrContents())
	}

	return nil
}
