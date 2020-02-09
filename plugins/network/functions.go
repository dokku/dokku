package network

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/dokku/dokku/plugins/common"

	sh "github.com/codeskyblue/go-sh"
)

// attachAppToNetwork attaches a container to a network
func attachAppToNetwork(containerID string, networkName string, appName string, phase string, processType string) {
	cmdParts := []string{
		common.DockerBin(),
		"network",
		"connect",
	}

	if phase == "deploy" {
		property := "tld"
		defaultValue := GetDefaultValue(property)
		tld := common.PropertyGetDefault("network", appName, property, defaultValue)

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
		common.LogFail(fmt.Sprintf("Unable to attach container to network: %v", err.Error()))
	}
}

// isConflictingPropertyValue returns true if the other attach property has a conflicting value
func isConflictingPropertyValue(appName string, property string, value string) bool {
	if value == "" {
		return false
	}

	otherProperty := "attach-post-create"
	if property == otherProperty {
		otherProperty = "attach-post-deploy"
	}

	defaultValue := GetDefaultValue(otherProperty)
	otherValue := common.PropertyGetDefault("network", appName, otherProperty, defaultValue)

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
	b, err := sh.Command(common.DockerBin(), "network", "list", "--format", "{{ .Name }}").Output()
	output := strings.TrimSpace(string(b[:]))

	networks := []string{}
	if err != nil {
		common.LogVerboseQuiet(output)
		return networks, err
	}

	networks = strings.Split(output, "\n")
	return networks, nil
}
