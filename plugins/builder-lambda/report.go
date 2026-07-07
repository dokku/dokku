package builderlambda

import (
	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the builder-lambda report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	var flags map[string]common.ReportFunc
	if appName == "--global" {
		flags = map[string]common.ReportFunc{
			"--builder-lambda-computed-lambdayml-path": reportComputedLambdaymlPath,
			"--builder-lambda-global-lambdayml-path":   reportGlobalLambdaymlPath,
		}
	} else {
		flags = map[string]common.ReportFunc{
			"--builder-lambda-computed-lambdayml-path": reportComputedLambdaymlPath,
			"--builder-lambda-global-lambdayml-path":   reportGlobalLambdaymlPath,
			"--builder-lambda-lambdayml-path":          reportLambdaymlPath,
		}
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp(common.ReportSingleAppInput{
		ReportType:              "builder-lambda",
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

func reportLambdaymlPath(appName string) string {
	return common.PropertyGet("builder-lambda", appName, "lambdayml-path")
}

func reportGlobalLambdaymlPath(appName string) string {
	return common.PropertyGet("builder-lambda", "--global", "lambdayml-path")
}

func reportComputedLambdaymlPath(appName string) string {
	value := reportLambdaymlPath(appName)
	if value == "" {
		value = reportGlobalLambdaymlPath(appName)
	}
	if value == "" {
		value = "lambda.yml"
	}

	return value
}
