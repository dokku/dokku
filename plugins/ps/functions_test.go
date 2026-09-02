package ps

import (
	"os"
	"testing"
)

// assertFormations fails the test when the specified formations do not match
// the expected process type and quantity pairs, in order
func assertFormations(t *testing.T, formations FormationSlice, expected []Formation) {
	t.Helper()

	if len(formations) != len(expected) {
		t.Fatalf("expected %d formations, got %d", len(expected), len(formations))
	}

	for i, formation := range formations {
		if formation.ProcessType != expected[i].ProcessType || formation.Quantity != expected[i].Quantity {
			t.Fatalf("expected %s=%d at index %d, got %s=%d", expected[i].ProcessType, expected[i].Quantity, i, formation.ProcessType, formation.Quantity)
		}
	}
}

func TestAppendClearedFormations(t *testing.T) {
	testCases := []struct {
		name          string
		processTuples []string
		existingTuple []string
		expected      []Formation
	}{
		{
			name:          "no existing formations",
			processTuples: []string{"web=1"},
			existingTuple: []string{},
			expected:      []Formation{{ProcessType: "web", Quantity: 1}},
		},
		{
			name:          "zeroes out unspecified process types",
			processTuples: []string{"web=1"},
			existingTuple: []string{"web=2", "worker=3"},
			expected:      []Formation{{ProcessType: "web", Quantity: 1}, {ProcessType: "worker", Quantity: 0}},
		},
		{
			name:          "does not overwrite specified process types",
			processTuples: []string{"web=1", "worker=2"},
			existingTuple: []string{"web=4", "worker=5"},
			expected:      []Formation{{ProcessType: "web", Quantity: 1}, {ProcessType: "worker", Quantity: 2}},
		},
		{
			name:          "zeroes out every process type when none are specified",
			processTuples: []string{},
			existingTuple: []string{"web=2", "worker=3"},
			expected:      []Formation{{ProcessType: "web", Quantity: 0}, {ProcessType: "worker", Quantity: 0}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			formations, err := parseProcessTuples(testCase.processTuples)
			if err != nil {
				t.Fatalf("unexpected error parsing process tuples: %v", err)
			}

			existingFormations, err := parseProcessTuples(testCase.existingTuple)
			if err != nil {
				t.Fatalf("unexpected error parsing existing process tuples: %v", err)
			}

			assertFormations(t, appendClearedFormations(formations, existingFormations), testCase.expected)
		})
	}
}

func TestDefaultProcessTuplesWithoutProcfile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ps-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	t.Setenv("DOKKU_LIB_ROOT", tmpDir)
	t.Setenv("DOKKU_PID", "")

	processTuples, err := defaultProcessTuples("test-app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(processTuples) != 1 || processTuples[0] != "web=1" {
		t.Fatalf("expected [web=1], got %v", processTuples)
	}
}
