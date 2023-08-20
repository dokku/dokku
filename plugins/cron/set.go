package cron

import "errors"

func validateSetValue(appName string, key string, value string) error {
	if key == "mailto" && appName != "--global" {
		return errors.New("Property cannot be specified on a per-app basis")
	}

	return nil
}
