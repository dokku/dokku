package builderdockerfile

import (
	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the builder-dockerfile report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	var flags map[string]common.ReportFunc
	if appName == "--global" {
		flags = map[string]common.ReportFunc{
			"--builder-dockerfile-computed-dockerfile-path": reportComputedDockerfilePath,
			"--builder-dockerfile-global-dockerfile-path":   reportGlobalDockerfilePath,
		}
	} else {
		flags = map[string]common.ReportFunc{
			"--builder-dockerfile-computed-dockerfile-path": reportComputedDockerfilePath,
			"--builder-dockerfile-global-dockerfile-path":   reportGlobalDockerfilePath,
			"--builder-dockerfile-dockerfile-path":          reportDockerfilePath,
		}
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp(common.ReportSingleAppInput{
		ReportType:              "builder-dockerfile",
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

func reportDockerfilePath(appName string) string {
	return common.PropertyGet("builder-dockerfile", appName, "dockerfile-path")
}

func reportGlobalDockerfilePath(appName string) string {
	return common.PropertyGet("builder-dockerfile", "--global", "dockerfile-path")
}

func reportComputedDockerfilePath(appName string) string {
	value := reportDockerfilePath(appName)
	if value == "" {
		value = reportGlobalDockerfilePath(appName)
	}
	if value == "" {
		value = "Dockerfile"
	}

	return value
}
