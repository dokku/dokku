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
	canScale := common.PropertyGetDefault("ps", appName, "can-scale", "true")
	return common.ToBool(canScale)
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

func hasProcfile(appName string) bool {
	procfilePath := getProcfilePath(appName)
	if common.FileExists(fmt.Sprintf("%s.%s.missing", procfilePath, os.Getenv("DOKKU_PID"))) {
		return false
	}

	if common.FileExists(fmt.Sprintf("%s.%s", procfilePath, os.Getenv("DOKKU_PID"))) {
		return true
	}

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

func parseProcessTuples(processTuples []string) (FormationSlice, error) {
	formations := FormationSlice{}

	for _, processTuple := range processTuples {
		s := strings.Split(processTuple, "=")
		if len(s) == 1 {
			return formations, fmt.Errorf("Missing count for process type %s", processTuple)
		}

		processType := s[0]
		quantity, err := strconv.Atoi(s[1])
		if err != nil {
			return formations, fmt.Errorf("Invalid count for process type %s", s[0])
		}

		formations = append(formations, &Formation{
			ProcessType: processType,
			Quantity:    quantity,
		})
	}

	return formations, nil
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

func getFormations(appName string) (FormationSlice, error) {
	formations := FormationSlice{}
	processTuples, err := common.PropertyListGet("ps", appName, "scale")
	if err != nil {
		return formations, err
	}

	oldProcessTuples, err := common.PropertyListGet("ps", appName, "scale.old")
	if err != nil {
		return formations, err
	}

	formations, err = parseProcessTuples(processTuples)
	if err != nil {
		return formations, err
	}

	oldFormations, err := parseProcessTuples(oldProcessTuples)
	if err != nil {
		return formations, err
	}

	return append(formations, oldFormations...), nil
}

func restorePrep() error {
	if err := common.PlugnTrigger("proxy-clear-config", []string{"--all"}...); err != nil {
		return fmt.Errorf("Error clearing proxy config: %s", err)
	}

	return nil
}

func scaleReport(appName string) error {
	formations, err := getFormations(appName)
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

	sort.Sort(formations)
	for _, formation := range formations {
		content = append(content, fmt.Sprintf("%s=%d", formation.ProcessType, formation.Quantity))
	}

	for _, line := range content {
		s := strings.Split(line, "=")
		common.Log(fmt.Sprintf("%s %s", common.RightPad(fmt.Sprintf("%s:", s[0]), 5, " "), s[1]))
	}

	return nil
}

func scaleSet(appName string, skipDeploy bool, clearExisting bool, processTuples []string) error {
	formations, err := parseProcessTuples(processTuples)
	if err != nil {
		return err
	}

	if err := updateScale(appName, clearExisting, formations); err != nil {
		return err
	}

	if skipDeploy {
		return nil
	}

	if !common.IsDeployed(appName) {
		return nil
	}

	imageTag, err := common.GetRunningImageTag(appName, "")
	if err != nil {
		return err
	}

	for _, formation := range formations {
		if err := common.PlugnTrigger("deploy", []string{appName, imageTag, formation.ProcessType}...); err != nil {
			return err
		}
	}

	return nil
}

func getProcessSpecificProcfile(appName string) string {
	existingProcfile := getProcfilePath(appName)
	processSpecificProcfile := fmt.Sprintf("%s.%s", existingProcfile, os.Getenv("DOKKU_PID"))
	if common.FileExists(processSpecificProcfile) {
		return processSpecificProcfile
	}

	return existingProcfile
}

func updateScale(appName string, clearExisting bool, formationUpdates FormationSlice) error {
	formations := FormationSlice{}
	if !clearExisting {
		processTuples, err := common.PropertyListGet("ps", appName, "scale")
		if err != nil {
			return err
		}

		formations, err = parseProcessTuples(processTuples)
		if err != nil {
			return err
		}
	}

	validProcessTypes := make(map[string]bool)
	if hasProcfile(appName) {
		var err error
		validProcessTypes, err = processesInProcfile(getProcessSpecificProcfile(appName))
		if err != nil {
			return err
		}
	}

	foundProcessTypes := map[string]bool{}
	updatedFormation := FormationSlice{}
	for _, formation := range formationUpdates {
		if hasProcfile(appName) && !validProcessTypes[formation.ProcessType] && formation.Quantity != 0 {
			return fmt.Errorf("%s is not a valid process name to scale up", formation.ProcessType)
		}

		foundProcessTypes[formation.ProcessType] = true
		updatedFormation = append(updatedFormation, &Formation{
			ProcessType: formation.ProcessType,
			Quantity:    formation.Quantity,
		})
	}

	for _, formation := range formations {
		if foundProcessTypes[formation.ProcessType] {
			continue
		}

		foundProcessTypes[formation.ProcessType] = true
		updatedFormation = append(updatedFormation, &Formation{
			ProcessType: formation.ProcessType,
			Quantity:    formation.Quantity,
		})
	}

	for processType := range validProcessTypes {
		if foundProcessTypes[processType] {
			continue
		}

		updatedFormation = append(updatedFormation, &Formation{
			ProcessType: processType,
			Quantity:    0,
		})
	}

	values := []string{}
	oldValues := []string{}
	for _, formation := range updatedFormation {
		if !validProcessTypes[formation.ProcessType] && formation.Quantity == 0 {
			oldValues = append(oldValues, fmt.Sprintf("%s=%d", formation.ProcessType, formation.Quantity))
			continue
		}

		values = append(values, fmt.Sprintf("%s=%d", formation.ProcessType, formation.Quantity))
	}

	if err := common.PropertyListWrite("ps", appName, "scale.old", oldValues); err != nil {
		return err
	}

	return common.PropertyListWrite("ps", appName, "scale", values)
}
