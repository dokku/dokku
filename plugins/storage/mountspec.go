package storage

import (
	"fmt"
	"strings"
)

// The key=value tokens a storage:mount --replace option list recognizes. Each
// one mirrors the flag the single-entry form uses, so every mount-time field
// has exactly one spelling across the two forms. Readonly keeps its docker
// spelling - the bare "ro" token this grammar already had, plus "rw" as its
// explicit opposite - rather than gaining a key of its own.
const (
	mountSpecKeyPhase       = "phase"
	mountSpecKeySubpath     = "volume-subpath"
	mountSpecKeyVolumeChown = "volume-chown"
)

// mountSpec is one parsed positional argument of the storage:mount --replace
// form. Every mount-time field an attachment carries except ProcessType is
// expressible per spec, so one call can vary them between mounts.
type mountSpec struct {
	EntryName     string
	ContainerPath string
	Phases        []string
	Subpath       string
	Readonly      bool
	VolumeOptions string
	VolumeChown   string
}

// parseMountSpec splits an "<entry>:<container-dir>[:<options>]" argument,
// where the first field names a storage entry rather than a host path and the
// option list carries every mount-time field rather than only the docker ones.
//
// It deliberately does not route through ParseMountPath, which hoists "ro" out
// of the list before any caller sees it. Keeping the whole grammar here leaves
// the ro/rw pair as a single rule in a single place, lets an error quote the
// token list the operator actually typed, and keeps the legacy colon form,
// storage:unmount and the migration on the grammar they already had.
func parseMountSpec(spec string) (mountSpec, error) {
	parts := strings.SplitN(spec, ":", 3)
	parsed := mountSpec{Phases: []string{PhaseDeploy, PhaseRun}}
	if len(parts) >= 1 {
		parsed.EntryName = parts[0]
	}
	if len(parts) >= 2 {
		parsed.ContainerPath = parts[1]
	}

	if parsed.EntryName == "" || parsed.ContainerPath == "" {
		return mountSpec{}, fmt.Errorf("Invalid mount specified: %s", spec)
	}

	// A trailing colon is the no-options case rather than an empty option.
	if len(parts) >= 3 && parts[2] != "" {
		if err := parseMountSpecOptions(&parsed, spec, parts[2]); err != nil {
			return mountSpec{}, err
		}
	}

	return parsed, nil
}

// parseMountSpecOptions consumes a spec's comma-separated option list, lifting
// the recognized tokens into their mount-time fields and rejoining what is left
// into VolumeOptions the way the colon form always has.
//
// Order within the list is not significant, so anything the list can say twice
// is rejected rather than resolved by position, and an unrecognized key is
// rejected rather than passed through - a typo that reaches VolumeOptions would
// otherwise surface only when docker refuses the container at deploy time.
func parseMountSpecOptions(parsed *mountSpec, spec string, options string) error {
	sawReadonly := false
	sawWritable := false
	seenKeys := map[string]bool{}
	seenPhases := map[string]bool{}
	remaining := []string{}

	for _, token := range strings.Split(options, ",") {
		if token == "" {
			return fmt.Errorf("Mount spec %q has an empty mount option", spec)
		}

		switch token {
		case "ro":
			sawReadonly = true
			continue
		case "rw":
			sawWritable = true
			continue
		}

		// A bare docker option carries no "=", and neither does a token
		// starting with one, which is no key at all. Both pass through.
		index := strings.Index(token, "=")
		if index <= 0 {
			remaining = append(remaining, token)
			continue
		}

		key, value := token[:index], token[index+1:]
		if key != mountSpecKeyPhase && key != mountSpecKeySubpath && key != mountSpecKeyVolumeChown {
			return fmt.Errorf("Mount spec %q specifies unknown key %q; supported keys are %s, %s and %s",
				spec, key, mountSpecKeyPhase, mountSpecKeyVolumeChown, mountSpecKeySubpath)
		}

		if value == "" {
			return fmt.Errorf("Mount spec %q has an empty value for %s", spec, key)
		}

		// A phase names one of a pair and so appears once per phase; every
		// other key names a single value and so appears at most once.
		if key != mountSpecKeyPhase {
			if seenKeys[key] {
				return fmt.Errorf("Mount spec %q specifies %s more than once", spec, key)
			}
			seenKeys[key] = true
		}

		switch key {
		case mountSpecKeyPhase:
			if value != PhaseDeploy && value != PhaseRun {
				return fmt.Errorf("Mount spec %q specifies unsupported phase %q (must be %q or %q)", spec, value, PhaseDeploy, PhaseRun)
			}
			if seenPhases[value] {
				return fmt.Errorf("Mount spec %q specifies phase %s more than once", spec, value)
			}
			seenPhases[value] = true
		case mountSpecKeySubpath:
			parsed.Subpath = value
		case mountSpecKeyVolumeChown:
			if err := ValidateChownOption(value); err != nil {
				return fmt.Errorf("Mount spec %q has an invalid %s value %q: %w", spec, key, value, err)
			}
			parsed.VolumeChown = value
		}
	}

	if sawReadonly && sawWritable {
		return fmt.Errorf("Mount spec %q sets both ro and rw", spec)
	}

	parsed.Readonly = sawReadonly
	parsed.VolumeOptions = strings.Join(remaining, ",")
	if len(seenPhases) > 0 {
		parsed.Phases = canonicalPhases(seenPhases)
	}

	return nil
}

// canonicalPhases renders a phase set in a fixed order, so two spec strings
// naming the same phases in a different order store identical JSON and render
// identically in storage:report.
func canonicalPhases(seen map[string]bool) []string {
	phases := []string{}
	for _, phase := range []string{PhaseDeploy, PhaseRun} {
		if seen[phase] {
			phases = append(phases, phase)
		}
	}
	return phases
}
