package network

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// CommandCreate is an alias for "docker network create"
func CommandCreate(networkName string) error {
	common.LogInfo1Quiet(fmt.Sprintf("Creating network %v", networkName))
	createCmd := common.NewShellCmd(strings.Join([]string{
		common.DockerBin(),
		"network",
		"create",
		"--attachable",
		"--label",
		fmt.Sprintf("com.dokku.network-name=%v", networkName),
		networkName,
	}, " "))
	var stderr bytes.Buffer
	createCmd.ShowOutput = false
	createCmd.Command.Stderr = &stderr
	_, err := createCmd.Output()

	if err != nil {
		err = errors.New(strings.TrimSpace(stderr.String()))
	}

	return err
}

// CommandDestroy is an alias for "docker network rm"
func CommandDestroy(networkName string, forceDestroy bool) error {
	if os.Getenv("DOKKU_APPS_FORCE_DELETE") == "1" {
		forceDestroy = true
	}

	if !forceDestroy {
		common.LogWarn("WARNING: Potentially Destructive Action")
		common.LogWarn(fmt.Sprintf("This command will destroy network %v.", networkName))
		common.LogWarn(fmt.Sprintf("To proceed, type \"%v\"", networkName))
		fmt.Print("> ")
		var response string
		_, err := fmt.Scanln(&response)
		if err != nil {
			return err
		}

		if response != networkName {
			common.LogStderr("Confirmation did not match test. Aborted.")
			os.Exit(1)
			return nil
		}
	}

	common.LogInfo1Quiet(fmt.Sprintf("Destroying network %v", networkName))
	destroyCmd := common.NewShellCmd(strings.Join([]string{
		common.DockerBin(),
		"network",
		"rm",
		networkName,
	}, " "))
	var stderr bytes.Buffer
	destroyCmd.ShowOutput = false
	destroyCmd.Command.Stderr = &stderr
	_, err := destroyCmd.Output()

	if err != nil {
		err = errors.New(strings.TrimSpace(stderr.String()))
	}

	return err
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
		common.LogFail("Network does not exist")
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
func CommandRebuildall() {
	apps, err := common.DokkuApps()
	if err != nil {
		common.LogFail(err.Error())
	}
	for _, appName := range apps {
		BuildConfig(appName)
	}
}

// CommandReport displays a network report for one or more apps
func CommandReport(appName string, infoFlag string) error {
	if strings.HasPrefix(appName, "--") {
		infoFlag = appName
		appName = ""
	}

	if len(appName) == 0 {
		apps, err := common.DokkuApps()
		if err != nil {
			return err
		}
		for _, appName := range apps {
			if err := ReportSingleApp(appName, infoFlag); err != nil {
				return err
			}
		}
		return nil
	}

	return ReportSingleApp(appName, infoFlag)
}

// CommandSet set or clear a network property for an app
func CommandSet(appName string, property string, value string) {
	if property == "bind-all-interfaces" && value == "" {
		value = "false"
	}

	attachProperites := map[string]bool{
		"attach-post-create": true,
		"attach-post-deploy": true,
	}

	invalidNetworks := map[string]bool{
		"host":   true,
		"bridge": true,
	}
	if attachProperites[property] {
		if invalidNetworks[value] {
			common.LogFail("Invalid network name specified for attach")
		}

		if isConflictingPropertyValue(appName, property, value) {
			common.LogFail("Network name already associated with this app")
		}
	}

	common.CommandPropertySet("network", appName, property, value, DefaultProperties)
}
