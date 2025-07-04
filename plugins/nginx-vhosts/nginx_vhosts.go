package nginxvhosts

import (
	"fmt"
	"os"

	"github.com/dokku/dokku/plugins/common"
)

func AppAccessLogFormat(appName string) string {
	return common.PropertyGet("nginx", appName, "access-log-format")
}

func ComputedAccessLogFormat(appName string) string {
	appValue := AppAccessLogFormat(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalAccessLogFormat()
}

func GlobalAccessLogFormat() string {
	return common.PropertyGet("nginx", "--global", "access-log-format")
}

func AppAccessLogPath(appName string) string {
	return common.PropertyGet("nginx", appName, "access-log-path")
}

func ComputedAccessLogPath(appName string) string {
	appValue := AppAccessLogPath(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalAccessLogPath(appName)
}

func GlobalAccessLogPath(appName string) string {
	defaultLogPath := fmt.Sprintf("%s/%s-access.log", getLogRoot(), appName)
	return common.PropertyGetDefault("nginx", "--global", "access-log-path", defaultLogPath)
}

func AppBindAddressIPv4(appName string) string {
	return common.PropertyGet("nginx", appName, "bind-address-ipv4")
}

func ComputedBindAddressIPv4(appName string) string {
	appValue := AppBindAddressIPv4(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalBindAddressIPv4()
}

func GlobalBindAddressIPv4() string {
	return common.PropertyGet("nginx", "--global", "bind-address-ipv4")
}

func AppBindAddressIPv6(appName string) string {
	return common.PropertyGet("nginx", appName, "bind-address-ipv6")
}

func ComputedBindAddressIPv6(appName string) string {
	appValue := AppBindAddressIPv6(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalBindAddressIPv6()
}

func GlobalBindAddressIPv6() string {
	return common.PropertyGetDefault("nginx", "--global", "bind-address-ipv6", "::")
}

func AppClientBodyTimeout(appName string) string {
	return common.PropertyGet("nginx", appName, "client-body-timeout")
}

func ComputedClientBodyTimeout(appName string) string {
	appValue := AppClientBodyTimeout(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalClientBodyTimeout()
}

func GlobalClientBodyTimeout() string {
	return common.PropertyGetDefault("nginx", "--global", "client-body-timeout", "60s")
}

func AppClientHeaderTimeout(appName string) string {
	return common.PropertyGet("nginx", appName, "client-header-timeout")
}

func ComputedClientHeaderTimeout(appName string) string {
	appValue := AppClientHeaderTimeout(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalClientHeaderTimeout()
}

func GlobalClientHeaderTimeout() string {
	return common.PropertyGetDefault("nginx", "--global", "client-header-timeout", "60s")
}

func AppClientMaxBodySize(appName string) string {
	return common.PropertyGet("nginx", appName, "client-max-body-size")
}

func ComputedClientMaxBodySize(appName string) string {
	appValue := AppClientMaxBodySize(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalClientMaxBodySize()
}

func GlobalClientMaxBodySize() string {
	return common.PropertyGetDefault("nginx", "--global", "client-max-body-size", "1m")
}

func AppDisableCustomConfig(appName string) string {
	return common.PropertyGet("nginx", appName, "disable-custom-config")
}

func ComputedDisableCustomConfig(appName string) string {
	appValue := AppDisableCustomConfig(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalDisableCustomConfig()
}

func GlobalDisableCustomConfig() string {
	return common.PropertyGetDefault("nginx", "--global", "disable-custom-config", "false")
}

func AppErrorLogPath(appName string) string {
	return common.PropertyGet("nginx", appName, "error-log-path")
}

func ComputedErrorLogPath(appName string) string {
	appValue := AppErrorLogPath(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalErrorLogPath(appName)
}

func GlobalErrorLogPath(appName string) string {
	defaultLogPath := fmt.Sprintf("%s/%s-error.log", getLogRoot(), appName)
	return common.PropertyGetDefault("nginx", "--global", "error-log-path", defaultLogPath)
}

func AppHSTSIncludeSubdomains(appName string) string {
	return common.PropertyGet("nginx", appName, "hsts-include-subdomains")
}

func ComputedHSTSIncludeSubdomains(appName string) string {
	appValue := AppHSTSIncludeSubdomains(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalHSTSIncludeSubdomains()
}

func GlobalHSTSIncludeSubdomains() string {
	return common.PropertyGetDefault("nginx", "--global", "hsts-include-subdomains", "true")
}

func AppHSTSMaxAge(appName string) string {
	return common.PropertyGet("nginx", appName, "hsts-max-age")
}

func ComputedHSTSMaxAge(appName string) string {
	appValue := AppHSTSMaxAge(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalHSTSMaxAge()
}

func GlobalHSTSMaxAge() string {
	return common.PropertyGetDefault("nginx", "--global", "hsts-max-age", "15724800")
}

func AppHSTSPreload(appName string) string {
	return common.PropertyGet("nginx", appName, "hsts-preload")
}

func ComputedHSTSPreload(appName string) string {
	appValue := AppHSTSPreload(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalHSTSPreload()
}

func GlobalHSTSPreload() string {
	return common.PropertyGetDefault("nginx", "--global", "hsts-preload", "false")
}

func AppHSTS(appName string) string {
	return common.PropertyGet("nginx", appName, "hsts")
}

func ComputedHSTS(appName string) string {
	appValue := AppHSTS(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalHSTS()
}

func GlobalHSTS() string {
	return common.PropertyGetDefault("nginx", "--global", "hsts", "true")
}

func AppNginxConfSigilPath(appName string) string {
	return common.PropertyGet("nginx", appName, "nginx-conf-sigil-path")
}

func ComputedNginxConfSigilPath(appName string) string {
	appValue := AppNginxConfSigilPath(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalNginxConfSigilPath()
}

func GlobalNginxConfSigilPath() string {
	return common.PropertyGetDefault("nginx", "--global", "nginx-conf-sigil-path", "nginx.conf.sigil")
}

func AppNginxServiceCommand(appName string) string {
	return common.PropertyGet("nginx", appName, "nginx-service-command")
}

func ComputedNginxServiceCommand(appName string) string {
	appValue := AppNginxServiceCommand(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalNginxServiceCommand()
}

func GlobalNginxServiceCommand() string {
	return common.PropertyGetDefault("nginx", "--global", "nginx-service-command", "")
}

func AppKeepaliveTimeout(appName string) string {
	return common.PropertyGet("nginx", appName, "keepalive-timeout")
}

func ComputedKeepaliveTimeout(appName string) string {
	appValue := AppKeepaliveTimeout(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalKeepaliveTimeout()
}

func GlobalKeepaliveTimeout() string {
	return common.PropertyGetDefault("nginx", "--global", "keepalive-timeout", "75s")
}

func AppLingeringTimeout(appName string) string {
	return common.PropertyGet("nginx", appName, "lingering-timeout")
}

func ComputedLingeringTimeout(appName string) string {
	appValue := AppLingeringTimeout(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalLingeringTimeout()
}

func GlobalLingeringTimeout() string {
	return common.PropertyGetDefault("nginx", "--global", "lingering-timeout", "5s")
}

func AppProxyBufferSize(appName string) string {
	return common.PropertyGet("nginx", appName, "proxy-buffer-size")
}

func ComputedProxyBufferSize(appName string) string {
	appValue := AppProxyBufferSize(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalProxyBufferSize()
}

func GlobalProxyBufferSize() string {
	return common.PropertyGetDefault("nginx", "--global", "proxy-buffer-size", fmt.Sprintf("%dk", os.Getpagesize()/1024))
}

func AppProxyBuffering(appName string) string {
	return common.PropertyGet("nginx", appName, "proxy-buffering")
}

func ComputedProxyBuffering(appName string) string {
	appValue := AppProxyBuffering(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalProxyBuffering()
}

func GlobalProxyBuffering() string {
	return common.PropertyGetDefault("nginx", "--global", "proxy-buffering", "on")
}

func AppProxyBuffers(appName string) string {
	return common.PropertyGet("nginx", appName, "proxy-buffers")
}

func ComputedProxyBuffers(appName string) string {
	appValue := AppProxyBuffers(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalProxyBuffers()
}

func GlobalProxyBuffers() string {
	return common.PropertyGetDefault("nginx", "--global", "proxy-buffers", fmt.Sprintf("8 %dk", os.Getpagesize()/1024))
}

func AppProxyBusyBuffersSize(appName string) string {
	return common.PropertyGet("nginx", appName, "proxy-busy-buffers-size")
}

func ComputedProxyBusyBuffersSize(appName string) string {
	appValue := AppProxyBusyBuffersSize(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalProxyBusyBuffersSize()
}

func GlobalProxyBusyBuffersSize() string {
	return common.PropertyGetDefault("nginx", "--global", "proxy-busy-buffers-size", fmt.Sprintf("%dk", (os.Getpagesize()/1024)*2))
}

func AppProxyConnectTimeout(appName string) string {
	return common.PropertyGet("nginx", appName, "proxy-connect-timeout")
}

func ComputedProxyConnectTimeout(appName string) string {
	appValue := AppProxyConnectTimeout(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalProxyConnectTimeout()
}

func GlobalProxyConnectTimeout() string {
	return common.PropertyGetDefault("nginx", "--global", "proxy-connect-timeout", "60s")
}

func AppProxyReadTimeout(appName string) string {
	return common.PropertyGet("nginx", appName, "proxy-read-timeout")
}

func ComputedProxyReadTimeout(appName string) string {
	appValue := AppProxyReadTimeout(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalProxyReadTimeout()
}

func GlobalProxyReadTimeout() string {
	return common.PropertyGetDefault("nginx", "--global", "proxy-read-timeout", "60s")
}

func AppProxySendTimeout(appName string) string {
	return common.PropertyGet("nginx", appName, "proxy-send-timeout")
}

func ComputedProxySendTimeout(appName string) string {
	appValue := AppProxySendTimeout(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalProxySendTimeout()
}

func GlobalProxySendTimeout() string {
	return common.PropertyGetDefault("nginx", "--global", "proxy-send-timeout", "60s")
}

func AppSendTimeout(appName string) string {
	return common.PropertyGet("nginx", appName, "send-timeout")
}

func ComputedSendTimeout(appName string) string {
	appValue := AppSendTimeout(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalSendTimeout()
}

func GlobalSendTimeout() string {
	return common.PropertyGetDefault("nginx", "--global", "send-timeout", "60s")
}

func AppUnderscoreInHeaders(appName string) string {
	return common.PropertyGet("nginx", appName, "underscore-in-headers")
}

func ComputedUnderscoreInHeaders(appName string) string {
	appValue := AppUnderscoreInHeaders(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalUnderscoreInHeaders()
}

func GlobalUnderscoreInHeaders() string {
	return common.PropertyGetDefault("nginx", "--global", "underscore-in-headers", "off")
}

func AppXForwardedForValue(appName string) string {
	return common.PropertyGet("nginx", appName, "x-forwarded-for-value")
}

func ComputedXForwardedForValue(appName string) string {
	appValue := AppXForwardedForValue(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalXForwardedForValue()
}

func GlobalXForwardedForValue() string {
	return common.PropertyGetDefault("nginx", "--global", "x-forwarded-for-value", "$remote_addr")
}

func AppXForwardedPortValue(appName string) string {
	return common.PropertyGet("nginx", appName, "x-forwarded-port-value")
}

func ComputedXForwardedPortValue(appName string) string {
	appValue := AppXForwardedPortValue(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalXForwardedPortValue()
}

func GlobalXForwardedPortValue() string {
	return common.PropertyGetDefault("nginx", "--global", "x-forwarded-port-value", "$server_port")
}

func AppXForwardedProtoValue(appName string) string {
	return common.PropertyGet("nginx", appName, "x-forwarded-proto-value")
}

func ComputedXForwardedProtoValue(appName string) string {
	appValue := AppXForwardedProtoValue(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalXForwardedProtoValue()
}

func GlobalXForwardedProtoValue() string {
	return common.PropertyGetDefault("nginx", "--global", "x-forwarded-proto-value", "$scheme")
}

func AppXForwardedSSL(appName string) string {
	return common.PropertyGet("nginx", appName, "x-forwarded-ssl")
}

func ComputedXForwardedSSL(appName string) string {
	appValue := AppXForwardedSSL(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalXForwardedSSL()
}

func GlobalXForwardedSSL() string {
	return common.PropertyGetDefault("nginx", "--global", "x-forwarded-ssl", "")
}
