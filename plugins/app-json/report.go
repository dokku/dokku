package appjson

import (
	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the builder report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	flags := map[string]common.ReportFunc{
		"--app-json-computed-selected": reportComputedAppjsonpath,
		"--app-json-global-selected":   reportGlobalAppjsonpath,
		"--app-json-selected":          reportAppjsonpath,
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	trimPrefix := false
	uppercaseFirstCharacter := true
	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp("app-json", appName, infoFlag, infoFlags, flagKeys, format, trimPrefix, uppercaseFirstCharacter)
}

func reportComputedAppjsonpath(appName string) string {
	value := reportAppjsonpath(appName)
	if value == "" {
		value = reportGlobalAppjsonpath(appName)
	}

	return value
}

func reportGlobalAppjsonpath(appName string) string {
	return common.PropertyGetDefault("app-json", "--global", "appjson-path", "app.json")
}

func reportAppjsonpath(appName string) string {
	return common.PropertyGet("app-json", appName, "appjson-path")
}
