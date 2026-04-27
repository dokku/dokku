package dockeroptions

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

func getPhaseFilePath(appName string, phase string) string {
	return filepath.Join(common.AppRoot(appName), "DOCKER_OPTIONS_"+strings.ToUpper(phase))
}

func copyPhaseFile(srcApp string, dstApp string, phase string) error {
	srcPath := getPhaseFilePath(srcApp, phase)
	if !common.FileExists(srcPath) {
		return nil
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("Unable to open docker options phase file %s.%s: %s", srcApp, phase, err.Error())
	}
	defer src.Close()

	dstPath := getPhaseFilePath(dstApp, phase)
	dst, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("Unable to create docker options phase file %s.%s: %s", dstApp, phase, err.Error())
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("Unable to copy docker options phase file %s.%s to %s.%s: %s", srcApp, phase, dstApp, phase, err.Error())
	}

	return nil
}

func removePhaseFile(appName string, phase string) error {
	if err := os.Remove(getPhaseFilePath(appName, phase)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
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
	return common.WriteSliceToFile(common.WriteSliceToFileInput{
		Filename: phaseFilePath,
		Lines:    options,
		Mode:     os.FileMode(0600),
	})
}
