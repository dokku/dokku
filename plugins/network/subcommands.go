package network

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// CommandCreate is an alias for "docker network create"
func CommandCreate(networkName string) error {
	common.LogInfo1Quiet(fmt.Sprintf("Creating network %v", networkName))
	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command: common.DockerBin(),
		Args:    []string{"network", "create", "--attachable", "--label", fmt.Sprintf("com.dokku.network-name=%v", networkName), networkName},
	})
	if err != nil {
		return fmt.Errorf("Unable to create network: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Unable to create network: %s", result.StderrContents())
	}

	return err
}

// CommandDestroy is an alias for "docker network rm"
func CommandDestroy(networkName string, forceDestroy bool) error {
	if os.Getenv("DOKKU_APPS_FORCE_DELETE") == "1" {
		forceDestroy = true
	}

	if !forceDestroy {
		if err := common.AskForDestructiveConfirmation(networkName, "network"); err != nil {
			return err
		}
	}

	common.LogInfo1Quiet(fmt.Sprintf("Destroying network %v", networkName))
	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command: common.DockerBin(),
		Args:    []string{"network", "rm", networkName},
	})
	if err != nil {
		return fmt.Errorf("Unable to destroy network: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Unable to destroy network: %s", result.StderrContents())
	}

	return nil
}

// CommandExists checks if a network exists
func CommandExists(networkName string) error {
	if networkName == "" {
		return errors.New("No network name specified")
	}

	exists, err := networkExists(networkName)
	if err != nil {
		return err
	}

	if exists {
		common.LogInfo1Quiet("Network exists")
	} else {
		return errors.New("Network does not exist")
	}

	return nil
}

// CommandInfo is an alias for "docker network inspect"
func CommandInfo() error {
	return nil
}

// CommandList is an alias for "docker network ls"
func CommandList() error {
	networks, err := listNetworks()
	if err != nil {
		return err
	}

	common.LogInfo2Quiet("Networks")
	for _, networkName := range networks {
		fmt.Println(networkName)
	}

	return nil
}

// CommandRebuildall rebuilds network settings for all apps
func CommandRebuildall() error {
	apps, err := common.DokkuApps()
	if err != nil {
		return err
	}

	for _, appName := range apps {
		err = BuildConfig(appName)
		if err != nil {
			return err
		}
	}

	return nil
}

// CommandReport displays a network report for one or more apps
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

// CommandSet set or clear a network property for an app
func CommandSet(appName string, property string, value string, values []string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	if property == "bind-all-interfaces" && value == "" {
		value = "false"
	}

	attachProperties := map[string]bool{
		"attach-post-create": true,
		"attach-post-deploy": true,
	}

	invalidNetworks := map[string]bool{
		"host":   true,
		"bridge": true,
	}
	if attachProperties[property] {
		for _, networkName := range values {
			if invalidNetworks[networkName] {
				return fmt.Errorf("Invalid network name specified for attach: %s", networkName)
			}

			if isConflictingPropertyValue(appName, property, networkName) {
				return fmt.Errorf("Network name already associated with this app: %s", networkName)
			}
		}
		value = strings.Join(values, ",")
	}

	common.CommandPropertySet("network", appName, property, value, DefaultProperties, GlobalProperties)
	return nil
}
