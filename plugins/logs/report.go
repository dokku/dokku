package logs

import (
	"fmt"
	"os"

	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the logs report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	os.Setenv("DOKKU_REPORT_FORMAT", format)
	os.Setenv("DOKKU_REPORT_FLAG", infoFlag)
	flags := map[string]common.ReportFunc{
		"--logs-computed-app-label-alias": reportComputedAppLabelAlias,
		"--logs-computed-max-size":        reportComputedMaxSize,
		"--logs-global-app-label-alias":   reportGlobalAppLabelAlias,
		"--logs-global-max-size":          reportGlobalMaxSize,
		"--logs-global-vector-sink":       reportGlobalVectorSink,
		"--logs-app-label-alias":          reportAppLabelAlias,
		"--logs-max-size":                 reportMaxSize,
		"--logs-vector-global-image":      reportVectorGlobalImage,
		"--logs-vector-sink":              reportVectorSink,
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	trimPrefix := false
	uppercaseFirstCharacter := true
	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp("logs", appName, infoFlag, infoFlags, flagKeys, format, trimPrefix, uppercaseFirstCharacter)
}

func reportComputedAppLabelAlias(appName string) string {
	value := reportAppLabelAlias(appName)
	if value == "" {
		value = reportGlobalAppLabelAlias(appName)
	}

	return value
}

func reportGlobalAppLabelAlias(appName string) string {
	return common.PropertyGetDefault("logs", "--global", "app-label-alias", AppLabelAlias)
}

func reportAppLabelAlias(appName string) string {
	return common.PropertyGet("logs", appName, "app-label-alias")
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

func reportVectorGlobalImage(appName string) string {
	return getComputedVectorImage()
}

func reportGlobalVectorSink(appName string) string {
	value := common.PropertyGet("logs", "--global", "vector-sink")
	if value == "" {
		return value
	}

	if os.Getenv("DOKKU_REPORT_FORMAT") != "stdout" {
		return value
	}

	if os.Getenv("DOKKU_REPORT_FLAG") == "--logs-global-vector-sink" {
		return value
	}

	// only show the schema and sanitize the rest
	sink, err := SinkValueToConfig("--global", value)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%s://redacted", sink["type"])
}

func reportMaxSize(appName string) string {
	return common.PropertyGet("logs", appName, "max-size")
}

func reportVectorSink(appName string) string {
	value := common.PropertyGet("logs", appName, "vector-sink")
	if value == "" {
		return value
	}

	if os.Getenv("DOKKU_REPORT_FORMAT") != "stdout" {
		return value
	}

	if os.Getenv("DOKKU_REPORT_FLAG") == "--logs-vector-sink" {
		return value
	}

	// only show the schema and sanitize the rest
	sink, err := SinkValueToConfig(appName, value)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%s://redacted", sink["type"])
}
