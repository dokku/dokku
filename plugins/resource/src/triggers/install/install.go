package main

import (
	"fmt"

	"github.com/dokku/dokku/plugins/common"
)

// runs the install step for the resource plugin
func main() {
	if err := common.PropertySetup("resource"); err != nil {
		common.LogFail(fmt.Sprintf("Unable to install the resource plugin: %s", err.Error()))
	}
}
