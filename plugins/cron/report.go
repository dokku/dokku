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
		"--cron-mailfrom":             reportMailfrom,
		"--cron-mailto":               reportMailto,
		"--cron-task-count":           reportTasks,
		"--cron-global-maintenance":   reportGlobalMaintenance,
		"--cron-computed-maintenance": reportComputedMaintenance,
		"--cron-maintenance":          reportMaintenance,
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

func reportMailfrom(_ string) string {
	return common.PropertyGet("cron", "--global", "mailfrom")
}

func reportMailto(_ string) string {
	return common.PropertyGet("cron", "--global", "mailto")
}

func reportTasks(appName string) string {
	c, _ := FetchCronTasks(FetchCronTasksInput{AppName: appName})
	return strconv.Itoa(len(c))
}

func reportGlobalMaintenance(_ string) string {
	return common.PropertyGet("cron", "--global", "maintenance")
}

func reportComputedMaintenance(appName string) string {
	maintenance := common.PropertyGet("cron", appName, "maintenance")
	if maintenance == "true" {
		return "true"
	}

	return common.PropertyGetDefault("cron", "--global", "maintenance", DefaultProperties["maintenance"])
}

func reportMaintenance(appName string) string {
	return common.PropertyGet("cron", appName, "maintenance")
}
