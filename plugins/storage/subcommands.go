package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	dockeroptions "github.com/dokku/dokku/plugins/docker-options"
)

const (
	helpHeader = `Usage: dokku storage[:COMMAND]

Manage mounted volumes

Additional commands:`

	helpContent = `
    storage:ensure-directory [--chown option] <directory>, Creates a persistent storage directory in the recommended storage path
    storage:list <app> [--format text|json], List bind mounts for app's container(s) (host:container)
    storage:mount <app> <host-dir:container-dir>, Create a new bind mount
    storage:report [<app>] [<flag>], Displays a storage report for one or more apps
    storage:unmount <app> <host-dir:container-dir>, Remove an existing bind mount`
)

// CommandHelp displays help for the storage plugin
func CommandHelp() error {
	common.CommandUsage(helpHeader, helpContent)
	return nil
}

// CommandEnsureDirectory creates a persistent storage directory
func CommandEnsureDirectory(directory string, chownFlag string) error {
	if err := ValidateDirectoryName(directory); err != nil {
		return err
	}

	chownID, err := resolveChownID(chownFlag)
	if err != nil {
		return err
	}

	storageDirectory := filepath.Join(GetStorageDirectory(), directory)
	common.LogInfo1(fmt.Sprintf("Ensuring %s exists", storageDirectory))

	if err := os.MkdirAll(storageDirectory, 0755); err != nil {
		return fmt.Errorf("Unable to create directory: %s", err.Error())
	}

	if chownID != "false" {
		common.LogVerboseQuiet(fmt.Sprintf("Setting directory ownership to %s:%s", chownID, chownID))

		pluginPath := common.MustGetEnv("PLUGIN_AVAILABLE_PATH")
		chownScript := filepath.Join(pluginPath, "storage", "bin", "chown-storage-dir")

		result, err := common.CallExecCommand(common.ExecCommandInput{
			Command: "sudo",
			Args:    []string{chownScript, directory, chownID},
		})
		if err != nil {
			return fmt.Errorf("Unable to set directory ownership: %s", err.Error())
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("Unable to set directory ownership: %s", result.StderrContents())
		}
	}

	common.LogVerboseQuiet("Directory ready for mounting")
	return nil
}

// resolveChownID converts a chown flag value to a numeric UID
func resolveChownID(chownFlag string) (string, error) {
	var chownID string

	switch chownFlag {
	case "herokuish":
		chownID = "32767"
	case "heroku":
		chownID = "1000"
	case "packeto":
		common.LogVerbose("Detected deprecated chown flag 'packeto'. Using 'paketo' instead. Please update your configuration.")
		chownID = "2000"
	case "paketo":
		chownID = "2000"
	case "root":
		chownID = "0"
	case "false":
		return "false", nil
	default:
		return "", errors.New("Unsupported chown permissions")
	}

	userns, err := isUserNamespacesEnabled()
	if err != nil {
		return "", err
	}

	if userns && chownID != "false" {
		uid := 0
		fmt.Sscanf(chownID, "%d", &uid)
		uid += 165536
		chownID = fmt.Sprintf("%d", uid)
	}

	return chownID, nil
}

// isUserNamespacesEnabled checks if Docker user namespaces are enabled
func isUserNamespacesEnabled() (bool, error) {
	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command: common.DockerBin(),
		Args:    []string{"info", "-f", "{{range .SecurityOptions}}{{if eq . \"name=userns\"}}true{{end}}{{end}}"},
	})
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(result.StdoutContents()) == "true", nil
}

// CommandMount creates a new bind mount for an app
func CommandMount(appName string, mountPath string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if err := VerifyPaths(mountPath); err != nil {
		return err
	}

	if CheckIfPathExists(appName, mountPath, MountPhases) {
		return errors.New("Mount path already exists.")
	}

	return dockeroptions.AddDockerOptionToPhases(appName, MountPhases, fmt.Sprintf("-v %s", mountPath))
}

// CommandUnmount removes an existing bind mount from an app
func CommandUnmount(appName string, mountPath string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if err := VerifyPaths(mountPath); err != nil {
		return err
	}

	if !CheckIfPathExists(appName, mountPath, MountPhases) {
		return errors.New("Mount path does not exist.")
	}

	return dockeroptions.RemoveDockerOptionFromPhases(appName, MountPhases, fmt.Sprintf("-v %s", mountPath))
}

// CommandList lists all bind mounts for an app
func CommandList(appName string, format string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if format != "text" && format != "json" {
		return errors.New("Invalid --format value specified")
	}

	if format == "text" {
		common.LogInfo1Quiet(fmt.Sprintf("%s volume bind-mounts:", appName))
	}

	results, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "storage-list",
		Args:    []string{appName, "deploy", format},
	})
	if err != nil {
		return err
	}

	output := results.StdoutContents()
	if format == "text" && output != "" {
		lines := strings.Split(strings.TrimSpace(output), "\n")
		for _, line := range lines {
			if line != "" {
				if os.Getenv("DOKKU_QUIET_OUTPUT") != "" {
					fmt.Println(line)
				} else {
					common.LogVerbose(line)
				}
			}
		}
	} else if output != "" {
		fmt.Println(output)
	}

	return nil
}

// CommandReport displays a storage report for one or more apps
func CommandReport(appName string, format string, infoFlag string) error {
	if appName == "" {
		apps, err := common.DokkuApps()
		if err != nil {
			if errors.Is(err, common.NoAppsExist) {
				common.LogWarn(err.Error())
				return nil
			}
			return err
		}
		for _, app := range apps {
			if err := ReportSingleApp(app, format, infoFlag); err != nil {
				return err
			}
		}
		return nil
	}

	return ReportSingleApp(appName, format, infoFlag)
}
