package dockeroptions

import (
	"fmt"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp displays the docker options report for a single app.
// Default-scope options are reported under fixed keys for each phase.
// Process-scoped options surface as dynamic per-process keys, one per
// configured process+phase combination.
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	var flags map[string]common.ReportFunc
	if appName == "--global" {
		flags = map[string]common.ReportFunc{}
	} else {
		flags = map[string]common.ReportFunc{
			"--docker-options-build":  reportBuildOptions,
			"--docker-options-deploy": reportDeployOptions,
			"--docker-options-run":    reportRunOptions,
		}
	}

	processTypes, err := ListProcessTypesWithOptions(appName)
	if err != nil {
		return err
	}
	for _, processType := range processTypes {
		processType := processType
		flagName := fmt.Sprintf("--docker-options-deploy.%s", processType)
		flags[flagName] = func(app string) string {
			return joinProcessPhaseOptions(app, processType, "deploy")
		}
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
	return joinProcessPhaseOptions(appName, DefaultProcessType, "build")
}

func reportDeployOptions(appName string) string {
	return joinProcessPhaseOptions(appName, DefaultProcessType, "deploy")
}

func reportRunOptions(appName string) string {
	return joinProcessPhaseOptions(appName, DefaultProcessType, "run")
}

func joinProcessPhaseOptions(appName, processType, phase string) string {
	options, err := GetDockerOptionsForProcessPhase(appName, processType, phase)
	if err != nil || len(options) == 0 {
		return ""
	}
	return strings.Join(options, " ")
}
