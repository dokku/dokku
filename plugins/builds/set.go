package builds

import (
	"fmt"
	"strconv"

	"github.com/dokku/dokku/plugins/common"
)

// CommandSet writes (or clears) a builds property for an app or globally.
func CommandSet(appName, property, value string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	if _, ok := DefaultProperties[property]; !ok {
		return fmt.Errorf("Invalid property %q (allowed: retention)", property)
	}

	if value != "" {
		if err := validateSetValue(property, value); err != nil {
			return err
		}
	}

	common.CommandPropertySet("builds", appName, property, value, DefaultProperties, GlobalProperties)
	return nil
}

func validateSetValue(property, value string) error {
	switch property {
	case "retention":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("Invalid retention %q: must be a positive integer", value)
		}
		if n < MinimumRetention {
			return fmt.Errorf("Invalid retention %d: must be >= %d", n, MinimumRetention)
		}
	}
	return nil
}
