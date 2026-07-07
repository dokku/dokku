package traefikvhosts

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

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

	// dns-provider-* env vars are dynamic; their values are masked in the default
	// stdout report, but shown for --format json or when queried explicitly by name
	dnsProviderVars, err := common.PropertyGetAllByPrefix("traefik", "--global", "dns-provider-")
	if err != nil {
		return err
	}
	for key, value := range dnsProviderVars {
		flagName := "--traefik-" + key
		realValue := value
		if format == "json" || infoFlag == flagName {
			flags[flagName] = func(string) string { return realValue }
		} else {
			flags[flagName] = func(string) string { return "*******" }
		}
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
