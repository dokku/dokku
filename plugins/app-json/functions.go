package appjson

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	shellquote "github.com/kballard/go-shellquote"
	"golang.org/x/sync/errgroup"
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

	if isHerokuishImage && !isCnbImage {
		script = append(script, []string{
			"if [[ -d '/cache' ]]; then",
			"  rm -f /tmp/cache;",
			"fi;",
		}...)
	}

	return []string{shell, "-c", strings.Join(script, " ")}
}

func getAppJSON(appName string) (AppJSON, error) {
	if !common.FileExists(GetAppjsonPath(appName)) {
		return AppJSON{}, nil
	}

	b, err := ioutil.ReadFile(GetAppjsonPath(appName))
	if err != nil {
		return AppJSON{}, fmt.Errorf("Cannot read app.json file: %v", err)
	}

	if strings.TrimSpace(string(b)) == "" {
		return AppJSON{}, nil
	}

	var appJSON AppJSON
	if err = json.Unmarshal(b, &appJSON); err != nil {
		return AppJSON{}, fmt.Errorf("Cannot parse app.json: %v", err)
	}

	return appJSON, nil
}

// getPhaseScript extracts app.json from app image and returns the appropriate json key/value
func getPhaseScript(appName string, phase string) (string, error) {
	appJSON, err := getAppJSON(appName)
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
func getReleaseCommand(appName string, image string) string {
	err := common.SuppressOutput(func() error {
		return common.PlugnTrigger("procfile-extract", []string{appName, image}...)
	})

	if err != nil {
		return ""
	}

	processType := "release"
	port := "5000"
	b, _ := common.PlugnTriggerOutput("procfile-get-command", []string{appName, processType, port}...)
	return strings.TrimSpace(string(b[:]))
}

func getDokkuAppShell(appName string) string {
	shell := "/bin/bash"
	globalShell := ""
	appShell := ""

	ctx := context.Background()
	errs, ctx := errgroup.WithContext(ctx)
	errs.Go(func() error {
		b, _ := common.PlugnTriggerOutput("config-get-global", []string{"DOKKU_APP_SHELL"}...)
		globalShell = strings.TrimSpace(string(b[:]))
		return nil
	})
	errs.Go(func() error {
		b, _ := common.PlugnTriggerOutput("config-get", []string{appName, "DOKKU_APP_SHELL"}...)
		appShell = strings.TrimSpace(string(b[:]))
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

func cleanupDeploymentContainer(appName string, containerID string, phase string) error {
	if phase != "predeploy" {
		os.Setenv("DOKKU_SKIP_IMAGE_RETIRE", "true")
	}
	return common.PlugnTrigger("scheduler-register-retired", []string{appName, containerID}...)
}

func executeScript(appName string, image string, imageTag string, phase string) error {
	phaseName := phase
	if phase == "heroku.postdeploy" {
		phaseName = "first deploy postdeploy"
	}
	common.LogInfo1(fmt.Sprintf("Checking for %s task", phaseName))
	command := ""
	phaseSource := ""
	if phase == "release" {
		command = getReleaseCommand(appName, image)
		phaseSource = "Procfile"
	} else {
		var err error
		phaseSource = "app.json"
		if command, err = getPhaseScript(appName, phase); err != nil {
			common.LogExclaim(err.Error())
		}
	}

	if command == "" {
		common.LogVerbose(fmt.Sprintf("No %s task found, skipping", phaseName))
		return nil
	}

	common.LogInfo1(fmt.Sprintf("Executing %s task from %s: %s", phaseName, phaseSource, command))
	isHerokuishImage := common.IsImageHerokuishBased(image, appName)
	isCnbImage := common.IsImageCnbBased(image)
	dockerfileEntrypoint := ""
	dockerfileCommand := ""
	if !isHerokuishImage {
		dockerfileEntrypoint, _ = getEntrypointFromImage(image)
		dockerfileCommand, _ = getCommandFromImage(image)
	}

	dokkuAppShell := getDokkuAppShell(appName)
	script := constructScript(command, dokkuAppShell, isHerokuishImage, isCnbImage, dockerfileEntrypoint)

	imageSourceType := "dockerfile"
	if isHerokuishImage {
		imageSourceType = "herokuish"
	} else if isCnbImage {
		imageSourceType = "pack"
	}

	cacheDir := fmt.Sprintf("%s/cache", common.AppRoot(appName))
	cacheHostDir := fmt.Sprintf("%s/cache", common.AppHostRoot(appName))
	if !common.DirectoryExists(cacheDir) {
		os.MkdirAll(cacheDir, 0755)
	}

	var dockerArgs []string
	if b, err := common.PlugnTriggerSetup("docker-args-deploy", []string{appName, imageTag}...).SetInput("").Output(); err == nil {
		words, err := shellquote.Split(strings.TrimSpace(string(b[:])))
		if err != nil {
			return err
		}

		dockerArgs = append(dockerArgs, words...)
	}

	if b, err := common.PlugnTriggerSetup("docker-args-process-deploy", []string{appName, imageSourceType, imageTag}...).SetInput("").Output(); err == nil {
		words, err := shellquote.Split(strings.TrimSpace(string(b[:])))
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
	dockerArgs = append(dockerArgs, "-v", cacheHostDir+":/cache")
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

	defer cleanupDeploymentContainer(appName, containerID, phase)

	if !waitForExecution(containerID) {
		common.LogInfo2Quiet(fmt.Sprintf("Start of %s %s task (%s) output", appName, phaseName, containerID[0:9]))
		common.LogVerboseQuietContainerLogs(containerID)
		common.LogInfo2Quiet(fmt.Sprintf("End of %s %s task (%s) output", appName, phaseName, containerID[0:9]))
		return fmt.Errorf("Execution of %s task failed: %s", phaseName, command)
	}

	common.LogInfo2Quiet(fmt.Sprintf("Start of %s %s task (%s) output", appName, phaseName, containerID[0:9]))
	common.LogVerboseQuietContainerLogs(containerID)
	common.LogInfo2Quiet(fmt.Sprintf("End of %s %s task (%s) output", appName, phaseName, containerID[0:9]))

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
	containerCommitCmd := common.NewShellCmdWithArgs(
		common.DockerBin(),
		commitArgs...,
	)
	containerCommitCmd.ShowOutput = false
	containerCommitCmd.Command.Stderr = os.Stderr
	if !containerCommitCmd.Execute() {
		return fmt.Errorf("Commiting of '%s' to image failed: %s", phase, command)
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

	b, err := common.PlugnTriggerOutput("config-export", []string{appName, "false", "true", "json"}...)
	if err != nil {
		return "", err
	}
	var env map[string]string
	if err := json.Unmarshal(b, &env); err != nil {
		return "", err
	}

	containerCreateCmd := common.NewShellCmdWithArgs(
		common.DockerBin(),
		arguments...,
	)
	var stderr bytes.Buffer
	containerCreateCmd.Env = env
	containerCreateCmd.ShowOutput = false
	containerCreateCmd.Command.Stderr = &stderr

	b, err = containerCreateCmd.Output()
	if err != nil {
		return "", errors.New(stderr.String())
	}

	containerID := strings.TrimSpace(string(b))
	err = common.PlugnTrigger("post-container-create", []string{"app", containerID, appName, phase}...)
	return containerID, err
}

func refreshAppJSON(appName string, image string) error {
	baseDirectory := filepath.Join(common.MustGetEnv("DOKKU_LIB_ROOT"), "data", "app-json")
	if !common.DirectoryExists(baseDirectory) {
		return errors.New("Run 'dokku plugin:install' to ensure the correct directories exist")
	}

	directory := GetAppjsonDirectory(appName)
	if !common.DirectoryExists(directory) {
		if err := os.MkdirAll(directory, 0755); err != nil {
			return err
		}
	}

	appjsonPath := GetAppjsonPath(appName)
	if common.FileExists(appjsonPath) {
		if err := os.Remove(appjsonPath); err != nil {
			return errors.New("Unable to remove previous app.json file")
		}
	}

	common.CopyFromImage(appName, image, reportComputedAppjsonpath(appName), appjsonPath)
	return nil
}

func setScale(appName string, image string) error {
	if err := injectDokkuScale(appName, image); err != nil {
		return err
	}

	appJSON, err := getAppJSON(appName)
	if err != nil {
		return err
	}

	skipDeploy := true
	clearExisting := false
	args := []string{appName, strconv.FormatBool(skipDeploy), strconv.FormatBool(clearExisting)}
	for processType, formation := range appJSON.Formation {
		args = append(args, fmt.Sprintf("%s=%d", processType, formation.Quantity))
	}

	if len(args) == 3 {
		return common.PlugnTrigger("ps-can-scale", []string{appName, "true"}...)
	}

	if err := common.PlugnTrigger("ps-can-scale", []string{appName, "false"}...); err != nil {
		return err
	}

	return common.PlugnTrigger("ps-set-scale", args...)
}

func injectDokkuScale(appName string, image string) error {
	appJSON, err := getAppJSON(appName)
	if err != nil {
		return err
	}

	baseDirectory := filepath.Join(common.MustGetEnv("DOKKU_LIB_ROOT"), "data", "app-json")
	if !common.DirectoryExists(baseDirectory) {
		return errors.New("Run 'dokku plugin:install' to ensure the correct directories exist")
	}

	dokkuScaleFile := filepath.Join(baseDirectory, "DOKKU_SCALE")
	previouslyExtracted := common.FileExists(dokkuScaleFile)
	if previouslyExtracted {
		os.Remove(dokkuScaleFile)
	}

	common.CopyFromImage(appName, image, "DOKKU_SCALE", dokkuScaleFile)

	if !common.FileExists(dokkuScaleFile) {
		return nil
	}

	lines, err := common.FileToSlice(dokkuScaleFile)
	if err != nil {
		return err
	}

	if appJSON.Formation == nil {
		common.LogWarn("Deprecated: Injecting scale settings from DOKKU_SCALE file. Use the 'formation' key the app.json file to specify scaling instead of 'DOKKU_SCALE'.")
		appJSON.Formation = make(map[string]Formation)
	} else {
		common.LogWarn("Deprecated: DOKKU_SCALE ignored in favor of 'formation' key in the app.json")
		return nil
	}

	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		procParts := strings.SplitN(line, "=", 2)
		if len(procParts) != 2 {
			continue
		}

		processType := procParts[0]
		quantity, err := strconv.Atoi(procParts[1])
		if err != nil {
			continue
		}

		appJSON.Formation[processType] = Formation{
			Quantity: quantity,
		}
	}

	b, err := json.Marshal(appJSON)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(GetAppjsonPath(appName), b, 0644)
}
