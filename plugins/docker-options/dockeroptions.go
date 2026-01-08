package dockeroptions

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/dokku/dokku/plugins/common"
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

// RemoveDockerOptionFromPhases removes a docker option from specified phases
func RemoveDockerOptionFromPhases(appName string, phases []string, option string) error {
	for _, phase := range phases {
		options, err := GetDockerOptionsForPhase(appName, phase)
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
		if err = writeDockerOptionsForPhase(appName, phase, newOptions); err != nil {
			return err
		}
	}
	return nil
}

// GetSpecifiedDockerOptionsForPhase returns the docker options for the specified phase that are in the desiredOptions list
// It expects desiredOptions to be a list of docker options that are in the format "--option"
// And will retrieve any lines that start with the desired option
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

			// match options that are in the format "--option=value"
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

			// match options that are in the format "--option value"
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
