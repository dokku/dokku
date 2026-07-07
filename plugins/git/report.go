package git

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the git report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	var flags map[string]common.ReportFunc
	if appName == "--global" {
		flags = map[string]common.ReportFunc{
			"--git-computed-archive-max-files": reportComputedArchiveMaxFiles,
			"--git-computed-archive-max-size":  reportComputedArchiveMaxSize,
			"--git-computed-deploy-branch":     reportComputedDeployBranch,
			"--git-computed-keep-git-dir":      reportComputedKeepGitDir,
			"--git-global-archive-max-files":   reportGlobalArchiveMaxFiles,
			"--git-global-archive-max-size":    reportGlobalArchiveMaxSize,
			"--git-global-deploy-branch":       reportGlobalDeployBranch,
			"--git-global-keep-git-dir":        reportGlobalKeepGitDir,
		}
	} else {
		flags = map[string]common.ReportFunc{
			"--git-computed-archive-max-files": reportComputedArchiveMaxFiles,
			"--git-computed-archive-max-size":  reportComputedArchiveMaxSize,
			"--git-computed-deploy-branch":     reportComputedDeployBranch,
			"--git-computed-keep-git-dir":      reportComputedKeepGitDir,
			"--git-deploy-branch":              reportDeployBranch,
			"--git-global-archive-max-files":   reportGlobalArchiveMaxFiles,
			"--git-global-archive-max-size":    reportGlobalArchiveMaxSize,
			"--git-global-deploy-branch":       reportGlobalDeployBranch,
			"--git-global-keep-git-dir":        reportGlobalKeepGitDir,
			"--git-keep-git-dir":               reportKeepGitDir,
			"--git-last-updated-at":            reportLastUpdatedAt,
			"--git-rev-env-var":                reportRevEnvVar,
			"--git-sha":                        reportSha,
			"--git-source-image":               reportSourceImage,
		}
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp(common.ReportSingleAppInput{
		ReportType:              "git",
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

func reportGlobalArchiveMaxFiles(appName string) string {
	return common.PropertyGet("git", "--global", "archive-max-files")
}

func reportComputedArchiveMaxFiles(appName string) string {
	return common.PropertyGetDefault("git", "--global", "archive-max-files", "10000")
}

func reportGlobalArchiveMaxSize(appName string) string {
	return common.PropertyGet("git", "--global", "archive-max-size")
}

func reportComputedArchiveMaxSize(appName string) string {
	return common.PropertyGetDefault("git", "--global", "archive-max-size", "1073741824")
}

func reportDeployBranch(appName string) string {
	return common.PropertyGet("git", appName, "deploy-branch")
}

func reportGlobalDeployBranch(appName string) string {
	return common.PropertyGet("git", "--global", "deploy-branch")
}

func reportComputedDeployBranch(appName string) string {
	if value := common.PropertyGet("git", appName, "deploy-branch"); value != "" {
		return value
	}
	if value := common.PropertyGet("git", "--global", "deploy-branch"); value != "" {
		return value
	}

	return "master"
}

func reportKeepGitDir(appName string) string {
	return common.PropertyGet("git", appName, "keep-git-dir")
}

func reportGlobalKeepGitDir(appName string) string {
	return common.PropertyGet("git", "--global", "keep-git-dir")
}

func reportComputedKeepGitDir(appName string) string {
	if value := common.PropertyGet("git", appName, "keep-git-dir"); value != "" {
		return value
	}
	if value := common.PropertyGet("git", "--global", "keep-git-dir"); value != "" {
		return value
	}

	return "false"
}

func reportRevEnvVar(appName string) string {
	return common.PropertyGetDefault("git", appName, "rev-env-var", "GIT_REV")
}

func reportSourceImage(appName string) string {
	return common.PropertyGet("git", appName, "source-image")
}

func reportSha(appName string) string {
	appRoot := filepath.Join(common.MustGetEnv("DOKKU_ROOT"), appName)
	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command:          "git",
		Args:             []string{"rev-parse", "HEAD"},
		WorkingDirectory: appRoot,
	})
	if err != nil {
		return ""
	}

	return result.StdoutContents()
}

func reportLastUpdatedAt(appName string) string {
	branch := reportComputedDeployBranch(appName)
	headFile := filepath.Join(common.MustGetEnv("DOKKU_ROOT"), appName, "refs", "heads", branch)
	info, err := os.Stat(headFile)
	if err != nil {
		return ""
	}

	return strconv.FormatInt(info.ModTime().Unix(), 10)
}
