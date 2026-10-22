package dockeroptions

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

func TestAddDockerOptionToPhasesConcurrent(t *testing.T) {
	const appName = "race-test"
	setupReportEnv(t, appName)

	const count = 20
	var wg sync.WaitGroup
	errs := make([]error, count)
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = AddDockerOptionToPhases(appName, []string{"deploy", "run"}, fmt.Sprintf("--label=race%d", i))
		}(i)
	}
	wg.Wait()
	if err := errors.Join(errs...); err != nil {
		t.Fatalf("AddDockerOptionToPhases: %v", err)
	}

	for _, phase := range []string{"deploy", "run"} {
		options, err := GetDockerOptionsForPhase(appName, phase)
		if err != nil {
			t.Fatalf("GetDockerOptionsForPhase(%s): %v", phase, err)
		}
		if len(options) != count {
			t.Fatalf("%s phase has %d options, want %d: %v", phase, len(options), count, options)
		}
	}
}

func TestRemoveDockerOptionFromPhasesConcurrent(t *testing.T) {
	const appName = "race-test"
	setupReportEnv(t, appName)

	const count = 20
	for i := 0; i < count; i++ {
		if err := AddDockerOptionToPhases(appName, []string{"deploy"}, fmt.Sprintf("--label=race%d", i)); err != nil {
			t.Fatalf("AddDockerOptionToPhases: %v", err)
		}
	}

	var wg sync.WaitGroup
	errs := make([]error, count)
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = RemoveDockerOptionFromPhases(appName, []string{"deploy"}, fmt.Sprintf("--label=race%d", i))
		}(i)
	}
	wg.Wait()
	if err := errors.Join(errs...); err != nil {
		t.Fatalf("RemoveDockerOptionFromPhases: %v", err)
	}

	options, err := GetDockerOptionsForPhase(appName, "deploy")
	if err != nil {
		t.Fatalf("GetDockerOptionsForPhase: %v", err)
	}
	if len(options) != 0 {
		t.Fatalf("deploy phase still has options: %v", options)
	}
}
