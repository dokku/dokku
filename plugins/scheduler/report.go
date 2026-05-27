package scheduler

import (
	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the scheduler report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	var flags map[string]common.ReportFunc
	if appName == "--global" {
		flags = map[string]common.ReportFunc{
			"--scheduler-computed-selected": reportComputedSelected,
			"--scheduler-global-selected":   reportGlobalSelected,
			"--scheduler-computed-shell":    reportComputedShell,
			"--scheduler-global-shell":      reportGlobalShell,
		}
	} else {
		flags = map[string]common.ReportFunc{
			"--scheduler-computed-selected": reportComputedSelected,
			"--scheduler-global-selected":   reportGlobalSelected,
			"--scheduler-selected":          reportSelected,
			"--scheduler-computed-shell":    reportComputedShell,
			"--scheduler-global-shell":      reportGlobalShell,
			"--scheduler-shell":             reportShell,
		}
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp(common.ReportSingleAppInput{
		ReportType:              "scheduler",
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

func reportComputedSelected(appName string) string {
	value := reportSelected(appName)
	if value == "" {
		value = reportGlobalSelected(appName)
	}
	if value == "" {
		value = "docker-local"
	}

	return value
}

func reportGlobalSelected(appName string) string {
	return common.PropertyGet("scheduler", "--global", "selected")
}

func reportSelected(appName string) string {
	return common.PropertyGet("scheduler", appName, "selected")
}

func reportComputedShell(appName string) string {
	value := reportShell(appName)
	if value == "" {
		value = reportGlobalShell(appName)
	}

	return value
}

func reportGlobalShell(appName string) string {
	return common.PropertyGet("scheduler", "--global", "shell")
}

func reportShell(appName string) string {
	return common.PropertyGet("scheduler", appName, "shell")
}
