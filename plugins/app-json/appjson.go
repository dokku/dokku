package appjson

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/dokku/dokku/plugins/common"
)

type AppJson struct {
	Scripts struct {
		Dokku string `json:"dokku"`
	} `json:"scripts"`
}

// getPhaseScript extracts app.json from app image and returns the appropriate json key/value
func getPhaseScript(appName string, image string, phase string) (string, error) {
	appJsonFile, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("dokku-%s-%s", common.MustGetEnv("DOKKU_PID"), "getPhaseScript"))
	if err != nil {
		return "", fmt.Errorf("Cannot create temporary file: %v", err)
	}

	defer os.Remove(appJsonFile.Name())

	common.CopyFromImage(appName, image, "app.json", appJsonFile.Name())
	if !common.FileExists(appJsonFile.Name()) {
		return "", nil
	}

	b, err := ioutil.ReadAll(appJsonFile)
	if err != nil {
		return "", fmt.Errorf("Cannot read app.json file: %v", err)
	}

	var appJson AppJson
	if err = json.Unmarshal(b, &appJson); err != nil {
		return "", fmt.Errorf("Cannot parse app.json: %v", err)
	}

	return appJson.Scripts.Dokku, nil
}

func executeScript(appName string, imageTag string, phase string) error {
	return nil
}
