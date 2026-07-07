package haproxyvhosts

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the haproxy report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	flags := map[string]common.ReportFunc{
		"--haproxy-computed-image":              reportComputedImage,
		"--haproxy-computed-letsencrypt-email":  reportComputedLetsencryptEmail,
		"--haproxy-computed-letsencrypt-server": reportComputedLetsencryptServer,
		"--haproxy-computed-log-level":          reportComputedLogLevel,
		"--haproxy-computed-refresh-conf":       reportComputedRefreshConf,
		"--haproxy-global-image":                reportGlobalImage,
		"--haproxy-global-letsencrypt-email":    reportGlobalLetsencryptEmail,
		"--haproxy-global-letsencrypt-server":   reportGlobalLetsencryptServer,
		"--haproxy-global-log-level":            reportGlobalLogLevel,
		"--haproxy-global-refresh-conf":         reportGlobalRefreshConf,
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp(common.ReportSingleAppInput{
		ReportType:              "haproxy",
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
	return common.PropertyGet("haproxy", "--global", "image")
}

func reportComputedImage(appName string) string {
	value := reportGlobalImage(appName)
	if value == "" {
		value = dockerfileFromImage("haproxy-vhosts")
	}

	return value
}

func reportGlobalLetsencryptEmail(appName string) string {
	return common.PropertyGet("haproxy", "--global", "letsencrypt-email")
}

func reportComputedLetsencryptEmail(appName string) string {
	return reportGlobalLetsencryptEmail(appName)
}

func reportGlobalLetsencryptServer(appName string) string {
	return common.PropertyGet("haproxy", "--global", "letsencrypt-server")
}

func reportComputedLetsencryptServer(appName string) string {
	value := reportGlobalLetsencryptServer(appName)
	if value == "" {
		value = "https://acme-v02.api.letsencrypt.org/directory"
	}

	return value
}

func reportGlobalLogLevel(appName string) string {
	return common.PropertyGet("haproxy", "--global", "log-level")
}

func reportComputedLogLevel(appName string) string {
	value := reportGlobalLogLevel(appName)
	if value == "" {
		value = "ERROR"
	}

	return strings.ToUpper(value)
}

func reportGlobalRefreshConf(appName string) string {
	return common.PropertyGet("haproxy", "--global", "refresh-conf")
}

func reportComputedRefreshConf(appName string) string {
	value := reportGlobalRefreshConf(appName)
	if value == "" {
		value = "10"
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
