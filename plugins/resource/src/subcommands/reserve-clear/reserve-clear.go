package main

import (
	"flag"
	"os"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/resource"
)

func main() {
	args := flag.NewFlagSet("resource:reserve-clear", flag.ExitOnError)
	processType := args.String("process-type", "", "process-type: A process type to clear")
	args.Parse(os.Args[2:])

	err := resource.CommandReserveClear(args.Args(), *processType)
	if err != nil {
		common.LogFail(err.Error())
	}
}
