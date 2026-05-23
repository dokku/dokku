package logs

import (
	"fmt"
	"os"

	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the logs report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	os.Setenv("DOKKU_REPORT_FORMAT", format)
	os.Setenv("DOKKU_REPORT_FLAG", infoFlag)
	var flags map[string]common.ReportFunc
	if appName == "--global" {
		flags = map[string]common.ReportFunc{
			"--logs-computed-app-label-alias": reportComputedAppLabelAlias,
			"--logs-computed-max-size":        reportComputedMaxSize,
			"--logs-computed-vector-image":    reportComputedVectorImage,
			"--logs-computed-vector-networks": reportComputedVectorNetworks,
			"--logs-computed-vector-sink":     reportComputedVectorSink,
			"--logs-global-app-label-alias":   reportGlobalAppLabelAlias,
			"--logs-global-max-size":          reportGlobalMaxSize,
			"--logs-global-vector-image":      reportGlobalVectorImage,
			"--logs-global-vector-networks":   reportGlobalVectorNetworks,
			"--logs-global-vector-sink":       reportGlobalVectorSink,
		}
	} else {
		flags = map[string]common.ReportFunc{
			"--logs-app-label-alias":          reportAppLabelAlias,
			"--logs-computed-app-label-alias": reportComputedAppLabelAlias,
			"--logs-computed-max-size":        reportComputedMaxSize,
			"--logs-computed-vector-image":    reportComputedVectorImage,
			"--logs-computed-vector-networks": reportComputedVectorNetworks,
			"--logs-computed-vector-sink":     reportComputedVectorSink,
			"--logs-global-app-label-alias":   reportGlobalAppLabelAlias,
			"--logs-global-max-size":          reportGlobalMaxSize,
			"--logs-global-vector-image":      reportGlobalVectorImage,
			"--logs-global-vector-networks":   reportGlobalVectorNetworks,
			"--logs-global-vector-sink":       reportGlobalVectorSink,
			"--logs-max-size":                 reportMaxSize,
			"--logs-vector-sink":              reportVectorSink,
		}
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
	if value == "" {
		value = AppLabelAlias
	}

	return value
}

func reportGlobalAppLabelAlias(appName string) string {
	return common.PropertyGet("logs", "--global", "app-label-alias")
}

func reportAppLabelAlias(appName string) string {
	return common.PropertyGet("logs", appName, "app-label-alias")
}

func reportComputedMaxSize(appName string) string {
	value := reportMaxSize(appName)
	if value == "" {
		value = reportGlobalMaxSize(appName)
	}
	if value == "" {
		value = MaxSize
	}

	return value
}

func reportGlobalMaxSize(appName string) string {
	return common.PropertyGet("logs", "--global", "max-size")
}

func reportGlobalVectorImage(appName string) string {
	return common.PropertyGet("logs", "--global", "vector-image")
}

func reportComputedVectorImage(appName string) string {
	return getComputedVectorImage()
}

func reportGlobalVectorNetworks(appName string) string {
	return common.PropertyGet("logs", "--global", "vector-networks")
}

func reportComputedVectorNetworks(appName string) string {
	return reportGlobalVectorNetworks(appName)
}

func reportComputedVectorSink(appName string) string {
	value := reportVectorSink(appName)
	if value == "" {
		value = reportGlobalVectorSink(appName)
	}
	return value
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
