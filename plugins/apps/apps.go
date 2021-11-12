package apps

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

var (
	// DefaultProperties is a map of all valid network properties with corresponding default property values
	DefaultProperties = map[string]string{
		"deploy-source":          "",
		"deploy-source-metadata": "",
	}

	// GlobalProperties is a map of all valid global network properties
	GlobalProperties = map[string]bool{
		"deploy-source":          true,
		"deploy-source-metadata": true,
	}
)

// DokkuApps returns a list of all local apps
func DokkuApps() ([]string, error) {
	apps := []string{}
	dokkuRoot := common.MustGetEnv("DOKKU_ROOT")
	files, err := ioutil.ReadDir(dokkuRoot)
	if err != nil {
		return apps, fmt.Errorf("You haven't deployed any applications yet")
	}

	for _, f := range files {
		appRoot := common.AppRoot(f.Name())
		if !common.DirectoryExists(appRoot) {
			continue
		}
		if strings.HasPrefix(f.Name(), ".") {
			continue
		}
		apps = append(apps, f.Name())
	}

	if len(apps) == 0 {
		return apps, fmt.Errorf("You haven't deployed any applications yet")
	}

	return apps, nil
}
