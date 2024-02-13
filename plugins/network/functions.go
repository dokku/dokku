package network

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// attachAppToNetwork attaches a container to a network
func attachAppToNetwork(containerID string, networkName string, appName string, phase string, processType string) error {
	if isContainerInNetwork(containerID, networkName) {
		return nil
	}

	cmdParts := []string{
		common.DockerBin(),
		"network",
		"connect",
	}

	if phase == "deploy" {
		tld := reportComputedTld(appName)

		networkAlias := fmt.Sprintf("%v.%v", appName, processType)
		if tld != "" {
			networkAlias = fmt.Sprintf("%v.%v", networkAlias, tld)
		}

		cmdParts = append(cmdParts, "--alias")
		cmdParts = append(cmdParts, networkAlias)

		hostname, err := common.DockerInspect(containerID, "{{ .Config.Hostname }}")
		if err != nil {
			common.LogWarn(err.Error())
		} else {
			cmdParts = append(cmdParts, "--alias")
			cmdParts = append(cmdParts, fmt.Sprintf("%v.%v", hostname, networkAlias))
		}
	}

	cmdParts = append(cmdParts, networkName)
	cmdParts = append(cmdParts, containerID)
	attachCmd := common.NewShellCmd(strings.Join(cmdParts, " "))
	var stderr bytes.Buffer
	attachCmd.ShowOutput = false
	attachCmd.Command.Stderr = &stderr
	_, err := attachCmd.Output()
	if err != nil {
		err = errors.New(strings.TrimSpace(stderr.String()))
		return fmt.Errorf("Unable to attach container to network: %v", err.Error())
	}

	return nil
}

// isContainerInNetwork returns true if the container is already attached to the specified network
func isContainerInNetwork(containerID string, networkName string) bool {
	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command:       common.DockerBin(),
		Args:          []string{"container", "inspect", "--format", "{{range $net, $v := .NetworkSettings.Networks}}{{println $net}}{{end}}", containerID},
		CaptureOutput: true,
	})

	if err != nil {
		common.LogVerboseQuiet(fmt.Sprintf("Error checking container networking status: %v", err.Error()))
		return false
	}
	if result.ExitCode != 0 {
		common.LogVerboseQuiet(fmt.Sprintf("Error checking container networking status: %v", result.StderrContents()))
		return false
	}

	for _, line := range strings.Split(result.StdoutContents(), "\n") {
		network := strings.TrimSpace(line)
		if network == "" {
			continue
		}

		if network == networkName {
			return true
		}
	}

	return false
}

// isConflictingPropertyValue returns true if the other attach property has a conflicting value
func isConflictingPropertyValue(appName string, property string, value string) bool {
	if value == "" {
		return false
	}

	otherValue := ""
	if property == "attach-post-create" {
		otherValue = reportComputedAttachPostDeploy(appName)
	} else {
		otherValue = reportComputedAttachPostCreate(appName)
	}

	return value == otherValue
}

// networkExists checks to see if a network exists
func networkExists(networkName string) (bool, error) {
	if networkName == "" {
		return false, errors.New("No network name specified")
	}

	exists := false

	networks, err := listNetworks()
	if err != nil {
		return false, err
	}

	for _, n := range networks {
		if networkName == n {
			exists = true
			break
		}
	}

	return exists, nil
}

// listNetworks returns a list of docker networks
func listNetworks() ([]string, error) {
	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command:       common.DockerBin(),
		Args:          []string{"network", "ls", "--format", "{{ .Name }}"},
		CaptureOutput: true,
	})
	if err != nil {
		common.LogVerboseQuiet(result.StderrContents())
		return []string{}, err
	}
	if result.ExitCode != 0 {
		common.LogVerboseQuiet(result.StderrContents())
		return []string{}, fmt.Errorf("Unable to list networks")
	}

	networks := strings.Split(result.StdoutContents(), "\n")
	return networks, nil
}
