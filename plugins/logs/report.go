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
		"--logs-vector-sink":        reportVectorSink,
		"--logs-global-vector-sink": reportGlobalVectorSink,
	}

	trimPrefix := false
	uppercaseFirstCharacter := true
	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp("logs", appName, infoFlag, infoFlags, trimPrefix, uppercaseFirstCharacter)
}

func reportVectorSink(appName string) string {
	return common.PropertyGet("logs", appName, "vector-sink")
}

func reportGlobalVectorSink(appName string) string {
	return common.PropertyGet("logs", "--global", "vector-sink")
}
