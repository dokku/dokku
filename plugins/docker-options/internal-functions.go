package dockeroptions

import (
	"errors"
	"fmt"
	"strings"
)

var availablePhases = []string{"build", "deploy", "run"}

func parsePhases(phasesArg string) ([]string, error) {
	if phasesArg == "" {
		return nil, nil
	}

	phases := strings.Split(phasesArg, ",")
	for _, phase := range phases {
		valid := false
		for _, allowed := range availablePhases {
			if phase == allowed {
				valid = true
				break
			}
		}
		if !valid {
			return nil, fmt.Errorf("Phase(s) must be one of [%s]", strings.Join(availablePhases, " "))
		}
	}

	return phases, nil
}

// ErrIfReservedFlagsUsed returns a not-yet-implemented error when --process or --global are passed.
// Both flags are reserved for the upcoming process-scoped docker options work (issue #2441).
func ErrIfReservedFlagsUsed(processes []string, global bool) error {
	if len(processes) == 0 && !global {
		return nil
	}

	if len(processes) > 0 && global {
		return errors.New("--process and --global flags are mutually exclusive")
	}

	return errors.New("--process and --global flags are reserved for a future release and not yet implemented")
}
