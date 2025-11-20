package appjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	shellquote "github.com/kballard/go-shellquote"
)

func constructScript(command string, shell string, isHerokuishImage bool, isCnbImage bool, dockerfileEntrypoint string) []string {
	nonSkippableEntrypoints := map[string]bool{
		"ENTRYPOINT [\"/tini\",\"--\"]":               true,
		"ENTRYPOINT [\"/bin/tini\",\"--\"]":           true,
		"ENTRYPOINT [\"/usr/bin/tini\",\"--\"]":       true,
		"ENTRYPOINT [\"/usr/local/bin/tini\",\"--\"]": true,
	}

	cannotSkip := nonSkippableEntrypoints[dockerfileEntrypoint]
	if dockerfileEntrypoint != "" && !cannotSkip {
		words, err := shellquote.Split(strings.TrimSpace(command))
		if err != nil {
			common.LogWarn(fmt.Sprintf("Skipping command construction for app with ENTRYPOINT: %v", err.Error()))
			return nil
		}
		return words
	}

	script := []string{"set -e;", "set -o pipefail || true;"}
	if os.Getenv("DOKKU_TRACE") == "1" {
		script = append(script, "set -x;")
	}

	if isHerokuishImage && !isCnbImage {
		script = append(script, []string{
			"if [[ -d '/app' ]]; then",
			"  export HOME=/app;",
			"  cd $HOME;",
			"fi;",
			"if [[ -d '/app/.profile.d' ]]; then",
			"  for file in /app/.profile.d/*; do source $file; done;",
			"fi;",
		}...)
	}

	if strings.HasPrefix(command, "/") {
		commandBin := strings.Split(command, " ")[0]
		script = append(script, []string{
			fmt.Sprintf("if [[ ! -x \"%s\" ]]; then", commandBin),
			"  echo specified binary is not executable;",
			"  exit 1;",
			"fi;",
		}...)
	}

	script = append(script, fmt.Sprintf("%s || exit 1;", command))

	return []string{shell, "-c", strings.Join(script, " ")}
}

func getAppJSONPath(appName string) string {
	directory := filepath.Join(common.MustGetEnv("DOKKU_LIB_ROOT"), "data", "app-json", appName)
	return filepath.Join(directory, "app.json")
}

func getProcessSpecificAppJSONPath(appName string) string {
	existingAppJSON := getAppJSONPath(appName)
	processSpecificAppJSON := fmt.Sprintf("%s.%s", existingAppJSON, os.Getenv("DOKKU_PID"))
	if common.FileExists(processSpecificAppJSON) {
		return processSpecificAppJSON
	}

	return existingAppJSON
}

// getPhaseScript extracts app.json from app image and returns the appropriate json key/value
func getPhaseScript(appName string, phase string) (string, error) {
	appJSON, err := GetAppJSON(appName)
	if err != nil {
		return "", err
	}

	if phase == "heroku.postdeploy" {
		return appJSON.Scripts.Postdeploy, nil
	}

	if phase == "predeploy" {
		return appJSON.Scripts.Dokku.Predeploy, nil
	}

	return appJSON.Scripts.Dokku.Postdeploy, nil
}

// getReleaseCommand extracts the release command from a given app's procfile
func getReleaseCommand(appName string) string {
	processType := "release"
	port := "5000"
	results, _ := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "procfile-get-command",
		Args:    []string{appName, processType, port},
	})
	return results.StdoutContents()
}

func hasAppJSON(appName string) bool {
	appJSONPath := getAppJSONPath(appName)
	if common.FileExists(fmt.Sprintf("%s.%s.missing", appJSONPath, os.Getenv("DOKKU_PID"))) {
		return false
	}

	if common.FileExists(fmt.Sprintf("%s.%s", appJSONPath, os.Getenv("DOKKU_PID"))) {
		return true
	}

	return common.FileExists(appJSONPath)
}

func cleanupDeploymentContainer(containerID string, phase string) error {
	if phase != "predeploy" {
		os.Setenv("DOKKU_SKIP_IMAGE_RETIRE", "true")
	}

	if !common.ContainerRemove(containerID) {
		return fmt.Errorf("Failed to remove %s execution container", phase)
	}

	return nil
}

func executeScript(appName string, image string, imageTag string, phase string) error {
	common.LogInfo1(fmt.Sprintf("Checking for %s task", phase))
	command := ""
	phaseSource := ""
	if phase == "release" {
		command = getReleaseCommand(appName)
		phaseSource = "Procfile"
	} else {
		var err error
		phaseSource = "app.json"
		if command, err = getPhaseScript(appName, phase); err != nil {
			common.LogExclaim(err.Error())
		}
	}

	if command == "" {
		common.LogVerbose(fmt.Sprintf("No %s task found, skipping", phase))
		return nil
	}

	if phase == "predeploy" {
		common.LogVerbose(fmt.Sprintf("Executing %s task from %s: %s", phase, phaseSource, command))
	} else {
		common.LogVerbose(fmt.Sprintf("Executing %s task from %s in ephemeral container: %s", phase, phaseSource, command))
	}

	isHerokuishImage := common.IsImageHerokuishBased(image, appName)
	isCnbImage := common.IsImageCnbBased(image)
	dockerfileEntrypoint := ""
	dockerfileCommand := ""
	if !isHerokuishImage {
		dockerfileEntrypoint, _ = getEntrypointFromImage(image)
		dockerfileCommand, _ = getCommandFromImage(image)
	}

	dokkuAppShell := common.GetDokkuAppShell(appName)
	script := constructScript(command, dokkuAppShell, isHerokuishImage, isCnbImage, dockerfileEntrypoint)

	imageSourceType := "dockerfile"
	if isHerokuishImage {
		imageSourceType = "herokuish"
	} else if isCnbImage {
		imageSourceType = "pack"
	}

	var dockerArgs []string
	results, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "docker-args-deploy",
		Args:    []string{appName, imageTag},
		Stdin:   strings.NewReader(""),
	})
	if err == nil {
		words, err := shellquote.Split(results.StdoutContents())
		if err != nil {
			return err
		}

		dockerArgs = append(dockerArgs, words...)
	}

	results, err = common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "docker-args-process-deploy",
		Args:    []string{appName, imageSourceType, imageTag},
		Stdin:   strings.NewReader(""),
	})
	if err == nil {
		words, err := shellquote.Split(results.StdoutContents())
		if err != nil {
			return err
		}

		dockerArgs = append(dockerArgs, words...)
	}

	filteredArgs := []string{
		"--cpus",
		"--gpus",
		"--memory",
		"--memory-reservation",
		"--memory-swap",
		"--publish",
		"--publish-all",
		"--restart",
		"-p",
		"-P",
	}
	for _, filteredArg := range filteredArgs {
		// re := regexp.MustCompile("--" + filteredArg + "=[0-9A-Za-z!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~]+ ")

		skipNext := false
		var filteredDockerArgs []string
		for _, dockerArg := range dockerArgs {
			if skipNext {
				skipNext = false
				continue
			}

			if strings.HasPrefix(dockerArg, filteredArg+"=") {
				continue
			}

			if dockerArg == filteredArg {
				skipNext = true
				continue
			}

			filteredDockerArgs = append(filteredDockerArgs, dockerArg)
		}

		dockerArgs = filteredDockerArgs
	}

	dockerArgs = append(dockerArgs, "--label=dokku_phase_script="+phase)
	if isHerokuishImage && !isCnbImage {
		dockerArgs = append(dockerArgs, "-v", fmt.Sprintf("cache-%s:/tmp/cache", appName))
	}
	if os.Getenv("DOKKU_TRACE") != "" {
		dockerArgs = append(dockerArgs, "--env", "DOKKU_TRACE="+os.Getenv("DOKKU_TRACE"))
	}
	if isCnbImage {
		// TODO: handle non-linux lifecycles
		// Ideally we don't have to override this but `pack` injects the web process
		// as the default entrypoint, so we need to specify the launcher so the script
		// runs as expected
		dockerArgs = append(dockerArgs, "--entrypoint=/cnb/lifecycle/launcher")
	}

	containerID, err := createdContainerID(appName, dockerArgs, image, script, phase)
	if err != nil {
		return fmt.Errorf("Failed to create %s execution container: %s", phase, err.Error())
	}

	defer cleanupDeploymentContainer(containerID, phase)

	if !waitForExecution(containerID) {
		common.LogInfo2Quiet(fmt.Sprintf("Start of %s %s task (%s) output", appName, phase, containerID[0:9]))
		common.LogVerboseQuietContainerLogs(containerID)
		common.LogInfo2Quiet(fmt.Sprintf("End of %s %s task (%s) output", appName, phase, containerID[0:9]))
		return fmt.Errorf("Execution of %s task failed: %s", phase, command)
	}

	common.LogInfo2Quiet(fmt.Sprintf("Start of %s %s task (%s) output", appName, phase, containerID[0:9]))
	common.LogVerboseQuietContainerLogs(containerID)
	common.LogInfo2Quiet(fmt.Sprintf("End of %s %s task (%s) output", appName, phase, containerID[0:9]))

	if phase != "predeploy" {
		return nil
	}

	commitArgs := []string{"container", "commit"}
	if !isHerokuishImage || isCnbImage {
		if dockerfileEntrypoint != "" {
			commitArgs = append(commitArgs, "--change", dockerfileEntrypoint)
		}

		if dockerfileCommand != "" {
			commitArgs = append(commitArgs, "--change", dockerfileCommand)
		}
	}

	commitArgs = append(commitArgs, []string{
		"--change",
		"LABEL org.label-schema.schema-version=1.0",
		"--change",
		"LABEL org.label-schema.vendor=dokku",
		"--change",
		fmt.Sprintf("LABEL com.dokku.app-name=%s", appName),
		"--change",
		fmt.Sprintf("LABEL com.dokku.%s-phase=true", phase),
	}...)
	commitArgs = append(commitArgs, containerID, image)
	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command:      common.DockerBin(),
		Args:         commitArgs,
		StreamStderr: true,
	})
	if err != nil {
		return fmt.Errorf("Committing of '%s' to image failed: %w", phase, err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("Committing of '%s' to image failed: %s", phase, command)
	}

	return nil
}

func getEntrypointFromImage(image string) (string, error) {
	output, err := common.DockerInspect(image, "{{json .Config.Entrypoint}}")
	if err != nil {
		return "", err
	}
	if output == "null" {
		return "", err
	}

	var entrypoint []string
	if err = json.Unmarshal([]byte(output), &entrypoint); err != nil {
		return "", err
	}

	if len(entrypoint) == 3 && entrypoint[0] == "/bin/sh" && entrypoint[1] == "-c" {
		return fmt.Sprintf("ENTRYPOINT %s", entrypoint[2]), nil
	}

	serializedEntrypoint, err := json.Marshal(entrypoint)
	return fmt.Sprintf("ENTRYPOINT %s", string(serializedEntrypoint)), err
}

func getCommandFromImage(image string) (string, error) {
	output, err := common.DockerInspect(image, "{{json .Config.Cmd}}")
	if err != nil {
		return "", err
	}
	if output == "null" {
		return "", err
	}

	var command []string
	if err = json.Unmarshal([]byte(output), &command); err != nil {
		return "", err
	}

	if len(command) == 3 && command[0] == "/bin/sh" && command[1] == "-c" {
		return fmt.Sprintf("CMD %s", command[2]), nil
	}

	serializedEntrypoint, err := json.Marshal(command)
	return fmt.Sprintf("CMD %s", string(serializedEntrypoint)), err
}

func waitForExecution(containerID string) bool {
	if !common.ContainerStart(containerID) {
		return false
	}

	return common.ContainerWait(containerID)
}

func createdContainerID(appName string, dockerArgs []string, image string, command []string, phase string) (string, error) {
	runLabelArgs := fmt.Sprintf("--label=com.dokku.app-name=%s", appName)

	arguments := strings.Split(common.MustGetEnv("DOKKU_GLOBAL_RUN_ARGS"), " ")
	arguments = append(arguments, runLabelArgs)
	arguments = append(arguments, dockerArgs...)

	arguments = append([]string{"container", "create"}, arguments...)
	arguments = append(arguments, image)
	arguments = append(arguments, command...)

	results, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "config-export",
		Args:    []string{appName, "false", "true", "json"},
	})
	if err != nil {
		return "", err
	}
	var env map[string]string
	if err := json.Unmarshal(results.StdoutBytes(), &env); err != nil {
		return "", err
	}

	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command: common.DockerBin(),
		Args:    arguments,
		Env:     env,
	})
	if err != nil {
		return "", err
	}
	if result.ExitCode != 0 {
		return "", errors.New(result.StderrContents())
	}

	containerID := result.StdoutContents()
	_, err = common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "post-container-create",
		Args:        []string{"app", containerID, appName, phase},
		StreamStdio: true,
	})
	return containerID, err
}

func setScale(appName string) error {
	appJSON, err := GetAppJSON(appName)
	if err != nil {
		return err
	}

	skipDeploy := true
	clearExisting := true
	args := []string{appName, strconv.FormatBool(skipDeploy), strconv.FormatBool(clearExisting)}
	for processType, formation := range appJSON.Formation {
		if formation.Quantity != nil {
			args = append(args, fmt.Sprintf("%s=%d", processType, *formation.Quantity))
		}
	}

	if len(args) == 3 {
		_, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
			Trigger:     "ps-can-scale",
			Args:        []string{appName, "true"},
			StreamStdio: true,
		})
		return err
	}

	_, err = common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "ps-can-scale",
		Args:        []string{appName, "false"},
		StreamStdio: true,
	})
	if err != nil {
		return err
	}

	_, err = common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "ps-set-scale",
		Args:        args,
		StreamStdio: true,
	})
	return err
}
