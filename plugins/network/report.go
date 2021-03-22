package network

import (
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the network report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	flags := map[string]common.ReportFunc{
		"--network-bind-all-interfaces":      reportBindAllInterfaces,
		"--network-attach-post-create":       reportAttachPostCreate,
		"--network-attach-post-deploy":       reportAttachPostDeploy,
		"--network-computed-initial-network": reportComputedInitialNetwork,
		"--network-global-initial-network":   reportGlobalInitialNetwork,
		"--network-initial-network":          reportInitialNetwork,
		"--network-web-listeners":            reportWebListeners,
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	trimPrefix := false
	uppercaseFirstCharacter := true
	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp("network", appName, infoFlag, infoFlags, flagKeys, format, trimPrefix, uppercaseFirstCharacter)
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

func reportComputedInitialNetwork(appName string) string {
	value := reportInitialNetwork(appName)
	if value == "" {
		value = reportGlobalInitialNetwork(appName)
	}

	return value
}

func reportGlobalInitialNetwork(appName string) string {
	return common.PropertyGet("network", "--global", "initial-network")
}

func reportInitialNetwork(appName string) string {
	return common.PropertyGet("network", appName, "initial-network")
}

func reportWebListeners(appName string) string {
	return strings.Join(GetListeners(appName, "web"), " ")
}
