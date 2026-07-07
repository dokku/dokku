package checks

import (
	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the checks report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	var flags map[string]common.ReportFunc
	if appName == "--global" {
		flags = map[string]common.ReportFunc{
			"--checks-computed-wait-to-retire": reportComputedWaitToRetire,
			"--checks-global-wait-to-retire":   reportGlobalWaitToRetire,
		}
	} else {
		flags = map[string]common.ReportFunc{
			"--checks-disabled-list":           reportDisabledList,
			"--checks-skipped-list":            reportSkippedList,
			"--checks-computed-wait-to-retire": reportComputedWaitToRetire,
			"--checks-global-wait-to-retire":   reportGlobalWaitToRetire,
			"--checks-wait-to-retire":          reportWaitToRetire,
		}
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp(common.ReportSingleAppInput{
		ReportType:              "checks",
		AppName:                 appName,
		InfoFlag:                infoFlag,
		InfoFlags:               infoFlags,
		InfoFlagKeys:            flagKeys,
		Format:                  format,
		TrimPrefix:              true,
		UppercaseFirstCharacter: true,
		EmitLegacyPrefix:        false,
	})
}

func reportDisabledList(appName string) string {
	value := common.PropertyGet("checks", appName, "disabled")
	if value == "" {
		value = "none"
	}

	return value
}

func reportSkippedList(appName string) string {
	value := common.PropertyGet("checks", appName, "skipped")
	if value == "" {
		value = "none"
	}

	return value
}

func reportWaitToRetire(appName string) string {
	return common.PropertyGet("checks", appName, "wait-to-retire")
}

func reportGlobalWaitToRetire(appName string) string {
	return common.PropertyGet("checks", "--global", "wait-to-retire")
}

func reportComputedWaitToRetire(appName string) string {
	value := reportWaitToRetire(appName)
	if value == "" {
		value = reportGlobalWaitToRetire(appName)
	}
	if value == "" {
		value = "60"
	}

	return value
}
