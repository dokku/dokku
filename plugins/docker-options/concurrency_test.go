package dockeroptions

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"
)

func setupTestApp(t *testing.T, appName string) {
	t.Helper()
	dokkuRoot := setupMigrationEnv(t)
	if err := os.MkdirAll(filepath.Join(dokkuRoot, appName), 0755); err != nil {
		t.Fatalf("failed to create app directory: %v", err)
	}
}

func TestConcurrentAddDockerOptions(t *testing.T) {
	appName := "race-test-app"
	setupTestApp(t, appName)

	const workerCount = 25
	var wg sync.WaitGroup
	errCh := make(chan error, workerCount)

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			opt := fmt.Sprintf("--label=race%d", id)
			if err := AddDockerOptionToPhases(appName, []string{"deploy"}, opt); err != nil {
				errCh <- fmt.Errorf("worker %d failed: %w", id, err)
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Fatalf("concurrent add error: %v", err)
	}

	options, err := GetDockerOptionsForProcessPhase(appName, DefaultProcessType, "deploy")
	if err != nil {
		t.Fatalf("failed to get docker options: %v", err)
	}

	if len(options) != workerCount {
		t.Fatalf("expected %d options, got %d: %v", workerCount, len(options), options)
	}

	for i := 0; i < workerCount; i++ {
		expectedOpt := fmt.Sprintf("--label=race%d", i)
		if !slices.Contains(options, expectedOpt) {
			t.Errorf("missing expected option %q in %v", expectedOpt, options)
		}
	}
}

func TestConcurrentCommandAdd(t *testing.T) {
	appName := "race-cmd-app"
	setupTestApp(t, appName)

	const workerCount = 20
	var wg sync.WaitGroup
	errCh := make(chan error, workerCount)

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			opt := fmt.Sprintf("--label=cmd%d", id)
			if err := CommandAdd(appName, nil, "deploy", opt); err != nil {
				errCh <- fmt.Errorf("CommandAdd worker %d failed: %w", id, err)
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Fatalf("concurrent CommandAdd error: %v", err)
	}

	options, err := GetDockerOptionsForProcessPhase(appName, DefaultProcessType, "deploy")
	if err != nil {
		t.Fatalf("failed to get docker options: %v", err)
	}

	if len(options) != workerCount {
		t.Fatalf("expected %d options, got %d: %v", workerCount, len(options), options)
	}

	for i := 0; i < workerCount; i++ {
		expectedOpt := fmt.Sprintf("--label=cmd%d", i)
		if !slices.Contains(options, expectedOpt) {
			t.Errorf("missing expected option %q in %v", expectedOpt, options)
		}
	}
}

func TestConcurrentAddAndRemoveDockerOptions(t *testing.T) {
	appName := "race-mixed-app"
	setupTestApp(t, appName)

	const initialCount = 10
	for i := 0; i < initialCount; i++ {
		if err := AddDockerOptionToPhases(appName, []string{"deploy"}, fmt.Sprintf("--label=initial%d", i)); err != nil {
			t.Fatalf("seed failed: %v", err)
		}
	}

	var wg sync.WaitGroup
	errCh := make(chan error, initialCount*2)

	// Concurrently remove initial options
	for i := 0; i < initialCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			opt := fmt.Sprintf("--label=initial%d", id)
			if err := RemoveDockerOptionFromPhases(appName, []string{"deploy"}, opt); err != nil {
				errCh <- fmt.Errorf("remove worker %d failed: %w", id, err)
			}
		}(i)
	}

	// Concurrently add new options
	for i := 0; i < initialCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			opt := fmt.Sprintf("--label=added%d", id)
			if err := AddDockerOptionToPhases(appName, []string{"deploy"}, opt); err != nil {
				errCh <- fmt.Errorf("add worker %d failed: %w", id, err)
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Fatalf("concurrent add/remove error: %v", err)
	}

	options, err := GetDockerOptionsForProcessPhase(appName, DefaultProcessType, "deploy")
	if err != nil {
		t.Fatalf("failed to get docker options: %v", err)
	}

	// Verify all initial options are removed
	for i := 0; i < initialCount; i++ {
		initOpt := fmt.Sprintf("--label=initial%d", i)
		if slices.Contains(options, initOpt) {
			t.Errorf("found removed option %q in %v", initOpt, options)
		}
	}

	// Verify all new options are present
	for i := 0; i < initialCount; i++ {
		addedOpt := fmt.Sprintf("--label=added%d", i)
		if !slices.Contains(options, addedOpt) {
			t.Errorf("missing added option %q in %v", addedOpt, options)
		}
	}

	if len(options) != initialCount {
		t.Fatalf("expected %d options, got %d: %v", initialCount, len(options), options)
	}
}

func TestConcurrentMultipleAppsIsolation(t *testing.T) {
	appA := "isolation-app-a"
	appB := "isolation-app-b"

	dokkuRoot := setupMigrationEnv(t)
	if err := os.MkdirAll(filepath.Join(dokkuRoot, appA), 0755); err != nil {
		t.Fatalf("failed to create appA directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dokkuRoot, appB), 0755); err != nil {
		t.Fatalf("failed to create appB directory: %v", err)
	}

	const workerCount = 15
	var wg sync.WaitGroup
	errCh := make(chan error, workerCount*2)

	for i := 0; i < workerCount; i++ {
		wg.Add(2)
		go func(id int) {
			defer wg.Done()
			if err := AddDockerOptionToPhases(appA, []string{"deploy"}, fmt.Sprintf("--label=appA-%d", id)); err != nil {
				errCh <- fmt.Errorf("appA worker %d failed: %w", id, err)
			}
		}(i)
		go func(id int) {
			defer wg.Done()
			if err := AddDockerOptionToPhases(appB, []string{"deploy"}, fmt.Sprintf("--label=appB-%d", id)); err != nil {
				errCh <- fmt.Errorf("appB worker %d failed: %w", id, err)
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Fatalf("concurrent multi-app error: %v", err)
	}

	optsA, err := GetDockerOptionsForProcessPhase(appA, DefaultProcessType, "deploy")
	if err != nil {
		t.Fatalf("failed to get appA options: %v", err)
	}
	optsB, err := GetDockerOptionsForProcessPhase(appB, DefaultProcessType, "deploy")
	if err != nil {
		t.Fatalf("failed to get appB options: %v", err)
	}

	if len(optsA) != workerCount {
		t.Errorf("appA: expected %d options, got %d: %v", workerCount, len(optsA), optsA)
	}
	if len(optsB) != workerCount {
		t.Errorf("appB: expected %d options, got %d: %v", workerCount, len(optsB), optsB)
	}

	for i := 0; i < workerCount; i++ {
		if !slices.Contains(optsA, fmt.Sprintf("--label=appA-%d", i)) {
			t.Errorf("appA missing expected option --label=appA-%d", i)
		}
		if !slices.Contains(optsB, fmt.Sprintf("--label=appB-%d", i)) {
			t.Errorf("appB missing expected option --label=appB-%d", i)
		}
	}
}

func TestConcurrentSetDockerOption(t *testing.T) {
	appName := "race-set-app"
	setupTestApp(t, appName)

	const workerCount = 15
	var wg sync.WaitGroup
	errCh := make(chan error, workerCount)

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if err := SetDockerOptionForPhases(appName, []string{"deploy"}, "restart", fmt.Sprintf("on-failure:%d", id)); err != nil {
				errCh <- fmt.Errorf("Set worker %d failed: %w", id, err)
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Fatalf("concurrent SetDockerOption error: %v", err)
	}

	options, err := GetDockerOptionsForProcessPhase(appName, DefaultProcessType, "deploy")
	if err != nil {
		t.Fatalf("failed to get docker options: %v", err)
	}

	if len(options) != 1 {
		t.Fatalf("expected exactly 1 restart option, got %d: %v", len(options), options)
	}

	if !slices.ContainsFunc(options, func(opt string) bool {
		return len(opt) > 10 && opt[:10] == "--restart="
	}) {
		t.Errorf("expected option starting with --restart=, got %v", options)
	}
}
