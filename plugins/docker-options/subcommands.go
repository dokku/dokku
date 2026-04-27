package dockeroptions

import (
	"errors"
	"fmt"

	"github.com/dokku/dokku/plugins/common"
)

// CommandAdd adds a docker option to the specified phases for an app
func CommandAdd(appName string, phasesArg string, option string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	phases, err := parsePhases(phasesArg)
	if err != nil {
		return err
	}

	if option == "" {
		return errors.New("Please specify docker options to add to the phase")
	}

	return AddDockerOptionToPhases(appName, phases, option)
}

// CommandRemove removes a docker option from the specified phases for an app
func CommandRemove(appName string, phasesArg string, option string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	phases, err := parsePhases(phasesArg)
	if err != nil {
		return err
	}

	if option == "" {
		return errors.New("Please specify docker options to remove from the phase")
	}

	return RemoveDockerOptionFromPhases(appName, phases, option)
}

// CommandClear removes all docker options for an app, optionally limited to a list of phases
func CommandClear(appName string, phasesArg string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if phasesArg == "" {
		common.LogInfo1(fmt.Sprintf("Clearing docker-options for %s on all phases", appName))
		for _, phase := range availablePhases {
			if err := removePhaseFile(appName, phase); err != nil {
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
		if err := removePhaseFile(appName, phase); err != nil {
			return err
		}
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
