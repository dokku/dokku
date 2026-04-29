package dockeroptions

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

var availablePhases = []string{"build", "deploy", "run"}

// processScopedPhases lists the phases that may be scoped via --process.
// Only deploy is scoped: build runs once per app (no process-type concept) and
// run is invoked for ad-hoc commands and cron tasks where the caller does not
// supply a Procfile process type.
var processScopedPhases = map[string]bool{
	"deploy": true,
}

func isValidPhase(phase string) bool {
	for _, p := range availablePhases {
		if p == phase {
			return true
		}
	}
	return false
}

func parsePhases(phasesArg string) ([]string, error) {
	if phasesArg == "" {
		return nil, nil
	}

	phases := strings.Split(phasesArg, ",")
	for _, phase := range phases {
		if !isValidPhase(phase) {
			return nil, fmt.Errorf("Phase(s) must be one of [%s]", strings.Join(availablePhases, " "))
		}
	}

	return phases, nil
}

// ValidateProcessFlag rejects invalid combinations of --process and the
// supplied phases. When processes is empty the call applies to the default
// scope and any phase is allowed.
func ValidateProcessFlag(processes []string, phases []string) error {
	if len(processes) == 0 {
		return nil
	}

	for _, processType := range processes {
		if processType == "" {
			return errors.New("--process value must not be empty")
		}
		if processType == DefaultProcessType {
			return fmt.Errorf("%q is reserved and cannot be used as a --process value", DefaultProcessType)
		}
	}

	for _, phase := range phases {
		if !processScopedPhases[phase] {
			return fmt.Errorf("--process is only supported for the deploy phase, got %q", phase)
		}
	}

	return nil
}

// WarnIfProcessNotInProcfile emits a warning (without failing) when the named
// process type is not present in the app's current Procfile. The check is
// best-effort: if the Procfile or procfile-util are unavailable, the function
// returns silently.
func WarnIfProcessNotInProcfile(appName, processType string) {
	procfilePath := filepath.Join(common.GetAppDataDirectory("ps", appName), "Procfile")
	if !common.FileExists(procfilePath) {
		return
	}

	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command: "procfile-util",
		Args:    []string{"list", "--procfile", procfilePath},
	})
	if err != nil || result.ExitCode != 0 {
		return
	}

	for _, line := range strings.Split(strings.TrimSpace(result.StdoutContents()), "\n") {
		if strings.TrimSpace(line) == processType {
			return
		}
	}

	common.LogWarn(fmt.Sprintf("Process type %q is not declared in the Procfile for %s", processType, appName))
}
