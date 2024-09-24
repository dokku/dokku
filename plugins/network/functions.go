package network

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dokku/dokku/plugins/common"
)

type DockerNetwork struct {
	CreatedAt time.Time
	Driver    string
	ID        string
	Internal  bool
	IPv6      bool
	Labels    map[string]string
	Name      string
	Scope     string
}

// attachAppToNetwork attaches a container to a network
func attachAppToNetwork(containerID string, networkName string, appName string, phase string, processType string) error {
	if isContainerInNetwork(containerID, networkName) {
		return nil
	}

	cmdParts := []string{
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

	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command: common.DockerBin(),
		Args:    cmdParts,
	})
	if err != nil {
		return fmt.Errorf("Unable to attach container to network: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Unable to attach container to network: %s", result.StderrContents())
	}

	return nil
}

// isContainerInNetwork returns true if the container is already attached to the specified network
func isContainerInNetwork(containerID string, networkName string) bool {
	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command: common.DockerBin(),
		Args:    []string{"container", "inspect", "--format", "{{range $net, $v := .NetworkSettings.Networks}}{{println $net}}{{end}}", containerID},
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

	networks, err := getNetworks()
	if err != nil {
		return false, err
	}

	for _, network := range networks {
		if networkName == network.Name {
			exists = true
			break
		}
	}

	return exists, nil
}

// getNetworks returns a list of docker networks
func getNetworks() (map[string]DockerNetwork, error) {
	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command: common.DockerBin(),
		Args:    []string{"network", "ls", "--format", "json"},
	})
	if err != nil {
		common.LogVerboseQuiet(result.StderrContents())
		return map[string]DockerNetwork{}, err
	}
	if result.ExitCode != 0 {
		common.LogVerboseQuiet(result.StderrContents())
		return map[string]DockerNetwork{}, fmt.Errorf("Unable to list networks")
	}

	networkLines := strings.Split(result.StdoutContents(), "\n")
	networks := map[string]DockerNetwork{}
	for _, line := range networkLines {
		if line == "" {
			continue
		}
		result := make(map[string]interface{})
		err := json.Unmarshal([]byte(line), &result)
		if err != nil {
			return map[string]DockerNetwork{}, err
		}

		network := DockerNetwork{
			Driver: result["Driver"].(string),
			ID:     result["ID"].(string),
			Name:   result["Name"].(string),
			Scope:  result["Scope"].(string),
			Labels: map[string]string{},
		}

		if createdAtVal := result["CreatedAt"].(string); createdAtVal != "" {
			createdAt, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", "2024-02-25 01:55:24.275184461 +0000 UTC")
			if err == nil {
				network.CreatedAt = createdAt
			}
		}

		if ipv6Val := result["IPv6"].(string); ipv6Val != "" {
			val, err := strconv.ParseBool(ipv6Val)
			if err == nil {
				network.IPv6 = val
			}
		}
		if internalVal := result["Internal"].(string); internalVal != "" {
			val, err := strconv.ParseBool(internalVal)
			if err == nil {
				network.Internal = val
			}
		}

		labels := strings.Split(result["Labels"].(string), ",")
		for _, v := range labels {
			parts := strings.SplitN(v, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := parts[0]
			value := parts[1]
			network.Labels[key] = value
		}

		networks[network.Name] = network
	}

	return networks, nil
}
