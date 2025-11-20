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
	"regexp"
	"slices"
	"sort"
	"strings"
	"syscall"

	"github.com/dokku/dokku/plugins/common"
	resty "github.com/go-resty/resty/v2"
	"github.com/ryanuber/columnize"
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
func CommandClusterAdd(profileName string, role string, remoteHost string, serverIP string, allowUknownHosts bool, taintScheduling bool, kubeletArgs []string) error {
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

	incomingProfile := NodeProfile{}

	if profileName != "" {
		properties := common.PropertyGetDefault("scheduler-k3s", "--global", fmt.Sprintf("node-profile-%s", profileName), "")
		if properties == "" {
			return fmt.Errorf("Node profile %s not found", profileName)
		}

		err = json.Unmarshal([]byte(properties), &incomingProfile)
		if err != nil {
			return fmt.Errorf("Unable to unmarshal node profile: %w", err)
		}
	}

	if role != "" {
		incomingProfile.Role = role
	}

	if allowUknownHosts {
		incomingProfile.AllowUknownHosts = allowUknownHosts
	}

	if taintScheduling {
		incomingProfile.TaintScheduling = taintScheduling
	}

	if len(kubeletArgs) > 0 {
		incomingProfile.KubeletArgs = kubeletArgs
	}

	if incomingProfile.Role != "server" && incomingProfile.Role != "worker" {
		return fmt.Errorf("Invalid role: %s", incomingProfile.Role)
	}

	token := getGlobalGlobalToken()
	if len(token) == 0 {
		return fmt.Errorf("Missing k3s token")
	}

	if incomingProfile.TaintScheduling && incomingProfile.Role == "worker" {
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

	common.LogInfo1(fmt.Sprintf("Joining %s to k3s cluster as %s", remoteHost, incomingProfile.Role))
	common.LogInfo2Quiet("Updating apt")
	aptUpdateCmd, err := common.CallSshCommand(common.SshCommandInput{
		Command: "apt-get",
		Args: []string{
			"update",
		},
		AllowUknownHosts: incomingProfile.AllowUknownHosts,
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
		AllowUknownHosts: incomingProfile.AllowUknownHosts,
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
		AllowUknownHosts: incomingProfile.AllowUknownHosts,
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
		AllowUknownHosts: incomingProfile.AllowUknownHosts,
		RemoteHost:       remoteHost,
		StreamStdio:      true,
	})
	if err != nil {
		return fmt.Errorf("Unable to call chmod command over ssh: %w", err)
	}
	if chmodCmd.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from chmod command over ssh: %d", chmodCmd.ExitCode)
	}

	common.LogInfo2Quiet("Ensuring compatible k3s version for node")
	lowestNodeVersion, err := clientset.GetLowestNodeVersion(ctx, ListNodesInput{
		LabelSelector: "node-role.kubernetes.io/master=true",
	})
	if err != nil {
		return fmt.Errorf("Unable to get lowest node version: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "k3s-installer-*.sh")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	scriptContent := fmt.Sprintf(`#!/usr/bin/env bash
set -x
export INSTALL_K3S_VERSION=%s

/tmp/k3s-installer.sh "$@"`, lowestNodeVersion)

	if _, err := tmpFile.WriteString(scriptContent); err != nil {
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}
	tmpFile.Close()

	sftpCopyCmd, err := common.CallSftpCopy(common.SftpCopyInput{
		AllowUknownHosts: incomingProfile.AllowUknownHosts,
		DestinationPath:  "/tmp/k3s-installer-executor.sh",
		RemoteHost:       remoteHost,
		SourcePath:       tmpFile.Name(),
	})
	if err != nil {
		return fmt.Errorf("Unable to copy installer script via sftp: %w", err)
	}
	if sftpCopyCmd.ExitErr != nil {
		return fmt.Errorf("Invalid exit code from sftp copy command: %d", sftpCopyCmd.ExitErr)
	}

	chmodExecutorCmd, err := common.CallSshCommand(common.SshCommandInput{
		Command: "chmod",
		Args: []string{
			"0755",
			"/tmp/k3s-installer-executor.sh",
		},
		AllowUknownHosts: incomingProfile.AllowUknownHosts,
		RemoteHost:       remoteHost,
		StreamStdio:      true,
	})
	if err != nil {
		return fmt.Errorf("Unable to make installer script executable via ssh: %w", err)
	}
	if chmodExecutorCmd.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from chmod command via ssh: %d", chmodExecutorCmd.ExitCode)
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

	if incomingProfile.Role == "server" {
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

	if incomingProfile.TaintScheduling {
		args = append(args, "--node-taint", "CriticalAddonsOnly=true:NoSchedule")
	}

	for _, kubeletArg := range incomingProfile.KubeletArgs {
		args = append(args, "--kubelet-arg", kubeletArg)
	}

	common.LogInfo2Quiet(fmt.Sprintf("Adding %s k3s cluster", nodeName))
	joinCmd, err := common.CallSshCommand(common.SshCommandInput{
		Command:          "/tmp/k3s-installer-executor.sh",
		Args:             args,
		AllowUknownHosts: incomingProfile.AllowUknownHosts,
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
	if incomingProfile.Role == "worker" {
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

// CommandProfilesAdd adds a node profile to the k3s cluster
func CommandProfilesAdd(profileName string, role string, allowUknownHosts bool, taintScheduling bool, kubeletArgs []string) error {
	if role != "server" && role != "worker" {
		return fmt.Errorf("Invalid role: %s", role)
	}

	if profileName == "" {
		return fmt.Errorf("Missing profile name")
	}

	// profile names must only contain alphanumeric characters and dashes and cannot start with a dash
	if !regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?$`).MatchString(profileName) {
		return fmt.Errorf("Invalid profile name, must only contain alphanumeric characters and dashes and cannot start with a dash: %s", profileName)
	}

	// ensure profile names are no longer than 32 characters
	if len(profileName) > 32 {
		return fmt.Errorf("Profile name is too long, must be less than 32 characters: %s", profileName)
	}

	profile := NodeProfile{
		Name:             profileName,
		Role:             role,
		AllowUknownHosts: allowUknownHosts,
		TaintScheduling:  taintScheduling,
		KubeletArgs:      kubeletArgs,
	}

	data, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("Unable to marshal node profile to json: %w", err)
	}

	if err := common.PropertyWrite("scheduler-k3s", "--global", fmt.Sprintf("node-profile-%s.json", profileName), string(data)); err != nil {
		return fmt.Errorf("Unable to write node profile: %w", err)
	}

	common.LogInfo1(fmt.Sprintf("Node profile %s added", profileName))

	return nil
}

// CommandProfilesList lists the node profiles in the k3s cluster
func CommandProfilesList(format string) error {
	if format != "stdout" && format != "json" {
		return fmt.Errorf("Invalid format: %s", format)
	}

	properties, err := common.PropertyGetAllByPrefix("scheduler-k3s", "--global", "node-profile-")
	if err != nil {
		return fmt.Errorf("Unable to get node profiles: %w", err)
	}

	output := []NodeProfile{}
	for _, data := range properties {
		var profile NodeProfile
		err := json.Unmarshal([]byte(data), &profile)
		if err != nil {
			return fmt.Errorf("Unable to unmarshal node profile: %w", err)
		}

		output = append(output, profile)
	}

	if format == "stdout" {
		lines := []string{"name|role"}
		for _, profile := range output {
			lines = append(lines, fmt.Sprintf("%s|%s", profile.Name, profile.Role))
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

// CommandProfilesRemove removes a node profile from the k3s cluster
func CommandProfilesRemove(profileName string) error {
	if profileName == "" {
		return fmt.Errorf("Missing profile name")
	}

	// profile names must only contain alphanumeric characters and dashes and cannot start with a dash
	if !regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?$`).MatchString(profileName) {
		return fmt.Errorf("Invalid profile name, must only contain alphanumeric characters and dashes and cannot start with a dash: %s", profileName)
	}

	// ensure profile names are no longer than 32 characters
	if len(profileName) > 32 {
		return fmt.Errorf("Profile name is too long, must be less than 32 characters: %s", profileName)
	}

	if err := common.PropertyDelete("scheduler-k3s", "--global", fmt.Sprintf("node-profile-%s.json", profileName)); err != nil {
		return fmt.Errorf("Unable to delete node profile: %w", err)
	}

	common.LogInfo1(fmt.Sprintf("Node profile %s removed", profileName))
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
