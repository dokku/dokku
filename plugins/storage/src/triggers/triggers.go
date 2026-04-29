package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/storage"
)

func main() {
	parts := strings.Split(os.Args[0], "/")
	trigger := parts[len(parts)-1]
	flag.Parse()

	var err error
	switch trigger {
	case "install":
		err = storage.TriggerInstall()
	case "storage-list":
		appName := flag.Arg(0)
		phase := flag.Arg(1)
		format := flag.Arg(2)
		err = storage.TriggerStorageList(appName, phase, format)
	case "storage-app-mounts":
		appName := flag.Arg(0)
		phase := flag.Arg(1)
		err = storage.TriggerStorageAppMounts(appName, phase)
	case "docker-args-deploy":
		appName := flag.Arg(0)
		err = storage.TriggerDockerArgs(appName, storage.PhaseDeploy)
	case "docker-args-run":
		appName := flag.Arg(0)
		err = storage.TriggerDockerArgs(appName, storage.PhaseRun)
	case "post-delete":
		appName := flag.Arg(0)
		err = storage.TriggerPostDelete(appName)
	case "post-app-clone-setup":
		oldName := flag.Arg(0)
		newName := flag.Arg(1)
		err = storage.TriggerPostAppCloneSetup(oldName, newName)
	case "post-app-rename-setup":
		oldName := flag.Arg(0)
		newName := flag.Arg(1)
		err = storage.TriggerPostAppRenameSetup(oldName, newName)
	default:
		err = fmt.Errorf("Invalid plugin trigger call: %s", trigger)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
