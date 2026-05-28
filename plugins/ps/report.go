package ps

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the ps report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	var flags map[string]common.ReportFunc
	if appName == "--global" {
		flags = map[string]common.ReportFunc{
			"--ps-computed-procfile-path":        reportComputedProcfilePath,
			"--ps-computed-restart-policy":       reportComputedRestartPolicy,
			"--ps-computed-skip-deploy":          reportComputedSkipDeploy,
			"--ps-computed-stop-timeout-seconds": reportComputedStopTimeoutSeconds,
			"--ps-global-procfile-path":          reportGlobalProcfilePath,
			"--ps-global-restart-policy":         reportGlobalRestartPolicy,
			"--ps-global-skip-deploy":            reportGlobalSkipDeploy,
			"--ps-global-stop-timeout-seconds":   reportGlobalStopTimeoutSeconds,
		}
	} else {
		flags = map[string]common.ReportFunc{
			"--deployed":                           reportDeployed,
			"--processes":                          reportProcesses,
			"--ps-can-scale":                       reportCanScale,
			"--ps-computed-dockerfile-start-cmd":   reportComputedDockerfileStartCmd,
			"--ps-computed-procfile-path":          reportComputedProcfilePath,
			"--ps-computed-restart-policy":         reportComputedRestartPolicy,
			"--ps-computed-skip-deploy":            reportComputedSkipDeploy,
			"--ps-computed-start-cmd":              reportComputedStartCmd,
			"--ps-computed-stop-timeout-seconds":   reportComputedStopTimeoutSeconds,
			"--ps-dockerfile-start-cmd":            reportDockerfileStartCmd,
			"--ps-global-procfile-path":            reportGlobalProcfilePath,
			"--ps-global-restart-policy":           reportGlobalRestartPolicy,
			"--ps-global-skip-deploy":              reportGlobalSkipDeploy,
			"--ps-global-stop-timeout-seconds":     reportGlobalStopTimeoutSeconds,
			"--ps-procfile-path":                   reportProcfilePath,
			"--ps-restart-policy":                  reportRestartPolicy,
			"--ps-skip-deploy":                     reportSkipDeploy,
			"--ps-start-cmd":                       reportStartCmd,
			"--ps-stop-timeout-seconds":            reportStopTimeoutSeconds,
			"--restore":                            reportRestore,
			"--running":                            reportRunningState,
		}

		extraFlags := addStatusFlags(appName, infoFlag)
		for flag, fn := range extraFlags {
			flags[flag] = fn
		}
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp(common.ReportSingleAppInput{
		ReportType:              "ps",
		AppName:                 appName,
		InfoFlag:                infoFlag,
		InfoFlags:               infoFlags,
		InfoFlagKeys:            flagKeys,
		Format:                  format,
		TrimPrefix:              true,
		UppercaseFirstCharacter: true,
		EmitLegacyPrefix:        true,
	})
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

func reportComputedDockerfileStartCmd(appName string) string {
	return reportDockerfileStartCmd(appName)
}

func reportDockerfileStartCmd(appName string) string {
	return common.PropertyGet("ps", appName, "dockerfile-start-cmd")
}

func reportComputedProcfilePath(appName string) string {
	value := reportProcfilePath(appName)
	if value == "" {
		value = reportGlobalProcfilePath(appName)
	}
	if value == "" {
		value = "Procfile"
	}

	return value
}

func reportGlobalProcfilePath(appName string) string {
	return common.PropertyGet("ps", "--global", "procfile-path")
}

func reportProcfilePath(appName string) string {
	return common.PropertyGetDefault("ps", appName, "procfile-path", "")
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

func reportComputedRestartPolicy(appName string) string {
	value := reportRestartPolicy(appName)
	if value == "" {
		value = reportGlobalRestartPolicy(appName)
	}
	if value == "" {
		value = DefaultProperties["restart-policy"]
	}

	return value
}

func reportGlobalRestartPolicy(appName string) string {
	return common.PropertyGet("ps", "--global", "restart-policy")
}

func reportRestartPolicy(appName string) string {
	return common.PropertyGet("ps", appName, "restart-policy")
}

func reportRestore(appName string) string {
	return common.PropertyGetDefault("ps", appName, "restore", "true")
}

func reportComputedSkipDeploy(appName string) string {
	value := reportSkipDeploy(appName)
	if value == "" {
		value = reportGlobalSkipDeploy(appName)
	}
	if value == "" {
		value = "false"
	}

	return value
}

func reportGlobalSkipDeploy(appName string) string {
	return common.PropertyGet("ps", "--global", "skip-deploy")
}

func reportSkipDeploy(appName string) string {
	return common.PropertyGet("ps", appName, "skip-deploy")
}

func reportComputedStartCmd(appName string) string {
	return reportStartCmd(appName)
}

func reportStartCmd(appName string) string {
	return common.PropertyGet("ps", appName, "start-cmd")
}

func reportRunningState(appName string) string {
	return getRunningState(appName)
}

func reportComputedStopTimeoutSeconds(appName string) string {
	value := reportStopTimeoutSeconds(appName)
	if value == "" {
		value = reportGlobalStopTimeoutSeconds(appName)
	}
	if value == "" {
		value = "30"
	}

	return value
}

func reportGlobalStopTimeoutSeconds(appName string) string {
	return common.PropertyGet("ps", "--global", "stop-timeout-seconds")
}

func reportStopTimeoutSeconds(appName string) string {
	return common.PropertyGet("ps", appName, "stop-timeout-seconds")
}
