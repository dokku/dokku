package appjson

// TriggerPostDeploy is a trigger to execute the postdeploy deployment task
func TriggerPostDeploy(appName string, imageTag string) error {
	if err := executeScript(appName, imageTag, "postdeploy"); err != nil {
		return err
	}
	return nil
}

// TriggerPreDeploy is a trigger to execute predeploy and release deployment tasks
func TriggerPreDeploy(appName string, imageTag string) error {
	if err := executeScript(appName, imageTag, "predeploy"); err != nil {
		return err
	}

	if err := executeScript(appName, imageTag, "release"); err != nil {
		return err
	}
	return nil
}
