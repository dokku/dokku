package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/ryanuber/columnize"
	"golang.org/x/sync/errgroup"
)

type errfunc func() error

var (
	// DefaultProperties is a map of all valid common properties with corresponding default property values
	DefaultProperties = map[string]string{
		"deployed": "false",
	}

	// GlobalProperties is a map of all valid global common properties
	GlobalProperties = map[string]bool{}
)

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

// GetAppScheduler fetches the scheduler for a given application
func GetAppScheduler(appName string) string {
	appScheduler := ""
	globalScheduler := ""

	ctx := context.Background()
	errs, ctx := errgroup.WithContext(ctx)

	if appName != "--global" {
		errs.Go(func() error {
			appScheduler = getAppScheduler(appName)
			return nil
		})
	}
	errs.Go(func() error {
		globalScheduler = GetGlobalScheduler()
		return nil
	})
	errs.Wait()

	if appScheduler == "" {
		appScheduler = globalScheduler
	}
	return appScheduler
}

func getAppScheduler(appName string) string {
	b, _ := PlugnTriggerOutput("scheduler-detect", []string{appName}...)
	value := strings.TrimSpace(string(b[:]))
	if value != "" {
		return value
	}
	return ""
}

// GetGlobalScheduler fetchs the global scheduler
func GetGlobalScheduler() string {
	b, _ := PlugnTriggerOutput("scheduler-detect", []string{"--global"}...)
	value := strings.TrimSpace(string(b[:]))
	if value != "" {
		return value
	}

	return "docker-local"
}

// GetDeployingAppImageName returns deploying image identifier for a given app, tag tuple. validate if tag is presented
func GetDeployingAppImageName(appName, imageTag, imageRepo string) (string, error) {
	imageRemoteRepository := ""
	newImageTag := ""
	newImageRepo := ""

	ctx := context.Background()
	errs, ctx := errgroup.WithContext(ctx)
	errs.Go(func() error {
		b, err := PlugnTriggerOutput("deployed-app-repository", []string{appName}...)
		if err == nil {
			imageRemoteRepository = strings.TrimSpace(string(b[:]))
		}
		return err
	})
	errs.Go(func() error {
		b, err := PlugnTriggerOutput("deployed-app-image-tag", []string{appName}...)
		if err == nil {
			newImageTag = strings.TrimSpace(string(b[:]))
		}
		return err
	})

	errs.Go(func() error {
		b, err := PlugnTriggerOutput("deployed-app-image-repo", []string{appName}...)
		if err == nil {
			newImageRepo = strings.TrimSpace(string(b[:]))
		}
		return err
	})

	if err := errs.Wait(); err != nil {
		return "", err
	}

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

	imageName := fmt.Sprintf("%s%s:%s", imageRemoteRepository, imageRepo, imageTag)
	if !VerifyImage(imageName) {
		return "", fmt.Errorf("App image (%s) not found", imageName)
	}
	return imageName, nil
}

// GetAppImageRepo is the central definition of a dokku image repo pattern
func GetAppImageRepo(appName string) string {
	return strings.Join([]string{"dokku", appName}, "/")
}

// GetAppContainerIDs returns a list of docker container ids for given app and optional container_type
func GetAppContainerIDs(appName string, containerType string) ([]string, error) {
	var containerIDs []string
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

// GetAppImageName returns image identifier for a given app, tag tuple. validate if tag is presented
func GetAppImageName(appName, imageTag, imageRepo string) (imageName string) {
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
	deployed := PropertyGetDefault("common", appName, "deployed", "")
	if deployed == "" {
		deployed = "false"
		scheduler := GetAppScheduler(appName)
		_, err := PlugnTriggerOutput("scheduler-is-deployed", []string{scheduler, appName}...)
		if err == nil {
			deployed = "true"
		}

		CommandPropertySet("common", appName, "deployed", deployed, DefaultProperties, GlobalProperties)
	}

	return deployed == "true"
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

// ParseReportArgs splits out flags from non-flags for input into report commands
func ParseReportArgs(pluginName string, arguments []string) ([]string, string, error) {
	osArgs := []string{}
	infoFlags := []string{}
	skipNext := false
	for i, argument := range arguments {
		if skipNext {
			skipNext = false
			continue
		}
		if argument == "--format" {
			osArgs = append(osArgs, argument, arguments[i+1])
			skipNext = true
			continue
		}
		if strings.HasPrefix(argument, "--") {
			infoFlags = append(infoFlags, argument)
		} else {
			osArgs = append(osArgs, argument)
		}
	}

	if len(infoFlags) == 0 {
		return osArgs, "", nil
	}
	if len(infoFlags) == 1 {
		return osArgs, infoFlags[0], nil
	}
	return osArgs, "", fmt.Errorf("%s:report command allows only a single flag", pluginName)
}

// ParseScaleOutput allows golang plugins to properly parse the output of ps-current-scale
func ParseScaleOutput(b []byte) (map[string]int, error) {
	scale := make(map[string]int)

	for _, line := range strings.Split(string(b), "\n") {
		s := strings.Split(line, "=")
		processType := s[0]
		count, err := strconv.Atoi(s[1])
		if err != nil {
			return scale, err
		}
		scale[processType] = count
	}

	return scale, nil
}

// ReportSingleApp is an internal function that displays a report for an app
func ReportSingleApp(reportType string, appName string, infoFlag string, infoFlags map[string]string, infoFlagKeys []string, format string, trimPrefix bool, uppercaseFirstCharacter bool) error {
	if format != "stdout" && infoFlag != "" {
		return errors.New("--format flag cannot be specified when specifying an info flag")
	}

	if format == "json" {
		data := map[string]string{}
		for key, value := range infoFlags {
			prefix := "--"
			if trimPrefix {
				prefix = fmt.Sprintf("--%v-", reportType)
			}

			// key = strings.Replace(strings.Replace(strings.TrimPrefix(key, prefix), "-", " ", -1), ".", " ", -1)
			data[strings.TrimPrefix(key, prefix)] = value
		}
		out, err := json.Marshal(data)
		if err != nil {
			return err
		}
		Log(string(out))
		return nil
	}

	length := 0
	flags := []string{}
	for key := range infoFlags {
		if len(key) > length {
			length = len(key)
		}
		flags = append(flags, key)
	}
	sort.Strings(flags)
	if length < 31 {
		length = 31
	}

	if len(infoFlag) == 0 {
		LogInfo2Quiet(fmt.Sprintf("%s %v information", appName, reportType))
		for _, k := range flags {
			v, ok := infoFlags[k]
			if !ok {
				continue
			}

			prefix := "--"
			if trimPrefix {
				prefix = fmt.Sprintf("--%v-", reportType)
			}

			key := strings.Replace(strings.Replace(strings.TrimPrefix(k, prefix), "-", " ", -1), ".", " ", -1)

			if uppercaseFirstCharacter {
				key = UcFirst(key)
			}

			LogVerbose(fmt.Sprintf("%s%s", RightPad(fmt.Sprintf("%s:", key), length, " "), v))
		}
		return nil
	}

	for _, k := range flags {
		if infoFlag == k {
			v, ok := infoFlags[k]
			if !ok {
				continue
			}
			fmt.Println(v)
			return nil
		}
	}

	sort.Strings(infoFlagKeys)
	return fmt.Errorf("Invalid flag passed, valid flags: %s", strings.Join(infoFlagKeys, ", "))
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
	b := []byte(text)
	re := regexp.MustCompile("(?s)#.*")
	b = re.ReplaceAll(b, nil)
	return strings.TrimSpace(string(b))
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

// ToInt returns an int value for a given string
func ToInt(s string, defaultValue int) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}

	return i
}

// UcFirst uppercases the first character in a string
func UcFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}

// IsValidAppName verifies that the app name matches naming restrictions
func IsValidAppName(appName string) error {
	if appName == "" {
		return errors.New("Please specify an app to run the command on")
	}

	r, _ := regexp.Compile("^[a-z0-9][^/:_A-Z]*$")
	if r.MatchString(appName) {
		return nil
	}

	return errors.New("App name must begin with lowercase alphanumeric character, and cannot include uppercase characters, colons, or underscores")
}

// isValidAppNameOld verifies that the app name matches the old naming restrictions
func isValidAppNameOld(appName string) error {
	if appName == "" {
		return errors.New("Please specify an app to run the command on")
	}

	r, _ := regexp.Compile("^[a-z0-9][^/:A-Z]*$")
	if r.MatchString(appName) {
		return nil
	}

	return errors.New("App name must begin with lowercase alphanumeric character, and cannot include uppercase characters, or colons")
}

// AppDoesNotExist wraps error to include the app name
// and is used to distinguish between a normal error and an error
// where the app is missing
type AppDoesNotExist struct {
	appName string
}

// ExitCode returns an exit code to use in case this error bubbles
// up into an os.Exit() call
func (err *AppDoesNotExist) ExitCode() int {
	return 20
}

// Error returns a standard non-existent app error
func (err *AppDoesNotExist) Error() string {
	return fmt.Sprintf("App %s does not exist", err.appName)
}

// VerifyAppName checks if an app conforming to either the old or new
// naming conventions exists
func VerifyAppName(appName string) error {
	newErr := IsValidAppName(appName)
	oldErr := isValidAppNameOld(appName)
	if newErr != nil && oldErr != nil {
		return newErr
	}

	appRoot := AppRoot(appName)
	if !DirectoryExists(appRoot) {
		return &AppDoesNotExist{appName}
	}

	return nil
}

func times(str string, n int) (out string) {
	for i := 0; i < n; i++ {
		out += str
	}
	return
}
