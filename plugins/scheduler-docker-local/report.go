package schedulerdockerlocal

import (
	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the scheduler-docker-local report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	var flags map[string]common.ReportFunc
	if appName == "--global" {
		flags = map[string]common.ReportFunc{
			"--scheduler-docker-local-computed-init-process":            reportComputedInitProcess,
			"--scheduler-docker-local-computed-parallel-schedule-count": reportComputedParallelScheduleCount,
			"--scheduler-docker-local-global-init-process":              reportGlobalInitProcess,
			"--scheduler-docker-local-global-parallel-schedule-count":   reportGlobalParallelScheduleCount,
		}
	} else {
		flags = map[string]common.ReportFunc{
			"--scheduler-docker-local-computed-init-process":            reportComputedInitProcess,
			"--scheduler-docker-local-computed-parallel-schedule-count": reportComputedParallelScheduleCount,
			"--scheduler-docker-local-global-init-process":              reportGlobalInitProcess,
			"--scheduler-docker-local-global-parallel-schedule-count":   reportGlobalParallelScheduleCount,
			"--scheduler-docker-local-init-process":                     reportInitProcess,
			"--scheduler-docker-local-parallel-schedule-count":          reportParallelScheduleCount,
		}
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp(common.ReportSingleAppInput{
		ReportType:              "scheduler-docker-local",
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

func reportInitProcess(appName string) string {
	return common.PropertyGet("scheduler-docker-local", appName, "init-process")
}

func reportGlobalInitProcess(appName string) string {
	return common.PropertyGet("scheduler-docker-local", "--global", "init-process")
}

func reportComputedInitProcess(appName string) string {
	value := reportInitProcess(appName)
	if value == "" {
		value = reportGlobalInitProcess(appName)
	}
	if value == "" {
		value = "true"
	}

	return value
}

func reportParallelScheduleCount(appName string) string {
	return common.PropertyGet("scheduler-docker-local", appName, "parallel-schedule-count")
}

func reportGlobalParallelScheduleCount(appName string) string {
	return common.PropertyGet("scheduler-docker-local", "--global", "parallel-schedule-count")
}

func reportComputedParallelScheduleCount(appName string) string {
	value := reportParallelScheduleCount(appName)
	if value == "" {
		value = reportGlobalParallelScheduleCount(appName)
	}
	if value == "" {
		value = "1"
	}

	return value
}
