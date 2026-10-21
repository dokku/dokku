package common

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// memoryValueRegex matches a docker-compatible memory value: a number with an
// optional unit (b, k, m, g, t, p in either case) optionally followed by i and/or b
var memoryValueRegex = regexp.MustCompile(`^(\d+(?:\.\d+)?)(?:([kKmMgGtTpP])[iI]?)?([bB])?$`)

// IsValidMemoryValue returns true if the value is a docker-compatible memory value
func IsValidMemoryValue(value string) bool {
	return memoryValueRegex.MatchString(value)
}

// IsUnitlessMemoryValue returns true if the value is a valid memory value without any unit suffix
func IsUnitlessMemoryValue(value string) bool {
	matches := memoryValueRegex.FindStringSubmatch(value)
	if matches == nil {
		return false
	}

	return matches[2] == "" && matches[3] == ""
}

// ParseMemoryBytes converts a docker-compatible memory value into bytes.
// Units are binary, and a value without a unit is treated as megabytes
func ParseMemoryBytes(value string) (int64, error) {
	matches := memoryValueRegex.FindStringSubmatch(value)
	if matches == nil {
		return 0, fmt.Errorf("Invalid memory value '%s'", value)
	}

	size, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, fmt.Errorf("Invalid memory value '%s': %w", value, err)
	}

	multipliers := map[string]float64{
		"k": 1 << 10,
		"m": 1 << 20,
		"g": 1 << 30,
		"t": 1 << 40,
		"p": 1 << 50,
	}

	unit := strings.ToLower(matches[2])
	if unit == "" && matches[3] == "" {
		unit = "m"
	}

	multiplier := float64(1)
	if unit != "" {
		multiplier = multipliers[unit]
	}

	return int64(size * multiplier), nil
}
