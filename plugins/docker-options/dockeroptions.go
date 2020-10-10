package dockeroptions

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

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

func getPhaseFilePath(appName string, phase string) string {
	return filepath.Join(common.MustGetEnv("DOKKU_ROOT"), appName, "DOCKER_OPTIONS_"+strings.ToUpper(phase))
}

func touchPhaseFile(appName string, phase string) error {
	phaseFilePath := getPhaseFilePath(appName, phase)

	_, err := os.Stat(phaseFilePath)
	if !os.IsNotExist(err) {
		return nil
	}

	file, err := os.Create(phaseFilePath)
	if err != nil {
		return fmt.Errorf("Unable to create docker options phase file %s.%s: %s", appName, phase, err.Error())
	}
	defer file.Close()

	return nil
}

func writeDockerOptionsForPhase(appName string, phase string, options []string) error {
	phaseFilePath := getPhaseFilePath(appName, phase)
	file, err := os.OpenFile(phaseFilePath, os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	w := bufio.NewWriter(file)
	for _, option := range options {
		fmt.Fprintln(w, option)
	}
	if err = w.Flush(); err != nil {
		return fmt.Errorf("Unable to update docker options phase file %s.%s: %s", appName, phase, err.Error())
	}

	file.Chmod(0600)
	common.SetPermissions(phaseFilePath, 0600)

	return nil
}
