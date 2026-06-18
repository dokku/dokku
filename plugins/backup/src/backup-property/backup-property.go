package main

import (
	"fmt"
	"os"

	"github.com/dokku/dokku/plugins/common"
)

// backup-property is a small helper invoked by bash plugins to export or import
// their property store as a generic docket recipe slice, so there is a single
// YAML serializer shared by every plugin.
//
// Usage:
//
//	backup-property export <plugin> <app|--global> <scopeDir> [taskType]
//	backup-property import <plugin> <app|--global> <scopeDir>
func main() {
	if len(os.Args) < 5 {
		common.LogFail("Usage: backup-property <export|import> <plugin> <app|--global> <scopeDir> [taskType]")
	}

	action := os.Args[1]
	pluginName := os.Args[2]
	appName := os.Args[3]
	scopeDir := os.Args[4]

	var err error
	switch action {
	case "export":
		taskType := ""
		if len(os.Args) >= 6 {
			taskType = os.Args[5]
		}
		err = common.BackupPropertyExport(pluginName, appName, scopeDir, taskType)
	case "import":
		err = common.BackupPropertyImport(pluginName, appName, scopeDir)
	default:
		err = fmt.Errorf("unknown backup-property action: %s", action)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
