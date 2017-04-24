package common

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"

	sh "github.com/codeskyblue/go-sh"
)

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
	return &ShellCmd{
		Command:       exec.Command(cmd, args...),
		CommandString: command,
		Args:          args,
		ShowOutput:    true,
	}
}

// Execute is a lightweight wrapper around exec.Command
func (sc *ShellCmd) Execute() bool {
	env := os.Environ()
	for k, v := range sc.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	sc.Command.Env = env
	if sc.ShowOutput {
		sc.Command.Stdout = os.Stdout
		sc.Command.Stderr = os.Stderr
	}
	err := sc.Command.Run()
	if err != nil {
		return false
	}
	return true
}

// Output is a lightweight wrapper around exec.Command.Output()
func (sc *ShellCmd) Output() ([]byte, error) {
	env := os.Environ()
	for k, v := range sc.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	sc.Command.Env = env
	if sc.ShowOutput {
		sc.Command.Stdout = os.Stdout
		sc.Command.Stderr = os.Stderr
	}
	return sc.Command.Output()
}

// VerifyAppName verifies app name format and app existence"
func VerifyAppName(appName string) (err error) {
	if appName == "" {
		return fmt.Errorf("App name must not be null")
	}
	dokkuRoot := MustGetEnv("DOKKU_ROOT")
	appRoot := strings.Join([]string{dokkuRoot, appName}, "/")
	if !DirectoryExists(appRoot) {
		return fmt.Errorf("App %s does not exist: %v\n", appName, err)
	}
	r, _ := regexp.Compile("^[a-z].*")
	if !r.MatchString(appName) {
		return fmt.Errorf("App name (%s) must begin with lowercase alphanumeric character\n", appName)
	}
	return err
}

// MustGetEnv returns env variable or fails if it's not set
func MustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		LogFail(fmt.Sprintf("%s not set!", key))
	}
	return value
}

// LogFail is the failure log formatter
// prints text to stderr and exits with status 1
func LogFail(text string) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf("FAILED: %s", text))
	os.Exit(1)
}

// GetDeployingAppImageName returns deploying image identifier for a given app, tag tuple. validate if tag is presented
func GetDeployingAppImageName(appName, imageTag, imageRepo string) (imageName string) {
	if appName == "" {
		LogFail("(GetDeployingAppImageName) APP must not be empty")
	}

	b, err := sh.Command("plugn", "trigger", "deployed-app-repository", appName).Output()
	if err != nil {
		LogFail(err.Error())
	}
	imageRemoteRepository := string(b[:])

	b, err = sh.Command("plugn", "trigger", "deployed-app-image-tag", appName).Output()
	if err != nil {
		LogFail(err.Error())
	}
	newImageTag := string(b[:])

	b, err = sh.Command("plugn", "trigger", "deployed-app-image-repo", appName).Output()
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
		LogFail(fmt.Sprintf("app image (%s) not found", imageName))
	}
	return
}

// GetAppImageRepo is the central definition of a dokku image repo pattern
func GetAppImageRepo(appName string) string {
	return strings.Join([]string{"dokku", appName}, "/")
}

// VerifyImage returns true if docker image exists in local repo
func VerifyImage(image string) bool {
	imageCmd := NewShellCmd(strings.Join([]string{"docker inspect", image}, " "))
	imageCmd.ShowOutput = false
	return imageCmd.Execute()
}

// ContainerIsRunning checks to see if a container is running
func ContainerIsRunning(containerId string) bool {
	b, err := sh.Command("docker", "inspect", "--format", "'{{.State.Running}}'", containerId).Output()
	if err != nil {
		return false
	}
	return string(b[:]) == "true"
}

// DirectoryExists returns if a path exists and is a directory
func DirectoryExists(filePath string) bool {
	fi, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	return fi.IsDir()
}

// DokkuApps returns a list of all local apps
func DokkuApps() ([]string, error) {
	var apps []string

	dokkuRoot := MustGetEnv("DOKKU_ROOT")
	files, err := ioutil.ReadDir(dokkuRoot)
	if err != nil {
		return apps, errors.New("You haven't deployed any applications yet")
	}

	for _, f := range files {
		appRoot := strings.Join([]string{dokkuRoot, f.Name()}, "/")
		if !DirectoryExists(appRoot) {
			continue
		}
		if f.Name() == "tls" || strings.HasPrefix(f.Name(), ".") {
			continue
		}
		apps = append(apps, f.Name())
	}

	if len(apps) == 0 {
		return apps, errors.New("You haven't deployed any applications yet")
	}

	return apps, nil
}

// FileToSlice reads in all the lines from a file into a string slice
func FileToSlice(filePath string) ([]string, error) {
	var lines []string
	f, err := os.Open(filePath)
	if err != nil {
		return lines, err
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
	return lines, err
}

// FileExists returns if a path exists and is a file
func FileExists(filePath string) bool {
	fi, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	return fi.Mode().IsRegular()
}

// GetAppImageName returnS image identifier for a given app, tag tuple. validate if tag is presented
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
			LogFail(fmt.Sprintf("app image (%s) not found", imageName))
		}
	}
	return
}

// return true if given app has a running container
func IsDeployed(appName string) bool {
	dokkuRoot := MustGetEnv("DOKKU_ROOT")
	appRoot := strings.Join([]string{dokkuRoot, appName}, "/")
	files, err := ioutil.ReadDir(appRoot)
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
func IsImageHerokuishBased(image string) bool {
	// circleci can't support --rm as they run lxc in lxc
	dockerArgs := ""
	if !FileExists("/home/ubuntu/.circlerc") {
		dockerArgs = "--rm"
	}

	dockerGlobalArgs := os.Getenv("DOKKU_GLOBAL_RUN_ARGS")
	parts := []string{"docker", "run", dockerGlobalArgs, "--entrypoint=\"/bin/sh\"", dockerArgs, image, "-c", "\"test -f /exec\""}

	var dockerCmdParts []string
	for _, str := range parts {
		if str != "" {
			dockerCmdParts = append(dockerCmdParts, str)
		}
	}

	dockerCmd := NewShellCmd(strings.Join(dockerCmdParts, " "))
	dockerCmd.ShowOutput = false
	return dockerCmd.Execute()
}

// LogInfo1 is the info1 header formatter
func LogInfo1(text string) {
	fmt.Fprintln(os.Stdout, fmt.Sprintf("-----> %s", text))
}

// LogInfo2Quiet is the info1 header formatter (with quiet option)
func LogInfo1Quiet(text string) {
	if os.Getenv("DOKKU_QUIET_OUTPUT") != "" {
		LogInfo1(text)
	}
}

// LogInfo2 is the info2 header formatter
func LogInfo2(text string) {
	fmt.Fprintln(os.Stdout, fmt.Sprintf("=====> %s", text))
}

// LogInfo2Quiet is the info2 header formatter (with quiet option)
func LogInfo2Quiet(text string) {
	if os.Getenv("DOKKU_QUIET_OUTPUT") != "" {
		LogInfo2(text)
	}
}

// LogWarn is the warning log formatter
func LogWarn(text string) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf(" !     %s", text))
}

// ReadFirstLine gets the first line of a file that has contents and returns it
// if there are no contents, an empty string is returned
// will also return an empty string if the file does not exist
func ReadFirstLine(filename string) string {
	if !FileExists(filename) {
		return ""
	}
	f, err := os.Open(filename)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		return text
	}
	return ""

}

// StripInlineComments removes bash-style comment from input line
func StripInlineComments(text string) string {
	var bytes = []byte(text)
	re := regexp.MustCompile("(?s)#.*")
	bytes = re.ReplaceAll(bytes, nil)
	return strings.TrimSpace(string(bytes))
}

// ToBool returns a bool value for a given string
func ToBool(s string) bool {
	return s == "true"
}
