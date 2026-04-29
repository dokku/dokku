package dockeroptions

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// TriggerInstall sets up the docker-options property directory and migrates
// any pre-existing DOCKER_OPTIONS_* files into property lists.
func TriggerInstall() error {
	if err := common.PropertySetup("docker-options"); err != nil {
		return fmt.Errorf("Unable to install the docker-options plugin: %v", err)
	}

	if err := migrateLegacyDockerOptionsFiles(); err != nil {
		return fmt.Errorf("Unable to migrate legacy docker-options files: %v", err)
	}

	return nil
}

// TriggerPostAppCloneSetup copies docker option properties from the source app
// to the cloned app.
func TriggerPostAppCloneSetup(oldAppName string, newAppName string) error {
	return common.PropertyClone("docker-options", oldAppName, newAppName)
}

// TriggerPostAppRenameSetup moves docker option properties from the old app
// name to the new one.
func TriggerPostAppRenameSetup(oldAppName string, newAppName string) error {
	if err := common.PropertyClone("docker-options", oldAppName, newAppName); err != nil {
		return err
	}
	return common.PropertyDestroy("docker-options", oldAppName)
}

// TriggerPostDelete removes the docker option properties for an app and
// cleans up any leftover migrated legacy files.
func TriggerPostDelete(appName string) error {
	if err := common.PropertyDestroy("docker-options", appName); err != nil {
		return err
	}
	removeMigratedLegacyFiles(appName)
	return nil
}

// TriggerDockerArgs implements the legacy docker-args-{build,deploy,run}
// triggers. It echoes stdin verbatim, then appends the default-scope options
// for the given phase after applying image-source-type filtering.
func TriggerDockerArgs(phase, appName, imageSourceType, processType string) error {
	stdin, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	if _, err := os.Stdout.Write(stdin); err != nil {
		return err
	}

	options, err := GetDockerOptionsForProcessPhase(appName, DefaultProcessType, phase)
	if err != nil {
		return err
	}

	emitFilteredOptions(phase, imageSourceType, options)
	return nil
}

// TriggerDockerArgsProcessDeploy implements docker-args-process-deploy. It
// echoes stdin verbatim, then appends the deploy-phase options for the
// process-specific scope (filtered by image source type). Default-scope
// options are emitted by TriggerDockerArgs (docker-args-deploy), which the
// scheduler invokes alongside the per-process trigger.
func TriggerDockerArgsProcessDeploy(appName, imageSourceType, processType string) error {
	stdin, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	if _, err := os.Stdout.Write(stdin); err != nil {
		return err
	}

	if processType == "" || processType == DefaultProcessType {
		return nil
	}

	options, err := GetDockerOptionsForProcessPhase(appName, processType, "deploy")
	if err != nil {
		return err
	}

	emitFilteredOptions("deploy", imageSourceType, options)
	return nil
}

// emitFilteredOptions writes filtered docker options to stdout, prefixed with
// a single space so concatenation with prior output stays well-formed. It
// reproduces the filtering historically performed by the bash docker-args-*
// triggers.
func emitFilteredOptions(phase, imageSourceType string, options []string) {
	for _, option := range options {
		option = strings.TrimSpace(option)
		if option == "" || strings.HasPrefix(option, "#") {
			continue
		}

		if strings.HasPrefix(option, "--restart") {
			if phase == "deploy" {
				fmt.Printf(" %s", option)
			}
			continue
		}

		switch imageSourceType {
		case "dockerfile", "nixpacks", "railpack":
			if hasAnyPrefix(option, "--link", "-v", "--volume") {
				continue
			}
		case "herokuish":
			if hasAnyPrefix(option, "--file", "--build-args") {
				continue
			}
		}

		fmt.Printf(" %s", option)
	}
}

func hasAnyPrefix(s string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if s == prefix || strings.HasPrefix(s, prefix+"=") || strings.HasPrefix(s, prefix+" ") {
			return true
		}
	}
	return false
}
