package builds

import (
	"strconv"
	"time"

	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp emits the report for one app (or --global).
func ReportSingleApp(appName, format, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	var flags map[string]common.ReportFunc
	if appName == "--global" {
		flags = map[string]common.ReportFunc{
			"--builds-global-retention": reportGlobalRetention,
		}
	} else {
		flags = map[string]common.ReportFunc{
			"--build-id":                  reportBuildID,
			"--build-kind":                reportBuildKind,
			"--build-status":              reportBuildStatus,
			"--build-pid":                 reportBuildPID,
			"--build-source":              reportBuildSource,
			"--build-started-at":          reportBuildStartedAt,
			"--build-finished-at":         reportBuildFinishedAt,
			"--build-exit-code":           reportBuildExitCode,
			"--builds-retention":          reportRetention,
			"--builds-global-retention":   reportGlobalRetention,
			"--builds-computed-retention": reportComputedRetention,
		}
	}

	flagKeys := make([]string, 0, len(flags))
	for k := range flags {
		flagKeys = append(flagKeys, k)
	}

	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp("builds", appName, infoFlag, infoFlags, flagKeys, format, false, true)
}

func mostRecentBuild(appName string) (Build, bool) {
	builds, err := FetchBuilds(appName)
	if err != nil || len(builds) == 0 {
		return Build{}, false
	}
	return builds[0], true
}

func reportBuildID(appName string) string {
	b, ok := mostRecentBuild(appName)
	if !ok {
		return ""
	}
	return b.ID
}

func reportBuildKind(appName string) string {
	b, ok := mostRecentBuild(appName)
	if !ok {
		return ""
	}
	return string(b.Kind)
}

func reportBuildStatus(appName string) string {
	b, ok := mostRecentBuild(appName)
	if !ok {
		return ""
	}
	return string(b.DisplayStatus())
}

func reportBuildPID(appName string) string {
	b, ok := mostRecentBuild(appName)
	if !ok {
		return ""
	}
	return strconv.Itoa(b.PID)
}

func reportBuildSource(appName string) string {
	b, ok := mostRecentBuild(appName)
	if !ok {
		return ""
	}
	return string(b.Source)
}

func reportBuildStartedAt(appName string) string {
	b, ok := mostRecentBuild(appName)
	if !ok {
		return ""
	}
	return b.StartedAt.Format(time.RFC3339)
}

func reportBuildFinishedAt(appName string) string {
	b, ok := mostRecentBuild(appName)
	if !ok || b.FinishedAt == nil {
		return ""
	}
	return b.FinishedAt.Format(time.RFC3339)
}

func reportBuildExitCode(appName string) string {
	b, ok := mostRecentBuild(appName)
	if !ok || b.ExitCode == nil {
		return ""
	}
	return strconv.Itoa(*b.ExitCode)
}

func reportRetention(appName string) string {
	return common.PropertyGet("builds", appName, "retention")
}

func reportGlobalRetention(_ string) string {
	return common.PropertyGet("builds", "--global", "retention")
}

func reportComputedRetention(appName string) string {
	return strconv.Itoa(ResolveRetention(appName))
}
