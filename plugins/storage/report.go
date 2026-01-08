package storage

import "github.com/dokku/dokku/plugins/common"

func ReportSingleApp(appName string, infoFlag string, format string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	flags := map[string]common.ReportFunc{
		"--storage-build-mounts":  reportBuildMounts,
		"--storage-deploy-mounts": reportDeployMounts,
		"--storage-run-mounts":    reportRunMounts,
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	trimPrefix := false
	uppercaseFirstCharacter := true
	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp("storage", appName, infoFlag, infoFlags, flagKeys, format, trimPrefix, uppercaseFirstCharacter)
}

func reportBuildMounts(appName string) string {
	return GetBindMountsForDisplay(appName, "build")
}

func reportDeployMounts(appName string) string {
	return GetBindMountsForDisplay(appName, "deploy")
}

func reportRunMounts(appName string) string {
	return GetBindMountsForDisplay(appName, "run")
}
