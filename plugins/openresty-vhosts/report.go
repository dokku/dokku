package openrestyvhosts

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

const openrestyLogRoot = "/var/log/nginx"

// openrestyProp describes a settable openresty property along with the computed
// default used when neither an app-scoped nor a global value is present.
type openrestyProp struct {
	name            string
	hasApp          bool
	computedDefault func(appName string) string
}

func staticDefault(value string) func(string) string {
	return func(string) string { return value }
}

func openrestyPageSizeKB() int {
	return os.Getpagesize() / 1024
}

func openrestyProps() []openrestyProp {
	pageSize := openrestyPageSizeKB()
	return []openrestyProp{
		{"access-log-format", true, staticDefault("")},
		{"access-log-path", true, func(appName string) string {
			return fmt.Sprintf("%s/%s-access.log", openrestyLogRoot, appName)
		}},
		{"bind-address-ipv4", true, staticDefault("")},
		{"bind-address-ipv6", true, staticDefault("::")},
		{"client-body-timeout", true, staticDefault("60s")},
		{"client-header-timeout", true, staticDefault("60s")},
		{"client-max-body-size", true, staticDefault("1m")},
		{"error-log-path", true, func(appName string) string {
			return fmt.Sprintf("%s/%s-error.log", openrestyLogRoot, appName)
		}},
		{"hsts-include-subdomains", true, staticDefault("true")},
		{"hsts-max-age", true, staticDefault("15724800")},
		{"hsts-preload", true, staticDefault("false")},
		{"image", false, func(string) string { return dockerfileFromImage("openresty-vhosts") }},
		{"keepalive-timeout", true, staticDefault("75s")},
		{"letsencrypt-email", false, staticDefault("")},
		{"letsencrypt-server", false, staticDefault("https://acme-v02.api.letsencrypt.org/directory")},
		{"lingering-timeout", true, staticDefault("5s")},
		{"log-level", false, staticDefault("ERROR")},
		{"proxy-buffer-size", true, staticDefault(fmt.Sprintf("%dk", pageSize))},
		{"proxy-buffering", true, staticDefault("on")},
		{"proxy-buffers", true, staticDefault(fmt.Sprintf("8 %dk", pageSize))},
		{"proxy-busy-buffers-size", true, staticDefault(fmt.Sprintf("%dk", pageSize*2))},
		{"proxy-connect-timeout", true, staticDefault("60s")},
		{"proxy-read-timeout", true, staticDefault("60s")},
		{"proxy-send-timeout", true, staticDefault("60s")},
		{"send-timeout", true, staticDefault("60s")},
		{"underscore-in-headers", true, staticDefault("off")},
		{"x-forwarded-for-value", true, staticDefault("$remote_addr")},
		{"x-forwarded-port-value", true, staticDefault("$server_port")},
		{"x-forwarded-proto-value", true, staticDefault("$scheme")},
		{"x-forwarded-ssl", true, staticDefault("")},
	}
}

// ReportSingleApp is an internal function that displays the openresty report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}
	isGlobal := appName == "--global"

	flags := map[string]common.ReportFunc{}
	for _, p := range openrestyProps() {
		prop := p.name
		hasApp := p.hasApp
		computedDefault := p.computedDefault
		if hasApp && !isGlobal {
			flags["--openresty-"+prop] = func(app string) string {
				return common.PropertyGet("openresty", app, prop)
			}
		}
		flags["--openresty-global-"+prop] = func(app string) string {
			return common.PropertyGet("openresty", "--global", prop)
		}
		flags["--openresty-computed-"+prop] = func(app string) string {
			value := ""
			if hasApp && app != "--global" {
				value = common.PropertyGet("openresty", app, prop)
			}
			if value == "" {
				value = common.PropertyGet("openresty", "--global", prop)
			}
			if value == "" {
				value = computedDefault(app)
			}
			return value
		}
	}

	flags["--openresty-global-allowed-letsencrypt-domains-func-base64"] = func(app string) string {
		return common.PropertyGet("openresty", "--global", "allowed-letsencrypt-domains-func-base64")
	}
	flags["--openresty-computed-allowed-letsencrypt-domains-func-base64"] = func(app string) string {
		value := common.PropertyGetDefault("openresty", "--global", "allowed-letsencrypt-domains-func-base64", "return true")
		return base64.StdEncoding.EncodeToString([]byte(value))
	}

	if !isGlobal {
		flags["--openresty-hsts"] = func(app string) string {
			return common.PropertyGet("openresty", app, "hsts")
		}
	}
	flags["--openresty-global-hsts"] = func(app string) string {
		return common.PropertyGet("openresty", "--global", "hsts")
	}
	flags["--openresty-computed-hsts"] = openrestyHstsIsEnabled

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp(common.ReportSingleAppInput{
		ReportType:              "openresty",
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

func openrestyHstsIsEnabled(appName string) string {
	value := common.PropertyGet("openresty", appName, "hsts")
	if value == "" {
		value = common.PropertyGetDefault("openresty", "--global", "hsts", "true")
	}

	return value
}

// dockerfileFromImage returns the image referenced by the FROM line of the named
// plugin's Dockerfile, or an empty string when it cannot be read
func dockerfileFromImage(pluginName string) string {
	pluginPath := os.Getenv("PLUGIN_AVAILABLE_PATH")
	data, err := os.ReadFile(filepath.Join(pluginPath, pluginName, "Dockerfile"))
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, "FROM") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return fields[1]
			}
		}
	}

	return ""
}
