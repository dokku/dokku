package scheduler_k3s

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/dokku/dokku/plugins/common"
	resty "github.com/go-resty/resty/v2"
	"github.com/ryanuber/columnize"
	"gopkg.in/yaml.v3"
)

// CommandAnnotationsSet set or clear a scheduler-k3s annotation for an app
func CommandAnnotationsSet(appName string, processType string, resourceType string, key string, value string) error {
	if resourceType == "" {
		return fmt.Errorf("Missing resource-type")
	}

	if processType == "" {
		processType = GlobalProcessType
	}

	property := fmt.Sprintf("%s.%s", processType, resourceType)
	annotationsList, err := common.PropertyListGet("scheduler-k3s", appName, property)
	if err != nil {
		return fmt.Errorf("Unable to get property list: %w", err)
	}

	annotations := []string{}
	for _, annotation := range annotationsList {
		parts := strings.SplitN(annotation, ": ", 2)
		if len(parts) != 2 {
			return fmt.Errorf("Invalid annotation: %s", annotation)
		}
		if key == parts[0] {
			continue
		}

		annotations = append(annotations, annotation)
	}

	if value != "" {
		annotations = append(annotations, fmt.Sprintf("%s: %s", key, value))
	}

	sort.Strings(annotations)
	if err := common.PropertyListWrite("scheduler-k3s", appName, property, annotations); err != nil {
		return fmt.Errorf("Unable to write property list: %w", err)
	}

	return nil
}

// CommandAutoscalingAuthSet set or clear a scheduler-k3s autoscaling keda trigger authentication object for an app
func CommandAutoscalingAuthSet(appName string, trigger string, metadata map[string]string, global bool) error {
	if global {
		appName = "--global"
	}

	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	if len(trigger) == 0 {
		return fmt.Errorf("Missing trigger type argument")
	}

	if len(metadata) == 0 {
		properties, err := common.PropertyGetAllByPrefix("scheduler-k3s", appName, fmt.Sprintf("%s%s.", TriggerAuthPropertyPrefix, trigger))
		if err != nil {
			return fmt.Errorf("Unable to get property list: %w", err)
		}

		for key := range properties {
			if err := common.PropertyDelete("scheduler-k3s", appName, key); err != nil {
				return fmt.Errorf("Unable to delete property: %w", err)
			}
		}

		if appName == "--global" {
			helmAgent, err := NewHelmAgent("keda", DeployLogPrinter)
			if err != nil {
				return fmt.Errorf("Unable to create helm agent: %w", err)
			}

			releaseName := fmt.Sprintf("keda-cluster-trigger-authentications-%s", trigger)
			if err := helmAgent.UninstallChart(releaseName); err != nil {
				return fmt.Errorf("Unable to uninstall chart: %w", err)
			}
		}

		return nil
	}

	for key, value := range metadata {
		if err := common.PropertyWrite("scheduler-k3s", appName, fmt.Sprintf("%s%s.%s", TriggerAuthPropertyPrefix, trigger, key), value); err != nil {
			return fmt.Errorf("Unable to set property: %w", err)
		}

		common.LogInfo1("Trigger authentication settings saved")
		common.LogVerbose("Resources will be created or updated on next deploy")
	}

	if appName == "--global" {
		err := applyKedaClusterTriggerAuthentications(context.Background(), trigger, metadata)
		if err != nil {
			return fmt.Errorf("Unable to install chart: %w", err)
		}
	}

	return nil
}

// CommandAutoscalingAuthReport displays a scheduler-k3s autoscaling keda trigger authentication report for one or more apps
func CommandAutoscalingAuthReport(appName string, format string, global bool, includeMetadata bool) error {
	if len(appName) == 0 && !global {
		return fmt.Errorf("Missing required app name or --global flag")
	}

	if len(appName) > 0 && global {
		return fmt.Errorf("Cannot specify both app name and --global flag")
	}

	if !global {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	if len(appName) > 0 {
		return ReportAutoscalingAuthSingleApp(appName, format, includeMetadata)
	}

	return ReportAutoscalingAuthSingleApp("--global", format, includeMetadata)
}

// CommandInitialize initializes a k3s cluster on the local server
func CommandInitialize(ingressClass string, serverIP string, taintScheduling bool) error {
	if ingressClass != "nginx" && ingressClass != "traefik" {
		return fmt.Errorf("Invalid ingress-class: %s", ingressClass)
	}

	if err := isK3sInstalled(); err == nil {
		return fmt.Errorf("k3s already installed, cannot re-initialize k3s")
	}

	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM)
	go func() {
		<-signals
		cancel()
	}()

	if serverIP == "" {
		var err error
		serverIP, err = getServerIP()
		if err != nil {
			return fmt.Errorf("Unable to get server ip address: %w", err)
		}

		common.LogVerboseQuiet(fmt.Sprintf("Using server ip address: %s", serverIP))
	}

	common.LogInfo1Quiet("Initializing k3s")

	common.LogInfo2Quiet("Updating apt")
	aptUpdateCmd, err := common.CallExecCommand(common.ExecCommandInput{
		Command: "apt-get",
		Args: []string{
			"update",
		},
		StreamStdio: true,
	})
	if err != nil {
		return fmt.Errorf("Unable to call apt-get update command: %w", err)
	}
	if aptUpdateCmd.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from apt-get update command: %d", aptUpdateCmd.ExitCode)
	}

	common.LogInfo2Quiet("Installing k3s dependencies")
	aptInstallCmd, err := common.CallExecCommand(common.ExecCommandInput{
		Command: "apt-get",
		Args: []string{
			"-y",
			"install",
			"ca-certificates",
			"curl",
			"open-iscsi",
			"nfs-common",
			"wireguard",
		},
		StreamStdio: true,
	})
	if err != nil {
		return fmt.Errorf("Unable to call apt-get install command: %w", err)
	}
	if aptInstallCmd.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from apt-get install command: %d", aptInstallCmd.ExitCode)
	}

	common.LogInfo2Quiet("Downloading k3s installer")
	client := resty.New()
	resp, err := client.R().
		SetContext(ctx).
		Get("https://get.k3s.io")
	if err != nil {
		return fmt.Errorf("Unable to download k3s installer: %w", err)
	}
	if resp == nil {
		return fmt.Errorf("Missing response from k3s installer download: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("Invalid status code for k3s installer script: %d", resp.StatusCode())
	}

	f, err := os.CreateTemp("", "sample")
	if err != nil {
		return fmt.Errorf("Unable to create temporary file for k3s installer: %w", err)
	}
	defer os.Remove(f.Name())

	if err := f.Close(); err != nil {
		return fmt.Errorf("Unable to close k3s installer file: %w", err)
	}

	err = common.WriteStringToFile(common.WriteStringToFileInput{
		Content:  resp.String(),
		Filename: f.Name(),
		Mode:     os.FileMode(0755),
	})
	if err != nil {
		return fmt.Errorf("Unable to write k3s installer to file: %w", err)
	}

	fi, err := os.Stat(f.Name())
	if err != nil {
		return fmt.Errorf("Unable to get k3s installer file size: %w", err)
	}

	if fi.Size() == 0 {
		return fmt.Errorf("Invalid k3s installer filesize")
	}

	token := getGlobalGlobalToken()
	if len(token) == 0 {
		n := 5
		b := make([]byte, n)
		if _, err := rand.Read(b); err != nil {
			return fmt.Errorf("Unable to generate random node name: %w", err)
		}

		token = strings.ToLower(fmt.Sprintf("%X", b))
		if err := CommandSet("--global", "token", token); err != nil {
			return fmt.Errorf("Unable to set k3s token: %w", err)
		}
	}

	nodeName := serverIP
	n := 5
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return fmt.Errorf("Unable to generate random node name: %w", err)
	}
	nodeName = strings.ReplaceAll(strings.ToLower(fmt.Sprintf("ip-%s-%s", nodeName, fmt.Sprintf("%X", b))), ".", "-")

	args := []string{
		// initialize the cluster
		"--cluster-init",
		// disable local-storage
		"--disable", "local-storage",
		// disable traefik so it can be installed separately
		"--disable", "traefik",
		// expose etcd metrics
		"--etcd-expose-metrics",
		// use wireguard for flannel
		"--flannel-backend=wireguard-native",
		// bind controller-manager to all interfaces
		"--kube-controller-manager-arg", "bind-address=0.0.0.0",
		// bind proxy metrics to all interfaces
		"--kube-proxy-arg", "metrics-bind-address=0.0.0.0",
		// bind scheduler to all interfaces
		"--kube-scheduler-arg", "bind-address=0.0.0.0",
		// gc terminated pods
		"--kube-controller-manager-arg", "terminated-pod-gc-threshold=10",
		// specify the node name
		"--node-name", nodeName,
		// allow access for the dokku user
		"--write-kubeconfig-mode", "0644",
		// specify a token
		"--token", token,
	}
	if taintScheduling {
		args = append(args, "--node-taint", "CriticalAddonsOnly=true:NoSchedule")
	}

	common.CommandPropertySet("scheduler-k3s", "--global", "ingress-class", ingressClass, DefaultProperties, GlobalProperties)
	if ingressClass == "nginx" {
		args = append(args, "--disable", "traefik")
	}

	common.LogInfo2Quiet("Running k3s installer")
	installerCmd, err := common.CallExecCommand(common.ExecCommandInput{
		Command:     f.Name(),
		Args:        args,
		StreamStdio: true,
	})
	if err != nil {
		return fmt.Errorf("Unable to call k3s installer command: %w", err)
	}
	if installerCmd.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from k3s installer command: %d", installerCmd.ExitCode)
	}

	clientset, err := NewKubernetesClient()
	if err != nil {
		return fmt.Errorf("Unable to create kubernetes client: %w", err)
	}

	common.LogInfo2Quiet("Waiting for node to exist")
	nodes, err := waitForNodeToExist(ctx, WaitForNodeToExistInput{
		Clientset:  clientset,
		NodeName:   nodeName,
		RetryCount: 20,
	})
	if err != nil {
		return fmt.Errorf("Error waiting for pod to exist: %w", err)
	}
	if len(nodes) == 0 {
		return fmt.Errorf("Unable to find node after initializing cluster, node will not be annotated/labeled appropriately access registry secrets")
	}

	for _, manifest := range KubernetesManifests {
		common.LogInfo2Quiet(fmt.Sprintf("Installing %s@%s", manifest.Name, manifest.Version))
		err = clientset.ApplyKubernetesManifest(ctx, ApplyKubernetesManifestInput{
			Manifest: manifest.Path,
		})
		if err != nil {
			return fmt.Errorf("Unable to apply kubernetes manifest: %w", err)
		}
	}

	for key, value := range ServerLabels {
		common.LogInfo2Quiet(fmt.Sprintf("Labeling node %s=%s", key, value))
		if err != nil {
			return fmt.Errorf("Unable to create kubernetes client: %w", err)
		}

		err = clientset.LabelNode(ctx, LabelNodeInput{
			Name:  nodeName,
			Key:   key,
			Value: value,
		})
		if err != nil {
			return fmt.Errorf("Unable to patch node: %w", err)
		}
	}

	common.LogInfo2Quiet("Installing helm charts")
	err = installHelmCharts(ctx, clientset, func(chart HelmChart) bool {
		if chart.ChartPath == "traefik" && ingressClass == "nginx" {
			return false
		}

		if chart.ChartPath == "ingress-nginx" && ingressClass == "traefik" {
			return false
		}

		return true
	})
	if err != nil {
		return fmt.Errorf("Unable to install helm charts: %w", err)
	}

	common.LogInfo2Quiet("Installing helper commands")
	err = installHelperCommands(ctx)
	if err != nil {
		return fmt.Errorf("Unable to install helper commands: %w", err)
	}

	common.LogVerboseQuiet("Done")

	return nil
}

// CommandClusterAdd adds a server to the k3s cluster
func CommandClusterAdd(role string, remoteHost string, serverIP string, allowUknownHosts bool, taintScheduling bool) error {
	if err := isK3sInstalled(); err != nil {
		return fmt.Errorf("k3s not installed, cannot add node to cluster: %w", err)
	}

	clientset, err := NewKubernetesClient()
	if err != nil {
		return fmt.Errorf("Unable to create kubernetes client: %w", err)
	}

	if err := clientset.Ping(); err != nil {
		return fmt.Errorf("kubernetes api not available, cannot add node to cluster: %w", err)
	}

	if role != "server" && role != "worker" {
		return fmt.Errorf("Invalid server-type: %s", role)
	}

	token := getGlobalGlobalToken()
	if len(token) == 0 {
		return fmt.Errorf("Missing k3s token")
	}

	if taintScheduling && role == "worker" {
		return fmt.Errorf("Taint scheduling can only be used on the server role")
	}

	if serverIP == "" {
		var err error
		serverIP, err = getServerIP()
		if err != nil {
			return fmt.Errorf("Unable to get server ip address: %w", err)
		}

		common.LogVerboseQuiet(fmt.Sprintf("Using server ip address: %s", serverIP))
	}

	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM)
	go func() {
		<-signals
		cancel()
	}()

	// todo: check if k3s is installed on the remote host

	k3sVersionCmd, err := common.CallExecCommand(common.ExecCommandInput{
		Command: "k3s",
		Args:    []string{"--version"},
	})
	if err != nil {
		return fmt.Errorf("Unable to call k3s version command: %w", err)
	}
	if k3sVersionCmd.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from k3s --version command: %d", k3sVersionCmd.ExitCode)
	}

	k3sVersion := ""
	k3sVersionLines := strings.Split(string(k3sVersionCmd.Stdout), "\n")
	if len(k3sVersionLines) > 0 {
		k3sVersionParts := strings.Split(k3sVersionLines[0], " ")
		if len(k3sVersionParts) != 4 {
			return fmt.Errorf("Unable to get k3s version from k3s --version: %s", k3sVersionCmd.Stdout)
		}
		k3sVersion = k3sVersionParts[2]
	}
	common.LogDebug(fmt.Sprintf("k3s version: %s", k3sVersion))

	common.LogInfo1(fmt.Sprintf("Joining %s to k3s cluster as %s", remoteHost, role))
	common.LogInfo2Quiet("Updating apt")
	aptUpdateCmd, err := common.CallSshCommand(common.SshCommandInput{
		Command: "apt-get",
		Args: []string{
			"update",
		},
		AllowUknownHosts: allowUknownHosts,
		RemoteHost:       remoteHost,
		StreamStdio:      true,
		Sudo:             true,
	})
	if err != nil {
		return fmt.Errorf("Unable to call apt-get update command over ssh: %w", err)
	}
	if aptUpdateCmd.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from apt-get update command over ssh: %d", aptUpdateCmd.ExitCode)
	}

	common.LogInfo2Quiet("Installing k3s dependencies")
	aptInstallCmd, err := common.CallSshCommand(common.SshCommandInput{
		Command: "apt-get",
		Args: []string{
			"-y",
			"install",
			"ca-certificates",
			"curl",
			"open-iscsi",
			"nfs-common",
			"wireguard",
		},
		AllowUknownHosts: allowUknownHosts,
		RemoteHost:       remoteHost,
		StreamStdio:      true,
		Sudo:             true,
	})
	if err != nil {
		return fmt.Errorf("Unable to call apt-get install command over ssh: %w", err)
	}
	if aptInstallCmd.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from apt-get install command over ssh: %d", aptInstallCmd.ExitCode)
	}

	common.LogInfo2Quiet("Downloading k3s installer")
	curlTask, err := common.CallSshCommand(common.SshCommandInput{
		Command: "curl",
		Args: []string{
			"-o /tmp/k3s-installer.sh",
			"https://get.k3s.io",
		},
		AllowUknownHosts: allowUknownHosts,
		RemoteHost:       remoteHost,
		StreamStdio:      true,
	})
	if err != nil {
		return fmt.Errorf("Unable to call curl command over ssh: %w", err)
	}
	if curlTask.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from curl command over ssh: %d", curlTask.ExitCode)
	}

	common.LogInfo2Quiet("Setting k3s installer permissions")
	chmodCmd, err := common.CallSshCommand(common.SshCommandInput{
		Command: "chmod",
		Args: []string{
			"0755",
			"/tmp/k3s-installer.sh",
		},
		AllowUknownHosts: allowUknownHosts,
		RemoteHost:       remoteHost,
		StreamStdio:      true,
	})
	if err != nil {
		return fmt.Errorf("Unable to call chmod command over ssh: %w", err)
	}
	if chmodCmd.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from chmod command over ssh: %d", chmodCmd.ExitCode)
	}

	u, err := url.Parse(remoteHost)
	if err != nil {
		return fmt.Errorf("failed to parse remote host: %w", err)
	}

	nodeName := u.Hostname()
	n := 5
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return fmt.Errorf("Unable to generate random node name: %w", err)
	}
	nodeName = strings.ReplaceAll(strings.ToLower(fmt.Sprintf("ip-%s-%s", nodeName, fmt.Sprintf("%X", b))), ".", "-")

	args := []string{
		// disable local-storage
		"--disable", "local-storage",
		// use wireguard for flannel
		"--flannel-backend=wireguard-native",
		// specify the node name
		"--node-name", nodeName,
		// server to connect to as the main
		"--server",
		fmt.Sprintf("https://%s:6443", serverIP),
		// specify a token
		"--token",
		token,
	}

	if role == "server" {
		args = append([]string{"server"}, args...)
		// expose etcd metrics
		args = append(args, "--etcd-expose-metrics")
		// bind controller-manager to all interfaces
		args = append(args, "--kube-controller-manager-arg", "bind-address=0.0.0.0")
		// bind proxy metrics to all interfaces
		args = append(args, "--kube-proxy-arg", "metrics-bind-address=0.0.0.0")
		// bind scheduler to all interfaces
		args = append(args, "--kube-scheduler-arg", "bind-address=0.0.0.0")
		// gc terminated pods
		args = append(args, "--kube-controller-manager-arg", "terminated-pod-gc-threshold=10")
		// allow access for the dokku user
		args = append(args, "--write-kubeconfig-mode", "0644")
	} else {
		// disable etcd on workers
		args = append(args, "--disable-etcd")
		// disable apiserver on workers
		args = append(args, "--disable-apiserver")
		// disable controller-manager on workers
		args = append(args, "--disable-controller-manager")
		// disable scheduler on workers
		args = append(args, "--disable-scheduler")
		// bind proxy metrics to all interfaces
		args = append(args, "--kube-proxy-arg", "metrics-bind-address=0.0.0.0")
	}

	if taintScheduling {
		args = append(args, "--node-taint", "CriticalAddonsOnly=true:NoSchedule")
	}

	common.LogInfo2Quiet(fmt.Sprintf("Adding %s k3s cluster", nodeName))
	joinCmd, err := common.CallSshCommand(common.SshCommandInput{
		Command:          "/tmp/k3s-installer.sh",
		Args:             args,
		AllowUknownHosts: allowUknownHosts,
		RemoteHost:       remoteHost,
		StreamStdio:      true,
		Sudo:             true,
	})
	if err != nil {
		return fmt.Errorf("Unable to call k3s installer command over ssh: %w", err)
	}
	if joinCmd.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from k3s installer command over ssh: %d", joinCmd.ExitCode)
	}

	common.LogInfo2Quiet("Waiting for node to exist")
	nodes, err := waitForNodeToExist(ctx, WaitForNodeToExistInput{
		Clientset:  clientset,
		NodeName:   nodeName,
		RetryCount: 20,
	})
	if err != nil {
		return fmt.Errorf("Error waiting for pod to exist: %w", err)
	}
	if len(nodes) == 0 {
		return fmt.Errorf("Unable to find node after joining cluster, node will not be annotated/labeled appropriately access registry secrets")
	}

	labels := ServerLabels
	if role == "worker" {
		labels = WorkerLabels
	}

	for key, value := range labels {
		common.LogInfo2Quiet(fmt.Sprintf("Labeling node %s=%s", key, value))
		if err != nil {
			return fmt.Errorf("Unable to create kubernetes client: %w", err)
		}

		err = clientset.LabelNode(ctx, LabelNodeInput{
			Name:  nodeName,
			Key:   key,
			Value: value,
		})
		if err != nil {
			return fmt.Errorf("Unable to patch node: %w", err)
		}
	}

	common.LogInfo2Quiet("Annotating node with connection information")
	err = clientset.AnnotateNode(ctx, AnnotateNodeInput{
		Name:  nodes[0].Name,
		Key:   "dokku.com/remote-host",
		Value: remoteHost,
	})
	if err != nil {
		return fmt.Errorf("Unable to patch node: %w", err)
	}

	common.LogVerboseQuiet("Done")
	return nil
}

// CommandClusterList lists the nodes in the k3s cluster
func CommandClusterList(format string) error {
	if format != "stdout" && format != "json" {
		return fmt.Errorf("Invalid format: %s", format)
	}

	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM)
	go func() {
		<-signals
		cancel()
	}()

	clientset, err := NewKubernetesClient()
	if err != nil {
		return fmt.Errorf("Unable to create kubernetes client: %w", err)
	}

	if err := clientset.Ping(); err != nil {
		return fmt.Errorf("kubernetes api not available, cannot list cluster nodes: %w", err)
	}

	nodes, err := clientset.ListNodes(ctx, ListNodesInput{})
	if err != nil {
		return fmt.Errorf("Unable to list nodes: %w", err)
	}

	output := []Node{}
	for _, node := range nodes {
		output = append(output, kubernetesNodeToNode(node))
	}

	if format == "stdout" {
		lines := []string{"name|ready|roles|version"}
		for _, node := range output {
			lines = append(lines, node.String())
		}

		columnized := columnize.SimpleFormat(lines)
		fmt.Println(columnized)
		return nil
	}

	b, err := json.Marshal(output)
	if err != nil {
		return fmt.Errorf("Unable to marshal json: %w", err)
	}

	fmt.Println(string(b))
	return nil
}

// CommandClusterRemove removes a node from the k3s cluster
func CommandClusterRemove(nodeName string) error {
	if err := isK3sInstalled(); err != nil {
		return fmt.Errorf("k3s not installed, cannot remove node from cluster: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM)
	go func() {
		<-signals
		cancel()
	}()

	common.LogInfo1Quiet(fmt.Sprintf("Removing %s from k3s cluster", nodeName))
	clientset, err := NewKubernetesClient()
	if err != nil {
		return fmt.Errorf("Unable to create kubernetes client: %w", err)
	}

	if err := clientset.Ping(); err != nil {
		return fmt.Errorf("kubernetes api not available: %w", err)
	}

	common.LogVerboseQuiet("Getting node remote connection information")
	node, err := clientset.GetNode(ctx, GetNodeInput{
		Name: nodeName,
	})
	if err != nil {
		return fmt.Errorf("Unable to get node: %w", err)
	}

	common.LogVerboseQuiet("Checking if node is a remote node managed by Dokku")
	if node.RemoteHost == "" {
		return fmt.Errorf("Node %s is not a remote node managed by Dokku", nodeName)
	}

	common.LogVerboseQuiet("Uninstalling k3s on remote host")
	removeCmd, err := common.CallSshCommand(common.SshCommandInput{
		Command:          "/usr/local/bin/k3s-uninstall.sh",
		Args:             []string{},
		AllowUknownHosts: true,
		RemoteHost:       node.RemoteHost,
		StreamStdio:      true,
		Sudo:             true,
	})
	if err != nil {
		return fmt.Errorf("Unable to call k3s uninstall command over ssh: %w", err)
	}

	if removeCmd.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from k3s uninstall command over ssh: %d", removeCmd.ExitCode)
	}

	common.LogVerboseQuiet("Deleting node from k3s cluster")
	err = clientset.DeleteNode(ctx, DeleteNodeInput{
		Name: nodeName,
	})
	if err != nil {
		return fmt.Errorf("Unable to delete node: %w", err)
	}

	common.LogVerboseQuiet("Done")

	return nil
}

// CommandEnsureCharts ensures that the required helm charts are installed
func CommandEnsureCharts(forceInstall bool, forceChartNames []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM)
	go func() {
		<-signals
		cancel()
	}()

	clientset, err := NewKubernetesClient()
	if err != nil {
		return fmt.Errorf("Unable to create kubernetes client: %w", err)
	}

	ingressClass := getGlobalIngressClass()

	namespacedHelmAgents := map[string]*HelmAgent{}
	for _, chart := range HelmCharts {
		_, ok := namespacedHelmAgents[chart.Namespace]
		if !ok {
			helmAgent, err := NewHelmAgent(chart.Namespace, DeployLogPrinter)
			if err != nil {
				common.LogWarn(fmt.Sprintf("Unable to create helm agent: %s", err.Error()))
				return err
			}
			namespacedHelmAgents[chart.Namespace] = helmAgent
		}
	}

	common.LogInfo2Quiet("Installing helm charts")
	err = installHelmCharts(ctx, clientset, func(chart HelmChart) bool {
		common.LogInfo1(fmt.Sprintf("Processing chart %s@%s", chart.ReleaseName, chart.Version))
		if chart.ChartPath == "traefik" && ingressClass == "nginx" {
			common.LogVerbose("Skipping chart due to ingress-class mismatch")
			return false
		}

		if chart.ChartPath == "ingress-nginx" && ingressClass == "traefik" {
			common.LogVerbose("Skipping chart due to ingress-class mismatch")
			return false
		}

		if forceInstall {
			common.LogVerbose("Force installing chart")
			return true
		}

		if len(forceChartNames) > 0 && slices.Contains(forceChartNames, chart.ReleaseName) {
			common.LogVerbose("Force installing chart due to flag")
			return true
		}

		helmAgent := namespacedHelmAgents[chart.Namespace]
		latestRevision, err := helmAgent.InstalledRevision(chart.ReleaseName)
		if err != nil {
			common.LogWarn(fmt.Sprintf("Unable to get installed revision: %s", err))
			return false
		}

		if latestRevision.Name == "" {
			common.LogVerbose("Installing missing chart")
			return true
		}

		if latestRevision.Version != chart.Version {
			common.LogVerbose(fmt.Sprintf("Installing chart due to version mismatch: %s != %s", latestRevision.AppVersion, chart.Version))
			return false
		}

		common.LogVerbose("Skipping chart: already installed")
		return false
	})
	if err != nil {
		return fmt.Errorf("Unable to install helm charts: %w", err)
	}

	for _, manifest := range KubernetesManifests {
		common.LogInfo2Quiet(fmt.Sprintf("Installing %s@%s", manifest.Name, manifest.Version))
		err = clientset.ApplyKubernetesManifest(ctx, ApplyKubernetesManifestInput{
			Manifest: manifest.Path,
		})
		if err != nil {
			return fmt.Errorf("Unable to apply kubernetes manifest: %w", err)
		}
	}

	common.LogInfo2Quiet("Done")

	return nil
}

// CommandLabelsSet set or clear a scheduler-k3s label for an app
func CommandLabelsSet(appName string, processType string, resourceType string, key string, value string) error {
	if resourceType == "" {
		return fmt.Errorf("Missing resource-type")
	}

	if processType == "" {
		processType = GlobalProcessType
	}

	property := fmt.Sprintf("labels.%s.%s", processType, resourceType)
	labelsList, err := common.PropertyListGet("scheduler-k3s", appName, property)
	if err != nil {
		return fmt.Errorf("Unable to get property list: %w", err)
	}

	labels := []string{}
	for _, annotation := range labelsList {
		parts := strings.SplitN(annotation, ": ", 2)
		if len(parts) != 2 {
			return fmt.Errorf("Invalid annotation: %s", annotation)
		}
		if key == parts[0] {
			continue
		}

		labels = append(labels, annotation)
	}

	if value != "" {
		labels = append(labels, fmt.Sprintf("%s: %s", key, value))
	}

	sort.Strings(labels)
	if err := common.PropertyListWrite("scheduler-k3s", appName, property, labels); err != nil {
		return fmt.Errorf("Unable to write property list: %w", err)
	}

	return nil
}

// CommandReport displays a scheduler-k3s report for one or more apps
func CommandReport(appName string, format string, infoFlag string) error {
	if len(appName) == 0 {
		apps, err := common.DokkuApps()
		if err != nil {
			if errors.Is(err, common.NoAppsExist) {
				common.LogWarn(err.Error())
				return nil
			}
			return err
		}
		for _, appName := range apps {
			if err := ReportSingleApp(appName, format, infoFlag); err != nil {
				return err
			}
		}
		return nil
	}

	return ReportSingleApp(appName, format, infoFlag)
}

// CommandSet set or clear a scheduler-k3s property for an app
func CommandSet(appName string, property string, value string) error {
	validProperties := DefaultProperties
	globalProperties := GlobalProperties
	if strings.HasPrefix(property, "chart.") {
		if appName != "--global" {
			return fmt.Errorf("Chart properties can only be set globally")
		}

		chartParts := strings.SplitN(property, ".", 3)
		if len(chartParts) != 3 {
			return fmt.Errorf("Invalid chart property, expected format: chart.$CHART_NAME.$PROPERTY: %s", property)
		}

		if chartParts[1] == "" {
			return fmt.Errorf("Invalid chart property, missing chart name")
		}

		chartName := chartParts[1]
		for _, chart := range HelmCharts {
			if chart.ReleaseName == chartName {
				validProperties[property] = ""
				globalProperties[property] = true
				break
			}
		}

		if _, ok := validProperties[property]; !ok {
			return fmt.Errorf("Invalid chart property, no matching chart found: %s", property)
		}
	}

	common.CommandPropertySet("scheduler-k3s", appName, property, value, validProperties, globalProperties)

	letsencryptProperties := map[string]bool{
		"letsencrypt-email-prod": true,
		"letsencrypt-email-stag": true,
		"letsencrypt-server":     true,
	}
	if appName == "--global" && letsencryptProperties[property] {
		return applyClusterIssuers(context.Background())
	}

	return nil
}

// CommandShowKubeconfig displays the kubeconfig file contents
func CommandShowKubeconfig() error {
	kubeconfigPath := getKubeconfigPath()
	if !common.FileExists(kubeconfigPath) {
		return fmt.Errorf("Kubeconfig file does not exist: %s", kubeconfigPath)
	}

	b, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("Unable to read kubeconfig file: %w", err)
	}

	fmt.Println(string(b))

	return nil
}

func CommandUninstall() error {
	if err := isK3sInstalled(); err != nil {
		return fmt.Errorf("k3s not installed, cannot uninstall: %w", err)
	}

	common.LogInfo1("Uninstalling k3s")
	uninstallerCmd, err := common.CallExecCommand(common.ExecCommandInput{
		Command:     "/usr/local/bin/k3s-uninstall.sh",
		StreamStdio: true,
	})
	if err != nil {
		return fmt.Errorf("Unable to call k3s uninstaller command: %w", err)
	}
	if uninstallerCmd.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from k3s uninstaller command: %d", uninstallerCmd.ExitCode)
	}

	common.LogInfo2Quiet("Removing k3s dependencies")
	return uninstallHelperCommands(context.Background())
}

func CommandAddPVC(pvcName string, namespace string, accessMode string, storageSize string, storageClass string) error {
	chartDir, err := os.MkdirTemp("", "pvc-chart-")
	if err != nil {
		return fmt.Errorf("Error creating pvc chart directory: %w", err)
	}
	defer os.RemoveAll(chartDir)

	if err := os.MkdirAll(filepath.Join(chartDir, "templates"), os.FileMode(0755)); err != nil {
		return fmt.Errorf("Error creating pvc chart templates directory: %w", err)
	}

	// create the chart.yaml
	chart := &Chart{
		ApiVersion: "v2",
		AppVersion: "1.0.0",
		Icon:       "https://dokku.com/assets/dokku-logo.svg",
		Name:       "PersistentVolumeClaim",
		Version:    "0.0.1",
	}

	err = writeYaml(WriteYamlInput{
		Object: chart,
		Path:   filepath.Join(chartDir, "Chart.yaml"),
	})
	if err != nil {
		return fmt.Errorf("Error writing PersistentVolumeClaim chart: %w", err)
	}

	// create the values.yaml
	values := PersistentVolumeClaim{
		Name:         pvcName,
		AccessMode:   accessMode,
		Storage:      storageSize,
		StorageClass: storageClass,
		Namespace:    namespace,
	}

	err = writeYaml(WriteYamlInput{
		Object: values,
		Path:   filepath.Join(chartDir, "values.yaml"),
	})
	if err != nil {
		return fmt.Errorf("Error writing chart: %w", err)
	}
	if os.Getenv("DOKKU_TRACE") == "1" {
		common.CatFile(filepath.Join(chartDir, "values.yaml"))
	}

	b, err := templates.ReadFile("templates/chart/pvc.yaml")
	if err != nil {
		return fmt.Errorf("Error reading PVC template: %w", err)
	}
	common.CatFile("templates/chart/pvc.yaml")

	// write pvc.yaml
	pvcFile := filepath.Join(chartDir, "templates", "pvc.yaml")
	err = os.WriteFile(pvcFile, []byte(b), os.FileMode(0644))
	if err != nil {
		return fmt.Errorf("Error writing cron job template: %w", err)
	}
	if os.Getenv("DOKKU_TRACE") == "1" {
		common.CatFile(pvcFile)
	}

	// add _helpers.tpl
	b, err = templates.ReadFile("templates/chart/_helpers.tpl")
	if err != nil {
		return fmt.Errorf("Error reading _helpers template: %w", err)
	}

	helpersFile := filepath.Join(chartDir, "templates", "_helpers.tpl")
	err = os.WriteFile(helpersFile, b, os.FileMode(0644))
	if err != nil {
		return fmt.Errorf("Error writing _helpers template: %w", err)
	}

	// install the chart
	helmAgent, err := NewHelmAgent(namespace, DeployLogPrinter)
	if err != nil {
		return fmt.Errorf("Error creating helm agent: %w", err)
	}

	chartPath, err := filepath.Abs(chartDir)
	if err != nil {
		return fmt.Errorf("Error getting chart path: %w", err)
	}

	timeoutDuration, err := time.ParseDuration("300s")
	if err != nil {
		return fmt.Errorf("Error parsing deploy timeout duration: %w", err)
	}

	err = helmAgent.InstallOrUpgradeChart(context.Background(), ChartInput{
		ChartPath:         chartPath,
		Namespace:         namespace,
		ReleaseName:       fmt.Sprintf("pvc-%s-%s", namespace, pvcName),
		RollbackOnFailure: true,
		Timeout:           timeoutDuration,
		Wait:              true,
	})
	if err != nil {
		return fmt.Errorf("Error installing pvc chart: %w", err)
	}

	common.LogInfo1Quiet("Applied pvc chart")

	return nil
}

func CommandRemovePVC(pvcName string, namespace string) error {
	clientset, err := NewKubernetesClient()
	if err != nil {
		if isK3sKubernetes() {
			if err := isK3sInstalled(); err != nil {
				common.LogWarn("k3s is not installed, skipping")
				return nil
			}
		}
		return fmt.Errorf("Error creating kubernetes client: %w", err)
	}

	if err := clientset.Ping(); err != nil {
		return fmt.Errorf("kubernetes api not available: %w", err)
	}

	pvcInput := PvcInput{
		Name:      pvcName,
		Namespace: namespace,
	}

	// Retrieve the PVC
	pvc, err := clientset.GetPvc(context.Background(), pvcInput)
	if err != nil {
		return fmt.Errorf("failed to get PVC %s in namespace %s: %w", pvcInput.Name, pvcInput.Namespace, err)
	}
	// Check if the annotation exists and is set to "true"
	if val, exists := pvc.Annotations["dokku.com/managed"]; !exists || val != "true" {
		return fmt.Errorf("PVC %s in namespace %s is not managed by dokku (annotation missing or incorrect)", pvcInput.Name, pvcInput.Namespace)
	}

	err = clientset.DeletePvc(context.Background(), pvcInput)
	if err != nil {
		return fmt.Errorf("Error deleting pvc: %w", err)
	}

	return nil
}

func CommandMountPVC(appName string, processType string, pvcName string, mountPath string, subPath string, readOnly bool, chown string) error {
	clientset, err := NewKubernetesClient()
	if err != nil {
		if isK3sKubernetes() {
			if err := isK3sInstalled(); err != nil {
				common.LogWarn("k3s is not installed, skipping")
				return nil
			}
		}
		return fmt.Errorf("Error creating kubernetes client: %w", err)
	}

	if err := clientset.Ping(); err != nil {
		return fmt.Errorf("kubernetes api not available: %w", err)
	}

	// 1. get namespace for the app
	namespace := getComputedNamespace(appName)
	// 2. check if pvcName exists in this namespace
	pvcInput := PvcInput{
		Name:      pvcName,
		Namespace: namespace,
	}
	// Retrieve the PVC
	_, err = clientset.GetPvc(context.Background(), pvcInput)
	if err != nil {
		return fmt.Errorf("failed to get PVC %s in namespace %s: %w", pvcInput.Name, pvcInput.Namespace, err)
	}
	// TODO: 2.2. maybe check if pvc has dokku.com/managed ??
	// 3. add to properties
	volume := ProcessVolume{
		Name:      fmt.Sprintf("%s-%s-%s", appName, processType, pvcName),
		Type:      "persistentVolumeClaim",
		ClaimName: pvcName,
		MountPath: mountPath,
	}
	if len(subPath) > 0 {
		volume.SubPath = subPath
	}
	if readOnly {
		volume.ReadOnly = readOnly
	}
	if len(chown) > 0 {
		volume.Chown = chown
	}
	// Create an empty slice of ProcessVolume
	var volumes []ProcessVolume
	// get already defined volumes and add above
	propertyName := fmt.Sprintf("volumes.%s", processType)
	err = yaml.Unmarshal([]byte(common.PropertyGet("scheduler-k3s", appName, propertyName)), &volumes)
	if err != nil {
		return fmt.Errorf("failed to decode YAML in properties: %w", err)
	}
	// Check if the volume already exists
	for _, v := range volumes {
		if v.Name == volume.Name {
			return fmt.Errorf("Volume %s already exists, skipping append.", volume.Name)
		}
	}
	volumes = append(volumes, volume)
	volumesYaml, err := yaml.Marshal(&volumes)
	if err != nil {
		return fmt.Errorf("failed to marshal PVC %s in namespace %s: %w", pvcName, namespace, err)
	}
	err = common.PropertyWrite("scheduler-k3s", appName, propertyName, string(volumesYaml))
	if err != nil {
		return fmt.Errorf("failed to store property PVC %s in namespace %s: %w", pvcName, namespace, err)
	}

	return nil
}

func CommandUnMountPVC(appName string, processType string, pvcName string, mountPath string) error {
	// Create an empty slice of ProcessVolume
	var volumes []ProcessVolume
	// get already defined volumes and add above
	propertyName := fmt.Sprintf("volumes.%s", processType)
	err := yaml.Unmarshal([]byte(common.PropertyGet("scheduler-k3s", appName, propertyName)), &volumes)
	if err != nil {
		return fmt.Errorf("failed to decode YAML in properties: %w", err)
	}

	// Create a new slice without the volume to delete
	filteredVolumes := []ProcessVolume{}
	volName := fmt.Sprintf("%s-%s-%s", appName, processType, pvcName)
	for _, v := range volumes {
		if v.Name != volName || v.MountPath != mountPath {
			filteredVolumes = append(filteredVolumes, v)
		}
	}
	volumesYaml, err := yaml.Marshal(&filteredVolumes)
	if err != nil {
		return fmt.Errorf("failed to marshal delete PVC %s: %w", pvcName, err)
	}
	err = common.PropertyWrite("scheduler-k3s", appName, propertyName, string(volumesYaml))
	if err != nil {
		return fmt.Errorf("failed to store property PVC %s: %w", pvcName, err)
	}

	return nil
}
