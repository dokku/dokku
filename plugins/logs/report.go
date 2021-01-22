package logs

import (
	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the logs report for one or more apps
func ReportSingleApp(appName, infoFlag string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	flags := map[string]common.ReportFunc{
		"--logs-computed-max-size":  reportComputedMaxSize,
		"--logs-global-max-size":    reportGlobalMaxSize,
		"--logs-global-vector-sink": reportGlobalVectorSink,
		"--logs-max-size":           reportMaxSize,
		"--logs-vector-sink":        reportVectorSink,
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	trimPrefix := false
	uppercaseFirstCharacter := true
	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp("logs", appName, infoFlag, infoFlags, flagKeys, trimPrefix, uppercaseFirstCharacter)
}

func reportComputedMaxSize(appName string) string {
	value := reportMaxSize(appName)
	if value == "" {
		value = reportGlobalMaxSize(appName)
	}

	return value
}

func reportGlobalMaxSize(appName string) string {
	return common.PropertyGetDefault("logs", "--global", "max-size", MaxSize)
}

func reportGlobalVectorSink(appName string) string {
	return common.PropertyGet("logs", "--global", "vector-sink")
}

func reportMaxSize(appName string) string {
	return common.PropertyGet("logs", appName, "max-size")
}

func reportVectorSink(appName string) string {
	return common.PropertyGet("logs", appName, "vector-sink")
}
