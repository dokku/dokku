package apps

import (
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the app report for one or more apps
func ReportSingleApp(appName, infoFlag string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	deploySource := ""
	if b, err := common.PlugnTriggerSetup("deploy-source", []string{appName}...).SetInput("").Output(); err != nil {
		deploySource = strings.TrimSpace(string(b[:]))
	}

	locked := "false"
	if appIsLocked(appName) {
		locked = "true"
	}

	infoFlags := map[string]string{
		"--app-dir":           common.AppRoot(appName),
		"--app-deploy-source": deploySource,
		"--app-locked":        locked,
	}

	trimPrefix := false
	uppercaseFirstCharacter := true
	return common.ReportSingleApp("app", appName, infoFlag, infoFlags, trimPrefix, uppercaseFirstCharacter)
}
