package traefikvhosts

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// valueMask is shown in place of credential values in the default stdout report
const valueMask = "*******"

// ReportSingleApp is an internal function that displays the traefik report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	flags := map[string]common.ReportFunc{
		"--traefik-computed-api-enabled":             reportComputedAPIEnabled,
		"--traefik-computed-api-entry-point":         reportComputedAPIEntryPoint,
		"--traefik-computed-api-entry-point-address": reportComputedAPIEntryPointAddress,
		"--traefik-computed-api-vhost":               reportComputedAPIVhost,
		"--traefik-computed-basic-auth-password":     reportComputedBasicAuthPassword,
		"--traefik-computed-basic-auth-username":     reportComputedBasicAuthUsername,
		"--traefik-computed-challenge-mode":          reportComputedChallengeMode,
		"--traefik-computed-dashboard-enabled":       reportComputedDashboardEnabled,
		"--traefik-computed-dns-provider":            reportComputedDNSProvider,
		"--traefik-computed-http-entry-point":        reportComputedHTTPEntryPoint,
		"--traefik-computed-https-entry-point":       reportComputedHTTPSEntryPoint,
		"--traefik-computed-image":                   reportComputedImage,
		"--traefik-computed-letsencrypt-email":       reportComputedLetsencryptEmail,
		"--traefik-computed-letsencrypt-server":      reportComputedLetsencryptServer,
		"--traefik-computed-log-level":               reportComputedLogLevel,
		"--traefik-global-api-enabled":               reportGlobalAPIEnabled,
		"--traefik-global-api-entry-point":           reportGlobalAPIEntryPoint,
		"--traefik-global-api-entry-point-address":   reportGlobalAPIEntryPointAddress,
		"--traefik-global-api-vhost":                 reportGlobalAPIVhost,
		"--traefik-global-basic-auth-password":       reportGlobalBasicAuthPassword,
		"--traefik-global-basic-auth-username":       reportGlobalBasicAuthUsername,
		"--traefik-global-challenge-mode":            reportGlobalChallengeMode,
		"--traefik-global-dashboard-enabled":         reportGlobalDashboardEnabled,
		"--traefik-global-dns-provider":              reportGlobalDNSProvider,
		"--traefik-global-http-entry-point":          reportGlobalHTTPEntryPoint,
		"--traefik-global-https-entry-point":         reportGlobalHTTPSEntryPoint,
		"--traefik-global-image":                     reportGlobalImage,
		"--traefik-global-letsencrypt-email":         reportGlobalLetsencryptEmail,
		"--traefik-global-letsencrypt-server":        reportGlobalLetsencryptServer,
		"--traefik-global-log-level":                 reportGlobalLogLevel,
	}

	for _, flagName := range []string{"--traefik-computed-basic-auth-password", "--traefik-global-basic-auth-password"} {
		flags[flagName] = maskedReportFunc(flags[flagName], flagName, format, infoFlag)
	}

	dnsProviderFlags, err := dnsProviderReportFlags(format, infoFlag)
	if err != nil {
		return err
	}
	for flagName, reportFunc := range dnsProviderFlags {
		flags[flagName] = reportFunc
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp(common.ReportSingleAppInput{
		ReportType:              "traefik",
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

// maskedReportFunc hides a credential value behind valueMask so the default stdout
// report - and thus the aggregate `dokku report` - never prints it. The raw value is
// returned for machine-readable output or when the flag is requested by name.
func maskedReportFunc(fn common.ReportFunc, flagName string, format string, infoFlag string) common.ReportFunc {
	if format == "json" || infoFlag == flagName {
		return fn
	}

	return func(appName string) string {
		if fn(appName) == "" {
			return ""
		}

		return valueMask
	}
}

// dnsProviderReportFlags returns report functions for the dynamic dns-provider-*
// properties, keyed by their global report flag
func dnsProviderReportFlags(format string, infoFlag string) (map[string]common.ReportFunc, error) {
	flags := map[string]common.ReportFunc{}
	properties, err := common.PropertyGetAllByPrefix("traefik", "--global", "dns-provider-")
	if err != nil {
		return flags, err
	}

	for property, value := range properties {
		flagName := "--traefik-global-" + property
		flags[flagName] = maskedReportFunc(func(string) string { return value }, flagName, format, infoFlag)
	}

	return flags, nil
}

func reportGlobalAPIEnabled(appName string) string {
	return common.PropertyGet("traefik", "--global", "api-enabled")
}

func reportComputedAPIEnabled(appName string) string {
	value := reportGlobalAPIEnabled(appName)
	if value == "" {
		value = "false"
	}

	return value
}

func reportGlobalAPIEntryPoint(appName string) string {
	return common.PropertyGet("traefik", "--global", "api-entry-point")
}

func reportComputedAPIEntryPoint(appName string) string {
	return reportGlobalAPIEntryPoint(appName)
}

func reportGlobalAPIEntryPointAddress(appName string) string {
	return common.PropertyGet("traefik", "--global", "api-entry-point-address")
}

func reportComputedAPIEntryPointAddress(appName string) string {
	return reportGlobalAPIEntryPointAddress(appName)
}

func reportGlobalAPIVhost(appName string) string {
	return common.PropertyGet("traefik", "--global", "api-vhost")
}

func reportComputedAPIVhost(appName string) string {
	value := reportGlobalAPIVhost(appName)
	if value == "" {
		value = "traefik.dokku.me"
	}

	return value
}

func reportGlobalBasicAuthPassword(appName string) string {
	return common.PropertyGet("traefik", "--global", "basic-auth-password")
}

func reportComputedBasicAuthPassword(appName string) string {
	return reportGlobalBasicAuthPassword(appName)
}

func reportGlobalBasicAuthUsername(appName string) string {
	return common.PropertyGet("traefik", "--global", "basic-auth-username")
}

func reportComputedBasicAuthUsername(appName string) string {
	return reportGlobalBasicAuthUsername(appName)
}

func reportGlobalChallengeMode(appName string) string {
	return common.PropertyGet("traefik", "--global", "challenge-mode")
}

func reportComputedChallengeMode(appName string) string {
	value := reportGlobalChallengeMode(appName)
	if value == "" {
		value = "tls"
	}

	return value
}

func reportGlobalDashboardEnabled(appName string) string {
	return common.PropertyGet("traefik", "--global", "dashboard-enabled")
}

func reportComputedDashboardEnabled(appName string) string {
	value := reportGlobalDashboardEnabled(appName)
	if value == "" {
		value = "false"
	}

	return value
}

func reportGlobalDNSProvider(appName string) string {
	return common.PropertyGet("traefik", "--global", "dns-provider")
}

func reportComputedDNSProvider(appName string) string {
	return reportGlobalDNSProvider(appName)
}

func reportGlobalHTTPEntryPoint(appName string) string {
	return common.PropertyGet("traefik", "--global", "http-entry-point")
}

func reportComputedHTTPEntryPoint(appName string) string {
	value := reportGlobalHTTPEntryPoint(appName)
	if value == "" {
		value = "http"
	}

	return value
}

func reportGlobalHTTPSEntryPoint(appName string) string {
	return common.PropertyGet("traefik", "--global", "https-entry-point")
}

func reportComputedHTTPSEntryPoint(appName string) string {
	value := reportGlobalHTTPSEntryPoint(appName)
	if value == "" {
		value = "https"
	}

	return value
}

func reportGlobalImage(appName string) string {
	return common.PropertyGet("traefik", "--global", "image")
}

func reportComputedImage(appName string) string {
	value := reportGlobalImage(appName)
	if value == "" {
		value = dockerfileFromImage("traefik-vhosts")
	}

	return value
}

func reportGlobalLetsencryptEmail(appName string) string {
	return common.PropertyGet("traefik", "--global", "letsencrypt-email")
}

func reportComputedLetsencryptEmail(appName string) string {
	return reportGlobalLetsencryptEmail(appName)
}

func reportGlobalLetsencryptServer(appName string) string {
	return common.PropertyGet("traefik", "--global", "letsencrypt-server")
}

func reportComputedLetsencryptServer(appName string) string {
	value := reportGlobalLetsencryptServer(appName)
	if value == "" {
		value = "https://acme-v02.api.letsencrypt.org/directory"
	}

	return value
}

func reportGlobalLogLevel(appName string) string {
	return common.PropertyGet("traefik", "--global", "log-level")
}

func reportComputedLogLevel(appName string) string {
	value := reportGlobalLogLevel(appName)
	if value == "" {
		value = "ERROR"
	}

	return strings.ToUpper(value)
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
