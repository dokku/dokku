package common

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

// ComposeUpInput is the input for the ComposeUp function
type ComposeUpInput struct {
	ProjectName string
	ComposeFile string
}

// ComposeUp executes a docker compose up command
func ComposeUp(input ComposeUpInput) error {
	result, err := CallExecCommand(ExecCommandInput{
		Command: DockerBin(),
		Args: []string{
			"compose",
			"--file", input.ComposeFile,
			"--project-name", input.ProjectName,
			"up",
			"--detach",
			"--quiet-pull",
		},
		StreamStdio:      true,
		WorkingDirectory: "/tmp",
	})

	if err != nil || result.ExitCode != 0 {
		return fmt.Errorf("Unable to start compose project: %s", result.Stderr)
	}

	return nil
}

// ComposeDownInput is the input for the ComposeDown function
type ComposeDownInput struct {
	ProjectName string
	ComposeFile string
}

// ComposeDown executes a docker compose down command
func ComposeDown(input ComposeDownInput) error {
	result, err := CallExecCommand(ExecCommandInput{
		Command: DockerBin(),
		Args: []string{
			"compose",
			"--file", input.ComposeFile,
			"--project-name", input.ProjectName,
			"down",
			"--remove-orphans",
		},
		StreamStdio:      true,
		WorkingDirectory: "/tmp",
	})

	if err != nil || result.ExitCode != 0 {
		return fmt.Errorf("Unable to stop %s: %s", input.ProjectName, result.Stderr)
	}

	return nil
}

// ContainerIsRunning checks to see if a container is running
func ContainerIsRunning(containerID string) bool {
	b, err := DockerInspect(containerID, "'{{.State.Running}}'")
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(b[:])) == "true"
}

// ContainerStart runs 'docker container start' against an existing container
func ContainerStart(containerID string) bool {
	result, err := CallExecCommand(ExecCommandInput{
		Command:      DockerBin(),
		Args:         []string{"container", "start", containerID},
		StreamStderr: true,
	})
	if err != nil {
		return false
	}
	return result.ExitCode == 0
}

// ContainerRemove runs 'docker container remove' against an existing container
func ContainerRemove(containerID string) bool {
	result, err := CallExecCommand(ExecCommandInput{
		Command:      DockerBin(),
		Args:         []string{"container", "remove", "-f", containerID},
		StreamStderr: true,
	})
	if err != nil {
		return false
	}
	return result.ExitCode == 0
}

// ContainerExists checks to see if a container exists
func ContainerExists(containerID string) bool {
	result, err := CallExecCommand(ExecCommandInput{
		Command: DockerBin(),
		Args:    []string{"container", "inspect", containerID},
	})
	if err != nil {
		return false
	}
	return result.ExitCode == 0
}

// ContainerWait runs 'docker container wait' against an existing container
func ContainerWait(containerID string) bool {
	result, err := CallExecCommand(ExecCommandInput{
		Command:      DockerBin(),
		Args:         []string{"container", "wait", containerID},
		StreamStderr: true,
	})
	if err != nil {
		return false
	}

	stdout := result.StdoutContents()
	if stdout != "0" {
		exitCode, err := strconv.Atoi(stdout)
		if err != nil {
			return false
		}

		return exitCode == 0
	}

	return result.ExitCode == 0
}

// ContainerWaitTilReady will wait timeout seconds and then check if a container is running
// returning an error if it is not running at the end of the timeout
func ContainerWaitTilReady(containerID string, timeout time.Duration) error {
	time.Sleep(timeout)

	if !ContainerIsRunning(containerID) {
		return fmt.Errorf("Container %s is not running", containerID)
	}

	return nil
}

// CopyDirFromImage copies a directory from named image to destination
func CopyDirFromImage(appName string, image string, source string, destination string) error {
	if !VerifyImage(image) {
		return fmt.Errorf("Invalid docker image for copying content")
	}

	if !IsAbsPath(source) {
		workDir := GetWorkingDir(appName, image)
		if workDir != "" {
			source = fmt.Sprintf("%s/%s", workDir, source)
		}
	}

	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("dokku-%s-%s", MustGetEnv("DOKKU_PID"), "CopyFromImage"))
	if err != nil {
		return fmt.Errorf("Error creating temporary directory: %v", err)
	}

	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			LogWarn(fmt.Sprintf("Error removing temporary directory %s: %v\n", tmpDir, err))
		}
	}()

	globalRunArgs := MustGetEnv("DOKKU_GLOBAL_RUN_ARGS")
	createLabelArgs := []string{"--label", fmt.Sprintf("com.dokku.app-name=%s", appName), globalRunArgs}
	containerID, err := DockerContainerCreate(image, createLabelArgs)
	if err != nil {
		return fmt.Errorf("Unable to create temporary container: %v", err)
	}
	defer ContainerRemove(containerID)

	// docker cp exits with status 1 when run as non-root user when it tries to chown the file
	// after successfully copying the file. Thus, we suppress stderr.
	// ref: https://github.com/dotcloud/docker/issues/3986
	result, err := CallExecCommand(ExecCommandInput{
		Command: DockerBin(),
		Args:    []string{"container", "cp", "--quiet", fmt.Sprintf("%s:%s", containerID, source), tmpDir},
	})
	if err != nil {
		return fmt.Errorf("Unable to copy file %s from image: %w", source, err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Unable to copy file %s from image: %v", source, result.StderrContents())
	}

	if !DirectoryExists(tmpDir) {
		return fmt.Errorf("Unable to copy file %s from image: %v", source, result.StderrContents())
	}

	files, err := os.ReadDir(tmpDir)
	if err != nil {
		return fmt.Errorf("Unable to read temporary directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		sourceFile := fmt.Sprintf("%s/%s", tmpDir, file.Name())
		destinationFile := fmt.Sprintf("%s/%s", destination, file.Name())
		// workaround when owner is root. seems to only happen when running inside docker
		CallExecCommand(ExecCommandInput{
			Command: "dos2unix",
			Args:    []string{"-l", "-n", sourceFile, destinationFile},
		}) // nolint: errcheck

		// add trailing newline for certain places where file parsing depends on it
		result, err = CallExecCommand(ExecCommandInput{
			Command: "tail",
			Args:    []string{"-c1", destination},
		})
		if err != nil || result.ExitCode != 0 {
			return fmt.Errorf("Unable to append trailing newline to copied file: %v", result.Stderr)
		}

		if result.Stdout != "" {
			f, err := os.OpenFile(destination, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := f.WriteString("\n"); err != nil {
				return fmt.Errorf("Unable to append trailing newline to copied file: %v", err)
			}
		}
	}

	return nil
}

// CopyFromImage copies a file from named image to destination
func CopyFromImage(appName string, image string, source string, destination string) error {
	if !VerifyImage(image) {
		return fmt.Errorf("Invalid docker image for copying content")
	}

	if !IsAbsPath(source) {
		workDir := GetWorkingDir(appName, image)
		if workDir != "" {
			source = fmt.Sprintf("%s/%s", workDir, source)
		}
	}

	globalRunArgs := MustGetEnv("DOKKU_GLOBAL_RUN_ARGS")
	createLabelArgs := []string{"--label", fmt.Sprintf("com.dokku.app-name=%s", appName), globalRunArgs}
	containerID, err := DockerContainerCreate(image, createLabelArgs)
	if err != nil {
		return fmt.Errorf("Unable to create temporary container: %v", err)
	}
	defer ContainerRemove(containerID)

	// docker cp exits with status 1 when run as non-root user when it tries to chown the file
	// after successfully copying the file. Thus, we suppress stderr.
	// ref: https://github.com/dotcloud/docker/issues/3986
	result, err := CallExecCommand(ExecCommandInput{
		Command: DockerBin(),
		Args:    []string{"container", "cp", "--quiet", fmt.Sprintf("%s:%s", containerID, source), "-"},
	})
	if err != nil {
		return fmt.Errorf("Unable to copy file %s from image: %w", source, err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Unable to copy file %s from image: %v", source, result.StderrContents())
	}

	tarContents := result.StdoutContents()
	if tarContents == "" {
		return fmt.Errorf("Unable to copy file %s from image", source)
	}

	// extract the contents via tar
	content, err := extractTarToString(tarContents)
	if err != nil {
		return fmt.Errorf("Unable to extract contents from tar: %v", err)
	}

	tmpFile, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("dokku-%s-%s", MustGetEnv("DOKKU_PID"), "CopyFromImage"))
	if err != nil {
		return fmt.Errorf("Cannot create temporary file: %v", err)
	}

	defer func() {
		if err := tmpFile.Close(); err != nil {
			LogWarn(fmt.Sprintf("Unable to close temporary file: %v", err))
		}
		if err := os.Remove(tmpFile.Name()); err != nil {
			LogWarn(fmt.Sprintf("Unable to remove temporary file: %v", err))
		}
	}()

	// write contents to tmpFile
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		return fmt.Errorf("Unable to write to temporary file: %v", err)
	}

	fi, err := os.Stat(tmpFile.Name())
	if err != nil {
		return err
	}

	if fi.Size() == 0 {
		return fmt.Errorf("Unable to copy file %s from image", source)
	}

	// workaround when owner is root. seems to only happen when running inside docker
	CallExecCommand(ExecCommandInput{
		Command: "dos2unix",
		Args:    []string{"-l", "-n", tmpFile.Name(), destination},
	}) // nolint: errcheck

	// add trailing newline for certain places where file parsing depends on it
	result, err = CallExecCommand(ExecCommandInput{
		Command: "tail",
		Args:    []string{"-c1", destination},
	})
	if err != nil || result.ExitCode != 0 {
		return fmt.Errorf("Unable to append trailing newline to copied file: %v", result.Stderr)
	}

	if result.Stdout != "" {
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

// Function to extract tar contents and return them as a string
func extractTarToString(in string) (string, error) {
	// Initialize a buffer to accumulate the extracted content
	var extractedContent bytes.Buffer

	// Create a tar reader from standard input
	tarReader := tar.NewReader(strings.NewReader(in))

	// Iterate through the files in the tar archive
	for {
		// Read the next header (file entry)
		_, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return "", fmt.Errorf("error reading tar header: %v", err)
		}

		// Write the content of the current file into the buffer
		_, err = io.Copy(&extractedContent, tarReader)
		if err != nil {
			return "", fmt.Errorf("error copying file content: %v", err)
		}
	}

	// Return the accumulated content as a string
	return strings.TrimSpace(extractedContent.String()), nil
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
			triggerName := "config-get"
			triggerArgs := []string{appName, "DOKKU_SKIP_CLEANUP"}
			if appName == "--global" {
				triggerName = "config-get-global"
				triggerArgs = []string{"DOKKU_SKIP_CLEANUP"}
			}

			results, _ := CallPlugnTrigger(PlugnTriggerInput{
				Trigger: triggerName,
				Args:    triggerArgs,
			})
			if results.StdoutContents() == "true" {
				skipCleanup = true
			}
		}

		if skipCleanup || os.Getenv("DOKKU_SKIP_CLEANUP") == "true" {
			LogInfo1("DOKKU_SKIP_CLEANUP set. Skipping dokku cleanup")
			return nil
		}
	}

	LogInfo1("Cleaning up...")
	if appName == "--global" {
		appName = ""
	}

	// delete all non-running and dead containers
	exitedContainerIDs, _ := listContainers("exited", appName)
	deadContainerIDs, _ := listContainers("dead", appName)
	containerIDs := append(exitedContainerIDs, deadContainerIDs...)

	if len(containerIDs) > 0 {
		DockerRemoveContainers(containerIDs)
	}

	// delete dangling images
	imageIDs, _ := ListDanglingImages(appName)
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
	args := []string{
		"container",
		"create",
	}

	args = append(args, containerCreateArgs...)
	args = append(args, image)

	result, err := CallExecCommand(ExecCommandInput{
		Command: DockerBin(),
		Args:    args,
	})
	if err != nil {
		return "", fmt.Errorf("Unable to create container: %w", err)
	}
	if result.ExitCode != 0 {
		return "", fmt.Errorf("Unable to create container: %s", result.StderrContents())
	}

	return result.StdoutContents(), nil
}

// DockerInspect runs an inspect command with a given format against a container or image ID
func DockerInspect(containerOrImageID, format string) (output string, err error) {
	result, err := CallExecCommand(ExecCommandInput{
		Command: DockerBin(),
		Args:    []string{"inspect", "--format", format, containerOrImageID},
	})
	if err != nil {
		return "", err
	}

	output = result.StdoutContents()
	if strings.HasPrefix(output, "'") && strings.HasSuffix(output, "'") {
		output = strings.TrimSuffix(strings.TrimPrefix(output, "'"), "'")
	}
	return
}

// GetWorkingDir returns the working directory for a given image
func GetWorkingDir(appName string, image string) string {
	if IsImageCnbBased(image) {
		return "/workspace"
	} else if IsImageHerokuishBased(image, appName) {
		return "/app"
	}

	workDir, _ := DockerInspect(image, "{{.Config.WorkingDir}}")
	return workDir
}

func IsComposeInstalled() bool {
	result, err := CallExecCommand(ExecCommandInput{
		Command: DockerBin(),
		Args:    []string{"info", "--format", "{{range .ClientInfo.Plugins}}{{if eq .Name \"compose\"}}true{{end}}{{end}}')"},
	})
	return err == nil && result.ExitCode == 0
}

// IsImageCnbBased returns true if app image is based on cnb
func IsImageCnbBased(image string) bool {
	if len(image) == 0 {
		return false
	}

	output, err := DockerInspect(image, "{{index .Config.Labels \"io.buildpacks.stack.id\" }}")
	if err != nil {
		return false
	}
	return output != ""
}

// IsImageHerokuishBased returns true if app image is based on herokuish
func IsImageHerokuishBased(image string, appName string) bool {
	if len(image) == 0 {
		return false
	}

	if IsImageCnbBased(image) {
		return true
	}

	dokkuAppUser := ""
	if len(appName) != 0 {
		results, err := CallPlugnTrigger(PlugnTriggerInput{
			Trigger: "config-get",
			Args:    []string{appName, "DOKKU_APP_USER"},
		})
		if err == nil {
			dokkuAppUser = results.StdoutContents()
		}
	}

	if len(dokkuAppUser) == 0 {
		dokkuAppUser = "herokuishuser"
	}

	output, err := DockerInspect(image, fmt.Sprintf("{{range .Config.Env}}{{if eq . \"USER=%s\" }}{{println .}}{{end}}{{end}}", dokkuAppUser))
	if err != nil {
		return false
	}
	return output != ""
}

// ListDanglingImages lists all dangling image ids for a given app
func ListDanglingImages(appName string) ([]string, error) {
	filters := []string{"dangling=true"}
	if appName != "" {
		filters = append(filters, []string{fmt.Sprintf("label=com.dokku.app-name=%v", appName)}...)
	}
	return DockerFilterImages(filters)
}

// RemoveImages removes images by ID
func RemoveImages(imageIDs []string) error {
	if len(imageIDs) == 0 {
		return nil
	}

	args := []string{
		"image",
		"rm",
	}

	args = append(args, imageIDs...)

	result, err := CallExecCommand(ExecCommandInput{
		Command: DockerBin(),
		Args:    args,
	})
	if err != nil {
		return fmt.Errorf("Unable to remove images: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Unable to remove images: %s", result.StderrContents())
	}

	return nil
}

// VerifyImage returns true if docker image exists in local repo
func VerifyImage(image string) bool {
	result, err := CallExecCommand(ExecCommandInput{
		Command: DockerBin(),
		Args:    []string{"image", "inspect", image},
	})
	return err == nil && result.ExitCode == 0
}

// DockerFilterContainers returns a slice of container IDs based on the passed in filters
func DockerFilterContainers(filters []string) ([]string, error) {
	args := []string{
		"container",
		"ls",
		"--quiet",
		"--all",
	}

	for _, filter := range filters {
		args = append(args, "--filter", filter)
	}

	result, err := CallExecCommand(ExecCommandInput{
		Command: DockerBin(),
		Args:    args,
	})
	if err != nil {
		return []string{}, fmt.Errorf("Unable to filter containers: %w", err)
	}
	if result.ExitCode != 0 {
		return []string{}, fmt.Errorf("Unable to filter containers: %s", result.StderrContents())
	}

	output := strings.Split(result.StdoutContents(), "\n")
	return output, nil
}

// DockerFilterImages returns a slice of image IDs based on the passed in filters
func DockerFilterImages(filters []string) ([]string, error) {
	args := []string{
		"image",
		"ls",
		"--quiet",
		"--all",
	}

	for _, filter := range filters {
		args = append(args, "--filter", filter)
	}

	result, err := CallExecCommand(ExecCommandInput{
		Command: DockerBin(),
		Args:    args,
	})
	if err != nil {
		return []string{}, fmt.Errorf("Unable to filter images: %w", err)
	}
	if result.ExitCode != 0 {
		return []string{}, fmt.Errorf("Unable to filter images: %s", result.StderrContents())
	}

	output := strings.Split(result.StdoutContents(), "\n")
	return output, nil
}

func listContainers(status string, appName string) ([]string, error) {
	filters := []string{
		fmt.Sprintf("status=%v", status),
		fmt.Sprintf("label=%v", os.Getenv("DOKKU_CONTAINER_LABEL")),
	}

	if appName != "" {
		filters = append(filters, fmt.Sprintf("label=com.dokku.app-name=%v", appName))
	}
	return DockerFilterContainers(filters)
}

func pruneUnusedImages(appName string) {
	args := []string{
		"image",
		"prune",
		"--all",
		"--force",
		"--filter",
		fmt.Sprintf("label=com.dokku.app-name=%v", appName),
	}

	CallExecCommand(ExecCommandInput{
		Command: DockerBin(),
		Args:    args,
	}) // nolint: errcheck
}

// DockerRemoveContainers will call `docker container rm` on the specified containers
func DockerRemoveContainers(containerIDs []string) {
	args := []string{
		"container",
		"rm",
	}

	args = append(args, containerIDs...)

	CallExecCommand(ExecCommandInput{
		Command: DockerBin(),
		Args:    args,
	}) // nolint: errcheck
}
