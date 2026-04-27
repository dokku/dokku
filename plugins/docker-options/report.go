package dockeroptions

import (
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp displays the docker options report for a single app
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	flags := map[string]common.ReportFunc{
		"--docker-options-build":  reportBuildOptions,
		"--docker-options-deploy": reportDeployOptions,
		"--docker-options-run":    reportRunOptions,
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	trimPrefix := false
	uppercaseFirstCharacter := true
	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp("docker options", appName, infoFlag, infoFlags, flagKeys, format, trimPrefix, uppercaseFirstCharacter)
}

func reportBuildOptions(appName string) string {
	return joinPhaseOptions(appName, "build")
}

func reportDeployOptions(appName string) string {
	return joinPhaseOptions(appName, "deploy")
}

func reportRunOptions(appName string) string {
	return joinPhaseOptions(appName, "run")
}

func joinPhaseOptions(appName string, phase string) string {
	options, err := GetDockerOptionsForPhase(appName, phase)
	if err != nil || len(options) == 0 {
		return ""
	}
	return strings.Join(options, " ")
}
