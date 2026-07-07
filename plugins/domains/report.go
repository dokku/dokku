package domains

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

var hostnameRegex = regexp.MustCompile(`^[a-z0-9.*-]*[a-z0-9*-]$`)

// ReportSingleApp is an internal function that displays the domains report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	var flags map[string]common.ReportFunc
	if appName == "--global" {
		flags = map[string]common.ReportFunc{
			"--domains-global-enabled": reportGlobalEnabled,
			"--domains-global-vhosts":  reportGlobalVhosts,
		}
	} else {
		flags = map[string]common.ReportFunc{
			"--domains-app-enabled":    reportAppEnabled,
			"--domains-app-vhosts":     reportAppVhosts,
			"--domains-global-enabled": reportGlobalEnabled,
			"--domains-global-vhosts":  reportGlobalVhosts,
		}
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp(common.ReportSingleAppInput{
		ReportType:              "domains",
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

func reportAppEnabled(appName string) string {
	results, _ := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "config-get",
		Args:    []string{appName, "NO_VHOST"},
	})
	if results.StdoutContents() == "1" {
		return "false"
	}

	return "true"
}

func reportAppVhosts(appName string) string {
	dokkuRoot := common.MustGetEnv("DOKKU_ROOT")
	return normalizeWhitespace(readFileString(filepath.Join(dokkuRoot, appName, "VHOST")))
}

func reportGlobalEnabled(appName string) string {
	if isGlobalVhostEnabled() {
		return "true"
	}

	return "false"
}

func reportGlobalVhosts(appName string) string {
	if !isGlobalVhostEnabled() {
		return ""
	}

	return normalizeWhitespace(globalVhostContents())
}

func globalVhostContents() string {
	dokkuRoot := common.MustGetEnv("DOKKU_ROOT")
	return readFileString(filepath.Join(dokkuRoot, "VHOST"))
}

func isGlobalVhostEnabled() bool {
	for _, line := range strings.Split(globalVhostContents(), "\n") {
		if isValidHostname(line) {
			return true
		}
	}

	return false
}

func isValidHostname(hostname string) bool {
	hostname = strings.ToLower(hostname)
	if hostname == "_" {
		return true
	}

	return hostnameRegex.MatchString(hostname)
}

func readFileString(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	return string(data)
}

func normalizeWhitespace(value string) string {
	return strings.Join(strings.Fields(value), " ")
}
