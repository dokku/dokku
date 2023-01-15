package repo

import (
	"fmt"
	"os"
	"strings"

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
	purgeCacheCmd := common.NewShellCmd(strings.Join([]string{
		common.DockerBin(),
		"volume",
		"rm", "-f", fmt.Sprintf("cache-%s", appName)}, " "))
	purgeCacheCmd.ShowOutput = false
	purgeCacheCmd.Command.Stderr = os.Stderr
	if !purgeCacheCmd.Execute() {
		exitCode := 1
		if purgeCacheCmd.ExitError != nil {
			exitCode = purgeCacheCmd.ExitError.ExitCode()
		}
		return &PurgeCacheFailed{exitCode}
	}

	return nil
}
