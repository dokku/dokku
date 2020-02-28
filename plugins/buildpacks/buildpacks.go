package buildpacks

import (
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the app report for one or more apps
func ReportSingleApp(appName, infoFlag string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	buildpacks, err := common.PropertyListGet("buildpacks", appName, "buildpacks")
	if err != nil {
		return err
	}

	infoFlags := map[string]string{
		"--buildpacks-list": strings.Join(buildpacks, ","),
	}

	trimPrefix := false
	uppercaseFirstCharacter := true
	return common.ReportSingleApp("buildpacks", appName, infoFlag, infoFlags, trimPrefix, uppercaseFirstCharacter)
}
