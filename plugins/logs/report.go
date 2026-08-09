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
			"--logs-computed-app-label-alias":  reportComputedAppLabelAlias,
			"--logs-computed-max-size":         reportComputedMaxSize,
			"--logs-computed-vector-cron-sink": reportComputedVectorCronSink,
			"--logs-computed-vector-image":     reportComputedVectorImage,
			"--logs-computed-vector-networks":  reportComputedVectorNetworks,
			"--logs-computed-vector-sink":      reportComputedVectorSink,
			"--logs-global-app-label-alias":    reportGlobalAppLabelAlias,
			"--logs-global-max-size":           reportGlobalMaxSize,
			"--logs-global-vector-cron-sink":   reportGlobalVectorCronSink,
			"--logs-global-vector-image":       reportGlobalVectorImage,
			"--logs-global-vector-networks":    reportGlobalVectorNetworks,
			"--logs-global-vector-sink":        reportGlobalVectorSink,
		}
	} else {
		flags = map[string]common.ReportFunc{
			"--logs-app-label-alias":           reportAppLabelAlias,
			"--logs-computed-app-label-alias":  reportComputedAppLabelAlias,
			"--logs-computed-max-size":         reportComputedMaxSize,
			"--logs-computed-vector-cron-sink": reportComputedVectorCronSink,
			"--logs-computed-vector-image":     reportComputedVectorImage,
			"--logs-computed-vector-networks":  reportComputedVectorNetworks,
			"--logs-computed-vector-sink":      reportComputedVectorSink,
			"--logs-global-app-label-alias":    reportGlobalAppLabelAlias,
			"--logs-global-max-size":           reportGlobalMaxSize,
			"--logs-global-vector-cron-sink":   reportGlobalVectorCronSink,
			"--logs-global-vector-image":       reportGlobalVectorImage,
			"--logs-global-vector-networks":    reportGlobalVectorNetworks,
			"--logs-global-vector-sink":        reportGlobalVectorSink,
			"--logs-max-size":                  reportMaxSize,
			"--logs-vector-cron-sink":          reportVectorCronSink,
			"--logs-vector-sink":               reportVectorSink,
		}
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp(common.ReportSingleAppInput{
		ReportType:              "logs",
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
	return redactedSink(common.PropertyGet("logs", "--global", "vector-sink"), "--logs-global-vector-sink")
}

func reportComputedVectorCronSink(appName string) string {
	value := reportVectorCronSink(appName)
	if value == "" {
		value = reportGlobalVectorCronSink(appName)
	}
	return value
}

func reportGlobalVectorCronSink(appName string) string {
	return redactedSink(common.PropertyGet("logs", "--global", "vector-cron-sink"), "--logs-global-vector-cron-sink")
}

func reportVectorCronSink(appName string) string {
	return redactedSink(common.PropertyGet("logs", appName, "vector-cron-sink"), "--logs-vector-cron-sink")
}

// redactedSink returns the sink value as-is for json reports or when the exact
// flag was requested, and otherwise reduces it to its scheme so that
// credentials embedded in the DSN are not printed in a general report
func redactedSink(value string, exactFlag string) string {
	if value == "" {
		return value
	}

	if os.Getenv("DOKKU_REPORT_FORMAT") != "stdout" {
		return value
	}

	if os.Getenv("DOKKU_REPORT_FLAG") == exactFlag {
		return value
	}

	sink, err := SinkValueToConfig(SinkValueToConfigInput{SinkValue: value})
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%s://redacted", sink["type"])
}

func reportMaxSize(appName string) string {
	return common.PropertyGet("logs", appName, "max-size")
}

func reportVectorSink(appName string) string {
	return redactedSink(common.PropertyGet("logs", appName, "vector-sink"), "--logs-vector-sink")
}
