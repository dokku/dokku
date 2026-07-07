package nginxvhosts

import (
	"os"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

type nginxReportProp struct {
	name       string
	appFn      common.ReportFunc
	computedFn common.ReportFunc
	globalFn   common.ReportFunc
}

func wrapGlobal(fn func() string) common.ReportFunc {
	return func(string) string { return fn() }
}

func nginxReportProps() []nginxReportProp {
	return []nginxReportProp{
		{"access-log-format", AppAccessLogFormat, ComputedAccessLogFormat, wrapGlobal(GlobalAccessLogFormat)},
		{"access-log-path", AppAccessLogPath, ComputedAccessLogPath, wrapGlobal(GlobalAccessLogPath)},
		{"bind-address-ipv4", AppBindAddressIPv4, ComputedBindAddressIPv4, wrapGlobal(GlobalBindAddressIPv4)},
		{"bind-address-ipv6", AppBindAddressIPv6, ComputedBindAddressIPv6, wrapGlobal(GlobalBindAddressIPv6)},
		{"client-body-timeout", AppClientBodyTimeout, ComputedClientBodyTimeout, wrapGlobal(GlobalClientBodyTimeout)},
		{"client-header-timeout", AppClientHeaderTimeout, ComputedClientHeaderTimeout, wrapGlobal(GlobalClientHeaderTimeout)},
		{"client-max-body-size", AppClientMaxBodySize, ComputedClientMaxBodySize, wrapGlobal(GlobalClientMaxBodySize)},
		{"disable-custom-config", AppDisableCustomConfig, ComputedDisableCustomConfig, wrapGlobal(GlobalDisableCustomConfig)},
		{"error-log-path", AppErrorLogPath, ComputedErrorLogPath, wrapGlobal(GlobalErrorLogPath)},
		{"hsts-include-subdomains", AppHSTSIncludeSubdomains, ComputedHSTSIncludeSubdomains, wrapGlobal(GlobalHSTSIncludeSubdomains)},
		{"hsts-max-age", AppHSTSMaxAge, ComputedHSTSMaxAge, wrapGlobal(GlobalHSTSMaxAge)},
		{"hsts-preload", AppHSTSPreload, ComputedHSTSPreload, wrapGlobal(GlobalHSTSPreload)},
		{"hsts", AppHSTS, ComputedHSTS, wrapGlobal(GlobalHSTS)},
		{"keepalive-timeout", AppKeepaliveTimeout, ComputedKeepaliveTimeout, wrapGlobal(GlobalKeepaliveTimeout)},
		{"lingering-timeout", AppLingeringTimeout, ComputedLingeringTimeout, wrapGlobal(GlobalLingeringTimeout)},
		{"nginx-conf-sigil-path", AppNginxConfSigilPath, ComputedNginxConfSigilPath, wrapGlobal(GlobalNginxConfSigilPath)},
		{"nginx-service-command", AppNginxServiceCommand, ComputedNginxServiceCommand, wrapGlobal(GlobalNginxServiceCommand)},
		{"proxy-buffer-size", AppProxyBufferSize, ComputedProxyBufferSize, wrapGlobal(GlobalProxyBufferSize)},
		{"proxy-buffering", AppProxyBuffering, ComputedProxyBuffering, wrapGlobal(GlobalProxyBuffering)},
		{"proxy-buffers", AppProxyBuffers, ComputedProxyBuffers, wrapGlobal(GlobalProxyBuffers)},
		{"proxy-busy-buffers-size", AppProxyBusyBuffersSize, ComputedProxyBusyBuffersSize, wrapGlobal(GlobalProxyBusyBuffersSize)},
		{"proxy-connect-timeout", AppProxyConnectTimeout, ComputedProxyConnectTimeout, wrapGlobal(GlobalProxyConnectTimeout)},
		{"proxy-keepalive", AppProxyKeepalive, ComputedProxyKeepalive, wrapGlobal(GlobalProxyKeepalive)},
		{"proxy-read-timeout", AppProxyReadTimeout, ComputedProxyReadTimeout, wrapGlobal(GlobalProxyReadTimeout)},
		{"proxy-send-timeout", AppProxySendTimeout, ComputedProxySendTimeout, wrapGlobal(GlobalProxySendTimeout)},
		{"send-timeout", AppSendTimeout, ComputedSendTimeout, wrapGlobal(GlobalSendTimeout)},
		{"underscore-in-headers", AppUnderscoreInHeaders, ComputedUnderscoreInHeaders, wrapGlobal(GlobalUnderscoreInHeaders)},
		{"x-forwarded-for-value", AppXForwardedForValue, ComputedXForwardedForValue, wrapGlobal(GlobalXForwardedForValue)},
		{"x-forwarded-port-value", AppXForwardedPortValue, ComputedXForwardedPortValue, wrapGlobal(GlobalXForwardedPortValue)},
		{"x-forwarded-proto-value", AppXForwardedProtoValue, ComputedXForwardedProtoValue, wrapGlobal(GlobalXForwardedProtoValue)},
		{"x-forwarded-ssl", AppXForwardedSSL, ComputedXForwardedSSL, wrapGlobal(GlobalXForwardedSSL)},
	}
}

// ReportSingleApp is an internal function that displays the nginx report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}
	isGlobal := appName == "--global"

	flags := map[string]common.ReportFunc{}
	for _, p := range nginxReportProps() {
		flags["--nginx-"+p.name] = p.appFn
		flags["--nginx-computed-"+p.name] = p.computedFn
		flags["--nginx-global-"+p.name] = p.globalFn
	}
	flags["--nginx-last-visited-at"] = reportLastVisitedAt

	if isGlobal {
		for key := range flags {
			if !strings.Contains(key, "-global-") {
				delete(flags, key)
			}
		}
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp(common.ReportSingleAppInput{
		ReportType:              "nginx",
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

func reportLastVisitedAt(appName string) string {
	logPath := AppAccessLogPath(appName)
	if logPath == "off" || logPath == "/dev/null" {
		return ""
	}

	info, err := os.Stat(logPath)
	if err != nil || !info.Mode().IsRegular() {
		return ""
	}

	return strconv.FormatInt(info.ModTime().Unix(), 10)
}
