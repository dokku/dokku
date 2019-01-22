package main

import (
	"fmt"

	"github.com/dokku/dokku/plugins/common"
)

// runs the install step for the buildpacks plugin
func main() {
	if err := common.PropertySetup("buildpacks"); err != nil {
		common.LogFail(fmt.Sprintf("Unable to install the buildpacks plugin: %s", err.Error()))
	}
}
