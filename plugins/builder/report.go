package builder

import (
	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the builder report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	flags := map[string]common.ReportFunc{
		"--builder-computed-selected":  reportComputedSelected,
		"--builder-global-selected":    reportGlobalSelected,
		"--builder-selected":           reportSelected,
		"--builder-computed-build-dir": reportComputedBuildDir,
		"--builder-global-build-dir":   reportGlobalBuildDir,
		"--builder-build-dir":          reportBuildDir,
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	trimPrefix := false
	uppercaseFirstCharacter := true
	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp("builder", appName, infoFlag, infoFlags, flagKeys, format, trimPrefix, uppercaseFirstCharacter)
}

func reportComputedSelected(appName string) string {
	value := reportSelected(appName)
	if value == "" {
		value = reportGlobalSelected(appName)
	}

	return value
}

func reportGlobalSelected(appName string) string {
	return common.PropertyGet("builder", "--global", "selected")
}

func reportSelected(appName string) string {
	return common.PropertyGet("builder", appName, "selected")
}

func reportComputedBuildDir(appName string) string {
	value := reportBuildDir(appName)
	if value == "" {
		value = reportGlobalBuildDir(appName)
	}

	return value
}
func reportGlobalBuildDir(appName string) string {
	return common.PropertyGet("builder", "--global", "build-dir")
}

func reportBuildDir(appName string) string {
	return common.PropertyGet("builder", appName, "build-dir")
}
