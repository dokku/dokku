package ports

import (
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the ports report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	flags := map[string]common.ReportFunc{
		"--ports-map":          reportPortMap,
		"--ports-map-detected": reportPortMapDetected,
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	trimPrefix := false
	uppercaseFirstCharacter := true
	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp("ports", appName, infoFlag, infoFlags, flagKeys, format, trimPrefix, uppercaseFirstCharacter)
}

func reportPortMap(appName string) string {
	var portMaps []string
	for _, portMap := range getPortMaps(appName) {
		portMaps = append(portMaps, portMap.String())
	}

	return strings.Join(portMaps, " ")
}

func reportPortMapDetected(appName string) string {
	var portMaps []string
	for _, portMap := range getDetectedPortMaps(appName) {
		portMaps = append(portMaps, portMap.String())
	}

	return strings.Join(portMaps, " ")
}
