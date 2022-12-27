package ports

import (
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the port report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	flags := map[string]common.ReportFunc{
		"--ports-map": reportPortMap,
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
	var portMap []string
	for _, pm := range getPortMaps(appName) {
		portMap = append(portMap, pm.String())
	}

	return strings.Join(portMap, " ")
}
