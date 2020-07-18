package appjson

func TriggerPostDeploy(appName string, imageTag string) error {
	if err := executeScript(appName, imageTag, "predeploy"); err != nil {
		return err
	}

	if err := executeScript(appName, imageTag, "release"); err != nil {
		return err
	}
	return nil
}

func TriggerPreDeploy(appName string, imageTag string) error {
	if err := executeScript(appName, imageTag, "postdeploy"); err != nil {
		return err
	}
	return nil
}
