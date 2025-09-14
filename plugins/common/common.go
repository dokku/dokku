package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/otiai10/copy"
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

// CorePostDeployPath is a file or directory that was extracted
type CorePostDeployPath struct {
	// IsDirectory is whether the source is a directory
	IsDirectory bool

	// Path is the name of the file or directory
	Path string
}

// CorePostDeployInput is the input for the CorePostDeploy function
type CorePostDeployInput struct {
	// AppName is the name of the app
	AppName string

	// Destination is the destination directory
	Destination string

	// PluginName is the name of the plugin that is deploying the file or directory
	PluginName string

	// ExtractedPaths is the list of paths that were extracted
	ExtractedPaths []CorePostDeployPath
}

// CorePostDeploy moves extracted paths to the destination directory
// and removes any existing files or directories that were not extracted
//
//	CorePostDeploy(CorePostDeployInput{
//		AppName: "my-app",
//		Destination: "/var/lib/dokku/data/my-app",
//		ExtractedPaths: []CorePostDeployPath{
//			{Name: "app.json", IsDirectory: false},
//			{Name: "kustomization", IsDirectory: true},
//		},
//	})
func CorePostDeploy(input CorePostDeployInput) error {
	if input.PluginName == "" {
		return fmt.Errorf("Missing required PluginName in CorePostDeploy")
	}

	if input.AppName == "" {
		return fmt.Errorf("Missing required AppName in CorePostDeploy for plugin %v", input.PluginName)
	}

	if input.Destination == "" {
		return fmt.Errorf("Missing required Destination in CorePostDeploy for plugin %v", input.PluginName)
	}

	for i, extractedPath := range input.ExtractedPaths {
		if extractedPath.Path == "" {
			return fmt.Errorf("Missing required Name in CorePostDeploy for index %v for plugin %v", i, input.PluginName)
		}

		existingPath := filepath.Join(input.Destination, extractedPath.Path)
		processSpecificPath := fmt.Sprintf("%s.%s", existingPath, os.Getenv("DOKKU_PID"))

		if extractedPath.IsDirectory {
			if DirectoryExists(processSpecificPath) {
				if err := os.RemoveAll(existingPath); err != nil {
					return err
				}

				if err := os.Rename(processSpecificPath, existingPath); err != nil {
					return err
				}
			} else if DirectoryExists(fmt.Sprintf("%s.missing", processSpecificPath)) {
				if err := os.RemoveAll(fmt.Sprintf("%s.missing", processSpecificPath)); err != nil {
					return err
				}

				if DirectoryExists(existingPath) {
					if err := os.RemoveAll(existingPath); err != nil {
						return err
					}
				}
			}
		} else {
			if FileExists(processSpecificPath) {
				if err := os.Rename(processSpecificPath, existingPath); err != nil {
					return err
				}
			} else if FileExists(fmt.Sprintf("%s.missing", processSpecificPath)) {
				if err := os.Remove(fmt.Sprintf("%s.missing", processSpecificPath)); err != nil {
					return err
				}

				if FileExists(existingPath) {
					if err := os.Remove(existingPath); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// CorePostExtractValidator is a function that validates the file or directory
type CorePostExtractValidator func(appName string, path string) error

// CorePostExtractToExtract is a file or directory to extract
type CorePostExtractToExtract struct {
	// Destination is an optional alias destination path
	// If not provided, the Path will be used as the destination
	Destination string

	// IsDirectory is whether the source is a directory
	IsDirectory bool

	// Name is the common name of the file or directory to extract
	Name string

	// Path is the path to the file or directory to extract
	Path string

	// Validator is a function that validates the file or directory
	Validator CorePostExtractValidator
}

// CorePostExtractInput is the input for the CorePostExtract function
type CorePostExtractInput struct {
	// AppName is the name of the app
	AppName string

	// BuildDir is the optional build directory to extract from
	BuildDir string

	// DestinationDir is the destination directory
	Destination string

	// PluginName is the name of the plugin that is extracting the file or directory
	PluginName string

	// SourceWorkDir is the source work directory
	SourceWorkDir string

	// ToExtract is a list of files or directories to extract
	ToExtract []CorePostExtractToExtract
}

// CorePostExtract extracts files or directories from a source work directory to a destination directory
//
//	CorePostExtract(CorePostExtractInput{
//		AppName: "my-app",
//		SourceWorkDir: "/tmp/my-app-source",
//		Destination: "/var/lib/dokku/data/my-app",
//		ToExtract: []CorePostExtractToExtract{
//			{Path: "app2.json", IsDirectory: false, Name: "app.json"},
//			{Path: "config/kustomize", IsDirectory: true, Destination: "kustomization"},
//		},
//	})
func CorePostExtract(input CorePostExtractInput) error {
	if input.PluginName == "" {
		return fmt.Errorf("Missing required PluginName in CorePostExtract")
	}

	if input.AppName == "" {
		return fmt.Errorf("Missing required AppName in CorePostExtract for plugin %v", input.PluginName)
	}

	if input.Destination == "" {
		return fmt.Errorf("Missing required Destination in CorePostExtract for plugin %v", input.PluginName)
	}

	if input.SourceWorkDir == "" {
		return fmt.Errorf("Missing required SourceWorkDir in CorePostExtract for plugin %v", input.PluginName)
	}

	results, _ := CallPlugnTrigger(PlugnTriggerInput{
		Trigger: "git-get-property",
		Args:    []string{input.AppName, "source-image"},
	})
	sourceImage := results.StdoutContents()

	for i, toExtract := range input.ToExtract {
		if toExtract.Name == "" {
			return fmt.Errorf("Name is required for index %v in CorePostExtract for plugin %v", i, input.PluginName)
		}

		if toExtract.Path == "" {
			return fmt.Errorf("Path is required for index %v in CorePostExtract for plugin %v", i, input.PluginName)
		}

		if toExtract.Destination == "" {
			toExtract.Destination = toExtract.Path
		}

		sourcePath := filepath.Join(input.SourceWorkDir, toExtract.Path)
		repoDefaultSourcePath := filepath.Join(input.SourceWorkDir, toExtract.Name)
		imageSourcePath := toExtract.Path
		if input.BuildDir != "" {
			sourcePath = filepath.Join(input.SourceWorkDir, input.BuildDir, toExtract.Path)
			repoDefaultSourcePath = filepath.Join(input.SourceWorkDir, input.BuildDir, toExtract.Name)
			imageSourcePath = filepath.Join(input.BuildDir, toExtract.Path)
		}

		destination := filepath.Join(input.Destination, toExtract.Destination)
		processSpecificDestination := fmt.Sprintf("%s.%s", destination, os.Getenv("DOKKU_PID"))
		missingDestination := fmt.Sprintf("%s.missing", processSpecificDestination)
		files, err := filepath.Glob(fmt.Sprintf("%s.*", destination))
		if err != nil {
			return err
		}
		for _, f := range files {
			if err := os.Remove(f); err != nil {
				return err
			}
		}

		// ignore if the path is empty
		if toExtract.Path == "" {
			if err := TouchFile(missingDestination); err != nil {
				return err
			}
			continue
		}

		if sourceImage == "" {
			// ignore if the file does not exist
			if toExtract.IsDirectory {
				if !DirectoryExists(sourcePath) {
					if sourcePath != repoDefaultSourcePath && DirectoryExists(repoDefaultSourcePath) {
						if err := os.RemoveAll(repoDefaultSourcePath); err != nil {
							return fmt.Errorf("Unable to remove existing %v: %s", toExtract.Name, err.Error())
						}
					}

					if err := TouchFile(missingDestination); err != nil {
						return err
					}
					continue
				}

				if err := Copy(sourcePath, processSpecificDestination); err != nil {
					return fmt.Errorf("Unable to extract %v from %v: %s", toExtract.Name, toExtract.Path, err.Error())
				}

				if sourcePath != repoDefaultSourcePath {
					if err := Copy(sourcePath, repoDefaultSourcePath); err != nil {
						return fmt.Errorf("Unable to move %v into place: %s", toExtract.Name, err.Error())
					}
				}
			} else {
				if !FileExists(sourcePath) {
					// delete the existing file if the user tried to override it with a non-existent file
					if sourcePath != repoDefaultSourcePath && FileExists(repoDefaultSourcePath) {
						if err := os.Remove(repoDefaultSourcePath); err != nil {
							return err
						}
					}
					if err := TouchFile(missingDestination); err != nil {
						return err
					}
					continue
				}

				if err := Copy(sourcePath, processSpecificDestination); err != nil {
					return fmt.Errorf("Unable to extract %v from %v: %v", toExtract.Name, toExtract.Path, err.Error())
				}

				if sourcePath != repoDefaultSourcePath {
					// ensure the file in the repo is the same as the one the user specified
					if err := copy.Copy(sourcePath, repoDefaultSourcePath); err != nil {
						return fmt.Errorf("Unable to move %v into place: %v", toExtract.Name, err.Error())
					}
				}
			}
		} else {
			if toExtract.IsDirectory {

				if err := CopyDirFromImage(input.AppName, sourceImage, imageSourcePath, processSpecificDestination); err != nil {
					return TouchFile(missingDestination)
				}
			} else {
				if err := CopyFromImage(input.AppName, sourceImage, imageSourcePath, processSpecificDestination); err != nil {
					return TouchFile(missingDestination)
				}
			}
		}

		// validate the file
		if toExtract.Validator != nil {
			if err := toExtract.Validator(input.AppName, processSpecificDestination); err != nil {
				return err
			}
		}
	}

	return nil
}

// EnvWrap wraps a func with a setenv call and resets the value at the end
func EnvWrap(fn func() error, environ map[string]string) error {
	oldEnviron := map[string]string{}
	for key, value := range environ {
		oldEnviron[key] = os.Getenv(key)
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	if err := fn(); err != nil {
		return err
	}

	for key, value := range oldEnviron {
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	return nil
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
	results, _ := CallPlugnTrigger(PlugnTriggerInput{
		Trigger: "scheduler-detect",
		Args:    []string{appName},
	})
	value := results.StdoutContents()
	if value != "" {
		return value
	}
	return ""
}

// GetGlobalScheduler fetchs the global scheduler
func GetGlobalScheduler() string {
	results, _ := CallPlugnTrigger(PlugnTriggerInput{
		Trigger: "scheduler-detect",
		Args:    []string{"--global"},
	})
	value := results.StdoutContents()
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
		results, err := CallPlugnTrigger(PlugnTriggerInput{
			Trigger: "deployed-app-repository",
			Args:    []string{appName},
		})
		if err == nil {
			imageRemoteRepository = results.StdoutContents()
		}
		return err
	})
	errs.Go(func() error {
		results, err := CallPlugnTrigger(PlugnTriggerInput{
			Trigger: "deployed-app-image-tag",
			Args:    []string{appName},
		})

		if err == nil {
			newImageTag = results.StdoutContents()
		}
		return err
	})

	errs.Go(func() error {
		results, err := CallPlugnTrigger(PlugnTriggerInput{
			Trigger: "deployed-app-image-repo",
			Args:    []string{appName},
		})
		if err == nil {
			newImageRepo = results.StdoutContents()
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

// GetRunningImageTag retrieves current deployed image tag for a given app
func GetRunningImageTag(appName string, imageTag string) (string, error) {
	results, err := CallPlugnTrigger(PlugnTriggerInput{
		Trigger: "deployed-app-image-tag",
		Args:    []string{appName},
	})
	if err != nil {
		return imageTag, err
	}
	newImageTag := results.StdoutContents()
	if newImageTag != "" {
		imageTag = newImageTag
	}
	if imageTag == "" {
		imageTag = "latest"
	}

	return imageTag, nil
}

// GetDokkuAppShell returns the shell for a given app
func GetDokkuAppShell(appName string) string {
	shell := "/bin/bash"
	globalShell := ""
	appShell := ""

	ctx := context.Background()
	errs, ctx := errgroup.WithContext(ctx)
	errs.Go(func() error {
		results, _ := CallPlugnTriggerWithContext(ctx, PlugnTriggerInput{
			Trigger: "config-get-global",
			Args:    []string{"DOKKU_APP_SHELL"},
		})
		globalShell = results.StdoutContents()
		return nil
	})
	errs.Go(func() error {
		results, _ := CallPlugnTriggerWithContext(ctx, PlugnTriggerInput{
			Trigger: "config-get",
			Args:    []string{appName, "DOKKU_APP_SHELL"},
		})
		appShell = results.StdoutContents()
		return nil
	})

	errs.Wait()
	if appShell != "" {
		shell = appShell
	} else if globalShell != "" {
		shell = globalShell
	}

	return shell
}

// DokkuApps returns a list of all local apps
func DokkuApps() ([]string, error) {
	apps, err := UnfilteredDokkuApps()
	if err != nil {
		return apps, err
	}

	return filterApps(apps)
}

// UnfilteredDokkuApps returns an unfiltered list of all local apps
func UnfilteredDokkuApps() ([]string, error) {
	apps := []string{}
	dokkuRoot := MustGetEnv("DOKKU_ROOT")
	files, err := os.ReadDir(dokkuRoot)
	if err != nil {
		return apps, NoAppsExist
	}

	for _, f := range files {
		appRoot := AppRoot(f.Name())
		if !DirectoryExists(appRoot) {
			continue
		}
		if strings.HasPrefix(f.Name(), ".") {
			continue
		}
		// skip apps that start with an uppercase letter
		if unicode.IsUpper(rune(f.Name()[0])) {
			continue
		}

		apps = append(apps, f.Name())
	}

	if len(apps) == 0 {
		return apps, NoAppsExist
	}

	return apps, nil
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
		_, err := CallPlugnTrigger(PlugnTriggerInput{
			Trigger: "scheduler-is-deployed",
			Args:    []string{scheduler, appName},
		})
		if err == nil {
			deployed = "true"
		}

		EnvWrap(func() error {
			return PropertyWrite("common", appName, "deployed", deployed)
		}, map[string]string{"DOKKU_QUIET_OUTPUT": "1"})
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
func ParseScaleOutput(b []byte) (map[string]int32, error) {
	scale := make(map[string]int32)

	for _, line := range strings.Split(string(b), "\n") {
		s := strings.SplitN(line, "=", 2)
		if len(s) != 2 {
			return scale, fmt.Errorf("invalid scale output stored by dokku: %v", line)
		}
		processType := s[0]
		count, err := strconv.ParseInt(s[1], 10, 32)
		if err != nil {
			return scale, err
		}
		scale[processType] = int32(count)
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
	out, _ := io.ReadAll(r)
	os.Stdout = rescueStdout

	if err != nil {
		fmt.Print(string(out[:]))
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

// NoAppsExist wraps error to include the app name
// and is used to distinguish between a normal error and an error
// where the app is missing
var NoAppsExist = errors.New("You haven't deployed any applications yet")

// VarArgs skips a number of incoming arguments, returning what is left over
func VarArgs(arguments []string, skip int) []string {
	if len(arguments) <= skip {
		return []string{}
	}

	return arguments[skip:]
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

	apps, _ := filterApps([]string{appName})
	if len(apps) != 1 {
		return &AppDoesNotExist{appName}
	}

	return nil
}
