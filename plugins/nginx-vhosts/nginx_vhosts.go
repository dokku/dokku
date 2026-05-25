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

	value := GlobalAccessLogPath()
	if value == "" {
		value = fmt.Sprintf("%s/%s-access.log", getLogRoot(), appName)
	}
	return value
}

func GlobalAccessLogPath() string {
	return common.PropertyGet("nginx", "--global", "access-log-path")
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

	value := GlobalBindAddressIPv6()
	if value == "" {
		value = "::"
	}
	return value
}

func GlobalBindAddressIPv6() string {
	return common.PropertyGet("nginx", "--global", "bind-address-ipv6")
}

func AppClientBodyTimeout(appName string) string {
	return common.PropertyGet("nginx", appName, "client-body-timeout")
}

func ComputedClientBodyTimeout(appName string) string {
	appValue := AppClientBodyTimeout(appName)
	if appValue != "" {
		return appValue
	}

	value := GlobalClientBodyTimeout()
	if value == "" {
		value = "60s"
	}
	return value
}

func GlobalClientBodyTimeout() string {
	return common.PropertyGet("nginx", "--global", "client-body-timeout")
}

func AppClientHeaderTimeout(appName string) string {
	return common.PropertyGet("nginx", appName, "client-header-timeout")
}

func ComputedClientHeaderTimeout(appName string) string {
	appValue := AppClientHeaderTimeout(appName)
	if appValue != "" {
		return appValue
	}

	value := GlobalClientHeaderTimeout()
	if value == "" {
		value = "60s"
	}
	return value
}

func GlobalClientHeaderTimeout() string {
	return common.PropertyGet("nginx", "--global", "client-header-timeout")
}

func AppClientMaxBodySize(appName string) string {
	return common.PropertyGet("nginx", appName, "client-max-body-size")
}

func ComputedClientMaxBodySize(appName string) string {
	appValue := AppClientMaxBodySize(appName)
	if appValue != "" {
		return appValue
	}

	value := GlobalClientMaxBodySize()
	if value == "" {
		value = "1m"
	}
	return value
}

func GlobalClientMaxBodySize() string {
	return common.PropertyGet("nginx", "--global", "client-max-body-size")
}

func AppDisableCustomConfig(appName string) string {
	return common.PropertyGet("nginx", appName, "disable-custom-config")
}

func ComputedDisableCustomConfig(appName string) string {
	appValue := AppDisableCustomConfig(appName)
	if appValue != "" {
		return appValue
	}

	value := GlobalDisableCustomConfig()
	if value == "" {
		value = "false"
	}
	return value
}

func GlobalDisableCustomConfig() string {
	return common.PropertyGet("nginx", "--global", "disable-custom-config")
}

func AppErrorLogPath(appName string) string {
	return common.PropertyGet("nginx", appName, "error-log-path")
}

func ComputedErrorLogPath(appName string) string {
	appValue := AppErrorLogPath(appName)
	if appValue != "" {
		return appValue
	}

	value := GlobalErrorLogPath()
	if value == "" {
		value = fmt.Sprintf("%s/%s-error.log", getLogRoot(), appName)
	}
	return value
}

func GlobalErrorLogPath() string {
	return common.PropertyGet("nginx", "--global", "error-log-path")
}

func AppHSTSIncludeSubdomains(appName string) string {
	return common.PropertyGet("nginx", appName, "hsts-include-subdomains")
}

func ComputedHSTSIncludeSubdomains(appName string) string {
	appValue := AppHSTSIncludeSubdomains(appName)
	if appValue != "" {
		return appValue
	}

	value := GlobalHSTSIncludeSubdomains()
	if value == "" {
		value = "true"
	}
	return value
}

func GlobalHSTSIncludeSubdomains() string {
	return common.PropertyGet("nginx", "--global", "hsts-include-subdomains")
}

func AppHSTSMaxAge(appName string) string {
	return common.PropertyGet("nginx", appName, "hsts-max-age")
}

func ComputedHSTSMaxAge(appName string) string {
	appValue := AppHSTSMaxAge(appName)
	if appValue != "" {
		return appValue
	}

	value := GlobalHSTSMaxAge()
	if value == "" {
		value = "15724800"
	}
	return value
}

func GlobalHSTSMaxAge() string {
	return common.PropertyGet("nginx", "--global", "hsts-max-age")
}

func AppHSTSPreload(appName string) string {
	return common.PropertyGet("nginx", appName, "hsts-preload")
}

func ComputedHSTSPreload(appName string) string {
	appValue := AppHSTSPreload(appName)
	if appValue != "" {
		return appValue
	}

	value := GlobalHSTSPreload()
	if value == "" {
		value = "false"
	}
	return value
}

func GlobalHSTSPreload() string {
	return common.PropertyGet("nginx", "--global", "hsts-preload")
}

func AppHSTS(appName string) string {
	return common.PropertyGet("nginx", appName, "hsts")
}

func ComputedHSTS(appName string) string {
	appValue := AppHSTS(appName)
	if appValue != "" {
		return appValue
	}

	value := GlobalHSTS()
	if value == "" {
		value = "true"
	}
	return value
}

func GlobalHSTS() string {
	return common.PropertyGet("nginx", "--global", "hsts")
}

func AppNginxConfSigilPath(appName string) string {
	return common.PropertyGet("nginx", appName, "nginx-conf-sigil-path")
}

func ComputedNginxConfSigilPath(appName string) string {
	appValue := AppNginxConfSigilPath(appName)
	if appValue != "" {
		return appValue
	}

	value := GlobalNginxConfSigilPath()
	if value == "" {
		value = "nginx.conf.sigil"
	}
	return value
}

func GlobalNginxConfSigilPath() string {
	return common.PropertyGet("nginx", "--global", "nginx-conf-sigil-path")
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
	return common.PropertyGet("nginx", "--global", "nginx-service-command")
}

func AppKeepaliveTimeout(appName string) string {
	return common.PropertyGet("nginx", appName, "keepalive-timeout")
}

func ComputedKeepaliveTimeout(appName string) string {
	appValue := AppKeepaliveTimeout(appName)
	if appValue != "" {
		return appValue
	}

	value := GlobalKeepaliveTimeout()
	if value == "" {
		value = "75s"
	}
	return value
}

func GlobalKeepaliveTimeout() string {
	return common.PropertyGet("nginx", "--global", "keepalive-timeout")
}

func AppLingeringTimeout(appName string) string {
	return common.PropertyGet("nginx", appName, "lingering-timeout")
}

func ComputedLingeringTimeout(appName string) string {
	appValue := AppLingeringTimeout(appName)
	if appValue != "" {
		return appValue
	}

	value := GlobalLingeringTimeout()
	if value == "" {
		value = "5s"
	}
	return value
}

func GlobalLingeringTimeout() string {
	return common.PropertyGet("nginx", "--global", "lingering-timeout")
}

func AppProxyBufferSize(appName string) string {
	return common.PropertyGet("nginx", appName, "proxy-buffer-size")
}

func ComputedProxyBufferSize(appName string) string {
	appValue := AppProxyBufferSize(appName)
	if appValue != "" {
		return appValue
	}

	value := GlobalProxyBufferSize()
	if value == "" {
		value = fmt.Sprintf("%dk", os.Getpagesize()/1024)
	}
	return value
}

func GlobalProxyBufferSize() string {
	return common.PropertyGet("nginx", "--global", "proxy-buffer-size")
}

func AppProxyBuffering(appName string) string {
	return common.PropertyGet("nginx", appName, "proxy-buffering")
}

func ComputedProxyBuffering(appName string) string {
	appValue := AppProxyBuffering(appName)
	if appValue != "" {
		return appValue
	}

	value := GlobalProxyBuffering()
	if value == "" {
		value = "on"
	}
	return value
}

func GlobalProxyBuffering() string {
	return common.PropertyGet("nginx", "--global", "proxy-buffering")
}

func AppProxyBuffers(appName string) string {
	return common.PropertyGet("nginx", appName, "proxy-buffers")
}

func ComputedProxyBuffers(appName string) string {
	appValue := AppProxyBuffers(appName)
	if appValue != "" {
		return appValue
	}

	value := GlobalProxyBuffers()
	if value == "" {
		value = fmt.Sprintf("8 %dk", os.Getpagesize()/1024)
	}
	return value
}

func GlobalProxyBuffers() string {
	return common.PropertyGet("nginx", "--global", "proxy-buffers")
}

func AppProxyBusyBuffersSize(appName string) string {
	return common.PropertyGet("nginx", appName, "proxy-busy-buffers-size")
}

func ComputedProxyBusyBuffersSize(appName string) string {
	appValue := AppProxyBusyBuffersSize(appName)
	if appValue != "" {
		return appValue
	}

	value := GlobalProxyBusyBuffersSize()
	if value == "" {
		value = fmt.Sprintf("%dk", (os.Getpagesize()/1024)*2)
	}
	return value
}

func GlobalProxyBusyBuffersSize() string {
	return common.PropertyGet("nginx", "--global", "proxy-busy-buffers-size")
}

func AppProxyConnectTimeout(appName string) string {
	return common.PropertyGet("nginx", appName, "proxy-connect-timeout")
}

func ComputedProxyConnectTimeout(appName string) string {
	appValue := AppProxyConnectTimeout(appName)
	if appValue != "" {
		return appValue
	}

	value := GlobalProxyConnectTimeout()
	if value == "" {
		value = "60s"
	}
	return value
}

func GlobalProxyConnectTimeout() string {
	return common.PropertyGet("nginx", "--global", "proxy-connect-timeout")
}

func AppProxyReadTimeout(appName string) string {
	return common.PropertyGet("nginx", appName, "proxy-read-timeout")
}

func ComputedProxyReadTimeout(appName string) string {
	appValue := AppProxyReadTimeout(appName)
	if appValue != "" {
		return appValue
	}

	value := GlobalProxyReadTimeout()
	if value == "" {
		value = "60s"
	}
	return value
}

func GlobalProxyReadTimeout() string {
	return common.PropertyGet("nginx", "--global", "proxy-read-timeout")
}

func AppProxySendTimeout(appName string) string {
	return common.PropertyGet("nginx", appName, "proxy-send-timeout")
}

func ComputedProxySendTimeout(appName string) string {
	appValue := AppProxySendTimeout(appName)
	if appValue != "" {
		return appValue
	}

	value := GlobalProxySendTimeout()
	if value == "" {
		value = "60s"
	}
	return value
}

func GlobalProxySendTimeout() string {
	return common.PropertyGet("nginx", "--global", "proxy-send-timeout")
}

func AppSendTimeout(appName string) string {
	return common.PropertyGet("nginx", appName, "send-timeout")
}

func ComputedSendTimeout(appName string) string {
	appValue := AppSendTimeout(appName)
	if appValue != "" {
		return appValue
	}

	value := GlobalSendTimeout()
	if value == "" {
		value = "60s"
	}
	return value
}

func GlobalSendTimeout() string {
	return common.PropertyGet("nginx", "--global", "send-timeout")
}

func AppUnderscoreInHeaders(appName string) string {
	return common.PropertyGet("nginx", appName, "underscore-in-headers")
}

func ComputedUnderscoreInHeaders(appName string) string {
	appValue := AppUnderscoreInHeaders(appName)
	if appValue != "" {
		return appValue
	}

	value := GlobalUnderscoreInHeaders()
	if value == "" {
		value = "off"
	}
	return value
}

func GlobalUnderscoreInHeaders() string {
	return common.PropertyGet("nginx", "--global", "underscore-in-headers")
}

func AppXForwardedForValue(appName string) string {
	return common.PropertyGet("nginx", appName, "x-forwarded-for-value")
}

func ComputedXForwardedForValue(appName string) string {
	appValue := AppXForwardedForValue(appName)
	if appValue != "" {
		return appValue
	}

	value := GlobalXForwardedForValue()
	if value == "" {
		value = "$remote_addr"
	}
	return value
}

func GlobalXForwardedForValue() string {
	return common.PropertyGet("nginx", "--global", "x-forwarded-for-value")
}

func AppXForwardedPortValue(appName string) string {
	return common.PropertyGet("nginx", appName, "x-forwarded-port-value")
}

func ComputedXForwardedPortValue(appName string) string {
	appValue := AppXForwardedPortValue(appName)
	if appValue != "" {
		return appValue
	}

	value := GlobalXForwardedPortValue()
	if value == "" {
		value = "$server_port"
	}
	return value
}

func GlobalXForwardedPortValue() string {
	return common.PropertyGet("nginx", "--global", "x-forwarded-port-value")
}

func AppXForwardedProtoValue(appName string) string {
	return common.PropertyGet("nginx", appName, "x-forwarded-proto-value")
}

func ComputedXForwardedProtoValue(appName string) string {
	appValue := AppXForwardedProtoValue(appName)
	if appValue != "" {
		return appValue
	}

	value := GlobalXForwardedProtoValue()
	if value == "" {
		value = "$scheme"
	}
	return value
}

func GlobalXForwardedProtoValue() string {
	return common.PropertyGet("nginx", "--global", "x-forwarded-proto-value")
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
	return common.PropertyGet("nginx", "--global", "x-forwarded-ssl")
}

func AppProxyKeepalive(appName string) string {
	return common.PropertyGet("nginx", appName, "proxy-keepalive")
}

func ComputedProxyKeepalive(appName string) string {
	appValue := AppProxyKeepalive(appName)
	if appValue != "" {
		return appValue
	}

	return GlobalProxyKeepalive()
}

func GlobalProxyKeepalive() string {
	return common.PropertyGet("nginx", "--global", "proxy-keepalive")
}
