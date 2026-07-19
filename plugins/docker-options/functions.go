package dockeroptions

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// CommandAdd adds a docker option to the specified phases for an app.
// When processes is empty the option is added to the default scope (apply
// to every container in the app); otherwise it is added to each named
// process type. The default and process flows are mutually exclusive
// because process scoping is only valid for the deploy phase.
//
// The option string is split on flag boundaries via SplitOptionString so a
// single invocation may carry multiple flags (e.g. `--build-arg X=Y --link a
// --link b`); each split flag becomes a separate stored option. A misplaced
// `--process` inside the option content is lifted into the process slice
// rather than stored as a docker option.
func CommandAdd(appName string, processes []string, phasesArg string, option string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	phases, err := parsePhases(phasesArg)
	if err != nil {
		return err
	}

	options, extractedProcesses, err := SplitOptionString(option)
	if err != nil {
		return err
	}

	if len(options) == 0 {
		return errors.New("Please specify docker options to add to the phase")
	}

	processes = dedupeProcesses(append(processes, extractedProcesses...))

	if err := ValidateProcessFlag(processes, phases); err != nil {
		return err
	}

	for _, processType := range processes {
		WarnIfProcessNotInProcfile(appName, processType)
	}

	for _, opt := range options {
		if len(processes) == 0 {
			if err := AddDockerOptionToPhases(appName, phases, opt); err != nil {
				return err
			}
			continue
		}
		if err := AddDockerOptionToProcessPhases(appName, processes, phases, opt); err != nil {
			return err
		}
	}

	return nil
}

// CommandRemove removes a docker option from the specified phases for an app.
// Process-flag handling matches CommandAdd, including splitting a multi-flag
// option string and lifting a misplaced `--process` into the process slice.
func CommandRemove(appName string, processes []string, phasesArg string, option string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	phases, err := parsePhases(phasesArg)
	if err != nil {
		return err
	}

	options, extractedProcesses, err := SplitOptionString(option)
	if err != nil {
		return err
	}

	if len(options) == 0 {
		return errors.New("Please specify docker options to remove from the phase")
	}

	processes = dedupeProcesses(append(processes, extractedProcesses...))

	if err := ValidateProcessFlag(processes, phases); err != nil {
		return err
	}

	for _, opt := range options {
		if len(processes) == 0 {
			if err := RemoveDockerOptionFromPhases(appName, phases, opt); err != nil {
				return err
			}
			continue
		}
		if err := RemoveDockerOptionFromProcessPhases(appName, processes, phases, opt); err != nil {
			return err
		}
	}

	return nil
}

// dedupeProcesses preserves order while collapsing repeated process names,
// which can occur when --process is specified both before the app (captured
// by pflag) and after (lifted by SplitOptionString).
func dedupeProcesses(processes []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(processes))
	for _, p := range processes {
		if seen[p] {
			continue
		}
		seen[p] = true
		result = append(result, p)
	}
	return result
}

// CommandClear removes all docker options for an app, optionally limited to
// a list of phases and/or specific process types. With no flags it clears
// the default scope across all phases; with --process flags it clears each
// named process type for the supplied (deploy-only) phases.
func CommandClear(appName string, processes []string, phasesArg string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if len(processes) == 0 {
		return clearDefaultScope(appName, phasesArg)
	}

	phases, err := parsePhases(phasesArg)
	if err != nil {
		return err
	}

	if len(phases) == 0 {
		phases = []string{"deploy"}
	}

	if err := ValidateProcessFlag(processes, phases); err != nil {
		return err
	}

	for _, processType := range processes {
		for _, phase := range phases {
			common.LogInfo1(fmt.Sprintf("Clearing docker-options for %s on phase %s for process %s", appName, phase, processType))
			if err := common.PropertyDelete("docker-options", appName, propertyKey(processType, phase)); err != nil {
				return err
			}
		}
	}

	return nil
}

func clearDefaultScope(appName string, phasesArg string) error {
	if phasesArg == "" {
		common.LogInfo1(fmt.Sprintf("Clearing docker-options for %s on all phases", appName))
		for _, phase := range availablePhases {
			if err := common.PropertyDelete("docker-options", appName, propertyKey(DefaultProcessType, phase)); err != nil {
				return err
			}
		}
		return nil
	}

	phases, err := parsePhases(phasesArg)
	if err != nil {
		return err
	}

	for _, phase := range phases {
		common.LogInfo1(fmt.Sprintf("Clearing docker-options for %s on phase %s", appName, phase))
		if err := common.PropertyDelete("docker-options", appName, propertyKey(DefaultProcessType, phase)); err != nil {
			return err
		}
	}

	return nil
}

// CommandList prints the docker options stored for a given process+phase, one
// option per line. An empty processType means the default scope.
func CommandList(appName, processType, phase string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if phase == "" {
		return errors.New("--phase is required")
	}

	if !isValidPhase(phase) {
		return fmt.Errorf("Phase must be one of [%s]", strings.Join(availablePhases, " "))
	}

	if processType != "" {
		if processType == DefaultProcessType {
			return fmt.Errorf("%q is reserved and cannot be used as a --process value", DefaultProcessType)
		}
		if !processScopedPhases[phase] {
			return fmt.Errorf("--process is only supported for the deploy phase, got %q", phase)
		}
	}

	options, err := GetDockerOptionsForProcessPhase(appName, processType, phase)
	if err != nil {
		return err
	}

	for _, option := range options {
		fmt.Println(option)
	}
	return nil
}

// CommandReport displays a docker-options report for one or more apps
func CommandReport(appName string, format string, infoFlag string) error {
	if appName == "" {
		apps, err := common.DokkuApps()
		if err != nil {
			if errors.Is(err, common.NoAppsExist) {
				common.LogWarn(err.Error())
				return nil
			}
			return err
		}
		for _, name := range apps {
			if err := ReportSingleApp(name, format, infoFlag); err != nil {
				return err
			}
		}
		return nil
	}

	return ReportSingleApp(appName, format, infoFlag)
}

// migratedPropertyKey returns the per-app property name recording that
// an app's DOCKER_OPTIONS_<PHASE> file was drained into the property
// store. Surfaces in the property store as `docker-options/<app>/migrated-<phase>`.
func migratedPropertyKey(phase string) string {
	return "migrated-" + strings.ToLower(phase)
}

// convertLegacyMigratedMarker drains a leftover DOCKER_OPTIONS_<PHASE>.migrated
// sentinel into the new per-phase property and removes the file. Runs
// as part of the upgrade-cycle pass; no-op when the file is absent.
// TODO(post-deprecation): remove this helper and its caller.
func convertLegacyMigratedMarker(appName, phase string) error {
	migratedPath := filepath.Join(common.AppRoot(appName), "DOCKER_OPTIONS_"+strings.ToUpper(phase)+".migrated")
	if !common.FileExists(migratedPath) {
		return nil
	}
	if err := common.PropertyWrite("docker-options", appName, migratedPropertyKey(phase), "true"); err != nil {
		return err
	}
	return os.Remove(migratedPath)
}

// migrateLegacyDockerOptionsFiles converts pre-properties
// DOCKER_OPTIONS_* files into property lists. Idempotency comes from
// two layers: a per-phase per-app `migrated-<phase>` property recording
// what was drained, and a global `migrated-from-files` short-circuit so
// the steady-state install path returns immediately.
//
// The upgrade-cycle conversion of leftover `.migrated` filesystem
// sentinels runs BEFORE the global gate so users upgrading from the
// previous release - who have `migrated-from-files` set AND `.migrated`
// files on disk - still get those files drained into properties.
func migrateLegacyDockerOptionsFiles() error {
	if common.PropertyGet("docker-options", "--global", "migrated-from-files") == "true" {
		return nil
	}

	apps, err := common.DokkuApps()
	if err != nil {
		if errors.Is(err, common.NoAppsExist) {
			return common.PropertyWrite("docker-options", "--global", "migrated-from-files", "true")
		}
		return err
	}

	// Upgrade-cycle conversion: drain any `.migrated` sentinels left
	// behind by the previous release into the new per-phase properties.
	// Always runs; cheap when nothing to do.
	for _, appName := range apps {
		for _, phase := range availablePhases {
			if err := convertLegacyMigratedMarker(appName, phase); err != nil {
				return err
			}
		}
	}

	for _, appName := range apps {
		for _, phase := range availablePhases {
			legacyPath := filepath.Join(common.AppRoot(appName), "DOCKER_OPTIONS_"+strings.ToUpper(phase))
			if !common.FileExists(legacyPath) {
				continue
			}

			lines, err := readLegacyOptionsFile(legacyPath)
			if err != nil {
				return err
			}

			if len(lines) > 0 {
				if err := common.PropertyListWrite("docker-options", appName, propertyKey(DefaultProcessType, phase), lines); err != nil {
					return err
				}
				if err := common.PropertyWrite("docker-options", appName, migratedPropertyKey(phase), "true"); err != nil {
					return err
				}
				common.LogInfo1(fmt.Sprintf("Migrated %s to docker-options properties", legacyPath))
			}

			if err := os.Remove(legacyPath); err != nil {
				return fmt.Errorf("Unable to remove migrated file %s: %s", legacyPath, err.Error())
			}
		}
	}

	return common.PropertyWrite("docker-options", "--global", "migrated-from-files", "true")
}

func readLegacyOptionsFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Unable to read %s: %s", path, err.Error())
	}

	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		lines = append(lines, trimmed)
	}
	return lines, nil
}

// traefikLabelMigrationKey gates the one-time pass that repairs docker-option
// labels whose backticks were stored with a stray leading backslash by an
// earlier release's option tokenizer (e.g. Traefik rules like Host(\`app\`)).
const traefikLabelMigrationKey = "migrated-traefik-backticks"

// migrateTraefikLabelBackticks repairs stored Traefik label options whose
// backticks carry a stray backslash (Host(\`app\`) instead of Host(`app`)) so
// they become valid on the next deploy. Only Traefik --label options (those
// whose label key begins with "traefik.") are considered, and only when they
// contain a backslash-escaped backtick, which correct storage never produces.
// It runs once, guarded by a global property.
func migrateTraefikLabelBackticks() error {
	if common.PropertyGet("docker-options", "--global", traefikLabelMigrationKey) == "true" {
		return nil
	}

	apps, err := common.DokkuApps()
	if err != nil {
		if errors.Is(err, common.NoAppsExist) {
			return common.PropertyWrite("docker-options", "--global", traefikLabelMigrationKey, "true")
		}
		return err
	}

	for _, appName := range apps {
		properties, err := common.PropertyGetAll("docker-options", appName)
		if err != nil {
			return err
		}

		for key := range properties {
			processType, phase, ok := splitPropertyKey(key)
			if !ok {
				continue
			}

			options, err := GetDockerOptionsForProcessPhase(appName, processType, phase)
			if err != nil {
				return err
			}

			changed := false
			for i, option := range options {
				if fixed, ok := repairTraefikLabelBackticks(option); ok {
					options[i] = fixed
					changed = true
				}
			}

			if changed {
				if err := writeDockerOptionsForProcessPhase(appName, processType, phase, options); err != nil {
					return err
				}
				common.LogInfo1(fmt.Sprintf("Repaired docker-options label backticks for %s (%s %s)", appName, processType, phase))
			}
		}
	}

	return common.PropertyWrite("docker-options", "--global", traefikLabelMigrationKey, "true")
}

// repairTraefikLabelBackticks removes the stray backslash before a backtick in
// a Traefik --label option (one whose label key begins with "traefik."),
// returning the repaired option and whether anything changed. Backticks are
// Traefik's rule syntax, so the repair is scoped to Traefik labels; non-label
// options, non-Traefik labels, and labels without a backslash-escaped backtick
// are left untouched so an intentional backslash elsewhere is never rewritten.
func repairTraefikLabelBackticks(option string) (string, bool) {
	spec, ok := labelSpec(option)
	if !ok || !strings.HasPrefix(spec, "traefik.") || !strings.Contains(option, "\\`") {
		return option, false
	}
	return strings.ReplaceAll(option, "\\`", "`"), true
}

// labelSpec returns the "key=value" portion of a stored --label/-l option with
// the flag and any surrounding shell-quoting removed so the label key can be
// inspected. It reports ok=false for options that do not set a docker label.
func labelSpec(option string) (string, bool) {
	s := strings.TrimLeft(option, " ")
	if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
		s = s[1 : len(s)-1]
	}
	switch {
	case strings.HasPrefix(s, "--label="):
		s = s[len("--label="):]
	case strings.HasPrefix(s, "--label "):
		s = s[len("--label "):]
	case strings.HasPrefix(s, "-l="):
		s = s[len("-l="):]
	case strings.HasPrefix(s, "-l "):
		s = s[len("-l "):]
	default:
		return "", false
	}
	s = strings.TrimPrefix(s, "'")
	s = strings.TrimPrefix(s, "\"")
	return s, true
}
