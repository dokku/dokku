package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	schedulerdockerlocal "github.com/dokku/dokku/plugins/scheduler-docker-local"
)

// main entrypoint to all triggers
func main() {
	parts := strings.Split(os.Args[0], "/")
	trigger := parts[len(parts)-1]
	flag.Parse()

	var err error
	switch trigger {
	case "scheduler-cron-write":
		scheduler := flag.Arg(0)
		err = schedulerdockerlocal.TriggerSchedulerCronWrite(scheduler)
	default:
		err = fmt.Errorf("Invalid plugin trigger call: %s", trigger)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
