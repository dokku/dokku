package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	scheduler_k3s "github.com/dokku/dokku/plugins/scheduler-k3s"

	flag "github.com/spf13/pflag"
)

// main entrypoint to all subcommands
func main() {
	parts := strings.Split(os.Args[0], "/")
	subcommand := parts[len(parts)-1]

	var err error
	switch subcommand {
	case "initialize":
		args := flag.NewFlagSet("scheduler-k3s:initialize", flag.ExitOnError)
		taintScheduling := args.Bool("taint-scheduling", false, "taint-scheduling: add a taint against scheduling app workloads")
		args.Parse(os.Args[2:])
		err = scheduler_k3s.CommandInitialize(*taintScheduling)
	case "cluster-add":
		args := flag.NewFlagSet("scheduler-k3s:cluster-add", flag.ExitOnError)
		allowUknownHosts := args.Bool("insecure-allow-unknown-hosts", false, "insecure-allow-unknown-hosts: allow unknown hosts")
		taintScheduling := args.Bool("taint-scheduling", false, "taint-scheduling: add a taint against scheduling app workloads")
		role := args.String("role", "worker", "role: [ server | worker ]")
		args.Parse(os.Args[2:])
		remoteHost := args.Arg(0)
		err = scheduler_k3s.CommandClusterAdd(*role, remoteHost, *allowUknownHosts, *taintScheduling)
	case "cluster-list":
		args := flag.NewFlagSet("scheduler-k3s:cluster-list", flag.ExitOnError)
		format := args.String("format", "stdout", "format: [ stdout | json ]")
		args.Parse(os.Args[2:])
		err = scheduler_k3s.CommandClusterList(*format)
	case "cluster-remove":
		args := flag.NewFlagSet("scheduler-k3s:cluster-remove", flag.ExitOnError)
		args.Parse(os.Args[2:])
		nodeName := args.Arg(0)
		err = scheduler_k3s.CommandClusterRemove(nodeName)
	case "report":
		args := flag.NewFlagSet("scheduler-k3s:report", flag.ExitOnError)
		format := args.String("format", "stdout", "format: [ stdout | json ]")
		osArgs, infoFlag, flagErr := common.ParseReportArgs("scheduler-k3s", os.Args[2:])
		if flagErr == nil {
			args.Parse(osArgs)
			appName := args.Arg(0)
			err = scheduler_k3s.CommandReport(appName, *format, infoFlag)
		}
	case "set":
		args := flag.NewFlagSet("scheduler-k3s:set", flag.ExitOnError)
		global := args.Bool("global", false, "--global: set a global property")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		property := args.Arg(1)
		value := args.Arg(2)
		if *global {
			appName = "--global"
			property = args.Arg(0)
			value = args.Arg(1)
		}
		err = scheduler_k3s.CommandSet(appName, property, value)
	case "show-kubeconfig":
		args := flag.NewFlagSet("scheduler-k3s:show-kubeconfig", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = scheduler_k3s.CommandShowKubeconfig()
	case "uninstall":
		args := flag.NewFlagSet("scheduler-k3s:uninstall", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = scheduler_k3s.CommandUninstall()
	default:
		err = fmt.Errorf("Invalid plugin subcommand call: %s", subcommand)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
