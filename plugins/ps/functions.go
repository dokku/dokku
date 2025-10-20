package ps

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	dockeroptions "github.com/dokku/dokku/plugins/docker-options"
)

func canScaleApp(appName string) bool {
	canScale := common.PropertyGetDefault("ps", appName, "can-scale", "true")
	return common.ToBool(canScale)
}

func getProcfileCommand(appName string, procfilePath string, processType string, port int) (string, error) {
	if !common.FileExists(procfilePath) {
		return "", errors.New("No procfile found")
	}

	configResult, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "config-export",
		Args:    []string{appName, "false", "true", "envfile"},
	})
	if err != nil {
		return "", err
	}

	// write the envfile to a temporary file
	tempFile, err := os.CreateTemp("", "envfile-*.env")
	if err != nil {
		return "", err
	}
	defer os.Remove(tempFile.Name())
	if _, err := tempFile.Write(configResult.StdoutBytes()); err != nil {
		return "", err
	}
	if err := tempFile.Close(); err != nil {
		return "", err
	}

	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command: "procfile-util",
		Args:    []string{"show", "--procfile", procfilePath, "--process-type", processType, "--default-port", strconv.Itoa(port), "--env-file", tempFile.Name()},
	})
	if err != nil {
		return "", fmt.Errorf("Error running procfile-util: %s", err)
	}
	if result.ExitCode != 0 {
		return "", fmt.Errorf("Error running procfile-util: %s", result.StderrContents())
	}

	return result.StdoutContents(), nil
}

func getProcfilePath(appName string) string {
	directory := common.GetAppDataDirectory("ps", appName)
	return filepath.Join(directory, "Procfile")
}

func getProcessSpecificProcfilePath(appName string) string {
	existingProcfile := getProcfilePath(appName)
	processSpecificProcfile := fmt.Sprintf("%s.%s", existingProcfile, os.Getenv("DOKKU_PID"))
	if common.FileExists(processSpecificProcfile) {
		return processSpecificProcfile
	}

	return existingProcfile
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
	results, _ := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "scheduler-app-status",
		Args:    []string{scheduler, appName},
	})
	count := strings.Split(results.StdoutContents(), " ")[0]
	return strconv.Atoi(count)
}

func getRunningState(appName string) string {
	scheduler := common.GetAppScheduler(appName)
	results, _ := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "scheduler-app-status",
		Args:    []string{scheduler, appName},
	})
	return strings.Split(results.StdoutContents(), " ")[1]
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

	foundFormations := map[string]bool{}
	for _, processTuple := range processTuples {
		s := strings.SplitN(processTuple, "=", 2)
		if len(s) == 1 {
			return formations, fmt.Errorf("Missing count for process type %s", processTuple)
		}

		processType := strings.TrimSpace(s[0])
		quantity, err := strconv.Atoi(strings.TrimSpace(s[1]))
		if err != nil {
			return formations, fmt.Errorf("Invalid count for process type %s", s[0])
		}

		if foundFormations[processType] {
			continue
		}

		foundFormations[processType] = true
		formations = append(formations, &Formation{
			ProcessType: processType,
			Quantity:    quantity,
		})
	}

	return formations, nil
}

func processesInProcfile(procfilePath string) (map[string]bool, error) {
	processes := map[string]bool{}

	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command: "procfile-util",
		Args:    []string{"list", "--procfile", procfilePath},
	})
	if err != nil {
		return processes, fmt.Errorf("Error listing processes: %s", err)
	}
	if result.ExitCode != 0 {
		return processes, fmt.Errorf("Error listing processes: %s", result.StderrContents())
	}

	for _, s := range strings.Split(result.StdoutContents(), "\n") {
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

	foundProcessTypes := map[string]bool{}
	for _, formation := range formations {
		foundProcessTypes[formation.ProcessType] = true
	}

	for _, formation := range oldFormations {
		if foundProcessTypes[formation.ProcessType] {
			continue
		}

		foundProcessTypes[formation.ProcessType] = true
		formations = append(formations, formation)
	}

	sort.Sort(formations)
	return formations, nil
}

func restorePrep() error {
	_, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "proxy-clear-config",
		Args:        []string{"--all"},
		StreamStdio: true,
	})
	if err != nil {
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

	content := []string{}
	if os.Getenv("DOKKU_QUIET_OUTPUT") == "" {
		content = append(content, "proctype=qty", "--------=---")
	}

	for _, formation := range formations {
		content = append(content, fmt.Sprintf("%s=%d", formation.ProcessType, formation.Quantity))
	}

	for _, line := range content {
		s := strings.Split(line, "=")
		common.Log(fmt.Sprintf("%s %s", common.RightPad(fmt.Sprintf("%s:", s[0]), 5, " "), s[1]))
	}

	return nil
}

// scaleSetInput is the input for the scaleSet function
type scaleSetInput struct {
	// appName is the name of the app to scale
	appName string

	// skipDeploy is a flag to skip the deploy phase
	skipDeploy bool

	// clearExisting is a flag to clear the existing scale
	clearExisting bool

	// processTuples is a list of process tuples to scale
	processTuples []string

	// deployOnlyChanged is a flag to deploy only the changed formations
	deployOnlyChanged bool
}

func scaleSet(input scaleSetInput) error {
	existingFormations, err := getFormations(input.appName)
	if err != nil {
		return err
	}

	formations, err := parseProcessTuples(input.processTuples)
	if err != nil {
		return err
	}

	if err := updateScale(input.appName, input.clearExisting, formations); err != nil {
		return err
	}

	if input.skipDeploy {
		return nil
	}

	if !common.IsDeployed(input.appName) {
		return nil
	}

	imageTag, err := common.GetRunningImageTag(input.appName, "")
	if err != nil {
		return err
	}

	changedFormations := FormationSlice{}
	if input.deployOnlyChanged {
		for _, formation := range formations {
			isChanged := true
			for _, existingFormation := range existingFormations {
				if existingFormation.ProcessType == formation.ProcessType {
					if existingFormation.Quantity == formation.Quantity {
						isChanged = false
						break
					}
				}
			}

			if isChanged {
				changedFormations = append(changedFormations, formation)
			}
		}
	} else {
		changedFormations = formations
	}

	for _, formation := range changedFormations {
		_, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
			Trigger:     "deploy",
			Args:        []string{input.appName, imageTag, formation.ProcessType},
			StreamStdio: true,
		})
		if err != nil {
			return err
		}
	}

	return nil
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
		validProcessTypes, err = processesInProcfile(getProcessSpecificProcfilePath(appName))
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
