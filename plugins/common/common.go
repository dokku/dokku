package common

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"unicode"

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
	if err := sc.Command.Run(); err != nil {
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
		LogFail(fmt.Sprintf("App image (%s) not found", imageName))
	}
	return
}

// GetAppImageRepo is the central definition of a dokku image repo pattern
func GetAppImageRepo(appName string) string {
	return strings.Join([]string{"dokku", appName}, "/")
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

// DockerInspect runs an inspect command with a given format against a container id
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

// IsDeployed returns true if given app has a running container
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

// StripInlineComments removes bash-style comment from input line
func StripInlineComments(text string) string {
	bytes := []byte(text)
	re := regexp.MustCompile("(?s)#.*")
	bytes = re.ReplaceAll(bytes, nil)
	return strings.TrimSpace(string(bytes))
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

// VerifyAppName verifies app name format and app existence"
func VerifyAppName(appName string) (err error) {
	if appName == "" {
		return fmt.Errorf("App name must not be null")
	}
	dokkuRoot := MustGetEnv("DOKKU_ROOT")
	appRoot := strings.Join([]string{dokkuRoot, appName}, "/")
	if !DirectoryExists(appRoot) {
		return fmt.Errorf("app %s does not exist", appName)
	}
	r, _ := regexp.Compile("^[a-z0-9].*")
	if !r.MatchString(appName) {
		return fmt.Errorf("app name (%s) must begin with lowercase alphanumeric character", appName)
	}
	return err
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

//PlugnTrigger fire the given plugn trigger with the given args
func PlugnTrigger(triggerName string, args ...string) error {
	shellArgs := make([]interface{}, len(args)+2)
	shellArgs[0] = "trigger"
	shellArgs[1] = triggerName
	for i, arg := range args {
		shellArgs[i+2] = arg
	}
	return sh.Command("plugn", shellArgs...).Run()
}
