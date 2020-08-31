package appjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

type AppJson struct {
	Scripts struct {
		Dokku struct {
			Predeploy  string `json:"predeploy"`
			Postdeploy string `json:"postdeploy"`
		} `json:"dokku"`
	} `json:"scripts"`
}

func constructScript(command string, isHerokuishImage bool) string {
	script := []string{"set -eo pipefail;"}
	if os.Getenv("DOKKU_TRACE") == "1" {
		script = append(script, "set -x;")
	}

	if isHerokuishImage {
		script = append(script, []string{
			"if [[ -d '/app' ]]; then",
			"  export HOME=/app;",
			"  cd $HOME;",
			"fi;",
			"if [[ -d '/app/.profile.d' ]]; then",
			"  for file in /app/.profile.d/*; do source $file; done;",
			"fi;",

			"if [[ -d '/cache' ]]; then",
			"  rm -rf /tmp/cache ;",
			"  ln -sf /cache /tmp/cache;",
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

	if isHerokuishImage {
		script = append(script, []string{
			"if [[ -d '/cache' ]]; then",
			"  rm -f /tmp/cache;",
			"fi;",
		}...)
	}

	return strings.Join(script, " ")
}

// getPhaseScript extracts app.json from app image and returns the appropriate json key/value
func getPhaseScript(appName string, image string, phase string) (string, error) {
	appJsonFile, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("dokku-%s-%s", common.MustGetEnv("DOKKU_PID"), "getPhaseScript"))
	if err != nil {
		return "", fmt.Errorf("Cannot create temporary file: %v", err)
	}

	defer os.Remove(appJsonFile.Name())

	common.CopyFromImage(appName, image, "app.json", appJsonFile.Name())
	if !common.FileExists(appJsonFile.Name()) {
		return "", nil
	}

	b, err := ioutil.ReadFile(appJsonFile.Name())
	if err != nil {
		return "", fmt.Errorf("Cannot read app.json file: %v", err)
	}

	if strings.TrimSpace(string(b)) == "" {
		return "", nil
	}

	var appJson AppJson
	if err = json.Unmarshal(b, &appJson); err != nil {
		return "", fmt.Errorf("Cannot parse app.json: %v", err)
	}

	if phase == "predeploy" {
		return appJson.Scripts.Dokku.Predeploy, nil
	}

	return appJson.Scripts.Dokku.Postdeploy, nil
}

// getReleaseCommand extracts the release command from a given app's procfile
func getReleaseCommand(appName string, image string) string {
	forceExtract := "true"
	if err := common.PlugnTrigger("procfile-extract", []string{appName, image, forceExtract}...); err != nil {
		return ""
	}

	processType := "release"
	port := "5000"
	b, _ := common.PlugnTriggerOutput("procfile-get-command", []string{appName, processType, port}...)
	return strings.TrimSpace(string(b[:]))
}

func executeScript(appName string, imageTag string, phase string) error {
	image := common.GetDeployingAppImageName(appName, imageTag, "")
	command := ""
	phaseSource := ""
	if phase == "release" {
		command = getReleaseCommand(appName, imageTag)
		phaseSource = "Procfile"
	} else {
		var err error
		phaseSource = "app.json"
		if command, err = getPhaseScript(appName, image, phase); err != nil {
			common.LogExclaim(err.Error())
		}
	}

	if command == "" {
		return nil
	}

	common.LogInfo1(fmt.Sprintf("Executing %s command from %s: %s", phase, phaseSource, command))
	isHerokuishImage := common.IsImageHerokuishBased(image, appName)
	script := constructScript(command, isHerokuishImage)

	imageSourceType := "dockerfile"
	if isHerokuishImage {
		imageSourceType = "herokuish"
	}

	cacheDir := fmt.Sprintf("%s/cache", common.AppRoot(appName))
	cacheHostDir := fmt.Sprintf("%s/cache", common.AppHostRoot(appName))
	if !common.DirectoryExists(cacheDir) {
		os.MkdirAll(cacheDir, 0755)
	}

	var dockerArgs []string
	if b, err := common.PlugnTriggerSetup("docker-args-deploy", []string{appName, imageTag}...).SetInput("").Output(); err != nil {
		dockerArgs = append(dockerArgs, strings.Split(strings.TrimSpace(string(b[:])), "\n")...)
	}

	if b, err := common.PlugnTriggerSetup("docker-args-process-deploy", []string{appName, imageSourceType, imageTag}...).SetInput("").Output(); err != nil {
		dockerArgs = append(dockerArgs, strings.Split(strings.TrimSpace(string(b[:])), "\n")...)
	}

	filteredArgs := []string{"restart", "cpus", "memory", "memory-swap", "memory-reservation", "gpus"}
	for _, filteredArg := range filteredArgs {
		// re := regexp.MustCompile("--" + filteredArg + "=[0-9A-Za-z!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~]+ ")

		var filteredDockerArgs []string
		for _, dockerArg := range dockerArgs {
			if !strings.HasPrefix(dockerArg, "--"+filteredArg+"=") {
				filteredDockerArgs = append(filteredDockerArgs, dockerArg)
			}
		}

		dockerArgs = filteredDockerArgs
	}

	dockerArgs = append(dockerArgs, "--label=dokku_phase_script="+phase)
	dockerArgs = append(dockerArgs, "-v", cacheHostDir+":/cache")
	if os.Getenv("DOKKU_TRACE") != "" {
		dockerArgs = append(dockerArgs, "--env", "DOKKU_TRACE="+os.Getenv("DOKKU_TRACE"))
	}

	dokkuAppShell := "/bin/bash"
	if b, _ := common.PlugnTriggerOutput("config-get-global", []string{"DOKKU_APP_SHELL"}...); strings.TrimSpace(string(b[:])) != "" {
		dokkuAppShell = strings.TrimSpace(string(b[:]))
	}

	if b, _ := common.PlugnTriggerOutput("config-get", []string{appName, "DOKKU_APP_SHELL"}...); strings.TrimSpace(string(b[:])) != "" {
		dokkuAppShell = strings.TrimSpace(string(b[:]))
	}

	containerCommand := []string{dokkuAppShell, "-c", script}
	containerID, _ := createdContainerID(appName, dockerArgs, image, containerCommand, phase)
	common.LogVerboseQuietContainerLogs(containerID)
	if !waitForExecution(containerID) {
		common.LogFail(fmt.Sprintf("Execution of '%s' command failed: %s", phase, command))
	}

	if phase != "predeploy" {
		return nil
	}

	commitArgs := []string{"container", "commit"}
	if !isHerokuishImage {
		dockerfileEntrypoint, _ := getEntrypointFromImage(image)
		if dockerfileEntrypoint != "" {
			commitArgs = append(commitArgs, "--change", dockerfileEntrypoint)
		}

		dockerfileCommand, _ := getCommandFromImage(image)
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
	containerCommitCmd := common.NewShellCmdWithArgs(
		common.DockerBin(),
		commitArgs...,
	)
	containerCommitCmd.ShowOutput = false
	containerCommitCmd.Command.Stderr = os.Stderr
	if !containerCommitCmd.Execute() {
		common.LogFail(fmt.Sprintf("Commiting of '%s' to image failed: %s", phase, command))
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
	containerStartCmd := common.NewShellCmdWithArgs(
		common.DockerBin(),
		"container",
		"start",
		containerID,
	)
	containerStartCmd.ShowOutput = false
	containerStartCmd.Command.Stderr = os.Stderr
	if !containerStartCmd.Execute() {
		return false
	}

	containerWaitCmd := common.NewShellCmdWithArgs(
		common.DockerBin(),
		"container",
		"wait",
		containerID,
	)

	containerWaitCmd.ShowOutput = false
	containerWaitCmd.Command.Stderr = os.Stderr
	b, err := containerWaitCmd.Output()
	if err != nil {
		return false
	}

	containerExitCode := strings.TrimSpace(string(b[:]))
	return containerExitCode == "0"
}

func createdContainerID(appName string, dockerArgs []string, image string, command []string, phase string) (string, error) {
	runLabelArgs := fmt.Sprintf("--label=com.dokku.app-name=%s", appName)

	arguments := strings.Split(common.MustGetEnv("DOKKU_GLOBAL_RUN_ARGS"), " ")
	arguments = append(arguments, runLabelArgs)
	arguments = append(arguments, dockerArgs...)

	arguments = append([]string{"container", "create"}, arguments...)
	arguments = append(arguments, image)
	arguments = append(arguments, command...)

	containerCreateCmd := common.NewShellCmdWithArgs(
		common.DockerBin(),
		arguments...,
	)
	var stderr bytes.Buffer
	containerCreateCmd.ShowOutput = false
	containerCreateCmd.Command.Stderr = &stderr

	b, err := containerCreateCmd.Output()
	if err != nil {
		return "", err
	}

	containerID := strings.TrimSpace(string(b))
	err = common.PlugnTrigger("post-container-create", []string{"app", appName, containerID, phase}...)
	return containerID, err
}
