package buildernixpacks

import (
	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the builder-nixpacks report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	var flags map[string]common.ReportFunc
	if appName == "--global" {
		flags = map[string]common.ReportFunc{
			"--builder-nixpacks-computed-nixpackstoml-path": reportComputedNixpackstomlPath,
			"--builder-nixpacks-global-nixpackstoml-path":   reportGlobalNixpackstomlPath,
		}
	} else {
		flags = map[string]common.ReportFunc{
			"--builder-nixpacks-computed-nixpackstoml-path": reportComputedNixpackstomlPath,
			"--builder-nixpacks-global-nixpackstoml-path":   reportGlobalNixpackstomlPath,
			"--builder-nixpacks-nixpackstoml-path":          reportNixpackstomlPath,
		}
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp(common.ReportSingleAppInput{
		ReportType:              "builder-nixpacks",
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

func reportNixpackstomlPath(appName string) string {
	return common.PropertyGet("builder-nixpacks", appName, "nixpackstoml-path")
}

func reportGlobalNixpackstomlPath(appName string) string {
	return common.PropertyGet("builder-nixpacks", "--global", "nixpackstoml-path")
}

func reportComputedNixpackstomlPath(appName string) string {
	value := reportNixpackstomlPath(appName)
	if value == "" {
		value = reportGlobalNixpackstomlPath(appName)
	}
	if value == "" {
		value = "nixpacks.toml"
	}

	return value
}
