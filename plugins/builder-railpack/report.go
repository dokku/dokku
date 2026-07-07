package builderrailpack

import (
	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the builder-railpack report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	var flags map[string]common.ReportFunc
	if appName == "--global" {
		flags = map[string]common.ReportFunc{
			"--builder-railpack-computed-railpackjson-path": reportComputedRailpackjsonPath,
			"--builder-railpack-global-railpackjson-path":   reportGlobalRailpackjsonPath,
		}
	} else {
		flags = map[string]common.ReportFunc{
			"--builder-railpack-computed-railpackjson-path": reportComputedRailpackjsonPath,
			"--builder-railpack-global-railpackjson-path":   reportGlobalRailpackjsonPath,
			"--builder-railpack-railpackjson-path":          reportRailpackjsonPath,
		}
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp(common.ReportSingleAppInput{
		ReportType:              "builder-railpack",
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

func reportRailpackjsonPath(appName string) string {
	return common.PropertyGet("builder-railpack", appName, "railpackjson-path")
}

func reportGlobalRailpackjsonPath(appName string) string {
	return common.PropertyGet("builder-railpack", "--global", "railpackjson-path")
}

func reportComputedRailpackjsonPath(appName string) string {
	value := reportRailpackjsonPath(appName)
	if value == "" {
		value = reportGlobalRailpackjsonPath(appName)
	}
	if value == "" {
		value = "railpack.json"
	}

	return value
}
