package builderpack

import (
	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the builder-pack report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	var flags map[string]common.ReportFunc
	if appName == "--global" {
		flags = map[string]common.ReportFunc{
			"--builder-pack-computed-projecttoml-path": reportComputedProjecttomlPath,
			"--builder-pack-global-projecttoml-path":   reportGlobalProjecttomlPath,
		}
	} else {
		flags = map[string]common.ReportFunc{
			"--builder-pack-computed-projecttoml-path": reportComputedProjecttomlPath,
			"--builder-pack-global-projecttoml-path":   reportGlobalProjecttomlPath,
			"--builder-pack-projecttoml-path":          reportProjecttomlPath,
		}
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp(common.ReportSingleAppInput{
		ReportType:              "builder-pack",
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

func reportProjecttomlPath(appName string) string {
	return common.PropertyGet("builder-pack", appName, "projecttoml-path")
}

func reportGlobalProjecttomlPath(appName string) string {
	return common.PropertyGet("builder-pack", "--global", "projecttoml-path")
}

func reportComputedProjecttomlPath(appName string) string {
	value := reportProjecttomlPath(appName)
	if value == "" {
		value = reportGlobalProjecttomlPath(appName)
	}
	if value == "" {
		value = "project.toml"
	}

	return value
}
