package ps

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	dockeroptions "github.com/dokku/dokku/plugins/docker-options"
	"github.com/ryanuber/columnize"
)

func canScaleApp(appName string) bool {
	return !common.FileExists(getScalefileExtractedPath(appName))
}

func extractProcfile(appName, image string, procfilePath string) error {
	if err := removeProcfile(appName); err != nil {
		return err
	}

	destination := getProcfilePath(appName)
	common.CopyFromImage(appName, image, "Procfile", destination)
	if !common.FileExists(destination) {
		common.LogInfo1Quiet("No Procfile found in app image")
		return nil
	}

	common.LogInfo1Quiet("App Procfile file found")
	checkCmd := common.NewShellCmd(strings.Join([]string{
		"procfile-util",
		"check",
		"--procfile",
		destination,
	}, " "))
	var stderr bytes.Buffer
	checkCmd.ShowOutput = false
	checkCmd.Command.Stderr = &stderr
	_, err := checkCmd.Output()

	if err != nil {
		return fmt.Errorf(strings.TrimSpace(stderr.String()))
	}

	return nil
}

func extractOrGenerateScalefile(appName string, image string) error {
	destination := getScalefilePath(appName)
	extracted := getScalefileExtractedPath(appName)
	previouslyExtracted := common.FileExists(extracted)

	if err := common.CopyFromImage(appName, image, "DOKKU_SCALE", extracted); err != nil {
		if previouslyExtracted {
			os.Remove(destination)
		}
		os.Remove(extracted)
	} else if err := common.CopyFile(extracted, destination); err != nil {
		return err
	}

	if common.FileExists(destination) {
		common.LogInfo1Quiet("DOKKU_SCALE file exists")
		return updateScalefile(appName, make(map[string]int))
	}

	common.LogInfo1Quiet("DOKKU_SCALE file not found in app image. Generating one based on Procfile...")
	if err := updateScalefile(appName, make(map[string]int)); err != nil {
		common.LogDebug(fmt.Sprintf("Error generating scale file: %s", err.Error()))
		return err
	}
	return nil
}

func generateScalefile(appName string) error {
	destination := getScalefilePath(appName)
	procfilePath := getProcfilePath(appName)
	content := []string{"web=1"}
	if !common.FileExists(procfilePath) {
		return common.WriteSliceToFile(destination, content)
	}

	lines, err := common.FileToSlice(procfilePath)
	if err != nil {
		return common.WriteSliceToFile(destination, content)
	}

	content = []string{}
	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ":")
		count := 0
		if parts[0] == "web" {
			count = 1
		}

		content = append(content, fmt.Sprintf("%s=%d", parts[0], count))
	}

	return common.WriteSliceToFile(destination, content)
}

func getProcessStatus(appName string) map[string]string {
	statuses := make(map[string]string)
	containerFiles := common.ListFilesWithPrefix(common.AppRoot(appName), "CONTAINER.")
	for _, filename := range containerFiles {
		containerID := common.ReadFirstLine(filename)
		containerStatus, _ := common.DockerInspect(containerID, "{{ .State.Status }}")
		process := strings.TrimPrefix(filename, fmt.Sprintf("%s/CONTAINER.", common.AppRoot(appName)))

		if containerStatus == "" {
			containerStatus = "missing"
		}

		statuses[process] = fmt.Sprintf("%s (CID: %s)", containerStatus, containerID[0:11])
	}

	return statuses
}

func getProcfileCommand(procfilePath string, processType string, port int) (string, error) {
	if !common.FileExists(procfilePath) {
		return "", errors.New("No procfile found")
	}

	shellCmd := common.NewShellCmd(strings.Join([]string{
		"procfile-util",
		"show",
		"--procfile",
		procfilePath,
		"--process-type",
		processType,
		"--default-port",
		strconv.Itoa(port),
	}, " "))
	var stderr bytes.Buffer
	shellCmd.ShowOutput = false
	shellCmd.Command.Stderr = &stderr
	b, err := shellCmd.Output()

	if err != nil {
		return "", fmt.Errorf(strings.TrimSpace(stderr.String()))
	}

	return strings.TrimSpace(string(b[:])), nil
}

func getProcfilePath(appName string) string {
	directory := filepath.Join(common.MustGetEnv("DOKKU_LIB_ROOT"), "data", "ps", appName)
	return filepath.Join(directory, "Procfile")
}

func getRestartPolicy(appName string) (string, error) {
	options, err := dockeroptions.GetDockerOptionsForPhase(appName, "deploy")
	if err != nil {
		return "", err
	}

	for _, option := range options {
		if strings.HasPrefix(option, "--restart=") {
			return strings.TrimPrefix(option, "--restart="), nil
		}
	}

	return "", nil
}

func getProcessCount(appName string) (int, error) {
	scheduler := common.GetAppScheduler(appName)
	b, _ := common.PlugnTriggerOutput("scheduler-app-status", []string{scheduler, appName}...)
	count := strings.Split(strings.TrimSpace(string(b[:])), " ")[0]
	return strconv.Atoi(count)
}

func getRunningState(appName string) string {
	scheduler := common.GetAppScheduler(appName)
	b, _ := common.PlugnTriggerOutput("scheduler-app-status", []string{scheduler, appName}...)
	return strings.Split(strings.TrimSpace(string(b[:])), " ")[1]
}

func getScalefilePath(appName string) string {
	return filepath.Join(common.AppRoot(appName), "DOKKU_SCALE")
}

func getScalefileExtractedPath(appName string) string {
	return filepath.Join(common.AppRoot(appName), "DOKKU_SCALE.extracted")
}

func hasScaleFile(appName string) bool {
	scalefilePath := getScalefilePath(appName)
	return common.FileExists(scalefilePath)
}

func hasProcfile(appName string) bool {
	procfilePath := getProcfilePath(appName)
	return common.FileExists(procfilePath)
}

func isValidRestartPolicy(policy string) bool {
	if policy == "" {
		return false
	}

	validRestartPolicies := map[string]bool{
		"no":             true,
		"always":         true,
		"unless-stopped": true,
		"on-failure":     true,
	}

	if _, ok := validRestartPolicies[policy]; ok {
		return true
	}

	return strings.HasPrefix(policy, "on-failure:")
}

func parseProcessTuples(processTuples []string) (map[string]int, error) {
	scale := make(map[string]int)

	for _, processTuple := range processTuples {
		s := strings.Split(processTuple, "=")
		if len(s) == 1 {
			return scale, fmt.Errorf("Missing count for process type %s", processTuple)
		}

		processType := s[0]
		count, err := strconv.Atoi(s[1])
		if err != nil {
			return scale, fmt.Errorf("Invalid count for process type %s", s[0])
		}

		scale[processType] = count
	}

	return scale, nil
}

func processesInProcfile(procfilePath string) (map[string]bool, error) {
	processes := map[string]bool{}

	shellCmd := common.NewShellCmd(strings.Join([]string{
		"procfile-util",
		"list",
		"--procfile",
		procfilePath,
	}, " "))
	var stderr bytes.Buffer
	shellCmd.ShowOutput = false
	shellCmd.Command.Stderr = &stderr
	b, err := shellCmd.Output()

	if err != nil {
		return processes, fmt.Errorf(strings.TrimSpace(stderr.String()))
	}

	for _, s := range strings.Split(strings.TrimSpace(string(b[:])), "\n") {
		processes[s] = true
	}

	return processes, nil
}

func readScaleFile(appName string) (map[string]int, error) {
	scale := make(map[string]int)
	if !hasScaleFile(appName) {
		return scale, errors.New("No DOKKU_SCALE file found")
	}

	scalefilePath := getScalefilePath(appName)
	lines, err := common.FileToSlice(scalefilePath)
	if err != nil {
		return scale, err
	}

	for i, line := range lines {
		s := strings.Split(line, "=")
		if len(s) != 2 {
			common.LogWarn(fmt.Sprintf("Invalid scale entry on line %d, skipping", i))
			continue
		}

		processType := s[0]
		count, err := strconv.Atoi(s[1])
		if err != nil {
			common.LogWarn(fmt.Sprintf("Invalid count on line %d, skipping", i))
			continue
		}

		scale[processType] = count
	}

	if len(scale) == 0 {
		common.LogWarn("No valid entries found in scale file, defaulting to web=1")
		scale["web"] = 1
	}

	return scale, nil
}

func removeProcfile(appName string) error {
	procfile := getProcfilePath(appName)
	if !common.FileExists(procfile) {
		return nil
	}

	return os.Remove(procfile)
}

func scaleReport(appName string) error {
	scalefilePath := getScalefilePath(appName)
	lines, err := common.FileToSlice(scalefilePath)
	if err != nil {
		return err
	}

	common.LogInfo1Quiet(fmt.Sprintf("Scaling for %s", appName))
	config := columnize.DefaultConfig()
	config.Delim = "="
	config.Glue = ": "
	config.Prefix = "    "
	config.Empty = ""

	content := []string{}
	if os.Getenv("DOKKU_QUIET_OUTPUT") == "" {
		content = append(content, "proctype=qty", "--------=---")
	}

	sort.Strings(lines)
	for _, line := range lines {
		content = append(content, line)
	}

	for _, line := range content {
		s := strings.Split(line, "=")
		common.Log(fmt.Sprintf("%s %s", common.RightPad(fmt.Sprintf("%s:", s[0]), 5, " "), s[1]))
	}

	return nil
}

func scaleSet(appName string, skipDeploy bool, processTuples []string) error {
	if !canScaleApp(appName) {
		return fmt.Errorf("App %s contains DOKKU_SCALE file and cannot be manually scaled", appName)
	}

	scale, err := parseProcessTuples(processTuples)
	if err != nil {
		return err
	}

	common.LogInfo1(fmt.Sprintf("Scaling %s processes: %s", appName, strings.Join(processTuples, " ")))
	if err := updateScalefile(appName, scale); err != nil {
		return err
	}

	if !common.IsDeployed(appName) {
		return nil
	}

	imageTag, err := common.GetRunningImageTag(appName)
	if err != nil {
		return err
	}

	if skipDeploy {
		return nil
	}

	return common.PlugnTrigger("release-and-deploy", []string{appName, imageTag}...)
}

func updateScalefile(appName string, scaleUpdates map[string]int) error {
	if !hasScaleFile(appName) {
		if err := generateScalefile(appName); err != nil {
			return err
		}
	}

	scale, err := readScaleFile(appName)
	if err != nil {
		return err
	}

	procfilePath := getProcfilePath(appName)
	procfileExists := hasProcfile(appName)
	validProcessTypes := make(map[string]bool)
	if procfileExists {
		validProcessTypes, err = processesInProcfile(procfilePath)
		if err != nil {
			return err
		}
	}

	for processType, count := range scaleUpdates {
		if procfileExists && !validProcessTypes[processType] && count != 0 {
			return fmt.Errorf("%s is not a valid process name to scale up", processType)
		}
		scale[processType] = count
	}

	for processType := range validProcessTypes {
		count, ok := scale[processType]
		if !ok {
			count = 0
		}
		scale[processType] = count
	}

	content := []string{}
	for processType, count := range scale {
		content = append(content, fmt.Sprintf("%s=%d", processType, count))
	}

	scalefilePath := getScalefilePath(appName)
	if err := common.WriteSliceToFile(scalefilePath, content); err != nil {
		return err
	}

	return nil
}
