package main

import (
	"fmt"
	"os"

	nginx_vhosts "github.com/dokku/dokku/plugins/nginx-vhosts"
	flag "github.com/spf13/pflag"
)

func main() {

	args := flag.NewFlagSet("ps:inspect", flag.ExitOnError)
	appName := args.String("app", "", "app: the app to inspect")
	global := args.Bool("global", false, "global: inspect global property")
	computed := args.Bool("computed", false, "computed: inspect computed property")
	err := args.Parse(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}

	property := args.Arg(0)
	var value string
	if *computed {
		value = computedValue(*appName, property)
	} else if *global {
		value = globalValue(*appName, property)
	} else {
		value = appValue(*appName, property)
	}

	fmt.Print(value)
}

func appValue(appName string, property string) string {
	var value string
	switch property {
	case "access-log-format":
		value = nginx_vhosts.AppAccessLogFormat(appName)
	case "access-log-path":
		value = nginx_vhosts.AppAccessLogPath(appName)
	case "bind-address-ipv4":
		value = nginx_vhosts.AppBindAddressIPv4(appName)
	case "bind-address-ipv6":
		value = nginx_vhosts.AppBindAddressIPv6(appName)
	case "client-max-body-size":
		value = nginx_vhosts.AppClientMaxBodySize(appName)
	case "disable-custom-config":
		value = nginx_vhosts.AppDisableCustomConfig(appName)
	case "error-log-path":
		value = nginx_vhosts.AppErrorLogPath(appName)
	case "hsts-include-subdomains":
		value = nginx_vhosts.AppHSTSIncludeSubdomains(appName)
	case "hsts-max-age":
		value = nginx_vhosts.AppHSTSMaxAge(appName)
	case "hsts-preload":
		value = nginx_vhosts.AppHSTSPreload(appName)
	case "hsts":
		value = nginx_vhosts.AppHSTS(appName)
	case "nginx-conf-sigil-path":
		value = nginx_vhosts.AppNginxConfSigilPath(appName)
	case "proxy-buffer-size":
		value = nginx_vhosts.AppProxyBufferSize(appName)
	case "proxy-buffering":
		value = nginx_vhosts.AppProxyBuffering(appName)
	case "proxy-buffers":
		value = nginx_vhosts.AppProxyBuffers(appName)
	case "proxy-busy-buffers-size":
		value = nginx_vhosts.AppProxyBusyBuffersSize(appName)
	case "proxy-read-timeout":
		value = nginx_vhosts.AppProxyReadTimeout(appName)
	case "x-forwarded-for-value":
		value = nginx_vhosts.AppXForwardedForValue(appName)
	case "x-forwarded-port-value":
		value = nginx_vhosts.AppXForwardedPortValue(appName)
	case "x-forwarded-proto-value":
		value = nginx_vhosts.AppXForwardedProtoValue(appName)
	case "x-forwarded-ssl":
		value = nginx_vhosts.AppXForwardedSSL(appName)
	}

	return value
}

func computedValue(appName string, property string) string {
	var value string
	switch property {
	case "access-log-format":
		value = nginx_vhosts.ComputedAccessLogFormat(appName)
	case "access-log-path":
		value = nginx_vhosts.ComputedAccessLogPath(appName)
	case "bind-address-ipv4":
		value = nginx_vhosts.ComputedBindAddressIPv4(appName)
	case "bind-address-ipv6":
		value = nginx_vhosts.ComputedBindAddressIPv6(appName)
	case "client-max-body-size":
		value = nginx_vhosts.ComputedClientMaxBodySize(appName)
	case "disable-custom-config":
		value = nginx_vhosts.ComputedDisableCustomConfig(appName)
	case "error-log-path":
		value = nginx_vhosts.ComputedErrorLogPath(appName)
	case "hsts-include-subdomains":
		value = nginx_vhosts.ComputedHSTSIncludeSubdomains(appName)
	case "hsts-max-age":
		value = nginx_vhosts.ComputedHSTSMaxAge(appName)
	case "hsts-preload":
		value = nginx_vhosts.ComputedHSTSPreload(appName)
	case "hsts":
		value = nginx_vhosts.ComputedHSTS(appName)
	case "nginx-conf-sigil-path":
		value = nginx_vhosts.ComputedNginxConfSigilPath(appName)
	case "proxy-buffer-size":
		value = nginx_vhosts.ComputedProxyBufferSize(appName)
	case "proxy-buffering":
		value = nginx_vhosts.ComputedProxyBuffering(appName)
	case "proxy-buffers":
		value = nginx_vhosts.ComputedProxyBuffers(appName)
	case "proxy-busy-buffers-size":
		value = nginx_vhosts.ComputedProxyBusyBuffersSize(appName)
	case "proxy-read-timeout":
		value = nginx_vhosts.ComputedProxyReadTimeout(appName)
	case "x-forwarded-for-value":
		value = nginx_vhosts.ComputedXForwardedForValue(appName)
	case "x-forwarded-port-value":
		value = nginx_vhosts.ComputedXForwardedPortValue(appName)
	case "x-forwarded-proto-value":
		value = nginx_vhosts.ComputedXForwardedProtoValue(appName)
	case "x-forwarded-ssl":
		value = nginx_vhosts.ComputedXForwardedSSL(appName)
	}

	return value
}

func globalValue(appName string, property string) string {
	var value string
	switch property {
	case "access-log-format":
		value = nginx_vhosts.GlobalAccessLogFormat()
	case "access-log-path":
		value = nginx_vhosts.GlobalAccessLogPath(appName)
	case "bind-address-ipv4":
		value = nginx_vhosts.GlobalBindAddressIPv4()
	case "bind-address-ipv6":
		value = nginx_vhosts.GlobalBindAddressIPv6()
	case "client-max-body-size":
		value = nginx_vhosts.GlobalClientMaxBodySize()
	case "disable-custom-config":
		value = nginx_vhosts.GlobalDisableCustomConfig()
	case "error-log-path":
		value = nginx_vhosts.GlobalErrorLogPath(appName)
	case "hsts-include-subdomains":
		value = nginx_vhosts.GlobalHSTSIncludeSubdomains()
	case "hsts-max-age":
		value = nginx_vhosts.GlobalHSTSMaxAge()
	case "hsts-preload":
		value = nginx_vhosts.GlobalHSTSPreload()
	case "hsts":
		value = nginx_vhosts.GlobalHSTS()
	case "nginx-conf-sigil-path":
		value = nginx_vhosts.GlobalNginxConfSigilPath()
	case "proxy-buffer-size":
		value = nginx_vhosts.GlobalProxyBufferSize()
	case "proxy-buffering":
		value = nginx_vhosts.GlobalProxyBuffering()
	case "proxy-buffers":
		value = nginx_vhosts.GlobalProxyBuffers()
	case "proxy-busy-buffers-size":
		value = nginx_vhosts.GlobalProxyBusyBuffersSize()
	case "proxy-read-timeout":
		value = nginx_vhosts.GlobalProxyReadTimeout()
	case "x-forwarded-for-value":
		value = nginx_vhosts.GlobalXForwardedForValue()
	case "x-forwarded-port-value":
		value = nginx_vhosts.GlobalXForwardedPortValue()
	case "x-forwarded-proto-value":
		value = nginx_vhosts.GlobalXForwardedProtoValue()
	case "x-forwarded-ssl":
		value = nginx_vhosts.GlobalXForwardedSSL()
	}

	return value
}
