package builds

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/dokku/dokku/plugins/common"

	"github.com/multiformats/go-base36"
)

const (
	DefaultRetention = 20
	MinimumRetention = 1
)

var (
	DefaultProperties = map[string]string{
		"retention": strconv.Itoa(DefaultRetention),
	}

	GlobalProperties = map[string]bool{
		"retention": true,
	}
)

// BuildKind represents whether a record is a full build or a deploy-only run.
type BuildKind string

const (
	BuildKindBuild  BuildKind = "build"
	BuildKindDeploy BuildKind = "deploy"
)

// Valid reports whether the kind is one of the recognized values.
func (k BuildKind) Valid() bool {
	return k == BuildKindBuild || k == BuildKindDeploy
}

// BuildStatus is the on-disk lifecycle status of a build record.
// BuildStatusAbandoned is intentionally NOT considered Valid - it is a
// display-only value computed at read time and never persisted.
type BuildStatus string

const (
	BuildStatusRunning   BuildStatus = "running"
	BuildStatusSucceeded BuildStatus = "succeeded"
	BuildStatusFailed    BuildStatus = "failed"
	BuildStatusCanceled  BuildStatus = "canceled"
	BuildStatusAbandoned BuildStatus = "abandoned"
)

// Valid reports whether the status is a persistable value.
func (s BuildStatus) Valid() bool {
	switch s {
	case BuildStatusRunning, BuildStatusSucceeded, BuildStatusFailed, BuildStatusCanceled:
		return true
	}
	return false
}

// IsTerminal reports whether the status represents a finalized build.
func (s BuildStatus) IsTerminal() bool {
	switch s {
	case BuildStatusSucceeded, BuildStatusFailed, BuildStatusCanceled:
		return true
	}
	return false
}

// BuildSource identifies the user-typed command (or internal trigger) that
// originated a build.
type BuildSource string

const (
	BuildSourceGitHook        BuildSource = "git-hook"
	BuildSourcePsRebuild      BuildSource = "ps:rebuild"
	BuildSourcePsRestart      BuildSource = "ps:restart"
	BuildSourcePsStart        BuildSource = "ps:start"
	BuildSourceDeploy         BuildSource = "deploy"
	BuildSourceConfigRedeploy BuildSource = "config-redeploy"
	BuildSourceGitSync        BuildSource = "git:sync"
	BuildSourceGitFromArchive BuildSource = "git:from-archive"
	BuildSourceGitFromImage   BuildSource = "git:from-image"
	BuildSourceGitLoadImage   BuildSource = "git:load-image"
	BuildSourceUnknown        BuildSource = "unknown"
)

var allBuildSources = []BuildSource{
	BuildSourceGitHook,
	BuildSourcePsRebuild,
	BuildSourcePsRestart,
	BuildSourcePsStart,
	BuildSourceDeploy,
	BuildSourceConfigRedeploy,
	BuildSourceGitSync,
	BuildSourceGitFromArchive,
	BuildSourceGitFromImage,
	BuildSourceGitLoadImage,
	BuildSourceUnknown,
}

// Valid reports whether the source is one of the recognized values.
func (s BuildSource) Valid() bool {
	for _, candidate := range allBuildSources {
		if s == candidate {
			return true
		}
	}
	return false
}

// DefaultKind maps a source to the kind of work it performs. Sources that
// produce a new image map to BuildKindBuild; sources that re-deploy an existing
// image map to BuildKindDeploy.
func (s BuildSource) DefaultKind() BuildKind {
	switch s {
	case BuildSourceGitHook,
		BuildSourcePsRebuild,
		BuildSourceGitSync,
		BuildSourceGitFromArchive,
		BuildSourceGitFromImage,
		BuildSourceGitLoadImage:
		return BuildKindBuild
	}
	return BuildKindDeploy
}

// Build is the persisted record for a single build/deploy.
type Build struct {
	ID         string      `json:"id"`
	App        string      `json:"app"`
	Kind       BuildKind   `json:"kind"`
	PID        int         `json:"pid"`
	StartedAt  time.Time   `json:"started_at"`
	FinishedAt *time.Time  `json:"finished_at,omitempty"`
	Status     BuildStatus `json:"status"`
	Source     BuildSource `json:"source"`
	ExitCode   *int        `json:"exit_code,omitempty"`
}

// DisplayStatus returns the status the operator should see, computing
// BuildStatusAbandoned for running records whose PID is no longer alive.
func (b Build) DisplayStatus() BuildStatus {
	if b.Status == BuildStatusRunning && !CheckPIDAlive(b.PID) {
		return BuildStatusAbandoned
	}
	return b.Status
}

// Duration returns the elapsed time between start and finish (or now, for
// in-flight builds).
func (b Build) Duration() time.Duration {
	end := time.Now().UTC()
	if b.FinishedAt != nil {
		end = *b.FinishedAt
	}
	if end.Before(b.StartedAt) {
		return 0
	}
	return end.Sub(b.StartedAt).Round(time.Second)
}

// LogPath returns the on-disk path of the per-build log file.
func (b Build) LogPath() string {
	return filepath.Join(common.GetAppDataDirectory("builds", b.App), b.ID+".log")
}

// GenerateBuildID produces a sortable base36 ULID-style id.
//
// Format: <8 base36 chars of ms timestamp><6 base36 chars of randomness>.
func GenerateBuildID() string {
	const tsLen = 8
	const randLen = 6

	ms := uint64(time.Now().UTC().UnixNano() / int64(time.Millisecond))
	tsBytes := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		tsBytes[i] = byte(ms & 0xff)
		ms >>= 8
	}

	tsEncoded := base36.EncodeToStringLc(tsBytes)
	if len(tsEncoded) > tsLen {
		tsEncoded = tsEncoded[len(tsEncoded)-tsLen:]
	}
	for len(tsEncoded) < tsLen {
		tsEncoded = "0" + tsEncoded
	}

	randBytes := make([]byte, 6)
	if _, err := rand.Read(randBytes); err != nil {
		// fall back to time-derived noise; uniqueness is degraded but never zero
		now := time.Now().UTC().UnixNano()
		for i := 0; i < len(randBytes); i++ {
			randBytes[i] = byte(now >> (8 * i))
		}
	}
	randEncoded := base36.EncodeToStringLc(randBytes)
	if len(randEncoded) > randLen {
		randEncoded = randEncoded[len(randEncoded)-randLen:]
	}
	for len(randEncoded) < randLen {
		randEncoded = "0" + randEncoded
	}

	return tsEncoded + randEncoded
}

// AppDataDir returns the directory that stores build records and logs for app.
func AppDataDir(appName string) string {
	return common.GetAppDataDirectory("builds", appName)
}

// RecordPath returns the on-disk path to the JSON record for a given build.
func RecordPath(appName, buildID string) string {
	return filepath.Join(AppDataDir(appName), buildID+".json")
}

// LogPathFor returns the on-disk path to the log file for a given build.
func LogPathFor(appName, buildID string) string {
	return filepath.Join(AppDataDir(appName), buildID+".log")
}

// WriteBuild writes a Build record to disk. It creates the data dir as needed.
func WriteBuild(b Build) error {
	if err := os.MkdirAll(AppDataDir(b.App), 0755); err != nil {
		return fmt.Errorf("create builds data dir: %w", err)
	}

	body, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal build record: %w", err)
	}

	tmp := RecordPath(b.App, b.ID) + ".tmp"
	if err := os.WriteFile(tmp, body, 0644); err != nil {
		return fmt.Errorf("write temp build record: %w", err)
	}
	if err := os.Rename(tmp, RecordPath(b.App, b.ID)); err != nil {
		return fmt.Errorf("rename build record into place: %w", err)
	}
	return nil
}

// ReadBuild loads the Build record from disk. Returns os.ErrNotExist when the
// record does not exist.
func ReadBuild(appName, buildID string) (Build, error) {
	var b Build
	body, err := os.ReadFile(RecordPath(appName, buildID))
	if err != nil {
		return b, err
	}
	if err := json.Unmarshal(body, &b); err != nil {
		return b, fmt.Errorf("parse build record %s: %w", buildID, err)
	}
	return b, nil
}

// FetchBuilds returns every recorded build for an app, sorted newest-first by
// StartedAt. It is read-only; callers that want to reap abandoned records
// should call ReapAbandonedBuilds explicitly via PruneAppBuilds or the
// builds:prune subcommand.
func FetchBuilds(appName string) ([]Build, error) {
	dir := AppDataDir(appName)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("list builds dir for %s: %w", appName, err)
	}

	builds := make([]Build, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		buildID := strings.TrimSuffix(name, ".json")
		b, err := ReadBuild(appName, buildID)
		if err != nil {
			common.LogWarn(fmt.Sprintf("Skipping unreadable build record %s/%s: %s", appName, name, err))
			continue
		}
		builds = append(builds, b)
	}

	sort.Slice(builds, func(i, j int) bool {
		return builds[i].StartedAt.After(builds[j].StartedAt)
	})
	return builds, nil
}

// FetchRunningBuilds returns the in-flight (display-status running) builds for
// an app. Records whose on-disk status is running but PID is dead are excluded.
func FetchRunningBuilds(appName string) ([]Build, error) {
	all, err := FetchBuilds(appName)
	if err != nil {
		return nil, err
	}

	running := make([]Build, 0, len(all))
	for _, b := range all {
		if b.Status == BuildStatusRunning && CheckPIDAlive(b.PID) {
			running = append(running, b)
		}
	}
	return running, nil
}

// CheckPIDAlive reports whether a process with the given pid currently exists.
// On unix, kill(pid, 0) returns nil if the process exists (or ESRCH if it does
// not). It treats invalid pids (<= 0) as dead.
func CheckPIDAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	if err := syscall.Kill(pid, 0); err != nil {
		if err == syscall.ESRCH {
			return false
		}
		// EPERM means the process exists but we don't have permission to
		// signal it - still alive from the operator's perspective.
		if err == syscall.EPERM {
			return true
		}
		return false
	}
	return true
}

// ResolveRetention returns the retention count for an app, cascading from
// per-app override → global override → DefaultRetention. Invalid (non-int or
// non-positive) property values fall through to the next level with a warning.
func ResolveRetention(appName string) int {
	if appName != "" && appName != "--global" {
		if n, ok := parseRetention(common.PropertyGet("builds", appName, "retention"), appName); ok {
			return n
		}
	}
	if n, ok := parseRetention(common.PropertyGet("builds", "--global", "retention"), "--global"); ok {
		return n
	}
	return DefaultRetention
}

func parseRetention(raw string, scope string) (int, bool) {
	if raw == "" {
		return 0, false
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		common.LogWarn(fmt.Sprintf("Ignoring non-integer builds retention %q for %s", raw, scope))
		return 0, false
	}
	if n < MinimumRetention {
		common.LogWarn(fmt.Sprintf("Ignoring builds retention %d below minimum %d for %s", n, MinimumRetention, scope))
		return 0, false
	}
	return n, true
}

// ReapAbandonedBuilds finalizes any record whose on-disk status is running but
// whose PID is no longer alive. Records are written with status=failed,
// exit_code=-1, finished_at=now. Returns the number of records reaped.
func ReapAbandonedBuilds(appName string) (int, error) {
	builds, err := FetchBuilds(appName)
	if err != nil {
		return 0, err
	}

	now := time.Now().UTC()
	exitCode := -1
	reaped := 0
	for _, b := range builds {
		if b.Status != BuildStatusRunning {
			continue
		}
		if CheckPIDAlive(b.PID) {
			continue
		}
		b.Status = BuildStatusFailed
		b.FinishedAt = &now
		b.ExitCode = &exitCode
		if err := WriteBuild(b); err != nil {
			common.LogWarn(fmt.Sprintf("Could not reap abandoned build %s/%s: %s", appName, b.ID, err))
			continue
		}
		reaped++
	}
	return reaped, nil
}

// PruneAppBuilds reaps abandoned records, then prunes finalized records beyond
// the retention cap. Live in-flight builds (status=running with alive PID) are
// always preserved.
func PruneAppBuilds(appName string) error {
	if _, err := ReapAbandonedBuilds(appName); err != nil {
		return err
	}

	builds, err := FetchBuilds(appName)
	if err != nil {
		return err
	}

	retention := ResolveRetention(appName)

	finalized := make([]Build, 0, len(builds))
	for _, b := range builds {
		if b.Status == BuildStatusRunning && CheckPIDAlive(b.PID) {
			continue
		}
		finalized = append(finalized, b)
	}

	if len(finalized) <= retention {
		return nil
	}

	for _, b := range finalized[retention:] {
		removeBuildFiles(appName, b.ID)
	}
	return nil
}

func removeBuildFiles(appName, buildID string) {
	if err := os.Remove(RecordPath(appName, buildID)); err != nil && !os.IsNotExist(err) {
		common.LogWarn(fmt.Sprintf("Could not remove build record %s/%s: %s", appName, buildID, err))
	}
	if err := os.Remove(LogPathFor(appName, buildID)); err != nil && !os.IsNotExist(err) {
		common.LogWarn(fmt.Sprintf("Could not remove build log %s/%s.log: %s", appName, buildID, err))
	}
}
