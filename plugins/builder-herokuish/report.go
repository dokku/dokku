package builderherokuish

import (
	"os"

	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the builder-herokuish report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	var flags map[string]common.ReportFunc
	if appName == "--global" {
		flags = map[string]common.ReportFunc{
			"--builder-herokuish-computed-allowed": reportComputedAllowed,
			"--builder-herokuish-global-allowed":   reportGlobalAllowed,
			"--builder-herokuish-computed-stack":   reportComputedStack,
			"--builder-herokuish-global-stack":     reportGlobalStack,
		}
	} else {
		flags = map[string]common.ReportFunc{
			"--builder-herokuish-computed-allowed": reportComputedAllowed,
			"--builder-herokuish-global-allowed":   reportGlobalAllowed,
			"--builder-herokuish-allowed":          reportAllowed,
			"--builder-herokuish-computed-stack":   reportComputedStack,
			"--builder-herokuish-global-stack":     reportGlobalStack,
			"--builder-herokuish-stack":            reportStack,
		}
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp(common.ReportSingleAppInput{
		ReportType:              "builder-herokuish",
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

func reportAllowed(appName string) string {
	return common.PropertyGet("builder-herokuish", appName, "allowed")
}

func reportGlobalAllowed(appName string) string {
	return common.PropertyGet("builder-herokuish", "--global", "allowed")
}

func reportComputedAllowed(appName string) string {
	allowed := reportAllowed(appName)
	if allowed == "" {
		allowed = reportGlobalAllowed(appName)
	}
	if allowed == "" {
		allowed = "true"
		if hostArchitecture() != "amd64" {
			allowed = "false"
		}
	}

	return allowed
}

func reportStack(appName string) string {
	return common.PropertyGet("builder-herokuish", appName, "stack")
}

func reportGlobalStack(appName string) string {
	return common.PropertyGet("builder-herokuish", "--global", "stack")
}

func reportComputedStack(appName string) string {
	stack := reportStack(appName)
	if stack == "" {
		stack = reportGlobalStack(appName)
	}
	if stack == "" {
		stack = os.Getenv("DOKKU_IMAGE")
	}
	if stack == "" {
		stack = "gliderlabs/herokuish:latest-24"
	}

	return stack
}

// hostArchitecture returns the dpkg architecture of the host, or an empty string
// when it cannot be determined (matching the bash `dpkg --print-architecture` check)
func hostArchitecture() string {
	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command: "dpkg",
		Args:    []string{"--print-architecture"},
	})
	if err != nil {
		return ""
	}

	return result.StdoutContents()
}
