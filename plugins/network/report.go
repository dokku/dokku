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
		"--network-bind-all-interfaces":          reportBindAllInterfaces,
		"--network-attach-post-create":           reportAttachPostCreate,
		"--network-attach-post-deploy":           reportAttachPostDeploy,
		"--network-computed-attach-post-create":  reportComputedAttachPostCreate,
		"--network-computed-attach-post-deploy":  reportComputedAttachPostDeploy,
		"--network-computed-bind-all-interfaces": reportComputedBindAllInterfaces,
		"--network-computed-initial-network":     reportComputedInitialNetwork,
		"--network-computed-tld":                 reportComputedTld,
		"--network-global-attach-post-create":    reportGlobalAttachPostCreate,
		"--network-global-attach-post-deploy":    reportGlobalAttachPostDeploy,
		"--network-global-bind-all-interfaces":   reportGlobalBindAllInterfaces,
		"--network-global-initial-network":       reportGlobalInitialNetwork,
		"--network-global-tld":                   reportGlobalTld,
		"--network-initial-network":              reportInitialNetwork,
		"--network-static-web-listener":          reportStaticWebListener,
		"--network-tld":                          reportTld,
		"--network-web-listeners":                reportWebListeners,
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

func reportAttachPostCreate(appName string) string {
	return common.PropertyGet("network", appName, "attach-post-create")
}

func reportAttachPostDeploy(appName string) string {
	return common.PropertyGet("network", appName, "attach-post-deploy")
}

func reportBindAllInterfaces(appName string) string {
	return common.PropertyGet("network", appName, "bind-all-interfaces")
}

func reportComputedAttachPostCreate(appName string) string {
	value := reportAttachPostCreate(appName)
	if value == "" {
		value = reportGlobalAttachPostCreate(appName)
	}

	return value
}

func reportComputedAttachPostDeploy(appName string) string {
	value := reportAttachPostDeploy(appName)
	if value == "" {
		value = reportGlobalAttachPostDeploy(appName)
	}

	return value
}

func reportComputedBindAllInterfaces(appName string) string {
	value := reportBindAllInterfaces(appName)
	if value == "" {
		value = reportGlobalBindAllInterfaces(appName)
	}

	return value
}

func reportComputedInitialNetwork(appName string) string {
	value := reportInitialNetwork(appName)
	if value == "" {
		value = reportGlobalInitialNetwork(appName)
	}

	return value
}

func reportComputedTld(appName string) string {
	value := reportTld(appName)
	if value == "" {
		value = reportGlobalTld(appName)
	}

	return value
}

func reportGlobalAttachPostCreate(appName string) string {
	return common.PropertyGet("network", "--global", "attach-post-create")
}

func reportGlobalAttachPostDeploy(appName string) string {
	return common.PropertyGet("network", "--global", "attach-post-deploy")
}

func reportGlobalBindAllInterfaces(appName string) string {
	return common.PropertyGetDefault("network", "--global", "bind-all-interfaces", "false")
}

func reportGlobalInitialNetwork(appName string) string {
	return common.PropertyGet("network", "--global", "initial-network")
}

func reportGlobalTld(appName string) string {
	return common.PropertyGet("network", "--global", "tld")
}

func reportInitialNetwork(appName string) string {
	return common.PropertyGet("network", appName, "initial-network")
}

func reportStaticWebListener(appName string) string {
	return common.PropertyGet("network", appName, "static-web-listener")
}

func reportTld(appName string) string {
	return common.PropertyGet("network", appName, "tld")
}

func reportWebListeners(appName string) string {
	return strings.Join(GetListeners(appName, "web"), " ")
}
