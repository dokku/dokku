package registry

import (
	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the registry report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	flags := map[string]common.ReportFunc{
		"--registry-computed-create-repository":      reportComputedCreateRepository,
		"--registry-global-create-repository":        reportGlobalCreateRepository,
		"--registry-create-repository":               reportCreateRepository,
		"--registry-computed-disable-delete-warning": reportComputedDisableDeleteWarning,
		"--registry-global-disable-delete-warning":   reportGlobalDisableDeleteWarning,
		"--registry-disable-delete-warning":          reportDisableDeleteWarning,
		"--registry-image-repo":                      reportImageRepo,
		"--registry-computed-push-on-release":        reportComputedPushOnRelease,
		"--registry-global-push-on-release":          reportGlobalPushOnRelease,
		"--registry-push-on-release":                 reportPushOnRelease,
		"--registry-computed-server":                 reportComputedServer,
		"--registry-global-server":                   reportGlobalServer,
		"--registry-server":                          reportServer,
		"--registry-tag-version":                     reportTagVersion,
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	trimPrefix := false
	uppercaseFirstCharacter := true
	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp("registry", appName, infoFlag, infoFlags, flagKeys, format, trimPrefix, uppercaseFirstCharacter)
}

func reportComputedCreateRepository(appName string) string {
	value := reportCreateRepository(appName)
	if value == "" {
		value = reportGlobalCreateRepository(appName)
	}

	if value == "" {
		value = DefaultProperties["create-repository"]
	}

	return value
}

func reportGlobalCreateRepository(appName string) string {
	return common.PropertyGet("registry", "--global", "create-repository")
}

func reportCreateRepository(appName string) string {
	return common.PropertyGet("registry", appName, "create-repository")
}

func reportComputedDisableDeleteWarning(appName string) string {
	value := reportDisableDeleteWarning(appName)
	if value == "" {
		value = reportGlobalDisableDeleteWarning(appName)
	}

	if value == "" {
		value = DefaultProperties["disable-delete-warning"]
	}

	return value
}

func reportGlobalDisableDeleteWarning(appName string) string {
	return common.PropertyGet("registry", "--global", "disable-delete-warning")
}

func reportDisableDeleteWarning(appName string) string {
	return common.PropertyGet("registry", appName, "disable-delete-warning")
}

func reportImageRepo(appName string) string {
	return common.PropertyGet("registry", appName, "image-repo")
}

func reportComputedPushOnRelease(appName string) string {
	value := reportPushOnRelease(appName)
	if value == "" {
		value = reportGlobalPushOnRelease(appName)
	}

	if value == "" {
		value = DefaultProperties["push-on-release"]
	}

	return value
}

func reportGlobalPushOnRelease(appName string) string {
	return common.PropertyGet("registry", "--global", "push-on-release")
}

func reportPushOnRelease(appName string) string {
	return common.PropertyGet("registry", appName, "push-on-release")
}

func reportComputedServer(appName string) string {
	value := reportServer(appName)
	if value == "" {
		value = reportGlobalServer(appName)
	}

	if value == "" {
		value = DefaultProperties["server"]
	}

	return value
}

func reportGlobalServer(appName string) string {
	return common.PropertyGet("registry", "--global", "server")
}

func reportServer(appName string) string {
	return common.PropertyGet("registry", appName, "server")
}

func reportTagVersion(appName string) string {
	return common.PropertyGet("registry", appName, "tag-version")
}
