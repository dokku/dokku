package builds

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/dokku/dokku/plugins/common"

	"github.com/ryanuber/columnize"
)

// listView holds a Build plus its computed display status, used by
// builds:list / builds:info JSON output.
type listView struct {
	Build
	DisplayStatus BuildStatus `json:"display_status"`
	Duration      string      `json:"duration"`
	LogPath       string      `json:"log_path"`
}

func toListView(b Build) listView {
	return listView{
		Build:         b,
		DisplayStatus: b.DisplayStatus(),
		Duration:      b.Duration().String(),
		LogPath:       b.LogPath(),
	}
}

// CommandList lists running builds across apps (no app arg) or running + recent
// records for one app.
func CommandList(appName string, format string, kindFilter string, statusFilter string) error {
	if format == "" {
		format = "stdout"
	}
	if format != "stdout" && format != "json" {
		return fmt.Errorf("Invalid format specified, supported formats: json, stdout")
	}

	if kindFilter != "" {
		k := BuildKind(kindFilter)
		if !k.Valid() {
			return fmt.Errorf("Invalid --kind value %q (allowed: build, deploy)", kindFilter)
		}
	}
	if statusFilter != "" {
		if !validStatusFilter(BuildStatus(statusFilter)) {
			return fmt.Errorf("Invalid --status value %q (allowed: running, succeeded, failed, canceled, abandoned)", statusFilter)
		}
	}

	if appName == "" {
		return commandListAllRunning(format, kindFilter, statusFilter)
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	all, err := FetchBuilds(appName)
	if err != nil {
		return err
	}

	retention := ResolveRetention(appName)
	views := make([]listView, 0, len(all))
	for _, b := range all {
		v := toListView(b)
		if !matchesKind(b, kindFilter) {
			continue
		}
		if !matchesStatus(v, statusFilter) {
			continue
		}
		views = append(views, v)
	}

	sort.SliceStable(views, func(i, j int) bool {
		// Live-running first, then by start time descending.
		iLive := views[i].Build.Status == BuildStatusRunning && views[i].DisplayStatus == BuildStatusRunning
		jLive := views[j].Build.Status == BuildStatusRunning && views[j].DisplayStatus == BuildStatusRunning
		if iLive != jLive {
			return iLive
		}
		return views[i].StartedAt.After(views[j].StartedAt)
	})

	if statusFilter == "" && kindFilter == "" && len(views) > retention {
		live := 0
		for _, v := range views {
			if v.DisplayStatus == BuildStatusRunning {
				live++
			}
		}
		cap := live + retention
		if cap < len(views) {
			views = views[:cap]
		}
	}

	if format == "json" {
		body, err := json.Marshal(views)
		if err != nil {
			return err
		}
		common.Log(string(body))
		return nil
	}

	if len(views) == 0 {
		fmt.Println("No builds recorded for this app")
		return nil
	}

	rows := []string{"Build ID | Kind | Status | PID | Source | Started | Duration"}
	for _, v := range views {
		rows = append(rows, fmt.Sprintf("%s | %s | %s | %d | %s | %s | %s",
			v.ID,
			v.Kind,
			v.DisplayStatus,
			v.PID,
			v.Source,
			v.StartedAt.Format(time.RFC3339),
			v.Duration,
		))
	}
	fmt.Println(columnize.SimpleFormat(rows))
	return nil
}

func commandListAllRunning(format, kindFilter, statusFilter string) error {
	apps, err := common.DokkuApps()
	if err != nil && !errors.Is(err, common.NoAppsExist) {
		return err
	}

	views := []listView{}
	for _, appName := range apps {
		running, err := FetchBuilds(appName)
		if err != nil {
			common.LogWarn(fmt.Sprintf("Could not read builds for %s: %s", appName, err))
			continue
		}
		for _, b := range running {
			if b.Status != BuildStatusRunning {
				continue
			}
			v := toListView(b)
			if !matchesKind(b, kindFilter) {
				continue
			}
			if !matchesStatus(v, statusFilter) {
				continue
			}
			views = append(views, v)
		}
	}

	sort.SliceStable(views, func(i, j int) bool {
		return views[i].StartedAt.After(views[j].StartedAt)
	})

	if format == "json" {
		body, err := json.Marshal(views)
		if err != nil {
			return err
		}
		common.Log(string(body))
		return nil
	}

	if len(views) == 0 {
		fmt.Println("No builds currently running")
		return nil
	}

	rows := []string{"App | Build ID | Kind | PID | Source | Started"}
	for _, v := range views {
		rows = append(rows, fmt.Sprintf("%s | %s | %s | %d | %s | %s",
			v.App,
			v.ID,
			v.Kind,
			v.PID,
			v.Source,
			v.StartedAt.Format(time.RFC3339),
		))
	}
	fmt.Println(columnize.SimpleFormat(rows))
	return nil
}

func matchesKind(b Build, filter string) bool {
	if filter == "" {
		return true
	}
	return string(b.Kind) == filter
}

func matchesStatus(v listView, filter string) bool {
	if filter == "" {
		return true
	}
	return string(v.DisplayStatus) == filter
}

func validStatusFilter(s BuildStatus) bool {
	switch s {
	case BuildStatusRunning, BuildStatusSucceeded, BuildStatusFailed, BuildStatusCanceled, BuildStatusAbandoned:
		return true
	}
	return false
}

// CommandInfo prints details for a single build record.
func CommandInfo(appName, buildID, format string) error {
	if format == "" {
		format = "stdout"
	}
	if format != "stdout" && format != "json" {
		return fmt.Errorf("Invalid format specified, supported formats: json, stdout")
	}
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}
	if buildID == "" {
		return errors.New("Please specify a build id")
	}

	b, err := ReadBuild(appName, buildID)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("No build record found for %s/%s", appName, buildID)
		}
		return err
	}
	v := toListView(b)

	if format == "json" {
		body, err := json.Marshal(v)
		if err != nil {
			return err
		}
		common.Log(string(body))
		return nil
	}

	rows := []string{
		fmt.Sprintf("Build ID:|%s", v.ID),
		fmt.Sprintf("App:|%s", v.App),
		fmt.Sprintf("Kind:|%s", v.Kind),
		fmt.Sprintf("Status:|%s", v.DisplayStatus),
		fmt.Sprintf("PID:|%d", v.PID),
		fmt.Sprintf("Source:|%s", v.Source),
		fmt.Sprintf("Started:|%s", v.StartedAt.Format(time.RFC3339)),
	}
	if v.FinishedAt != nil {
		rows = append(rows, fmt.Sprintf("Finished:|%s", v.FinishedAt.Format(time.RFC3339)))
	}
	rows = append(rows, fmt.Sprintf("Duration:|%s", v.Duration))
	if v.ExitCode != nil {
		rows = append(rows, fmt.Sprintf("Exit Code:|%d", *v.ExitCode))
	}
	rows = append(rows, fmt.Sprintf("Log:|%s", v.LogPath))
	fmt.Println(columnize.Format(rows, &columnize.Config{Delim: "|"}))
	return nil
}

// CommandCancel cancels the in-flight build for an app.
func CommandCancel(appName string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	lockPath := deployLockPath(appName)
	body, err := os.ReadFile(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			common.LogInfo1("App not currently deploying")
			return nil
		}
		return err
	}
	buildID := strings.TrimSpace(string(body))
	if buildID == "" {
		common.LogInfo1("No matching app deploy found")
		return nil
	}

	b, err := ReadBuild(appName, buildID)
	if err != nil {
		if os.IsNotExist(err) {
			common.LogInfo1(fmt.Sprintf("No build record for %s, removing stale lock file", buildID))
			_ = os.Remove(lockPath)
			return nil
		}
		return err
	}

	if b.Status != BuildStatusRunning {
		common.LogInfo1(fmt.Sprintf("Build %s is no longer running (status=%s); leaving record untouched", buildID, b.Status))
		return nil
	}

	if !CheckPIDAlive(b.PID) {
		common.LogInfo1("Build was already terminated, marking failed")
		now := time.Now().UTC()
		exitCode := -1
		b.Status = BuildStatusFailed
		b.FinishedAt = &now
		b.ExitCode = &exitCode
		if err := WriteBuild(b); err != nil {
			return err
		}
		_ = os.Remove(lockPath)
		return nil
	}

	common.LogInfo1("Killing app deploy")
	if err := killProcessGroup(b.PID); err != nil {
		common.LogWarn(fmt.Sprintf("Failed to signal process group %d: %s", b.PID, err))
	}

	current, err := ReadBuild(appName, buildID)
	if err == nil && current.Status == BuildStatusRunning {
		now := time.Now().UTC()
		exitCode := -1
		current.Status = BuildStatusCanceled
		current.FinishedAt = &now
		current.ExitCode = &exitCode
		if err := WriteBuild(current); err != nil {
			return err
		}
	}

	_ = os.Remove(lockPath)
	return nil
}

func killProcessGroup(pid int) error {
	pgid, err := syscall.Getpgid(pid)
	if err != nil {
		return err
	}
	return syscall.Kill(-pgid, syscall.SIGQUIT)
}

func deployLockPath(appName string) string {
	return filepath.Join(common.GetAppDataDirectory("apps", appName), ".deploy.lock")
}

// CommandOutput streams the build log for a given build (or the current one).
func CommandOutput(appName, buildID string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if buildID == "" || buildID == "current" {
		body, err := os.ReadFile(deployLockPath(appName))
		if err != nil {
			if os.IsNotExist(err) {
				common.LogInfo1("App not currently deploying")
				return nil
			}
			return err
		}
		buildID = strings.TrimSpace(string(body))
		if buildID == "" {
			common.LogInfo1("No matching app deploy found")
			return nil
		}
	}

	logPath := LogPathFor(appName, buildID)
	if _, err := os.Stat(logPath); err != nil {
		if os.IsNotExist(err) {
			return outputViaJournalctl(buildID)
		}
		return err
	}

	b, err := ReadBuild(appName, buildID)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if err == nil && b.Status == BuildStatusRunning && CheckPIDAlive(b.PID) {
		return execStreaming("tail", "-n", "1000", "-f", logPath)
	}
	return execStreaming("cat", logPath)
}

func outputViaJournalctl(buildID string) error {
	if _, err := exec.LookPath("journalctl"); err != nil {
		return fmt.Errorf("log file missing and journalctl not available for build %s", buildID)
	}
	return execStreaming("journalctl", "-n", "1000", "-a", "-o", "cat", fmt.Sprintf("SYSLOG_IDENTIFIER=dokku-%s", buildID))
}

func execStreaming(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// CommandPrune manually invokes PruneAppBuilds for one or every app.
func CommandPrune(appName string, allApps bool) error {
	if allApps {
		apps, err := common.DokkuApps()
		if err != nil {
			if errors.Is(err, common.NoAppsExist) {
				common.LogWarn(err.Error())
				return nil
			}
			return err
		}
		for _, app := range apps {
			if err := PruneAppBuilds(app); err != nil {
				common.LogWarn(fmt.Sprintf("Prune failed for %s: %s", app, err))
			}
		}
		return nil
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}
	return PruneAppBuilds(appName)
}

// CommandReport displays a build report for one or more apps.
func CommandReport(appName, format, infoFlag string) error {
	if appName == "" {
		apps, err := common.DokkuApps()
		if err != nil {
			if errors.Is(err, common.NoAppsExist) {
				common.LogWarn(err.Error())
				return nil
			}
			return err
		}
		for _, app := range apps {
			if err := ReportSingleApp(app, format, infoFlag); err != nil {
				return err
			}
		}
		return nil
	}
	return ReportSingleApp(appName, format, infoFlag)
}
