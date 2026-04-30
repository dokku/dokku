package schedulerdockerlocal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/storage"
)

// dockerNamedVolumeRegexp mirrors storage.dockerNamedVolumeRegexp and is
// used to decide whether a HostPath should be statted (filesystem) or
// looked up via `docker volume inspect`.
var dockerNamedVolumeRegexp = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]+$`)

// StorageExecInput captures the inputs forwarded by the storage plugin
// over the scheduler-storage-exec trigger.
type StorageExecInput struct {
	EntryName   string
	Image       string
	Interactive bool
	Tty         bool
	AsUser      string
	Command     []string
}

// TriggerSchedulerStorageExec runs an interactive or non-interactive
// command in a throwaway docker container that mounts the storage
// entry's host path or named volume at /data. Returns the exit code of
// docker run via os.Exit so the caller's status mirrors the underlying
// tool.
func TriggerSchedulerStorageExec(scheduler string, input StorageExecInput) error {
	if scheduler != "docker-local" {
		// Not for us; let other handlers respond.
		return nil
	}

	if !storage.EntryExists(input.EntryName) {
		return fmt.Errorf("storage entry %q does not exist", input.EntryName)
	}
	entry, err := storage.LoadEntry(input.EntryName)
	if err != nil {
		return err
	}
	if entry.Scheduler != storage.SchedulerDockerLocal {
		return fmt.Errorf("storage entry %q has scheduler %q, not docker-local", entry.Name, entry.Scheduler)
	}

	if err := preflightDockerLocalSource(entry.HostPath); err != nil {
		return err
	}

	args, err := buildDockerExecArgs(entry, input)
	if err != nil {
		return err
	}

	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command:     common.DockerBin(),
		Args:        args,
		StreamStdio: true,
	})
	// CallExecCommand wraps non-zero exit as an error; for storage:exec
	// the docker-run exit code is the signal we want to forward to the
	// caller. Propagate it before falling through to the err return,
	// which would otherwise be collapsed to exit 1 by the dispatcher.
	if result.ExitCode != 0 {
		os.Exit(result.ExitCode)
	}
	if err != nil {
		return err
	}
	return nil
}

// preflightDockerLocalSource fails fast when the host path or named
// volume backing the entry doesn't exist, so the user gets a real error
// instead of a generic "docker: ..." line.
func preflightDockerLocalSource(hostPath string) error {
	if hostPath == "" {
		return errors.New("storage entry is missing a host path")
	}
	if filepath.IsAbs(hostPath) {
		info, err := os.Stat(hostPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("host path %s does not exist", hostPath)
			}
			return fmt.Errorf("unable to stat host path %s: %w", hostPath, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("host path %s exists but is not a directory", hostPath)
		}
		return nil
	}
	if !dockerNamedVolumeRegexp.MatchString(hostPath) {
		return fmt.Errorf("host path %q is neither absolute nor a valid docker volume token", hostPath)
	}
	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command: common.DockerBin(),
		Args:    []string{"volume", "inspect", hostPath},
	})
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("docker named volume %q is not present: %s", hostPath, strings.TrimSpace(result.StderrContents()))
	}
	return nil
}

// buildDockerExecArgs assembles the `docker run` argv for a storage:exec
// invocation. Split out so the unit test can exercise it without
// actually invoking docker.
func buildDockerExecArgs(entry *storage.Entry, input StorageExecInput) ([]string, error) {
	cmd := input.Command
	if len(cmd) == 0 {
		cmd = []string{"sh", "-c", "command -v bash >/dev/null 2>&1 && exec bash || exec sh"}
	}

	args := []string{"run", "--rm"}
	if input.Tty {
		args = append(args, "-it")
	} else if input.Interactive {
		args = append(args, "-i")
	}

	user, err := resolveUser(entry.Chown, input.AsUser)
	if err != nil {
		return nil, err
	}
	if user != "" {
		args = append(args, "--user", user)
	}

	args = append(args, "-v", fmt.Sprintf("%s:/data", entry.HostPath))
	args = append(args, input.Image)
	args = append(args, cmd...)
	return args, nil
}

// resolveUser turns the entry's Chown setting plus an optional --as-user
// override into a docker --user value. Returns an empty string when the
// container should run as the image's default user.
func resolveUser(chown string, asUser string) (string, error) {
	if asUser != "" {
		return fmt.Sprintf("%s:%s", asUser, asUser), nil
	}
	if chown == "" || chown == "false" {
		return "", nil
	}
	uid, err := storage.ResolveChownID(chown)
	if err != nil {
		return "", err
	}
	if uid == "" || uid == "false" {
		return "", nil
	}
	return fmt.Sprintf("%s:%s", uid, uid), nil
}
