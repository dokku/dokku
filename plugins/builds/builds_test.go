package builds

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func setupTestRoot(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "builds-test")
	if err != nil {
		t.Fatalf("mkdirtemp: %v", err)
	}
	t.Setenv("DOKKU_LIB_ROOT", tmpDir)
	t.Setenv("PLUGIN_PATH", filepath.Join(tmpDir, "plugins"))
	t.Setenv("PLUGIN_ENABLED_PATH", filepath.Join(tmpDir, "plugins", "enabled"))
	t.Cleanup(func() { os.RemoveAll(tmpDir) })
	return tmpDir
}

func TestBuildKindValid(t *testing.T) {
	cases := map[BuildKind]bool{
		BuildKindBuild:    true,
		BuildKindDeploy:   true,
		BuildKind(""):     false,
		BuildKind("nope"): false,
	}
	for k, want := range cases {
		if got := k.Valid(); got != want {
			t.Errorf("BuildKind(%q).Valid() = %v, want %v", string(k), got, want)
		}
	}
}

func TestBuildStatusValid(t *testing.T) {
	cases := map[BuildStatus]bool{
		BuildStatusRunning:   true,
		BuildStatusSucceeded: true,
		BuildStatusFailed:    true,
		BuildStatusCanceled:  true,
		BuildStatusAbandoned: false, // intentionally not persistable
		BuildStatus(""):      false,
		BuildStatus("nope"):  false,
	}
	for s, want := range cases {
		if got := s.Valid(); got != want {
			t.Errorf("BuildStatus(%q).Valid() = %v, want %v", string(s), got, want)
		}
	}
}

func TestBuildStatusIsTerminal(t *testing.T) {
	cases := map[BuildStatus]bool{
		BuildStatusRunning:   false,
		BuildStatusSucceeded: true,
		BuildStatusFailed:    true,
		BuildStatusCanceled:  true,
		BuildStatusAbandoned: false,
	}
	for s, want := range cases {
		if got := s.IsTerminal(); got != want {
			t.Errorf("BuildStatus(%q).IsTerminal() = %v, want %v", string(s), got, want)
		}
	}
}

func TestBuildSourceValid(t *testing.T) {
	for _, src := range allBuildSources {
		if !src.Valid() {
			t.Errorf("expected %q to be valid", string(src))
		}
	}
	for _, bad := range []BuildSource{"", "garbage", "git:", "ps:something"} {
		if bad.Valid() {
			t.Errorf("expected %q to be invalid", string(bad))
		}
	}
}

func TestBuildSourceDefaultKind(t *testing.T) {
	build := []BuildSource{
		BuildSourceGitHook, BuildSourcePsRebuild, BuildSourceGitSync,
		BuildSourceGitFromArchive, BuildSourceGitFromImage, BuildSourceGitLoadImage,
	}
	deploy := []BuildSource{
		BuildSourcePsRestart, BuildSourcePsStart, BuildSourceDeploy,
		BuildSourceConfigRedeploy, BuildSourceUnknown,
	}
	for _, s := range build {
		if got := s.DefaultKind(); got != BuildKindBuild {
			t.Errorf("%q.DefaultKind() = %v, want build", string(s), got)
		}
	}
	for _, s := range deploy {
		if got := s.DefaultKind(); got != BuildKindDeploy {
			t.Errorf("%q.DefaultKind() = %v, want deploy", string(s), got)
		}
	}
}

func TestGenerateBuildIDUniquenessAndShape(t *testing.T) {
	const samples = 200
	seen := make(map[string]struct{}, samples)
	for i := 0; i < samples; i++ {
		id := GenerateBuildID()
		if len(id) != 14 {
			t.Fatalf("expected len 14, got %d for %q", len(id), id)
		}
		for _, r := range id {
			if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'z')) {
				t.Fatalf("non-base36 char %q in %q", r, id)
			}
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("duplicate id %q in %d samples", id, samples)
		}
		seen[id] = struct{}{}
	}
}

func TestWriteAndReadBuildRoundTrip(t *testing.T) {
	setupTestRoot(t)

	now := time.Now().UTC().Truncate(time.Second)
	exit := 0
	finished := now.Add(time.Minute)
	want := Build{
		ID:         GenerateBuildID(),
		App:        "myapp",
		Kind:       BuildKindBuild,
		PID:        12345,
		StartedAt:  now,
		FinishedAt: &finished,
		Status:     BuildStatusSucceeded,
		Source:     BuildSourceGitHook,
		ExitCode:   &exit,
	}

	if err := WriteBuild(want); err != nil {
		t.Fatalf("WriteBuild: %v", err)
	}

	got, err := ReadBuild(want.App, want.ID)
	if err != nil {
		t.Fatalf("ReadBuild: %v", err)
	}

	if got.ID != want.ID || got.Status != want.Status || got.Kind != want.Kind || got.Source != want.Source {
		t.Errorf("round-trip mismatch: got %+v want %+v", got, want)
	}
	if got.ExitCode == nil || *got.ExitCode != 0 {
		t.Errorf("exit code mismatch: got %v want 0", got.ExitCode)
	}
	if got.FinishedAt == nil || !got.FinishedAt.Equal(finished) {
		t.Errorf("finished_at mismatch: got %v want %v", got.FinishedAt, finished)
	}
}

func TestFetchBuildsSortsNewestFirst(t *testing.T) {
	setupTestRoot(t)

	base := time.Now().UTC().Truncate(time.Second)
	for i, offset := range []time.Duration{2 * time.Hour, 1 * time.Hour, 0} {
		b := Build{
			ID:        "build-" + strconv.Itoa(i),
			App:       "myapp",
			Kind:      BuildKindBuild,
			PID:       100 + i,
			StartedAt: base.Add(-offset),
			Status:    BuildStatusSucceeded,
			Source:    BuildSourceGitHook,
		}
		if err := WriteBuild(b); err != nil {
			t.Fatalf("write %d: %v", i, err)
		}
	}

	builds, err := FetchBuilds("myapp")
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if len(builds) != 3 {
		t.Fatalf("got %d builds, want 3", len(builds))
	}
	for i := 0; i < len(builds)-1; i++ {
		if !builds[i].StartedAt.After(builds[i+1].StartedAt) {
			t.Errorf("not sorted descending at %d: %v vs %v", i, builds[i].StartedAt, builds[i+1].StartedAt)
		}
	}
}

func TestFetchBuildsMissingDir(t *testing.T) {
	setupTestRoot(t)
	builds, err := FetchBuilds("never-deployed")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if len(builds) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(builds))
	}
}

func TestCheckPIDAlive(t *testing.T) {
	if CheckPIDAlive(os.Getpid()) != true {
		t.Error("expected current pid to be alive")
	}
	if CheckPIDAlive(0) {
		t.Error("expected pid 0 to be dead")
	}
	if CheckPIDAlive(-1) {
		t.Error("expected pid -1 to be dead")
	}
	// PID 1 is init/launchd; it always exists. We don't have permission to
	// signal it from a non-root user, but CheckPIDAlive should treat EPERM
	// as alive.
	if !CheckPIDAlive(1) {
		t.Error("expected pid 1 (init) to be reported alive")
	}
}

func TestReapAbandonedBuilds(t *testing.T) {
	setupTestRoot(t)
	app := "reapapp"

	// 1: alive (this test process), should NOT be reaped.
	alive := Build{
		ID:        "alive",
		App:       app,
		Kind:      BuildKindBuild,
		PID:       os.Getpid(),
		StartedAt: time.Now().UTC(),
		Status:    BuildStatusRunning,
		Source:    BuildSourceGitHook,
	}
	// 2: dead pid, status=running, should be reaped.
	dead := Build{
		ID:        "dead",
		App:       app,
		Kind:      BuildKindBuild,
		PID:       1, // signal-eperm but we treat as alive; use a clearly-dead one
		StartedAt: time.Now().UTC(),
		Status:    BuildStatusRunning,
		Source:    BuildSourceGitHook,
	}
	// pick a pid that shouldn't exist
	dead.PID = pickDeadPID(t)
	// 3: terminal, should not be touched.
	finished := time.Now().UTC()
	exit := 0
	terminal := Build{
		ID:         "terminal",
		App:        app,
		Kind:       BuildKindBuild,
		PID:        9999,
		StartedAt:  time.Now().UTC().Add(-time.Hour),
		FinishedAt: &finished,
		Status:     BuildStatusSucceeded,
		Source:     BuildSourceGitHook,
		ExitCode:   &exit,
	}

	for _, b := range []Build{alive, dead, terminal} {
		if err := WriteBuild(b); err != nil {
			t.Fatalf("write %s: %v", b.ID, err)
		}
	}

	reaped, err := ReapAbandonedBuilds(app)
	if err != nil {
		t.Fatalf("reap: %v", err)
	}
	if reaped != 1 {
		t.Errorf("expected 1 reap, got %d", reaped)
	}

	got, err := ReadBuild(app, "dead")
	if err != nil {
		t.Fatalf("read dead: %v", err)
	}
	if got.Status != BuildStatusFailed {
		t.Errorf("dead.Status = %v, want failed", got.Status)
	}
	if got.ExitCode == nil || *got.ExitCode != -1 {
		t.Errorf("dead.ExitCode = %v, want -1", got.ExitCode)
	}

	stillAlive, err := ReadBuild(app, "alive")
	if err != nil {
		t.Fatalf("read alive: %v", err)
	}
	if stillAlive.Status != BuildStatusRunning {
		t.Errorf("alive.Status = %v, want running", stillAlive.Status)
	}

	stillTerminal, err := ReadBuild(app, "terminal")
	if err != nil {
		t.Fatalf("read terminal: %v", err)
	}
	if stillTerminal.Status != BuildStatusSucceeded {
		t.Errorf("terminal.Status = %v, want succeeded", stillTerminal.Status)
	}
}

func pickDeadPID(t *testing.T) int {
	t.Helper()
	for candidate := 99999; candidate < 200000; candidate++ {
		if !CheckPIDAlive(candidate) {
			return candidate
		}
	}
	t.Fatal("could not find a dead pid candidate")
	return 0
}

func TestPruneAppBuildsRespectsRetentionAndProtectsLive(t *testing.T) {
	setupTestRoot(t)
	app := "pruneapp"

	now := time.Now().UTC()

	// 1 live in-flight build (status=running, alive PID).
	live := Build{
		ID:        "live",
		App:       app,
		Kind:      BuildKindBuild,
		PID:       os.Getpid(),
		StartedAt: now,
		Status:    BuildStatusRunning,
		Source:    BuildSourceGitHook,
	}
	if err := WriteBuild(live); err != nil {
		t.Fatalf("write live: %v", err)
	}

	// Force retention down to 2 via property.
	if err := os.MkdirAll(filepath.Join(os.Getenv("DOKKU_LIB_ROOT"), "config", "builds", app), 0755); err != nil {
		t.Fatalf("setup property dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(os.Getenv("DOKKU_LIB_ROOT"), "config", "builds", app, "retention"), []byte("2"), 0644); err != nil {
		t.Fatalf("write retention property: %v", err)
	}

	// 5 finalized builds with descending start times.
	exit := 0
	for i := 0; i < 5; i++ {
		finished := now.Add(-time.Duration(i+1) * time.Minute)
		started := finished.Add(-time.Second)
		b := Build{
			ID:         "old-" + strconv.Itoa(i),
			App:        app,
			Kind:       BuildKindBuild,
			PID:        9000 + i,
			StartedAt:  started,
			FinishedAt: &finished,
			Status:     BuildStatusSucceeded,
			Source:     BuildSourceGitHook,
			ExitCode:   &exit,
		}
		if err := WriteBuild(b); err != nil {
			t.Fatalf("write old-%d: %v", i, err)
		}
		// Touch a log file so the prune deletes that too.
		_ = os.WriteFile(b.LogPath(), []byte("log"), 0644)
	}

	if err := PruneAppBuilds(app); err != nil {
		t.Fatalf("prune: %v", err)
	}

	remaining, err := FetchBuilds(app)
	if err != nil {
		t.Fatalf("fetch after prune: %v", err)
	}

	// live + 2 retained = 3
	if len(remaining) != 3 {
		t.Fatalf("expected 3 records after prune, got %d (%v)", len(remaining), remainingIDs(remaining))
	}

	// live record must survive
	foundLive := false
	for _, r := range remaining {
		if r.ID == "live" {
			foundLive = true
			if r.Status != BuildStatusRunning {
				t.Errorf("live.Status changed to %v", r.Status)
			}
		}
	}
	if !foundLive {
		t.Errorf("live record was pruned (remaining=%v)", remainingIDs(remaining))
	}

	// the two newest old-* records (old-0 and old-1) should remain
	expectedKept := map[string]bool{"old-0": true, "old-1": true}
	for _, r := range remaining {
		if r.ID == "live" {
			continue
		}
		if !expectedKept[r.ID] {
			t.Errorf("unexpected surviving record %s", r.ID)
		}
	}

	// log files for pruned old-* should be gone; for kept ones, still present.
	for i := 2; i < 5; i++ {
		path := filepath.Join(AppDataDir(app), "old-"+strconv.Itoa(i)+".log")
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("expected log %s removed, got err=%v", path, err)
		}
	}
}

func remainingIDs(builds []Build) []string {
	ids := make([]string, 0, len(builds))
	for _, b := range builds {
		ids = append(ids, b.ID)
	}
	return ids
}

func TestResolveRetentionCascade(t *testing.T) {
	tmp := setupTestRoot(t)
	app := "retapp"

	// no overrides: default
	if got := ResolveRetention(app); got != DefaultRetention {
		t.Errorf("no overrides: got %d want %d", got, DefaultRetention)
	}

	// global override
	globalDir := filepath.Join(tmp, "config", "builds", "--global")
	if err := os.MkdirAll(globalDir, 0755); err != nil {
		t.Fatalf("mkdir global: %v", err)
	}
	if err := os.WriteFile(filepath.Join(globalDir, "retention"), []byte("7"), 0644); err != nil {
		t.Fatalf("write global retention: %v", err)
	}
	if got := ResolveRetention(app); got != 7 {
		t.Errorf("global override: got %d want 7", got)
	}

	// per-app override beats global
	appDir := filepath.Join(tmp, "config", "builds", app)
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatalf("mkdir app: %v", err)
	}
	if err := os.WriteFile(filepath.Join(appDir, "retention"), []byte("3"), 0644); err != nil {
		t.Fatalf("write app retention: %v", err)
	}
	if got := ResolveRetention(app); got != 3 {
		t.Errorf("per-app override: got %d want 3", got)
	}

	// invalid app value falls back to global
	if err := os.WriteFile(filepath.Join(appDir, "retention"), []byte("not-a-number"), 0644); err != nil {
		t.Fatalf("write invalid: %v", err)
	}
	if got := ResolveRetention(app); got != 7 {
		t.Errorf("invalid app value: got %d want 7 (global)", got)
	}

	// negative value also falls through
	if err := os.WriteFile(filepath.Join(appDir, "retention"), []byte("0"), 0644); err != nil {
		t.Fatalf("write zero: %v", err)
	}
	if got := ResolveRetention(app); got != 7 {
		t.Errorf("zero app value: got %d want 7 (global)", got)
	}
}

func TestDisplayStatusComputesAbandoned(t *testing.T) {
	setupTestRoot(t)
	abandoned := Build{
		ID:        "x",
		App:       "myapp",
		Kind:      BuildKindBuild,
		PID:       pickDeadPID(t),
		StartedAt: time.Now().UTC(),
		Status:    BuildStatusRunning,
		Source:    BuildSourceGitHook,
	}
	if got := abandoned.DisplayStatus(); got != BuildStatusAbandoned {
		t.Errorf("DisplayStatus on dead-PID running record = %v, want abandoned", got)
	}

	live := abandoned
	live.PID = os.Getpid()
	if got := live.DisplayStatus(); got != BuildStatusRunning {
		t.Errorf("DisplayStatus on alive-PID running record = %v, want running", got)
	}

	exit := 0
	finished := time.Now().UTC()
	terminal := Build{
		ID:         "y",
		App:        "myapp",
		Kind:       BuildKindBuild,
		PID:        os.Getpid(),
		StartedAt:  time.Now().UTC(),
		FinishedAt: &finished,
		Status:     BuildStatusSucceeded,
		Source:     BuildSourceGitHook,
		ExitCode:   &exit,
	}
	if got := terminal.DisplayStatus(); got != BuildStatusSucceeded {
		t.Errorf("terminal DisplayStatus = %v, want succeeded", got)
	}
}

func TestRecordStartFinalizeIdempotency(t *testing.T) {
	setupTestRoot(t)
	app := "idem"
	id := GenerateBuildID()

	if err := TriggerBuildsRecordStart(app, id, strconv.Itoa(os.Getpid()), string(BuildSourceGitHook)); err != nil {
		t.Fatalf("record-start: %v", err)
	}

	// First finalize: succeeded
	if err := TriggerBuildsRecordFinalize(app, id, "0"); err != nil {
		t.Fatalf("first finalize: %v", err)
	}
	first, err := ReadBuild(app, id)
	if err != nil {
		t.Fatalf("read after first finalize: %v", err)
	}
	if first.Status != BuildStatusSucceeded {
		t.Errorf("first finalize Status = %v, want succeeded", first.Status)
	}

	// Manually mark canceled, then call finalize with non-zero exit; the
	// canceled status should win because cancel finalized it first.
	canceled := first
	canceled.Status = BuildStatusCanceled
	if err := WriteBuild(canceled); err != nil {
		t.Fatalf("write canceled: %v", err)
	}
	if err := TriggerBuildsRecordFinalize(app, id, "1"); err != nil {
		t.Fatalf("second finalize: %v", err)
	}
	final, err := ReadBuild(app, id)
	if err != nil {
		t.Fatalf("read after second finalize: %v", err)
	}
	if final.Status != BuildStatusCanceled {
		t.Errorf("idempotent finalize Status = %v, want canceled", final.Status)
	}
}

func TestRecordStartUnknownSourceCoercedToUnknown(t *testing.T) {
	setupTestRoot(t)
	app := "src"
	id := GenerateBuildID()
	if err := TriggerBuildsRecordStart(app, id, strconv.Itoa(os.Getpid()), "made-up-source"); err != nil {
		t.Fatalf("record-start: %v", err)
	}
	got, err := ReadBuild(app, id)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got.Source != BuildSourceUnknown {
		t.Errorf("Source = %v, want unknown", got.Source)
	}
	if got.Kind != BuildKindDeploy {
		t.Errorf("Kind = %v, want deploy (DefaultKind for unknown)", got.Kind)
	}
}

func TestBuildJSONShape(t *testing.T) {
	setupTestRoot(t)
	now := time.Now().UTC().Truncate(time.Second)
	b := Build{
		ID:        "abc",
		App:       "j",
		Kind:      BuildKindDeploy,
		PID:       100,
		StartedAt: now,
		Status:    BuildStatusRunning,
		Source:    BuildSourcePsRestart,
	}
	if err := WriteBuild(b); err != nil {
		t.Fatalf("write: %v", err)
	}
	raw, err := os.ReadFile(RecordPath(b.App, b.ID))
	if err != nil {
		t.Fatalf("read raw: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"id", "app", "kind", "pid", "started_at", "status", "source"} {
		if _, ok := parsed[key]; !ok {
			t.Errorf("expected key %q in JSON, got %s", key, string(raw))
		}
	}
	for _, omittable := range []string{"finished_at", "exit_code"} {
		if _, ok := parsed[omittable]; ok {
			t.Errorf("expected key %q to be omitted while running, got %s", omittable, string(raw))
		}
	}
	if !strings.Contains(string(raw), `"kind": "deploy"`) {
		t.Errorf("kind not serialized as enum string: %s", string(raw))
	}
}
