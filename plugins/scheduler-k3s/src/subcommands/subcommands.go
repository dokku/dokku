package main

import (
	"errors"
	"fmt"
	"os"
	"slices"
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
		if *global {
			appName = "--global"
			trigger = args.Arg(0)
		}
		err = scheduler_k3s.CommandAutoscalingAuthSet(appName, trigger, *metadata, *global)
	case "autoscaling-auth:report":
		args := flag.NewFlagSet("scheduler-k3s:autoscaling-auth:report", flag.ExitOnError)
		global := args.Bool("global", false, "--global: show a global report")
		includeMetadata := args.Bool("include-metadata", false, "--include-metadata: include metadata in the report")
		format := args.String("format", "stdout", "format: [ stdout | json ]")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = scheduler_k3s.CommandAutoscalingAuthReport(appName, *format, *global, *includeMetadata)
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
	case "ensure-charts":
		args := flag.NewFlagSet("scheduler-k3s:ensure-charts", flag.ExitOnError)
		forceInstall := args.Bool("force", false, "--force: force install all charts")
		chartNames := args.StringSlice("charts", []string{}, "--charts: comma separated list of chart names to force install")
		args.Parse(os.Args[2:])
		err = scheduler_k3s.CommandEnsureCharts(*forceInstall, *chartNames)
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
	case "add-pvc":
		args := flag.NewFlagSet("scheduler-k3s:add-pvc", flag.ExitOnError)
		accessMode := args.String("access-mode", "ReadWriteOnce", "--access-mode: access mode default ReadWriteOnce")
		namespace := args.String("namespace", "default", "--namespace: default")
		storageClass := args.String("storage-class-name", "", "--storage-class-name: e.g. longhorn")
		args.Parse(os.Args[2:])
		// check accessMode
		accessModes := []string{"ReadWriteOnce", "ReadWriteMany", "ReadOnlyMany"}
		if !slices.Contains(accessModes, *accessMode) {
			err = errors.New("Please specify PVC access mode as either ReadWriteOnce, ReadOnlyMany,  ReadWriteMany")
			break
		}
		pvcName := args.Arg(0)
		storageSize := args.Arg(1)
		err = scheduler_k3s.CommandAddPVC(pvcName, *namespace, *accessMode, storageSize, *storageClass)
	case "remove-pvc":
		args := flag.NewFlagSet("scheduler-k3s:remove-pvc", flag.ExitOnError)
		namespace := args.String("namespace", "default", "--namespace: default")
		args.Parse(os.Args[2:])
		pvcName := args.Arg(0)
		err = scheduler_k3s.CommandRemovePVC(pvcName, *namespace)
	case "mount":
		args := flag.NewFlagSet("scheduler-k3s:mount", flag.ExitOnError)
		subPath := args.String("subpath", "", "--subpath: ")
		readOnly := args.Bool("readonly", false, "--readonly: false")
		processType := args.String("process-type", "web", "--process-type: web")
		chown := args.String("chown", "", "--chown: UID:GID")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		pvcName := args.Arg(1)
		mountPath := args.Arg(2)
		err = scheduler_k3s.CommandMountPVC(appName, *processType, pvcName, mountPath, *subPath, *readOnly, *chown)
	case "unmount":
		args := flag.NewFlagSet("scheduler-k3s:unmount", flag.ExitOnError)
		processType := args.String("process-type", "web", "--process-type: web")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		pvcName := args.Arg(1)
		mountPath := args.Arg(2)
		err = scheduler_k3s.CommandUnMountPVC(appName, *processType, pvcName, mountPath)
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
