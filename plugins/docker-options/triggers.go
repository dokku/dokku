package dockeroptions

// TriggerPostAppCloneSetup copies docker option files from the source app to the cloned app
func TriggerPostAppCloneSetup(oldAppName string, newAppName string) error {
	for _, phase := range availablePhases {
		if err := copyPhaseFile(oldAppName, newAppName, phase); err != nil {
			return err
		}
	}
	return nil
}

// TriggerPostAppRenameSetup copies docker option files from the old app name to the new app name
func TriggerPostAppRenameSetup(oldAppName string, newAppName string) error {
	for _, phase := range availablePhases {
		if err := copyPhaseFile(oldAppName, newAppName, phase); err != nil {
			return err
		}
	}
	return nil
}

// TriggerPostDelete deletes the docker option files for an app
func TriggerPostDelete(appName string) error {
	for _, phase := range availablePhases {
		if err := removePhaseFile(appName, phase); err != nil {
			return err
		}
	}
	return nil
}
