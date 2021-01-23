package logs

import (
	"errors"
	"fmt"
	"strconv"
)

func validateSetValue(appName string, key string, value string) error {
	if key == "max-size" {
		return validateMaxSize(appName, value)
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
		return errors.New("Invalid max-size value, must be a number followed by a unit of measure [k, m, d]")
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

	_, err := valueToConfig(appName, value)
	if err != nil {
		return err
	}

	return nil
}
