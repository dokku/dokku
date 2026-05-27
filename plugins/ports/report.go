package ports

import (
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the ports report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	var flags map[string]common.ReportFunc
	if appName == "--global" {
		flags = map[string]common.ReportFunc{}
	} else {
		flags = map[string]common.ReportFunc{
			"--ports-map":          reportPortMap,
			"--ports-map-detected": reportPortMapDetected,
		}
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp(common.ReportSingleAppInput{
		ReportType:              "ports",
		AppName:                 appName,
		InfoFlag:                infoFlag,
		InfoFlags:               infoFlags,
		InfoFlagKeys:            flagKeys,
		Format:                  format,
		TrimPrefix:              true,
		UppercaseFirstCharacter: true,
		EmitLegacyPrefix:        true,
	})
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
