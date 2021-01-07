package ps

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the ps report for one or more apps
func ReportSingleApp(appName, infoFlag string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	flags := map[string]common.ReportFunc{
		"--deployed":          reportDeployed,
		"--processes":         reportProcesses,
		"--ps-can-scale":      reportCanScale,
		"--ps-restart-policy": reportRestartPolicy,
		"--restore":           reportRestore,
		"--running":           reportRunningState,
	}

	extraFlags := addStatusFlags(appName, infoFlag)
	for flag, fn := range extraFlags {
		flags[flag] = fn
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	trimPrefix := false
	uppercaseFirstCharacter := true
	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp("ps", appName, infoFlag, infoFlags, flagKeys, trimPrefix, uppercaseFirstCharacter)
}

func addStatusFlags(appName string, infoFlag string) map[string]common.ReportFunc {
	flags := map[string]common.ReportFunc{}

	if infoFlag != "" && !strings.HasPrefix(infoFlag, "--status-") {
		return flags
	}

	scheduler := common.GetAppScheduler(appName)
	if scheduler != "docker-local" {
		return flags
	}

	containerFiles := common.ListFilesWithPrefix(common.AppRoot(appName), "CONTAINER.")
	for _, filename := range containerFiles {
		// See https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		containerFilePath := filename
		process := strings.TrimPrefix(filename, fmt.Sprintf("%s/CONTAINER.", common.AppRoot(appName)))

		flags[fmt.Sprintf("--status-%s", process)] = func(appName string) string {
			containerID := common.ReadFirstLine(containerFilePath)
			containerStatus, _ := common.DockerInspect(containerID, "{{ .State.Status }}")

			if containerStatus == "" {
				containerStatus = "missing"
			}

			return fmt.Sprintf("%s (CID: %s)", containerStatus, containerID[0:11])
		}
	}

	return flags
}

func reportCanScale(appName string) string {
	canScale := "false"
	if canScaleApp(appName) {
		canScale = "true"
	}

	return canScale
}

func reportDeployed(appName string) string {
	deployed := "false"
	if common.IsDeployed(appName) {
		deployed = "true"
	}

	return deployed
}

func reportProcesses(appName string) string {
	count, err := getProcessCount(appName)
	if err != nil {
		count = -1
	}

	return strconv.Itoa(count)
}

func reportRestartPolicy(appName string) string {
	policy, _ := getRestartPolicy(appName)
	if policy == "" {
		policy = DefaultProperties["restart-policy"]
	}

	return policy
}

func reportRestore(appName string) string {
	b, _ := common.PlugnTriggerOutput("config-get", []string{appName, "DOKKU_APP_RESTORE"}...)
	restore := strings.TrimSpace(string(b[:]))
	if restore == "0" {
		restore = "false"
	} else {
		restore = "true"
	}

	return restore
}

func reportRunningState(appName string) string {
	return getRunningState(appName)
}
