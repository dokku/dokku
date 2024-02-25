package ports

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// CommandList is a cmd wrapper to list port mappings for an app
func CommandList(appName string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	return listAppPortMaps(appName)
}

// CommandAdd adds port mappings to an app
func CommandAdd(appName string, portMapStrings []string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if len(portMapStrings) == 0 {
		return errors.New("No port mapping specified")
	}

	portMaps, err := parsePortMapString(strings.Join(portMapStrings, " "))
	if err != nil {
		return err
	}

	if err := reusesSchemeHostPort(portMaps); err != nil {
		return fmt.Errorf("Error validating new port mappings: %s", err)
	}

	existingPortMaps := getPortMaps(appName)
	allPortMaps := append(existingPortMaps, portMaps...)
	if err := reusesSchemeHostPort(allPortMaps); err != nil {
		return fmt.Errorf("Error validating all port mappings: %s", err)
	}

	if err := addPortMaps(appName, portMaps); err != nil {
		return err
	}

	return common.PlugnTrigger("post-proxy-ports-update", []string{appName, "add"}...)
}

// CommandClear clears all port mappings for an app
func CommandClear(appName string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if err := clearPorts(appName); err != nil {
		return err
	}

	return common.PlugnTrigger("post-proxy-ports-update", []string{appName, "clear"}...)
}

// CommandRemove removes specific port mappings from an app
func CommandRemove(appName string, portMapStrings []string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if len(portMapStrings) == 0 {
		return errors.New("No port mapping specified")
	}

	portMaps, err := parsePortMapString(strings.Join(portMapStrings, " "))
	if err != nil {
		return err
	}

	if err := removePortMaps(appName, portMaps); err != nil {
		return err
	}

	return common.PlugnTrigger("post-proxy-ports-update", []string{appName, "remove"}...)
}

// CommandSet sets port mappings for an app
func CommandSet(appName string, portMapStrings []string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if len(portMapStrings) == 0 {
		return errors.New("No port mapping specified")
	}

	portMaps, err := parsePortMapString(strings.Join(portMapStrings, " "))
	if err != nil {
		return err
	}

	if err := reusesSchemeHostPort(portMaps); err != nil {
		return fmt.Errorf("Error validating port mappings: %s", err)
	}

	if err := setPortMaps(appName, portMaps); err != nil {
		return err
	}

	return common.PlugnTrigger("post-proxy-ports-update", []string{appName, "set"}...)
}

// CommandReport displays a ports report for one or more apps
func CommandReport(appName string, format string, infoFlag string) error {
	if len(appName) == 0 {
		apps, err := common.DokkuApps()
		if err != nil {
			return err
		}
		for _, appName := range apps {
			if err := ReportSingleApp(appName, format, infoFlag); err != nil {
				return err
			}
		}
		return nil
	}

	return ReportSingleApp(appName, format, infoFlag)
}
