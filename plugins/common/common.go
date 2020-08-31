package common

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"unicode"

	sh "github.com/codeskyblue/go-sh"
	columnize "github.com/ryanuber/columnize"
)

type errfunc func() error

// ShellCmd represents a shell command to be run for dokku
type ShellCmd struct {
	Env           map[string]string
	Command       *exec.Cmd
	CommandString string
	Args          []string
	ShowOutput    bool
}

// NewShellCmd returns a new ShellCmd struct
func NewShellCmd(command string) *ShellCmd {
	items := strings.Split(command, " ")
	cmd := items[0]
	args := items[1:]
	return NewShellCmdWithArgs(cmd, args...)
}

// NewShellCmdWithArgs returns a new ShellCmd struct
func NewShellCmdWithArgs(cmd string, args ...string) *ShellCmd {
	commandString := strings.Join(append([]string{cmd}, args...), " ")

	return &ShellCmd{
		Command:       exec.Command(cmd, args...),
		CommandString: commandString,
		Args:          args,
		ShowOutput:    true,
	}
}

func (sc *ShellCmd) setup() {
	env := os.Environ()
	for k, v := range sc.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	sc.Command.Env = env
	if sc.ShowOutput {
		sc.Command.Stdout = os.Stdout
		sc.Command.Stderr = os.Stderr
	}
}

// Execute is a lightweight wrapper around exec.Command
func (sc *ShellCmd) Execute() bool {
	sc.setup()

	if err := sc.Command.Run(); err != nil {
		return false
	}
	return true
}

// Output is a lightweight wrapper around exec.Command.Output()
func (sc *ShellCmd) Output() ([]byte, error) {
	sc.setup()
	return sc.Command.Output()
}

// CombinedOutput is a lightweight wrapper around exec.Command.CombinedOutput()
func (sc *ShellCmd) CombinedOutput() ([]byte, error) {
	sc.setup()
	return sc.Command.CombinedOutput()
}

// AppRoot returns the app root path
func AppRoot(appName string) string {
	dokkuRoot := MustGetEnv("DOKKU_ROOT")
	return fmt.Sprintf("%v/%v", dokkuRoot, appName)
}

// AppHostRoot returns the app root path
func AppHostRoot(appName string) string {
	dokkuHostRoot := MustGetEnv("DOKKU_HOST_ROOT")
	return fmt.Sprintf("%v/%v", dokkuHostRoot, appName)
}

// AskForDestructiveConfirmation checks for confirmation on destructive actions
func AskForDestructiveConfirmation(name string, objectType string) error {
	LogWarn("WARNING: Potentially Destructive Action")
	LogWarn(fmt.Sprintf("This command will destroy %v %v.", objectType, name))
	LogWarn(fmt.Sprintf("To proceed, type \"%v\"", name))
	fmt.Print("> ")
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		return err
	}

	if response != name {
		LogStderr("Confirmation did not match test. Aborted.")
		os.Exit(1)
		return nil
	}

	return nil
}

// CommandUsage outputs help for a command
func CommandUsage(helpHeader string, helpContent string) {
	config := columnize.DefaultConfig()
	config.Delim = ","
	config.Prefix = "    "
	config.Empty = ""
	content := strings.Split(helpContent, "\n")[1:]
	fmt.Println(helpHeader)
	fmt.Println(columnize.Format(content, config))
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

// DockerCleanup cleans up all exited/dead containers and removes all dangling images
func DockerCleanup(appName string, forceCleanup bool) error {
	if !forceCleanup {
		skipCleanup := false
		if appName != "" {
			b, _ := PlugnTriggerOutput("config-get", []string{appName, "DOKKU_SKIP_CLEANUP"}...)
			output := string(b[:])
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

// GetAppScheduler fetches the scheduler for a given application
func GetAppScheduler(appName string) string {
	if appName == "--global" {
		appName = ""
	}

	b, _ := PlugnTriggerOutput("config-get", []string{appName, "DOKKU_SCHEDULER"}...)
	value := string(b[:])
	if value != "" {
		return value
	}

	b, _ = PlugnTriggerOutput("config-get-global", []string{"DOKKU_SCHEDULER"}...)
	value = string(b[:])
	if value != "" {
		return value
	}

	return "docker-local"
}

// GetDeployingAppImageName returns deploying image identifier for a given app, tag tuple. validate if tag is presented
func GetDeployingAppImageName(appName, imageTag, imageRepo string) (imageName string) {
	if appName == "" {
		LogFail("(GetDeployingAppImageName) APP must not be empty")
	}

	b, err := PlugnTriggerOutput("deployed-app-repository", []string{appName}...)
	if err != nil {
		LogFail(err.Error())
	}
	imageRemoteRepository := string(b[:])

	b, err = PlugnTriggerOutput("deployed-app-image-tag", []string{appName}...)
	if err != nil {
		LogFail(err.Error())
	}
	newImageTag := string(b[:])

	b, err = PlugnTriggerOutput("deployed-app-image-repo", []string{appName}...)
	if err != nil {
		LogFail(err.Error())
	}
	newImageRepo := string(b[:])

	if newImageRepo != "" {
		imageRepo = newImageRepo
	}
	if newImageTag != "" {
		imageTag = newImageTag
	}
	if imageRepo == "" {
		imageRepo = GetAppImageRepo(appName)
	}
	if imageTag == "" {
		imageTag = "latest"
	}

	imageName = fmt.Sprintf("%s%s:%s", imageRemoteRepository, imageRepo, imageTag)
	if !VerifyImage(imageName) {
		LogFail(fmt.Sprintf("App image (%s) not found", imageName))
	}
	return
}

// GetAppImageRepo is the central definition of a dokku image repo pattern
func GetAppImageRepo(appName string) string {
	return strings.Join([]string{"dokku", appName}, "/")
}

// GetAppContainerIDs returns a list of docker container ids for given app and optional container_type
func GetAppContainerIDs(appName string, containerType string) ([]string, error) {
	var containerIDs []string
	if err := VerifyAppName(appName); err != nil {
		return containerIDs, err
	}

	appRoot := AppRoot(appName)
	containerFilePath := fmt.Sprintf("%v/CONTAINER", appRoot)
	_, err := os.Stat(containerFilePath)
	if !os.IsNotExist(err) {
		containerIDs = append(containerIDs, ReadFirstLine(containerFilePath))
	}

	containerPattern := fmt.Sprintf("%v/CONTAINER.*", appRoot)
	if containerType != "" {
		containerPattern = fmt.Sprintf("%v/CONTAINER.%v.*", appRoot, containerType)
		if strings.Contains(".", containerType) {
			containerPattern = fmt.Sprintf("%v/CONTAINER.%v", appRoot, containerType)
		}
	}

	files, _ := filepath.Glob(containerPattern)
	for _, containerFile := range files {
		containerIDs = append(containerIDs, ReadFirstLine(containerFile))
	}

	return containerIDs, nil
}

// GetAppRunningContainerIDs return a list of running docker container ids for given app and optional container_type
func GetAppRunningContainerIDs(appName string, containerType string) ([]string, error) {
	var runningContainerIDs []string
	if err := VerifyAppName(appName); err != nil {
		return runningContainerIDs, err
	}

	if !IsDeployed(appName) {
		LogFail(fmt.Sprintf("App %v has not been deployed", appName))
	}

	containerIDs, err := GetAppContainerIDs(appName, containerType)
	if err != nil {
		return runningContainerIDs, nil
	}
	for _, containerID := range containerIDs {
		if ContainerIsRunning(containerID) {
			runningContainerIDs = append(runningContainerIDs, containerID)
		}
	}

	return runningContainerIDs, nil
}

// GetRunningImageTag retrieves current image tag for a given app and returns empty string if no deployed containers are found
func GetRunningImageTag(appName string) (string, error) {
	if err := VerifyAppName(appName); err != nil {
		return "", err
	}

	containerIDs, err := GetAppContainerIDs(appName, "")
	if err != nil {
		return "", err
	}

	for _, containerID := range containerIDs {
		if image, err := DockerInspect(containerID, "{{ .Config.Image }}"); err == nil {
			return strings.Split(image, ":")[1], nil
		}
	}

	return "", errors.New("No image tag found")
}

// ContainerIsRunning checks to see if a container is running
func ContainerIsRunning(containerID string) bool {
	b, err := DockerInspect(containerID, "'{{.State.Running}}'")
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(b[:])) == "true"
}

// DirectoryExists returns if a path exists and is a directory
func DirectoryExists(filePath string) bool {
	fi, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	return fi.IsDir()
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

// DokkuApps returns a list of all local apps
func DokkuApps() (apps []string, err error) {
	dokkuRoot := MustGetEnv("DOKKU_ROOT")
	files, err := ioutil.ReadDir(dokkuRoot)
	if err != nil {
		err = fmt.Errorf("You haven't deployed any applications yet")
		return
	}

	for _, f := range files {
		appRoot := AppRoot(f.Name())
		if !DirectoryExists(appRoot) {
			continue
		}
		if strings.HasPrefix(f.Name(), ".") {
			continue
		}
		apps = append(apps, f.Name())
	}

	if len(apps) == 0 {
		err = fmt.Errorf("You haven't deployed any applications yet")
		return
	}

	return
}

// FileToSlice reads in all the lines from a file into a string slice
func FileToSlice(filePath string) (lines []string, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		lines = append(lines, text)
	}
	err = scanner.Err()
	return
}

// FileExists returns if a path exists and is a file
func FileExists(filePath string) bool {
	fi, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	return fi.Mode().IsRegular()
}

// GetAppImageName returns image identifier for a given app, tag tuple. validate if tag is presented
func GetAppImageName(appName, imageTag, imageRepo string) (imageName string) {
	err := VerifyAppName(appName)
	if err != nil {
		LogFail(err.Error())
	}

	if imageRepo == "" {
		imageRepo = GetAppImageRepo(appName)
	}

	if imageTag == "" {
		imageName = fmt.Sprintf("%v:latest", imageRepo)
	} else {
		imageName = fmt.Sprintf("%v:%v", imageRepo, imageTag)
		if !VerifyImage(imageName) {
			LogFail(fmt.Sprintf("App image (%s) not found", imageName))
		}
	}
	return
}

// IsAbsPath returns 0 if input path is absolute
func IsAbsPath(path string) bool {
	return strings.HasPrefix(path, "/")
}

// IsDeployed returns true if given app has a running container
func IsDeployed(appName string) bool {
	files, err := ioutil.ReadDir(AppRoot(appName))
	if err != nil {
		return false
	}

	for _, f := range files {
		if f.Name() == "CONTAINER" || strings.HasPrefix(f.Name(), "CONTAINER.") {
			return true
		}
	}
	return false
}

// IsImageHerokuishBased returns true if app image is based on herokuish
func IsImageHerokuishBased(image string, appName string) bool {
	output, err := DockerInspect(image, "{{range .Config.Env}}{{if eq . \"USER=herokuishuser\" }}{{println .}}{{end}}{{end}}")
	if err != nil {
		return false
	}
	return output != ""
}

// MustGetEnv returns env variable or fails if it's not set
func MustGetEnv(key string) (val string) {
	val = os.Getenv(key)
	if val == "" {
		LogFail(fmt.Sprintf("%s not set!", key))
	}
	return
}

// GetenvWithDefault returns env variable or defaultValue if it's not set
func GetenvWithDefault(key string, defaultValue string) (val string) {
	val = os.Getenv(key)
	if val == "" {
		val = defaultValue
	}
	return
}

// ReadFirstLine gets the first line of a file that has contents and returns it
// if there are no contents, an empty string is returned
// will also return an empty string if the file does not exist
func ReadFirstLine(filename string) (text string) {
	if !FileExists(filename) {
		return
	}
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if text = strings.TrimSpace(scanner.Text()); text == "" {
			continue
		}
		return
	}
	return
}

// ReportSingleApp is an internal function that displays a report for an app
func ReportSingleApp(reportType string, appName string, infoFlag string, infoFlags map[string]string, trimPrefix bool, uppercaseFirstCharacter bool) error {
	flags := []string{}
	for key := range infoFlags {
		flags = append(flags, key)
	}
	sort.Strings(flags)

	if len(infoFlag) == 0 {
		LogInfo2Quiet(fmt.Sprintf("%s %v information", appName, reportType))
		for _, k := range flags {
			v := infoFlags[k]
			prefix := "--"
			if trimPrefix {
				prefix = fmt.Sprintf("--%v-", reportType)
			}

			key := strings.Replace(strings.Replace(strings.TrimPrefix(k, prefix), "-", " ", -1), ".", " ", -1)

			if uppercaseFirstCharacter {
				key = UcFirst(key)
			}

			LogVerbose(fmt.Sprintf("%s%s", RightPad(fmt.Sprintf("%s:", key), 31, " "), v))
		}
		return nil
	}

	for _, k := range flags {
		if infoFlag == k {
			v := infoFlags[k]
			fmt.Println(v)
			return nil
		}
	}

	keys := reflect.ValueOf(infoFlags).MapKeys()
	strkeys := make([]string, len(keys))
	for i := 0; i < len(keys); i++ {
		strkeys[i] = keys[i].String()
	}

	return fmt.Errorf("Invalid flag passed, valid flags: %s", strings.Join(strkeys, ", "))
}

// RightPad right-pads the string with pad up to len runes
func RightPad(str string, length int, pad string) string {
	return str + times(pad, length-len(str))
}

// ShiftString removes the first and returns that entry as well as the rest of the list
func ShiftString(a []string) (string, []string) {
	if len(a) == 0 {
		return "", a
	}

	return a[0], a[1:]
}

// StripInlineComments removes bash-style comment from input line
func StripInlineComments(text string) string {
	bytes := []byte(text)
	re := regexp.MustCompile("(?s)#.*")
	bytes = re.ReplaceAll(bytes, nil)
	return strings.TrimSpace(string(bytes))
}

// SuppressOutput suppresses the output of a function unless there is an error
func SuppressOutput(f errfunc) error {
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := f()

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout

	if err != nil {
		fmt.Printf(string(out[:]))
	}

	return err
}

// ToBool returns a bool value for a given string
func ToBool(s string) bool {
	return s == "true"
}

// UcFirst uppercases the first character in a string
func UcFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}

// IsValidAppName verifies app name format
func IsValidAppName(appName string) error {
	if appName == "" {
		return fmt.Errorf("APP must not be null")
	}

	r, _ := regexp.Compile("^[a-z0-9][^/:A-Z]*$")
	if r.MatchString(appName) {
		return nil
	}

	appRoot := AppRoot(appName)
	if DirectoryExists(appRoot) {
		os.RemoveAll(appRoot)
	}

	return errors.New("App name must begin with lowercase alphanumeric character, and cannot include uppercase characters or colons")
}

// VerifyAppName verifies app name format and app existence"
func VerifyAppName(appName string) error {
	if err := IsValidAppName(appName); err != nil {
		return err
	}

	appRoot := AppRoot(appName)
	if !DirectoryExists(appRoot) {
		return fmt.Errorf("App %s does not exist", appName)
	}

	return nil
}

// VerifyImage returns true if docker image exists in local repo
func VerifyImage(image string) bool {
	imageCmd := NewShellCmd(strings.Join([]string{DockerBin(), "image", "inspect", image}, " "))
	imageCmd.ShowOutput = false
	return imageCmd.Execute()
}

// DockerBin returns a string which contains a path to the current docker binary
func DockerBin() string {
	dockerBin := os.Getenv("DOCKER_BIN")
	if dockerBin == "" {
		dockerBin = "docker"
	}

	return dockerBin
}

// PlugnTrigger fire the given plugn trigger with the given args
func PlugnTrigger(triggerName string, args ...string) error {
	return PlugnTriggerSetup(triggerName, args...).Run()
}

// PlugnTriggerOutput fire the given plugn trigger with the given args
func PlugnTriggerOutput(triggerName string, args ...string) ([]byte, error) {
	return PlugnTriggerSetup(triggerName, args...).Output()
}

// PlugnTriggerSetup sets up a plugn trigger call
func PlugnTriggerSetup(triggerName string, args ...string) *sh.Session {
	shellArgs := make([]interface{}, len(args)+2)
	shellArgs[0] = "trigger"
	shellArgs[1] = triggerName
	for i, arg := range args {
		shellArgs[i+2] = arg
	}
	return sh.Command("plugn", shellArgs...)
}

func times(str string, n int) (out string) {
	for i := 0; i < n; i++ {
		out += str
	}
	return
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
