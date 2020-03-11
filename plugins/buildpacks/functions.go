package buildpacks

import (
	"errors"
)

func getAppName(args []string) (appName string, err error) {
	if len(args) >= 1 {
		appName = args[0]
	} else {
		err = errors.New("Please specify an app to run the command on")
	}

	return
}
