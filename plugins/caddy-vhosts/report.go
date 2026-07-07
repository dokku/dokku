package caddyvhosts

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the caddy report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	flags := map[string]common.ReportFunc{
		"--caddy-computed-image":              reportComputedImage,
		"--caddy-computed-letsencrypt-email":  reportComputedLetsencryptEmail,
		"--caddy-computed-letsencrypt-server": reportComputedLetsencryptServer,
		"--caddy-computed-log-level":          reportComputedLogLevel,
		"--caddy-computed-polling-interval":   reportComputedPollingInterval,
		"--caddy-computed-tls-internal":       reportComputedTLSInternal,
		"--caddy-global-image":                reportGlobalImage,
		"--caddy-global-letsencrypt-email":    reportGlobalLetsencryptEmail,
		"--caddy-global-letsencrypt-server":   reportGlobalLetsencryptServer,
		"--caddy-global-log-level":            reportGlobalLogLevel,
		"--caddy-global-polling-interval":     reportGlobalPollingInterval,
		"--caddy-global-tls-internal":         reportGlobalTLSInternal,
	}
	if appName != "--global" {
		flags["--caddy-tls-internal"] = reportTLSInternal
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp(common.ReportSingleAppInput{
		ReportType:              "caddy",
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

func reportGlobalImage(appName string) string {
	return common.PropertyGet("caddy", "--global", "image")
}

func reportComputedImage(appName string) string {
	value := reportGlobalImage(appName)
	if value == "" {
		value = dockerfileFromImage("caddy-vhosts")
	}

	return value
}

func reportGlobalLetsencryptEmail(appName string) string {
	return common.PropertyGet("caddy", "--global", "letsencrypt-email")
}

func reportComputedLetsencryptEmail(appName string) string {
	return reportGlobalLetsencryptEmail(appName)
}

func reportGlobalLetsencryptServer(appName string) string {
	return common.PropertyGet("caddy", "--global", "letsencrypt-server")
}

func reportComputedLetsencryptServer(appName string) string {
	value := reportGlobalLetsencryptServer(appName)
	if value == "" {
		value = "https://acme-v02.api.letsencrypt.org/directory"
	}

	return value
}

func reportGlobalLogLevel(appName string) string {
	return common.PropertyGet("caddy", "--global", "log-level")
}

func reportComputedLogLevel(appName string) string {
	value := reportGlobalLogLevel(appName)
	if value == "" {
		value = "ERROR"
	}

	return strings.ToUpper(value)
}

func reportGlobalPollingInterval(appName string) string {
	return common.PropertyGet("caddy", "--global", "polling-interval")
}

func reportComputedPollingInterval(appName string) string {
	value := reportGlobalPollingInterval(appName)
	if value == "" {
		value = "5s"
	}

	return value
}

func reportTLSInternal(appName string) string {
	return common.PropertyGet("caddy", appName, "tls-internal")
}

func reportGlobalTLSInternal(appName string) string {
	return common.PropertyGet("caddy", "--global", "tls-internal")
}

func reportComputedTLSInternal(appName string) string {
	value := reportTLSInternal(appName)
	if value == "" {
		value = reportGlobalTLSInternal(appName)
	}
	if value == "" {
		value = "false"
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
