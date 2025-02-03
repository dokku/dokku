package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

const (
	helpHeader = `Usage: dokku scheduler-k3s[:COMMAND]

Manage scheduler-k3s settings for an app

Additional commands:`

	helpContent = `
    scheduler-k3s:autoscaling-auth:set <app|--global> <trigger> [<--metadata key=value>...], Set or clear a scheduler-k3s autoscaling keda trigger authentication resource for an app
    scheduler-k3s:annotations:set <app|--global> <property> (<value>) [--process-type PROCESS_TYPE] <--resource-type RESOURCE_TYPE>, Set or clear an annotation for a given app/process-type/resource-type combination
    scheduler-k3s:cluster-add [--insecure-allow-unknown-hosts] [--server-ip SERVER_IP] [--taint-scheduling] <ssh://user@host:port>, Adds a server node to a Dokku-managed cluster
    scheduler-k3s:cluster-list [--format json|stdout], Lists all nodes in a Dokku-managed cluster
    scheduler-k3s:cluster-remove [node-id], Removes client node to a Dokku-managed cluster
	scheduler-k3s:ensure-charts, Ensures the k3s charts are installed
    scheduler-k3s:initialize [--server-ip SERVER_IP] [--taint-scheduling], Initializes a cluster
    scheduler-k3s:labels:set <app|--global> <property> (<value>) [--process-type PROCESS_TYPE] <--resource-type RESOURCE_TYPE>, Set or clear a label for a given app/process-type/resource-type combination
    scheduler-k3s:report [<app>] [<flag>], Displays a scheduler-k3s report for one or more apps
    scheduler-k3s:set <app> <property> (<value>), Set or clear a scheduler-k3s property for an app
    scheduler-k3s:show-kubeconfig, Displays the kubeconfig for remote usage
    scheduler-k3s:add-pvc <name> <size> [--access-mode <mode>] [--namespace <namespace>] [--storage-class-name <class-name>], Adds a Persistent Volume Claim (PVC)
	scheduler-k3s:remove-pvc <name> [--namespace <namespace>], Remove Persistent Volume Claim in Namespace
	scheduler-k3s:mount <app> <pvc_name> </container/path> [--process-type PROCESS_TYPE] [--subpath SUBPATH] [--readonly] [--chown option], Mount a Volume to Container Path for an app. Default processtype: web
	scheduler-k3s:unmount <app> <pvc_name> </container/path>, Unmount a Volume from an app
	scheduler-k3s:uninstall, Uninstalls k3s from the Dokku server`
)

func main() {
	cmd := ""
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "scheduler-k3s", "scheduler-k3s:help":
		usage()
	case "help":
		result, err := common.CallExecCommand(common.ExecCommandInput{
			Command: "ps",
			Args:    []string{"-o", "command=", strconv.Itoa(os.Getppid())},
		})
		if err == nil && strings.Contains(result.StdoutContents(), "--all") {
			fmt.Println(helpContent)
		} else {
			fmt.Print("\n    scheduler-k3s, Manage scheduler-k3s settings for an app\n")
		}
	default:
		dokkuNotImplementExitCode, err := strconv.Atoi(os.Getenv("DOKKU_NOT_IMPLEMENTED_EXIT"))
		if err != nil {
			fmt.Println("failed to retrieve DOKKU_NOT_IMPLEMENTED_EXIT environment variable")
			dokkuNotImplementExitCode = 10
		}
		os.Exit(dokkuNotImplementExitCode)
	}
}

func usage() {
	common.CommandUsage(helpHeader, helpContent)
}
