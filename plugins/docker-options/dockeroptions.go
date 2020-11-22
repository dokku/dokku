package dockeroptions

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

// SetDockerOptionForPhases sets an option to specified phases
func SetDockerOptionForPhases(appName string, phases []string, name string, value string) error {
	for _, phase := range phases {
		if err := touchPhaseFile(appName, phase); err != nil {
			return err
		}

		options, err := GetDockerOptionsForPhase(appName, phase)
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
		if err = writeDockerOptionsForPhase(appName, phase, newOptions); err != nil {
			return err
		}
	}
	return nil
}

// AddDockerOptionToPhases adds docker option to specified phases
func AddDockerOptionToPhases(appName string, phases []string, option string) error {
	for _, phase := range phases {
		if err := touchPhaseFile(appName, phase); err != nil {
			return err
		}

		options, err := GetDockerOptionsForPhase(appName, phase)
		if err != nil {
			return err
		}

		options = append(options, option)
		sort.Strings(options)
		if err = writeDockerOptionsForPhase(appName, phase, options); err != nil {
			return err
		}
	}
	return nil
}

// GetDockerOptionsForPhase returns the docker options for the specified phase
func GetDockerOptionsForPhase(appName string, phase string) ([]string, error) {
	options := []string{}

	if err := touchPhaseFile(appName, phase); err != nil {
		return options, err
	}

	phaseFilePath := getPhaseFilePath(appName, phase)
	file, err := os.Open(phaseFilePath)
	if err != nil {
		return options, fmt.Errorf("Unable to open docker options phase file %s.%s: %s", appName, appName, err.Error())
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		options = append(options, strings.TrimSpace(scanner.Text()))
	}

	if err = scanner.Err(); err != nil {
		return options, fmt.Errorf("Unable to read docker options phase file %s.%s: %s", appName, appName, err.Error())
	}

	return options, nil
}
