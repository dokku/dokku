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
	purgeCacheCmd := common.NewShellCmd(strings.Join([]string{
		common.DockerBin(),
		"volume",
		"rm", "-f", fmt.Sprintf("cache-%s", appName)}, " "))
	purgeCacheCmd.ShowOutput = false
	purgeCacheCmd.Command.Stderr = os.Stderr
	if !purgeCacheCmd.Execute() {
		return &PurgeCacheFailed{purgeCacheCmd.ExitError.ExitCode()}
	}

	return nil
}
