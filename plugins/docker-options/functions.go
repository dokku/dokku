package dockeroptions

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

func getPhaseFilePath(appName string, phase string) string {
	return filepath.Join(common.AppRoot(appName), "DOCKER_OPTIONS_"+strings.ToUpper(phase))
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
	return common.WriteSliceToFile(phaseFilePath, options)
}
