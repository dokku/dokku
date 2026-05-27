package cron

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the cron report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	var flags map[string]common.ReportFunc
	if appName == "--global" {
		flags = map[string]common.ReportFunc{
			"--cron-computed-mailfrom":    reportComputedMailfrom,
			"--cron-computed-mailto":      reportComputedMailto,
			"--cron-computed-maintenance": reportComputedMaintenance,
			"--cron-global-mailfrom":      reportGlobalMailfrom,
			"--cron-global-mailto":        reportGlobalMailto,
			"--cron-global-maintenance":   reportGlobalMaintenance,
		}
	} else {
		flags = map[string]common.ReportFunc{
			"--cron-computed-mailfrom":    reportComputedMailfrom,
			"--cron-computed-mailto":      reportComputedMailto,
			"--cron-computed-maintenance": reportComputedMaintenance,
			"--cron-global-mailfrom":      reportGlobalMailfrom,
			"--cron-global-mailto":        reportGlobalMailto,
			"--cron-global-maintenance":   reportGlobalMaintenance,
			"--cron-maintenance":          reportMaintenance,
			"--cron-task-count":           reportTasks,
		}

		extraFlags := addCronMaintenanceFlags(appName, infoFlag)
		for flag, fn := range extraFlags {
			flags[flag] = fn
		}
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp(common.ReportSingleAppInput{
		ReportType:              "cron",
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

func addCronMaintenanceFlags(appName string, infoFlag string) map[string]common.ReportFunc {
	flags := map[string]common.ReportFunc{}

	properties, err := common.PropertyGetAllByPrefix("cron", appName, MaintenancePropertyPrefix)
	if err != nil {
		return flags
	}

	for property, value := range properties {
		key := strings.Replace(property, MaintenancePropertyPrefix, "", 1)
		flags[fmt.Sprintf("--cron-maintenance-%s", key)] = func(appName string) string {
			return value
		}
	}

	return flags
}

func reportGlobalMailfrom(_ string) string {
	return common.PropertyGet("cron", "--global", "mailfrom")
}

func reportGlobalMailto(_ string) string {
	return common.PropertyGet("cron", "--global", "mailto")
}

func reportComputedMailfrom(_ string) string {
	return common.PropertyGetDefault("cron", "--global", "mailfrom", DefaultProperties["mailfrom"])
}

func reportComputedMailto(_ string) string {
	return common.PropertyGetDefault("cron", "--global", "mailto", DefaultProperties["mailto"])
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
