package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/builds"
	"github.com/dokku/dokku/plugins/common"
)

func main() {
	parts := strings.Split(os.Args[0], "/")
	trigger := parts[len(parts)-1]
	flag.Parse()

	var err error
	switch trigger {
	case "builds-generate-id":
		err = builds.TriggerBuildsGenerateID()
	case "builds-record-finalize":
		appName := flag.Arg(0)
		buildID := flag.Arg(1)
		exitStr := flag.Arg(2)
		err = builds.TriggerBuildsRecordFinalize(appName, buildID, exitStr)
	case "builds-record-start":
		appName := flag.Arg(0)
		buildID := flag.Arg(1)
		pid := flag.Arg(2)
		source := flag.Arg(3)
		err = builds.TriggerBuildsRecordStart(appName, buildID, pid, source)
	case "install":
		err = builds.TriggerInstall()
	case "post-app-rename-setup":
		oldAppName := flag.Arg(0)
		newAppName := flag.Arg(1)
		err = builds.TriggerPostAppRenameSetup(oldAppName, newAppName)
	case "post-delete":
		appName := flag.Arg(0)
		err = builds.TriggerPostDelete(appName)
	default:
		err = fmt.Errorf("Invalid plugin trigger call: %s", trigger)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
