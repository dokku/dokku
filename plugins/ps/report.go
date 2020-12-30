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

	infoFlags := common.CollectReport(appName, flags)
	scheduler := common.GetAppScheduler(appName)
	if scheduler == "docker-local" {
		processStatus := getProcessStatus(appName)
		for process, value := range processStatus {
			infoFlags[fmt.Sprintf("--status-%s", process)] = value
		}
	}

	trimPrefix := false
	uppercaseFirstCharacter := true
	return common.ReportSingleApp("ps", appName, infoFlag, infoFlags, trimPrefix, uppercaseFirstCharacter)
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
