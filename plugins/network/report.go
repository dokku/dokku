package network

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
		"--network-bind-all-interfaces": reportBindAllInterfaces,
		"--network-attach-post-create":  reportAttachPostCreate,
		"--network-attach-post-deploy":  reportAttachPostDeploy,
		"--network-web-listeners":       reportWebListeners,
	}

	trimPrefix := false
	uppercaseFirstCharacter := true
	infoFlags := common.CollectReport(appName, flags)
	return common.ReportSingleApp("network", appName, infoFlag, infoFlags, trimPrefix, uppercaseFirstCharacter)
}

func reportBindAllInterfaces(appName string) string {
	return common.PropertyGet("network", appName, "bind-all-interfaces")
}

func reportAttachPostCreate(appName string) string {
	return common.PropertyGet("network", appName, "attach-post-create")
}

func reportAttachPostDeploy(appName string) string {
	return common.PropertyGet("network", appName, "attach-post-deploy")
}

func reportWebListeners(appName string) string {
	return strings.Join(GetListeners(appName, "web"), " ")
}
