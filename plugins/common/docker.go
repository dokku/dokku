package common

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/codeskyblue/go-sh"
)

// ContainerIsRunning checks to see if a container is running
func ContainerIsRunning(containerID string) bool {
	b, err := DockerInspect(containerID, "'{{.State.Running}}'")
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(b[:])) == "true"
}

// CopyFromImage copies a file from named image to destination
func CopyFromImage(appName string, image string, source string, destination string) error {
	if err := VerifyAppName(appName); err != nil {
		return err
	}

	if !VerifyImage(image) {
		return fmt.Errorf("Invalid docker image for copying content")
	}

	workDir := ""
	if !IsAbsPath(source) {
		if IsImageHerokuishBased(image, appName) {
			workDir = "/app"
		} else {
			workDir, _ = DockerInspect(image, "{{.Config.WorkingDir}}")
		}

		if workDir != "" {
			source = fmt.Sprintf("%s/%s", workDir, source)
		}
	}

	tmpFile, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("dokku-%s-%s", MustGetEnv("DOKKU_PID"), "CopyFromImage"))
	if err != nil {
		return fmt.Errorf("Cannot create temporary file: %v", err)
	}

	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	globalRunArgs := MustGetEnv("DOKKU_GLOBAL_RUN_ARGS")
	createLabelArgs := []string{"--label", fmt.Sprintf("com.dokku.app-name=%s", appName), globalRunArgs}
	containerID, err := DockerContainerCreate(image, createLabelArgs)
	if err != nil {
		return fmt.Errorf("Unable to create temporary container: %v", err)
	}

	// docker cp exits with status 1 when run as non-root user when it tries to chown the file
	// after successfully copying the file. Thus, we suppress stderr.
	// ref: https://github.com/dotcloud/docker/issues/3986
	containerCopyCmd := NewShellCmd(strings.Join([]string{
		DockerBin(),
		"container",
		"cp",
		fmt.Sprintf("%s:%s", containerID, source),
		tmpFile.Name(),
	}, " "))
	containerCopyCmd.ShowOutput = false
	fileCopied := containerCopyCmd.Execute()

	containerRemoveCmd := NewShellCmd(strings.Join([]string{
		DockerBin(),
		"container",
		"rm",
		"--force",
		containerID,
	}, " "))
	containerRemoveCmd.ShowOutput = false
	containerRemoveCmd.Execute()

	if !fileCopied {
		return fmt.Errorf("Unable to copy file %s from image", source)
	}

	fi, err := os.Stat(tmpFile.Name())
	if err != nil {
		return err
	}

	if fi.Size() == 0 {
		return fmt.Errorf("Unable to copy file %s from image", source)
	}

	// workaround for CHECKS file when owner is root. seems to only happen when running inside docker
	dos2unixCmd := NewShellCmd(strings.Join([]string{
		"dos2unix",
		"-l",
		"-n",
		tmpFile.Name(),
		destination,
	}, " "))
	dos2unixCmd.ShowOutput = false
	dos2unixCmd.Execute()

	// add trailing newline for certain places where file parsing depends on it
	b, err := sh.Command("tail", "-c1", destination).Output()
	if string(b) != "" {
		f, err := os.OpenFile(destination, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := f.WriteString("\n"); err != nil {
			return fmt.Errorf("Unable to append trailing newline to copied file: %v", err)
		}
	}

	return nil
}

// DockerBin returns a string which contains a path to the current docker binary
func DockerBin() string {
	dockerBin := os.Getenv("DOCKER_BIN")
	if dockerBin == "" {
		dockerBin = "docker"
	}

	return dockerBin
}

// DockerCleanup cleans up all exited/dead containers and removes all dangling images
func DockerCleanup(appName string, forceCleanup bool) error {
	if !forceCleanup {
		skipCleanup := false
		if appName != "" {
			b, _ := PlugnTriggerOutput("config-get", []string{appName, "DOKKU_SKIP_CLEANUP"}...)
			output := strings.TrimSpace(string(b[:]))
			if output == "true" {
				skipCleanup = true
			}
		}

		if skipCleanup || os.Getenv("DOKKU_SKIP_CLEANUP") == "true" {
			LogInfo1("DOKKU_SKIP_CLEANUP set. Skipping dokku cleanup")
			return nil
		}
	}

	LogInfo1("Cleaning up...")
	scheduler := GetAppScheduler(appName)
	if appName == "--global" {
		appName = ""
	}

	forceCleanupArg := "false"
	if forceCleanup {
		forceCleanupArg = "true"
	}

	if err := PlugnTrigger("scheduler-docker-cleanup", []string{scheduler, appName, forceCleanupArg}...); err != nil {
		return fmt.Errorf("Failure while cleaning up app: %s", err)
	}

	// delete all non-running and dead containers
	exitedContainerIDs, _ := listContainers("exited", appName)
	deadContainerIDs, _ := listContainers("dead", appName)
	containerIDs := append(exitedContainerIDs, deadContainerIDs...)

	if len(containerIDs) > 0 {
		removeContainers(containerIDs)
	}

	// delete dangling images
	imageIDs, _ := listDanglingImages(appName)
	if len(imageIDs) > 0 {
		RemoveImages(imageIDs)
	}

	if appName != "" {
		// delete unused images
		pruneUnusedImages(appName)
	}

	return nil
}

// DockerContainerCreate creates a new container and returns the container ID
func DockerContainerCreate(image string, containerCreateArgs []string) (string, error) {
	cmd := []string{
		DockerBin(),
		"container",
		"create",
	}

	cmd = append(cmd, containerCreateArgs...)
	cmd = append(cmd, image)

	var stderr bytes.Buffer
	containerCreateCmd := NewShellCmd(strings.Join(cmd, " "))
	containerCreateCmd.ShowOutput = false
	containerCreateCmd.Command.Stderr = &stderr
	b, err := containerCreateCmd.Output()
	if err != nil {
		return "", errors.New(strings.TrimSpace(stderr.String()))
	}

	return strings.TrimSpace(string(b[:])), nil
}

// DockerInspect runs an inspect command with a given format against a container or image ID
func DockerInspect(containerOrImageID, format string) (output string, err error) {
	b, err := sh.Command(DockerBin(), "inspect", "--format", format, containerOrImageID).Output()
	if err != nil {
		return "", err
	}
	output = strings.TrimSpace(string(b[:]))
	if strings.HasPrefix(output, "'") && strings.HasSuffix(output, "'") {
		output = strings.TrimSuffix(strings.TrimPrefix(output, "'"), "'")
	}
	return
}

// IsImageHerokuishBased returns true if app image is based on herokuish
func IsImageHerokuishBased(image string, appName string) bool {
	output, err := DockerInspect(image, "{{range .Config.Env}}{{if eq . \"USER=herokuishuser\" }}{{println .}}{{end}}{{end}}")
	if err != nil {
		return false
	}
	return output != ""
}

// RemoveImages removes images by ID
func RemoveImages(imageIDs []string) {
	command := []string{
		DockerBin(),
		"image",
		"rm",
	}

	command = append(command, imageIDs...)

	var stderr bytes.Buffer
	rmCmd := NewShellCmd(strings.Join(command, " "))
	rmCmd.ShowOutput = false
	rmCmd.Command.Stderr = &stderr
	rmCmd.Execute()
}

// VerifyImage returns true if docker image exists in local repo
func VerifyImage(image string) bool {
	imageCmd := NewShellCmd(strings.Join([]string{DockerBin(), "image", "inspect", image}, " "))
	imageCmd.ShowOutput = false
	return imageCmd.Execute()
}

func listContainers(status string, appName string) ([]string, error) {
	command := []string{
		DockerBin(),
		"container",
		"list",
		"--quiet",
		"--all",
		"--filter",
		fmt.Sprintf("status=%v", status),
		"--filter",
		fmt.Sprintf("label=%v", os.Getenv("DOKKU_CONTAINER_LABEL")),
	}

	if appName != "" {
		command = append(command, []string{"--filter", fmt.Sprintf("label=com.dokku.app-name=%v", appName)}...)
	}

	var stderr bytes.Buffer
	listCmd := NewShellCmd(strings.Join(command, " "))
	listCmd.ShowOutput = false
	listCmd.Command.Stderr = &stderr
	b, err := listCmd.Output()

	if err != nil {
		return []string{}, errors.New(strings.TrimSpace(stderr.String()))
	}

	output := strings.Split(strings.TrimSpace(string(b[:])), "\n")
	return output, nil
}

func listDanglingImages(appName string) ([]string, error) {
	command := []string{
		DockerBin(),
		"image",
		"list",
		"--quiet",
		"--filter",
		"dangling=true",
	}

	if appName != "" {
		command = append(command, []string{"--filter", fmt.Sprintf("label=com.dokku.app-name=%v", appName)}...)
	}

	var stderr bytes.Buffer
	listCmd := NewShellCmd(strings.Join(command, " "))
	listCmd.ShowOutput = false
	listCmd.Command.Stderr = &stderr
	b, err := listCmd.Output()

	if err != nil {
		return []string{}, errors.New(strings.TrimSpace(stderr.String()))
	}

	output := strings.Split(strings.TrimSpace(string(b[:])), "\n")
	return output, nil
}

func pruneUnusedImages(appName string) {
	command := []string{
		DockerBin(),
		"image",
		"prune",
		"--all",
		"--force",
		"--filter",
		fmt.Sprintf("label=com.dokku.app-name=%v", appName),
	}

	var stderr bytes.Buffer
	pruneCmd := NewShellCmd(strings.Join(command, " "))
	pruneCmd.ShowOutput = false
	pruneCmd.Command.Stderr = &stderr
	pruneCmd.Execute()
}

func removeContainers(containerIDs []string) {
	command := []string{
		DockerBin(),
		"container",
		"rm",
	}

	command = append(command, containerIDs...)

	var stderr bytes.Buffer
	rmCmd := NewShellCmd(strings.Join(command, " "))
	rmCmd.ShowOutput = false
	rmCmd.Command.Stderr = &stderr
	rmCmd.Execute()
}
