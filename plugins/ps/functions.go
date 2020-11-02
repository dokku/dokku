package ps

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	dockeroptions "github.com/dokku/dokku/plugins/docker-options"
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
	if err := common.CopyFromImage(appName, image, "DOKKU_SCALE", destination); err != nil {
		os.Remove(extracted)
		os.Remove(destination)
	} else if err := common.CopyFile(destination, extracted); err != nil {
		return err
	}

	if common.FileExists(destination) {
		return nil
	}

	common.LogInfo1Quiet("DOKKU_SCALE file not found in app image. Generating one based on Procfile...")
	if err := generateScalefile(appName, destination); err != nil {
		common.LogDebug(fmt.Sprintf("Error: %s", err.Error()))
		return err
	}
	return nil
}

func generateScalefile(appName string, destination string) error {
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

func removeProcfile(appName string) error {
	procfile := getProcfilePath(appName)
	if !common.FileExists(procfile) {
		return nil
	}

	return os.Remove(procfile)
}

func updateScalefile(appName string, processTuples []string) error {
	procfilePath := getProcfilePath(appName)
	scalefilePath := getScalefilePath(appName)
	lines, err := common.FileToSlice(scalefilePath)
	if err != nil {
		return err
	}

	processTypes, err := processesInProcfile(procfilePath)
	if err != nil {
		return err
	}

	newLines := []string{}
	for processType := range processTypes {
		found := false
		for _, line := range lines {
			if strings.HasPrefix(line, fmt.Sprintf("%s=", processType)) {
				newLines = append(newLines, line)
				found = true
				break
			}
		}

		if !found {
			newLines = append(newLines, fmt.Sprintf("%s=0", processType))
		}
	}

	scale := []string{}
	for _, processTuple := range processTuples {
		s := strings.Split(processTuple, "=")
		if len(s) == 1 {
			return fmt.Errorf("Missing count for process type %s", processTuple)
		}

		processType := s[0]
		count, err := strconv.Atoi(s[1])
		if err != nil {
			return fmt.Errorf("Invalid count for process type %s", s[0])
		}

		if _, ok := processTypes[processType]; !ok {
		    if count != 0 && len(processTypes) == 0 {
		      	return fmt.Errorf("%s is not a valid process name to scale up", processType)
		    }
		}

		scale = append(scale, fmt.Sprintf("%s=%d", processType, count))
	}

	for _, line := range newLines {
		s := strings.Split(line, "=")
		processType := s[0]

		found := false
		for _, s := range scale {
			if strings.HasPrefix(s, fmt.Sprintf("%s=", processType)) {
				found = true
				break
			}
		}

		if !found {
			scale = append(scale, line)
		}

	}

	if err := common.WriteSliceToFile(scalefilePath, scale); err != nil {
		return err
	}

	return err
}
