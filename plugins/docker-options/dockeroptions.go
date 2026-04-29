package dockeroptions

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"mvdan.cc/sh/v3/shell"
)

// DefaultProcessType is the sentinel process-type key used for options that
// apply to every container in an app (i.e. options not scoped to a specific
// Procfile process type).
const DefaultProcessType = "_default_"

// SplitOptionString shell-tokenizes input, groups tokens on flag boundaries,
// and returns one re-serialized option per group. A docker-options subcommand
// flag (currently just --process) that lands inside the option content -
// because the user typed it after the app name, where pflag's
// SetInterspersed(false) hands it back as positional - is lifted into the
// returned processes slice rather than stored as a docker option. The caller
// merges those processes with whatever pflag already captured. Empty or
// whitespace-only input returns empty slices.
func SplitOptionString(input string) (options []string, processes []string, err error) {
	if strings.TrimSpace(input) == "" {
		return nil, nil, nil
	}

	fields, err := shell.Fields(input, func(string) string { return "" })
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to parse docker option: %s", err.Error())
	}

	var current []string
	flush := func() error {
		if len(current) == 0 {
			return nil
		}
		head := current[0]
		if head == "--process" {
			if len(current) < 2 {
				return fmt.Errorf("--process requires a value")
			}
			if len(current) > 2 {
				return fmt.Errorf("--process accepts a single value, got %d", len(current)-1)
			}
			processes = append(processes, current[1])
			current = nil
			return nil
		}
		if strings.HasPrefix(head, "--process=") {
			if len(current) > 1 {
				return fmt.Errorf("--process=value cannot be followed by additional tokens")
			}
			processes = append(processes, head[len("--process="):])
			current = nil
			return nil
		}
		options = append(options, joinShellTokens(current))
		current = nil
		return nil
	}

	for _, tok := range fields {
		if isFlagToken(tok) && len(current) > 0 {
			if err := flush(); err != nil {
				return nil, nil, err
			}
		}
		current = append(current, tok)
	}
	if err := flush(); err != nil {
		return nil, nil, err
	}

	return options, processes, nil
}

// isFlagToken reports whether tok looks like a CLI flag (long or short) rather
// than a value. Treating any token that begins with `-` and has more than one
// character as a flag matches docker's flag conventions and avoids the need
// for a per-flag whitelist.
func isFlagToken(tok string) bool {
	return len(tok) > 1 && tok[0] == '-'
}

// joinShellTokens joins tokens into a single string suitable for storage,
// shell-quoting any token that contains characters the bash scheduler's
// `eval` re-tokenization would interpret. The stored line round-trips through
// `eval set -- "$line"` back to the original token slice.
func joinShellTokens(tokens []string) string {
	parts := make([]string, len(tokens))
	for i, tok := range tokens {
		parts[i] = quoteShellArg(tok)
	}
	return strings.Join(parts, " ")
}

// quoteShellArg returns s wrapped in single quotes when it contains characters
// the shell would otherwise interpret (whitespace, quotes, expansion sigils,
// globs, redirections, etc.). Embedded single quotes are escaped with the
// standard `'\''` close-escape-open sequence. Tokens free of such characters
// are returned verbatim so the stored representation stays human-readable for
// the common case.
func quoteShellArg(s string) string {
	if s == "" {
		return "''"
	}
	if !needsShellQuoting(s) {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func needsShellQuoting(s string) bool {
	for _, r := range s {
		switch r {
		case ' ', '\t', '\n', '"', '\'', '$', '`', '\\',
			'*', '?', '[', ']', '<', '>', '|', '&', ';',
			'(', ')', '{', '}', '!', '#', '~':
			return true
		}
	}
	return false
}

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
