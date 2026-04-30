package main

import (
	"fmt"
	"os"
	"strings"

	flag "github.com/spf13/pflag"

	"github.com/dokku/dokku/plugins/common"
	schedulerdockerlocal "github.com/dokku/dokku/plugins/scheduler-docker-local"
)

// main entrypoint to all triggers
func main() {
	parts := strings.Split(os.Args[0], "/")
	trigger := parts[len(parts)-1]

	interactive := flag.Bool("interactive", false, "--interactive: stdin is open")
	tty := flag.Bool("tty", false, "--tty: stdin is a terminal")
	asUser := flag.String("as-user", "", "--as-user: numeric uid override")
	flag.Parse()

	var err error
	switch trigger {
	case "scheduler-cron-write":
		scheduler := flag.Arg(0)
		err = schedulerdockerlocal.TriggerSchedulerCronWrite(scheduler)
	case "scheduler-storage-exec":
		args := flag.Args()
		if len(args) < 3 {
			err = fmt.Errorf("scheduler-storage-exec requires <scheduler> <entry-name> <image>")
			break
		}
		scheduler := args[0]
		entryName := args[1]
		image := args[2]
		var cmd []string
		if len(args) > 3 {
			cmd = args[3:]
		}
		err = schedulerdockerlocal.TriggerSchedulerStorageExec(scheduler, schedulerdockerlocal.StorageExecInput{
			EntryName:   entryName,
			Image:       image,
			Interactive: *interactive,
			Tty:         *tty,
			AsUser:      *asUser,
			Command:     cmd,
		})
	default:
		err = fmt.Errorf("Invalid plugin trigger call: %s", trigger)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
