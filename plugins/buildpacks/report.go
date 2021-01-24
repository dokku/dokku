package buildpacks

import (
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the buildpacks report for one or more apps
func ReportSingleApp(appName, infoFlag string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	flags := map[string]common.ReportFunc{
		"--buildpacks-computed-stack": reportComputedStack,
		"--buildpacks-global-stack":   reportGlobalStack,
		"--buildpacks-list":           reportList,
		"--buildpacks-stack":          reportStack,
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	trimPrefix := false
	uppercaseFirstCharacter := true
	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp("buildpacks", appName, infoFlag, infoFlags, flagKeys, trimPrefix, uppercaseFirstCharacter)
}

func reportComputedStack(appName string) string {
	if stack := common.PropertyGetDefault("buildpacks", appName, "stack", ""); stack != "" {
		return stack
	}

	if stack := common.PropertyGetDefault("buildpacks", "--global", "stack", ""); stack != "" {
		return stack
	}

	b, _ := common.PlugnTriggerOutput("config-get", []string{appName, "DOKKU_IMAGE"}...)
	if dokkuImage := strings.TrimSpace(string(b[:])); dokkuImage != "" {
		common.LogWarn("Deprecated: use buildpacks:set-property instead of specifying DOKKU_IMAGE environment variable")
		return dokkuImage
	}

	return os.Getenv("DOKKU_IMAGE")
}

func reportGlobalStack(appName string) string {
	return common.PropertyGetDefault("buildpacks", "--global", "stack", "")
}

func reportList(appName string) string {
	buildpacks, err := common.PropertyListGet("buildpacks", appName, "buildpacks")
	if err != nil {
		return ""
	}

	return strings.Join(buildpacks, ",")
}

func reportStack(appName string) string {
	return common.PropertyGetDefault("buildpacks", appName, "stack", "")
}
