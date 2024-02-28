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
	case "annotations:set":
		args := flag.NewFlagSet("scheduler-k3s:annotations:set", flag.ExitOnError)
		global := args.Bool("global", false, "--global: set a global property")
		processType := args.String("process-type", "", "--process-type: scope to process-type")
		resourceType := args.String("resource-type", "", "--resource-type: scope to resource-type")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		property := args.Arg(1)
		value := args.Arg(2)
		if *global {
			appName = "--global"
			property = args.Arg(0)
			value = args.Arg(1)
		}

		err = scheduler_k3s.CommandAnnotationsSet(appName, *processType, *resourceType, property, value)
	case "autoscaling-auth:set":
		args := flag.NewFlagSet("scheduler-k3s:autoscaling-auth:set", flag.ExitOnError)
		global := args.Bool("global", false, "--global: set a global property")
		metadata := args.StringToString("metadata", map[string]string{}, "--metadata: a key=value map of parameter metadata")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		trigger := args.Arg(1)
		err = scheduler_k3s.CommandAutoscalingAuthSet(appName, trigger, *metadata, *global)
	case "cluster-add":
		args := flag.NewFlagSet("scheduler-k3s:cluster-add", flag.ExitOnError)
		allowUknownHosts := args.Bool("insecure-allow-unknown-hosts", false, "insecure-allow-unknown-hosts: allow unknown hosts")
		taintScheduling := args.Bool("taint-scheduling", false, "taint-scheduling: add a taint against scheduling app workloads")
		serverIP := args.String("server-ip", "", "server-ip: IP address of the dokku server node")
		role := args.String("role", "worker", "role: [ server | worker ]")
		args.Parse(os.Args[2:])
		remoteHost := args.Arg(0)
		err = scheduler_k3s.CommandClusterAdd(*role, remoteHost, *serverIP, *allowUknownHosts, *taintScheduling)
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
	case "initialize":
		args := flag.NewFlagSet("scheduler-k3s:initialize", flag.ExitOnError)
		taintScheduling := args.Bool("taint-scheduling", false, "taint-scheduling: add a taint against scheduling app workloads")
		serverIP := args.String("server-ip", "", "server-ip: IP address of the dokku server node")
		ingressClass := args.String("ingress-class", "traefik", "ingress-class: ingress-class to use for all outbound traffic")
		args.Parse(os.Args[2:])
		err = scheduler_k3s.CommandInitialize(*ingressClass, *serverIP, *taintScheduling)
	case "labels:set":
		args := flag.NewFlagSet("scheduler-k3s:labels:set", flag.ExitOnError)
		global := args.Bool("global", false, "--global: set a global property")
		processType := args.String("process-type", "", "--process-type: scope to process-type")
		resourceType := args.String("resource-type", "", "--resource-type: scope to resource-type")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		property := args.Arg(1)
		value := args.Arg(2)
		if *global {
			appName = "--global"
			property = args.Arg(0)
			value = args.Arg(1)
		}

		err = scheduler_k3s.CommandLabelsSet(appName, *processType, *resourceType, property, value)
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
