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

	flags := map[string]common.ReportFunc{
		"--app-dir":           reportDir,
		"--app-deploy-source": reportDeploySource,
		"--app-locked":        reportLocked,
	}

	trimPrefix := false
	uppercaseFirstCharacter := true
	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp("app", appName, infoFlag, infoFlags, trimPrefix, uppercaseFirstCharacter)
}

func reportDir(appName string) string {
	return common.AppRoot(appName)
}

func reportDeploySource(appName string) string {
	deploySource := ""
	if b, err := common.PlugnTriggerSetup("deploy-source", []string{appName}...).SetInput("").Output(); err != nil {
		deploySource = strings.TrimSpace(string(b[:]))
	}

	return deploySource
}

func reportLocked(appName string) string {
	locked := "false"
	if appIsLocked(appName) {
		locked = "true"
	}

	return locked
}
