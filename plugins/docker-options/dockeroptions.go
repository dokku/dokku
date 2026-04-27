package dockeroptions

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// DefaultProcessType is the sentinel process-type key used for options that
// apply to every container in an app (i.e. options not scoped to a specific
// Procfile process type).
const DefaultProcessType = "_default_"

func propertyKey(processType, phase string) string {
	if processType == "" {
		processType = DefaultProcessType
	}
	return fmt.Sprintf("%s.%s", processType, phase)
}

// SetDockerOptionForPhases sets a `--name=value` option in the default scope
// for the specified phases, replacing any existing entry with the same name.
func SetDockerOptionForPhases(appName string, phases []string, name string, value string) error {
	return SetDockerOptionForProcessPhases(appName, []string{DefaultProcessType}, phases, name, value)
}

// SetDockerOptionForProcessPhases sets a `--name=value` option for the specified
// process types and phases, replacing any existing entry with the same name.
func SetDockerOptionForProcessPhases(appName string, processTypes []string, phases []string, name string, value string) error {
	if len(processTypes) == 0 {
		processTypes = []string{DefaultProcessType}
	}
	for _, processType := range processTypes {
		for _, phase := range phases {
			options, err := GetDockerOptionsForProcessPhase(appName, processType, phase)
			if err != nil {
				return err
			}

			newOptions := []string{}
			for _, option := range options {
				if strings.HasPrefix(option, fmt.Sprintf("--%s=", name)) {
					continue
				}
				newOptions = append(newOptions, option)
			}

			newOptions = append(newOptions, fmt.Sprintf("--%s=%s", name, value))
			sort.Strings(newOptions)
			if err := writeDockerOptionsForProcessPhase(appName, processType, phase, newOptions); err != nil {
				return err
			}
		}
	}
	return nil
}

// AddDockerOptionToPhases adds an option to the default scope for the specified phases.
func AddDockerOptionToPhases(appName string, phases []string, option string) error {
	return AddDockerOptionToProcessPhases(appName, []string{DefaultProcessType}, phases, option)
}

// AddDockerOptionToProcessPhases adds an option to the specified process types and phases.
func AddDockerOptionToProcessPhases(appName string, processTypes []string, phases []string, option string) error {
	if len(processTypes) == 0 {
		processTypes = []string{DefaultProcessType}
	}
	for _, processType := range processTypes {
		for _, phase := range phases {
			options, err := GetDockerOptionsForProcessPhase(appName, processType, phase)
			if err != nil {
				return err
			}

			options = append(options, option)
			sort.Strings(options)
			if err := writeDockerOptionsForProcessPhase(appName, processType, phase, options); err != nil {
				return err
			}
		}
	}
	return nil
}

// GetDockerOptionsForPhase returns the docker options stored under the default
// scope for the specified phase.
func GetDockerOptionsForPhase(appName string, phase string) ([]string, error) {
	return GetDockerOptionsForProcessPhase(appName, DefaultProcessType, phase)
}

// GetDockerOptionsForProcessPhase returns the docker options stored under the
// given process-type scope for the specified phase. An empty processType is
// treated as the default scope.
func GetDockerOptionsForProcessPhase(appName, processType, phase string) ([]string, error) {
	options, err := common.PropertyListGet("docker-options", appName, propertyKey(processType, phase))
	if err != nil {
		return nil, fmt.Errorf("Unable to read docker options for %s.%s.%s: %s", appName, processType, phase, err.Error())
	}

	trimmed := make([]string, 0, len(options))
	for _, option := range options {
		option = strings.TrimSpace(option)
		if option == "" {
			continue
		}
		trimmed = append(trimmed, option)
	}
	return trimmed, nil
}

// RemoveDockerOptionFromPhases removes an option from the default scope for the specified phases.
func RemoveDockerOptionFromPhases(appName string, phases []string, option string) error {
	return RemoveDockerOptionFromProcessPhases(appName, []string{DefaultProcessType}, phases, option)
}

// RemoveDockerOptionFromProcessPhases removes an option from the specified process types and phases.
func RemoveDockerOptionFromProcessPhases(appName string, processTypes []string, phases []string, option string) error {
	if len(processTypes) == 0 {
		processTypes = []string{DefaultProcessType}
	}
	for _, processType := range processTypes {
		for _, phase := range phases {
			options, err := GetDockerOptionsForProcessPhase(appName, processType, phase)
			if err != nil {
				return err
			}

			newOptions := []string{}
			for _, opt := range options {
				if opt != option {
					newOptions = append(newOptions, opt)
				}
			}

			sort.Strings(newOptions)
			if err := writeDockerOptionsForProcessPhase(appName, processType, phase, newOptions); err != nil {
				return err
			}
		}
	}
	return nil
}

// GetSpecifiedDockerOptionsForPhase returns the docker options for the specified
// phase (default scope) that are in the desiredOptions list. It expects
// desiredOptions entries in the form "--option" and matches against options
// stored as "--option", "--option=value", or "--option value".
func GetSpecifiedDockerOptionsForPhase(appName string, phase string, desiredOptions []string) (map[string][]string, error) {
	foundOptions := map[string][]string{}
	options, err := GetDockerOptionsForPhase(appName, phase)
	if err != nil {
		return foundOptions, err
	}

	for _, option := range options {
		for _, desiredOption := range desiredOptions {
			if option == desiredOption {
				foundOptions[desiredOption] = []string{}
				break
			}

			if strings.HasPrefix(option, fmt.Sprintf("%s=", desiredOption)) {
				if _, ok := foundOptions[desiredOption]; !ok {
					foundOptions[desiredOption] = []string{}
				}

				parts := strings.SplitN(option, "=", 2)
				if len(parts) != 2 {
					common.LogWarn(fmt.Sprintf("Invalid docker option found for %s: %s", appName, option))
					continue
				}

				foundOptions[desiredOption] = append(foundOptions[desiredOption], parts[1])
				break
			}

			if strings.HasPrefix(option, fmt.Sprintf("%s ", desiredOption)) {
				if _, ok := foundOptions[desiredOption]; !ok {
					foundOptions[desiredOption] = []string{}
				}

				parts := strings.SplitN(option, " ", 2)
				if len(parts) != 2 {
					common.LogWarn(fmt.Sprintf("Invalid docker option found for %s: %s", appName, option))
					continue
				}

				foundOptions[desiredOption] = append(foundOptions[desiredOption], parts[1])
				break
			}
		}
	}

	return foundOptions, nil
}

// ListProcessTypesWithOptions returns the sorted list of process types that
// have at least one option configured, excluding DefaultProcessType.
func ListProcessTypesWithOptions(appName string) ([]string, error) {
	properties, err := common.PropertyGetAll("docker-options", appName)
	if err != nil {
		return nil, err
	}

	seen := map[string]bool{}
	for key := range properties {
		processType, _, ok := splitPropertyKey(key)
		if !ok {
			continue
		}
		if processType == DefaultProcessType {
			continue
		}
		seen[processType] = true
	}

	processTypes := make([]string, 0, len(seen))
	for processType := range seen {
		processTypes = append(processTypes, processType)
	}
	sort.Strings(processTypes)
	return processTypes, nil
}

func splitPropertyKey(key string) (processType, phase string, ok bool) {
	idx := strings.LastIndex(key, ".")
	if idx <= 0 || idx == len(key)-1 {
		return "", "", false
	}
	processType = key[:idx]
	phase = key[idx+1:]
	if !isValidPhase(phase) {
		return "", "", false
	}
	return processType, phase, true
}

func writeDockerOptionsForProcessPhase(appName, processType, phase string, options []string) error {
	return common.PropertyListWrite("docker-options", appName, propertyKey(processType, phase), options)
}
