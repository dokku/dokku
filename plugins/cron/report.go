package cron

import (
	"strconv"

	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the cron report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	flags := map[string]common.ReportFunc{
		"--cron-task-count": reportTasks,
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	trimPrefix := false
	uppercaseFirstCharacter := true
	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp("cron", appName, infoFlag, infoFlags, flagKeys, format, trimPrefix, uppercaseFirstCharacter)
}

func reportTasks(appName string) string {
	c, _ := fetchCronEntries(appName)
	return strconv.Itoa(len(c))
}
