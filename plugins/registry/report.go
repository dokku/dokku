package registry

import (
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the registry report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	var flags map[string]common.ReportFunc
	if appName == "--global" {
		flags = map[string]common.ReportFunc{
			"--registry-computed-push-on-release":     reportComputedPushOnRelease,
			"--registry-global-push-on-release":       reportGlobalPushOnRelease,
			"--registry-computed-server":              reportComputedServer,
			"--registry-global-server":                reportGlobalServer,
			"--registry-computed-image-repo-template": reportComputedImageRepoTemplate,
			"--registry-global-image-repo-template":   reportGlobalImageRepoTemplate,
			"--registry-computed-push-extra-tags":     reportComputedPushExtraTags,
			"--registry-global-push-extra-tags":       reportGlobalPushExtraTags,
		}
	} else {
		flags = map[string]common.ReportFunc{
			"--registry-computed-image-repo":          reportComputedImageRepo,
			"--registry-image-repo":                   reportImageRepo,
			"--registry-computed-push-on-release":     reportComputedPushOnRelease,
			"--registry-global-push-on-release":       reportGlobalPushOnRelease,
			"--registry-push-on-release":              reportPushOnRelease,
			"--registry-computed-server":              reportComputedServer,
			"--registry-global-server":                reportGlobalServer,
			"--registry-computed-image-repo-template": reportComputedImageRepoTemplate,
			"--registry-global-image-repo-template":   reportGlobalImageRepoTemplate,
			"--registry-image-repo-template":          reportImageRepoTemplate,
			"--registry-server":                       reportServer,
			"--registry-tag-version":                  reportTagVersion,
			"--registry-push-extra-tags":              reportPushExtraTags,
			"--registry-global-push-extra-tags":       reportGlobalPushExtraTags,
			"--registry-computed-push-extra-tags":     reportComputedPushExtraTags,
		}
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

func reportComputedImageRepo(appName string) string {
	imageRepo := strings.TrimSpace(reportImageRepo(appName))
	if imageRepo == "" {
		imageRepo, _ = getImageRepoFromTemplate(appName)
	}

	if imageRepo == "" {
		imageRepo = common.GetAppImageRepo(appName)
	}

	return imageRepo
}

func reportImageRepo(appName string) string {
	return common.PropertyGet("registry", appName, "image-repo")
}

func reportComputedPushOnRelease(appName string) string {
	value := reportPushOnRelease(appName)
	value = strings.TrimSpace(value)
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
	server := getRegistryServerForApp(appName)
	return strings.TrimSpace(server)
}

func reportImageRepoTemplate(appName string) string {
	return common.PropertyGet("registry", appName, "image-repo-template")
}

func reportGlobalImageRepoTemplate(appName string) string {
	return common.PropertyGet("registry", "--global", "image-repo-template")
}

func reportComputedImageRepoTemplate(appName string) string {
	value := strings.TrimSpace(reportImageRepoTemplate(appName))
	if value == "" {
		value = reportGlobalImageRepoTemplate(appName)
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
	tagVersion := common.PropertyGet("registry", appName, "tag-version")
	return strings.TrimSpace(tagVersion)
}

func reportPushExtraTags(appName string) string {
	return common.PropertyGet("registry", appName, "push-extra-tags")
}

func reportGlobalPushExtraTags(appName string) string {
	return common.PropertyGet("registry", "--global", "push-extra-tags")
}

func reportComputedPushExtraTags(appName string) string {
	value := strings.TrimSpace(reportPushExtraTags(appName))
	if value == "" {
		value = reportGlobalPushExtraTags(appName)
	}
	if value == "" {
		value = DefaultProperties["push-extra-tags"]
	}
	return value
}
