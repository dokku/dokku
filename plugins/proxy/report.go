package proxy

import (
	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the proxy report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	var flags map[string]common.ReportFunc
	if appName == "--global" {
		flags = map[string]common.ReportFunc{
			"--proxy-computed-proxy-port":     reportComputedProxyPort,
			"--proxy-computed-proxy-ssl-port": reportComputedProxySSLPort,
			"--proxy-computed-type":           reportComputedType,
			"--proxy-global-proxy-port":       reportGlobalProxyPort,
			"--proxy-global-proxy-ssl-port":   reportGlobalProxySSLPort,
			"--proxy-global-type":             reportGlobalType,
		}
	} else {
		flags = map[string]common.ReportFunc{
			"--proxy-computed-disabled":       reportComputedDisabled,
			"--proxy-computed-proxy-port":     reportComputedProxyPort,
			"--proxy-computed-proxy-ssl-port": reportComputedProxySSLPort,
			"--proxy-computed-type":           reportComputedType,
			"--proxy-disabled":                reportDisabled,
			"--proxy-enabled":                 reportEnabled,
			"--proxy-global-proxy-port":       reportGlobalProxyPort,
			"--proxy-global-proxy-ssl-port":   reportGlobalProxySSLPort,
			"--proxy-global-type":             reportGlobalType,
			"--proxy-proxy-port":              reportProxyPort,
			"--proxy-proxy-ssl-port":          reportProxySSLPort,
			"--proxy-type":                    reportType,
		}
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp(common.ReportSingleAppInput{
		ReportType:              "proxy",
		AppName:                 appName,
		InfoFlag:                infoFlag,
		InfoFlags:               infoFlags,
		InfoFlagKeys:            flagKeys,
		Format:                  format,
		TrimPrefix:              true,
		UppercaseFirstCharacter: true,
		EmitLegacyPrefix:        true,
	})
}

func reportEnabled(appName string) string {
	proxyEnabled := "false"
	if IsAppProxyEnabled(appName) {
		proxyEnabled = "true"
	}

	return proxyEnabled
}

func reportComputedType(appName string) string {
	return getComputedProxyType(appName)
}

func reportGlobalType(appName string) string {
	return getGlobalProxyType()
}

func reportType(appName string) string {
	return getAppProxyType(appName)
}

func reportDisabled(appName string) string {
	return getAppDisabled(appName)
}

func reportComputedDisabled(appName string) string {
	return getComputedDisabled(appName)
}

func reportProxyPort(appName string) string {
	return getAppProxyPort(appName)
}

func reportGlobalProxyPort(appName string) string {
	return getGlobalProxyPort()
}

func reportComputedProxyPort(appName string) string {
	return getComputedProxyPort(appName)
}

func reportProxySSLPort(appName string) string {
	return getAppProxySSLPort(appName)
}

func reportGlobalProxySSLPort(appName string) string {
	return getGlobalProxySSLPort()
}

func reportComputedProxySSLPort(appName string) string {
	return getComputedProxySSLPort(appName)
}
