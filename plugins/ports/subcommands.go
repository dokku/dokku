package ports

import (
	"errors"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
)

// CommandList is a cmd wrapper to list port mappings for an app
func CommandList(appName string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	return listAppPorts(appName)
}

// CommandAdd adds port mappings to an app
func CommandAdd(appName string, portMaps []string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if len(portMaps) == 0 {
		return errors.New("No port mapping specified")
	}

	portMap, err := parsePortMapString(strings.Join(portMaps, " "))
	if err != nil {
		return err
	}

	if err := addPorts(appName, portMap); err != nil {
		return err
	}

	return common.PlugnTrigger("post-ports-update", []string{appName, "add"}...)
}

// CommandClear clears all port mappings for an app
func CommandClear(appName string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	keys := []string{"DOKKU_PROXY_PORT_MAP"}
	if err := config.UnsetMany(appName, keys, false); err != nil {
		return err
	}

	return common.PlugnTrigger("post-ports-update", []string{appName, "clear"}...)
}

// CommandRemove removes specific port mappings from an app
func CommandRemove(appName string, portMaps []string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if len(portMaps) == 0 {
		return errors.New("No port mapping specified")
	}

	portMap, err := parsePortMapString(strings.Join(portMaps, " "))
	if err != nil {
		return err
	}

	if err := removePorts(appName, portMap); err != nil {
		return err
	}

	return common.PlugnTrigger("post-ports-update", []string{appName, "remove"}...)
}

// CommandSet sets port mappings for an app
func CommandSet(appName string, portMaps []string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if len(portMaps) == 0 {
		return errors.New("No port mapping specified")
	}

	portMap, err := parsePortMapString(strings.Join(portMaps, " "))
	if err != nil {
		return err
	}

	if err := setPorts(appName, portMap); err != nil {
		return err
	}

	return common.PlugnTrigger("post-ports-update", []string{appName, "set"}...)
}

// CommandReport displays a report for one or more apps
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
