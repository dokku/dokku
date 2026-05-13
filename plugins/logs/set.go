package logs

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

func validateSetValue(appName string, key string, value string) error {
	if key == "max-size" {
		return validateMaxSize(appName, value)
	}

	if key == "vector-image" {
		return validateVectorImage(appName, value)
	}

	if key == "vector-networks" {
		return validateVectorNetworks(appName, value)
	}

	if key == "vector-sink" {
		return validateVectorSink(appName, value)
	}

	return nil
}

func validateMaxSize(appName string, value string) error {
	if value == "" {
		return nil
	}

	if value == "unlimited" {
		return nil
	}

	last := value[len(value)-1:]
	if last != "k" && last != "m" && last != "g" {
		return errors.New("Invalid max-size unit measure, value must end in any of [k, m, g]")
	}

	if len(value) < 2 {
		return errors.New("Invalid max-size value, must be a number followed by a unit of measure [k, m, g]")
	}

	number := value[:len(value)-1]
	if _, err := strconv.Atoi(number); err != nil {
		return fmt.Errorf("Invalid max-size value, unable to convert number to int: %s", err.Error())
	}

	return nil
}

func validateVectorSink(appName string, value string) error {
	if value == "" {
		return nil
	}

	_, err := SinkValueToConfig(appName, value)
	if err != nil {
		return err
	}

	return nil
}

func validateVectorImage(appName string, value string) error {
	if appName != "--global" {
		return errors.New("vector-image may only be set globally with --global")
	}

	return nil
}

func validateVectorNetworks(appName string, value string) error {
	if appName != "--global" {
		return errors.New("vector-networks may only be set globally with --global")
	}

	if value == "" {
		return nil
	}

	for _, name := range strings.Split(value, ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			return errors.New("Invalid vector-networks value, empty entry in comma-separated list")
		}

		if name == "bridge" {
			return errors.New("Invalid vector-networks value, \"bridge\" is not a valid entry for vector-networks")
		}

		result, err := common.CallExecCommand(common.ExecCommandInput{
			Command: common.DockerBin(),
			Args:    []string{"network", "inspect", name},
		})
		if err != nil || result.ExitCode != 0 {
			return fmt.Errorf("Network %q does not exist", name)
		}
	}

	return nil
}
