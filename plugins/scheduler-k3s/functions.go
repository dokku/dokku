package scheduler_k3s

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	appjson "github.com/dokku/dokku/plugins/app-json"
	"github.com/dokku/dokku/plugins/common"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/util/term"
	"k8s.io/kubernetes/pkg/client/conditions"
	"mvdan.cc/sh/v3/shell"
)

// EnterPodInput contains all the information needed to enter a pod
type EnterPodInput struct {
	// Clientset is the kubernetes clientset
	Clientset KubernetesClient

	// Command is the command to run
	Command []string

	// Entrypoint is the entrypoint to run
	Entrypoint string

	// SelectedContainerName is the container name to enter
	SelectedContainerName string

	// SelectedPod is the pod to enter
	SelectedPod v1.Pod

	// WaitTimeout is the timeout to wait for the pod to be ready
	WaitTimeout int
}

// Node contains information about a node
type Node struct {
	// Name is the name of the node
	Name string

	// Roles is the roles of the node
	Roles []string

	// Ready is whether the node is ready
	Ready bool

	// RemoteHost is the remote host
	RemoteHost string

	// Version is the version of the node
	Version string
}

// String returns a string representation of the node
func (n Node) String() string {
	return fmt.Sprintf("%s|%s|%s|%s", n.Name, strconv.FormatBool(n.Ready), strings.Join(n.Roles, ","), n.Version)
}

// StartCommandInput contains all the information needed to get the start command
type StartCommandInput struct {
	// AppName is the name of the app
	AppName string
	// ProcessType is the process type
	ProcessType string
	// ImageSourceType is the image source type
	ImageSourceType string
	// Port is the port
	Port int32
	// Env is the environment variables
	Env map[string]string
}

// StartCommandOutput contains the start command
type StartCommandOutput struct {
	// Command is the start command
	Command []string
}

type WaitForNodeToExistInput struct {
	Clientset  KubernetesClient
	Namespace  string
	RetryCount int
	NodeName   string
}

type WaitForPodBySelectorRunningInput struct {
	Clientset     KubernetesClient
	Namespace     string
	LabelSelector string
	PodName       string
	Timeout       int
	Waiter        func(ctx context.Context, clientset KubernetesClient, podName, namespace string) wait.ConditionWithContextFunc
}

type WaitForPodToExistInput struct {
	Clientset     KubernetesClient
	Namespace     string
	RetryCount    int
	PodName       string
	LabelSelector string
}

func createKubernetesNamespace(ctx context.Context, namespaceName string) error {
	clientset, err := NewKubernetesClient()
	if err != nil {
		return err
	}

	namespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
			Annotations: map[string]string{
				"dokku.com/managed": "true",
			},
			Labels: map[string]string{
				"dokku.com/managed": "true",
			},
		},
	}
	_, err = clientset.CreateNamespace(ctx, CreateNamespaceInput{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	return nil
}

func enterPod(ctx context.Context, input EnterPodInput) error {
	coreclient, err := corev1client.NewForConfig(&input.Clientset.RestConfig)
	if err != nil {
		return fmt.Errorf("Error creating corev1 client: %w", err)
	}

	labelSelector := []string{}
	for k, v := range input.SelectedPod.Labels {
		labelSelector = append(labelSelector, fmt.Sprintf("%s=%s", k, v))
	}

	if input.WaitTimeout > 0 {
		input.WaitTimeout = 5
	}

	err = waitForPodBySelectorRunning(ctx, WaitForPodBySelectorRunningInput{
		Clientset:     input.Clientset,
		Namespace:     input.SelectedPod.Namespace,
		LabelSelector: strings.Join(labelSelector, ","),
		PodName:       input.SelectedPod.Name,
		Timeout:       input.WaitTimeout,
		Waiter:        isPodReady,
	})
	if err != nil {
		return fmt.Errorf("Error waiting for pod to be ready: %w", err)
	}

	defaultContainerName, hasDefaultContainer := input.SelectedPod.Annotations["kubectl.kubernetes.io/default-container"]
	if input.SelectedContainerName == "" && hasDefaultContainer {
		input.SelectedContainerName = defaultContainerName
	}
	if input.SelectedContainerName == "" {
		return fmt.Errorf("No container specified and no default container found")
	}

	req := coreclient.RESTClient().Post().
		Resource("pods").
		Namespace(input.SelectedPod.Namespace).
		Name(input.SelectedPod.Name).
		SubResource("exec")

	req.Param("container", input.SelectedContainerName)
	req.Param("stdin", "true")
	req.Param("stdout", "true")
	req.Param("stderr", "true")
	req.Param("tty", "true")

	if input.Entrypoint != "" {
		req.Param("command", input.Entrypoint)
	}
	for _, cmd := range input.Command {
		req.Param("command", cmd)
	}

	t := term.TTY{
		In:  os.Stdin,
		Out: os.Stdout,
		Raw: true,
	}
	size := t.GetSize()
	sizeQueue := t.MonitorSize(size)

	return t.Safe(func() error {
		exec, err := remotecommand.NewSPDYExecutor(&input.Clientset.RestConfig, "POST", req.URL())
		if err != nil {
			return fmt.Errorf("Error creating executor: %w", err)
		}

		return exec.StreamWithContext(ctx, remotecommand.StreamOptions{
			Stdin:             os.Stdin,
			Stdout:            os.Stdout,
			Stderr:            os.Stderr,
			Tty:               true,
			TerminalSizeQueue: sizeQueue,
		})
	})
}

func extractStartCommand(input StartCommandInput) string {
	command := ""
	if input.ImageSourceType == "herokuish" {
		return "/start " + input.ProcessType
	}

	resp, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:       "config-get",
		Args:          []string{input.AppName, "DOKKU_START_CMD"},
		CaptureOutput: true,
		StreamStdio:   false,
	})
	if err == nil && resp.ExitCode == 0 && len(resp.Stdout) > 0 {
		command = strings.TrimSpace(resp.Stdout)
	}

	if input.ImageSourceType == "dockerfile" {
		resp, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
			Trigger:       "config-get",
			Args:          []string{input.AppName, "DOKKU_DOCKERFILE_START_CMD"},
			CaptureOutput: true,
			StreamStdio:   false,
		})
		if err == nil && resp.ExitCode == 0 && len(resp.Stdout) > 0 {
			command = strings.TrimSpace(resp.Stdout)
		}
	}

	if command == "" {
		procfileStartCommand, _ := common.PlugnTriggerOutputAsString("procfile-get-command", []string{input.AppName, input.ProcessType, fmt.Sprint(input.Port)}...)
		if procfileStartCommand != "" {
			command = procfileStartCommand
		}
	}

	return command
}

func getDeployTimeout(appName string) string {
	return common.PropertyGetDefault("scheduler-k3s", appName, "deploy-timeout", "")
}

func getGlobalDeployTimeout() string {
	return common.PropertyGetDefault("scheduler-k3s", "--global", "deploy-timeout", "300s")
}

func getComputedDeployTimeout(appName string) string {
	deployTimeout := getDeployTimeout(appName)
	if deployTimeout == "" {
		deployTimeout = getGlobalDeployTimeout()
	}

	return deployTimeout
}

func getImagePullSecrets(appName string) string {
	return common.PropertyGetDefault("scheduler-k3s", appName, "image-pull-secrets", "")
}

func getGlobalImagePullSecrets() string {
	return common.PropertyGetDefault("scheduler-k3s", "--global", "image-pull-secrets", "")
}

func getComputedImagePullSecrets(appName string) string {
	imagePullSecrets := getImagePullSecrets(appName)
	if imagePullSecrets == "" {
		imagePullSecrets = getGlobalImagePullSecrets()
	}

	return imagePullSecrets
}

func getLetsencryptServer(appName string) string {
	return common.PropertyGetDefault("scheduler-k3s", appName, "letsencrypt-server", "")
}

func getGlobalLetsencryptServer(appName string) string {
	return common.PropertyGetDefault("scheduler-k3s", appName, "letsencrypt-server", "prod")
}

func getComputedLetsencryptServer(appName string) string {
	letsencryptServer := getLetsencryptServer(appName)
	if letsencryptServer == "" {
		letsencryptServer = getGlobalLetsencryptServer(appName)
	}

	return letsencryptServer
}

func getGlobalLetsencryptEmailProd() string {
	return common.PropertyGetDefault("scheduler-k3s", "--global", "letsencrypt-email-prod", "")
}

func getGlobalLetsencryptEmailStag() string {
	return common.PropertyGetDefault("scheduler-k3s", "--global", "letsencrypt-email-stag", "")
}

func getNamespace(appName string) string {
	return common.PropertyGetDefault("scheduler-k3s", appName, "namespace", "")
}

func getGlobalNamespace() string {
	return common.PropertyGetDefault("scheduler-k3s", "--global", "namespace", "default")
}

func getComputedNamespace(appName string) string {
	namespace := getNamespace(appName)
	if namespace == "" {
		namespace = getGlobalNamespace()
	}

	return namespace
}

func getGlobalNetworkInterface() string {
	return common.PropertyGetDefault("scheduler-k3s", "--global", "network-interface", "eth0")
}

func getRollbackOnFailure(appName string) string {
	return common.PropertyGetDefault("scheduler-k3s", appName, "rollback-on-failure", "")
}

func getGlobalRollbackOnFailure() string {
	return common.PropertyGetDefault("scheduler-k3s", "--global", "rollback-on-failure", "false")
}

func getComputedRollbackOnFailure(appName string) string {
	rollbackOnFailure := getRollbackOnFailure(appName)
	if rollbackOnFailure == "" {
		rollbackOnFailure = getGlobalRollbackOnFailure()
	}

	return rollbackOnFailure
}

func getGlobalGlobalToken() string {
	return common.PropertyGet("scheduler-k3s", "--global", "token")
}

func getProcessHealtchecks(healthchecks []appjson.Healthcheck, primaryPort int32) ProcessHealthchecks {
	if len(healthchecks) == 0 {
		return ProcessHealthchecks{}
	}

	livenessChecks := []ProcessHealthcheck{}
	readinessChecks := []ProcessHealthcheck{}
	startupChecks := []ProcessHealthcheck{}
	uptimeSeconds := []int32{}
	for _, healthcheck := range healthchecks {
		probe := ProcessHealthcheck{
			InitialDelaySeconds: healthcheck.InitialDelay,
			PeriodSeconds:       healthcheck.Wait,
			TimeoutSeconds:      healthcheck.Timeout,
			FailureThreshold:    healthcheck.Attempts,
			SuccessThreshold:    int32(1),
		}
		if len(healthcheck.Command) > 0 {
			probe.Exec = &ExecHealthcheck{
				Command: healthcheck.Command,
			}
		} else if healthcheck.Listening {
			probe.TCPSocket = &TCPHealthcheck{
				Port: primaryPort,
			}
			for _, header := range healthcheck.HTTPHeaders {
				if header.Name == "Host" {
					probe.TCPSocket.Host = header.Value
				}
			}
		} else if healthcheck.Path != "" {
			probe.HTTPGet = &HTTPHealthcheck{
				Path:        healthcheck.Path,
				Port:        primaryPort,
				HTTPHeaders: []HTTPHeader{},
			}

			if healthcheck.Scheme != "" {
				probe.HTTPGet.Scheme = URIScheme(strings.ToUpper(healthcheck.Scheme))
			}

			for _, header := range healthcheck.HTTPHeaders {
				probe.HTTPGet.HTTPHeaders = append(probe.HTTPGet.HTTPHeaders, HTTPHeader{
					Name:  header.Name,
					Value: header.Value,
				})
			}
		} else if healthcheck.Uptime > 0 {
			uptimeSeconds = append(uptimeSeconds, healthcheck.Uptime)
		}

		if healthcheck.Type == appjson.HealthcheckType_Liveness {
			livenessChecks = append(livenessChecks, probe)
		} else if healthcheck.Type == appjson.HealthcheckType_Readiness {
			readinessChecks = append(readinessChecks, probe)
		} else if healthcheck.Type == appjson.HealthcheckType_Startup {
			startupChecks = append(startupChecks, probe)
		}
	}
	if len(livenessChecks) > 1 {
		common.LogWarn("Multiple liveness checks are not supported, only the first one will be used")
	}
	if len(readinessChecks) > 1 {
		common.LogWarn("Multiple readiness checks are not supported, only the first one will be used")
	}
	if len(startupChecks) > 1 {
		common.LogWarn("Multiple startup checks are not supported, only the first one will be used")
	}
	if len(uptimeSeconds) > 1 {
		common.LogWarn("Multiple uptime checks are not supported, only the first one will be used")
	}

	processHealthchecks := ProcessHealthchecks{}
	if len(livenessChecks) > 0 {
		processHealthchecks.Liveness = livenessChecks[0]
	}
	if len(readinessChecks) > 0 {
		processHealthchecks.Readiness = readinessChecks[0]
	}
	if len(startupChecks) > 0 {
		processHealthchecks.Startup = startupChecks[0]
	}
	if len(uptimeSeconds) > 0 {
		processHealthchecks.MinReadySeconds = uptimeSeconds[0]
	}

	return processHealthchecks
}

func getProcessResources(appName string, processType string) (ProcessResourcesMap, error) {
	processResources := ProcessResourcesMap{
		Limits: ProcessResources{
			CPU:    "1000m",
			Memory: "512Mi",
		},
		Requests: ProcessResources{
			CPU:    "1000m",
			Memory: "512Mi",
		},
	}
	cpuLimit, err := common.PlugnTriggerOutputAsString("resource-get-property", []string{appName, processType, "limit", "cpu"}...)
	if err != nil && cpuLimit != "" && cpuLimit != "0" {
		_, err := resource.ParseQuantity(cpuLimit)
		if err != nil {
			return ProcessResourcesMap{}, fmt.Errorf("Error parsing cpu limit: %w", err)
		}
		processResources.Limits.CPU = cpuLimit
	}
	nvidiaGpuLimit, err := common.PlugnTriggerOutputAsString("resource-get-property", []string{appName, processType, "limit", "nvidia-gpu"}...)
	if err != nil && nvidiaGpuLimit != "" && nvidiaGpuLimit != "0" {
		_, err := resource.ParseQuantity(nvidiaGpuLimit)
		if err != nil {
			return ProcessResourcesMap{}, fmt.Errorf("Error parsing nvidia-gpu limit: %w", err)
		}
		processResources.Limits.NvidiaGPU = nvidiaGpuLimit
	}
	memoryLimit, err := common.PlugnTriggerOutputAsString("resource-get-property", []string{appName, processType, "limit", "memory"}...)
	if err != nil && memoryLimit != "" && memoryLimit != "0" {
		_, err := resource.ParseQuantity(memoryLimit)
		if err != nil {
			return ProcessResourcesMap{}, fmt.Errorf("Error parsing memory limit: %w", err)
		}
		processResources.Limits.Memory = memoryLimit
	}

	cpuRequest, err := common.PlugnTriggerOutputAsString("resource-get-property", []string{appName, processType, "reserve", "cpu"}...)
	if err != nil && cpuRequest != "" && cpuRequest != "0" {
		_, err := resource.ParseQuantity(cpuRequest)
		if err != nil {
			return ProcessResourcesMap{}, fmt.Errorf("Error parsing cpu request: %w", err)
		}
		processResources.Requests.CPU = cpuRequest
	}
	memoryRequest, err := common.PlugnTriggerOutputAsString("resource-get-property", []string{appName, processType, "reserve", "memory"}...)
	if err != nil && memoryRequest != "" && memoryRequest != "0" {
		_, err := resource.ParseQuantity(memoryRequest)
		if err != nil {
			return ProcessResourcesMap{}, fmt.Errorf("Error parsing memory request: %w", err)
		}
		processResources.Requests.Memory = memoryRequest
	}

	return processResources, nil
}

func getStartCommand(input StartCommandInput) (StartCommandOutput, error) {
	command := extractStartCommand(input)
	fields, err := shell.Fields(command, func(name string) string {
		if name == "PORT" {
			return fmt.Sprint(input.Port)
		}

		return input.Env[name]
	})
	if err != nil {
		return StartCommandOutput{}, err
	}

	return StartCommandOutput{
		Command: fields,
	}, nil
}

func installHelmCharts(ctx context.Context, clientset KubernetesClient) error {
	for _, repo := range HelmRepositories {
		helmAgent, err := NewHelmAgent("default", DeployLogPrinter)
		if err != nil {
			return fmt.Errorf("Error creating helm agent: %w", err)
		}

		err = helmAgent.AddRepository(ctx, AddRepositoryInput(repo))
		if err != nil {
			return fmt.Errorf("Error adding helm repository %s: %w", repo.Name, err)
		}
	}

	for _, chart := range HelmCharts {
		if chart.CreateNamespace {
			namespace := corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: chart.Namespace,
					Annotations: map[string]string{
						"dokku.com/managed": "true",
					},
					Labels: map[string]string{
						"dokku.com/managed": "true",
					},
				},
			}
			_, err := clientset.CreateNamespace(ctx, CreateNamespaceInput{
				Name: namespace,
			})
			if err != nil {
				return fmt.Errorf("Error creating namespace %s: %w", chart.Namespace, err)
			}
		}

		contents, err := templates.ReadFile(fmt.Sprintf("templates/helm-config/%s.yaml", chart.ReleaseName))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("Error reading values file %s: %w", chart.ReleaseName, err)
		}

		var values map[string]interface{}
		if len(contents) > 0 {
			err = yaml.Unmarshal(contents, &values)
			if err != nil {
				return fmt.Errorf("Error unmarshalling values file: %w", err)
			}
		}

		helmAgent, err := NewHelmAgent(chart.Namespace, DeployLogPrinter)
		if err != nil {
			return fmt.Errorf("Error creating helm agent: %w", err)
		}

		err = helmAgent.InstallOrUpgradeChart(ctx, ChartInput{
			ChartPath:   chart.ChartPath,
			Namespace:   chart.Namespace,
			ReleaseName: chart.ReleaseName,
			RepoURL:     chart.RepoURL,
			Values:      values,
			Version:     chart.Version,
		})
		if err != nil {
			return fmt.Errorf("Error installing chart %s: %w", chart.ChartPath, err)
		}
	}
	return nil
}

func isK3sInstalled() error {
	if !common.FileExists("/usr/local/bin/k3s") {
		return fmt.Errorf("k3s binary is not available")
	}

	if !common.FileExists(RegistryConfigPath) {
		return fmt.Errorf("k3s registry config is not available")
	}

	if !common.FileExists(KubeConfigPath) {
		return fmt.Errorf("k3s kubeconfig is not available")
	}

	return nil
}

func isPodReady(ctx context.Context, clientset KubernetesClient, podName, namespace string) wait.ConditionWithContextFunc {
	return func(ctx context.Context) (bool, error) {
		fmt.Printf(".")

		pod, err := clientset.GetPod(ctx, GetPodInput{
			Name:      podName,
			Namespace: namespace,
		})
		if err != nil {
			return false, err
		}

		switch pod.Status.Phase {
		case v1.PodRunning:
			return true, nil
		case v1.PodFailed, v1.PodSucceeded:
			return false, conditions.ErrPodCompleted
		}
		return false, nil
	}
}

// kubernetesNodeToNode converts a kubernetes node to a Node
func kubernetesNodeToNode(node v1.Node) Node {
	roles := []string{}
	if len(node.Labels["kubernetes.io/role"]) > 0 {
		roles = append(roles, node.Labels["kubernetes.io/role"])
	} else {
		for k, v := range node.Labels {
			if strings.HasPrefix(k, "node-role.kubernetes.io/") && v == "true" {
				roles = append(roles, strings.TrimPrefix(k, "node-role.kubernetes.io/"))
			}
		}
	}

	sort.Strings(roles)

	ready := false
	for _, condition := range node.Status.Conditions {
		if condition.Type == "Ready" {
			ready = condition.Status == "True"
			break
		}
	}

	remoteHost := ""
	if val, ok := node.Annotations["dokku.com/remote-host"]; ok {
		remoteHost = val
	}

	return Node{
		Name:       node.Name,
		Roles:      roles,
		Ready:      ready,
		RemoteHost: remoteHost,
		Version:    node.Status.NodeInfo.KubeletVersion,
	}
}

func waitForPodBySelectorRunning(ctx context.Context, input WaitForPodBySelectorRunningInput) error {
	pods, err := waitForPodToExist(ctx, WaitForPodToExistInput{
		Clientset:     input.Clientset,
		LabelSelector: input.LabelSelector,
		Namespace:     input.Namespace,
		PodName:       input.PodName,
		RetryCount:    3,
	})
	if err != nil {
		return fmt.Errorf("Error waiting for pod to exist: %w", err)
	}

	if len(pods) == 0 {
		return fmt.Errorf("no pods in %s with selector %s", input.Namespace, input.LabelSelector)
	}

	timeout := time.Duration(input.Timeout) * time.Second
	for _, pod := range pods {
		if input.PodName != "" && pod.Name != input.PodName {
			break
		}

		if err := wait.PollUntilContextTimeout(ctx, time.Second, timeout, false, input.Waiter(ctx, input.Clientset, pod.Name, pod.Namespace)); err != nil {
			print("\n")
			return fmt.Errorf("Error waiting for pod to be ready: %w", err)
		}
	}
	print("\n")
	return nil
}

func waitForNodeToExist(ctx context.Context, input WaitForNodeToExistInput) ([]v1.Node, error) {
	var matchingNodes []v1.Node
	var err error
	for i := 0; i < input.RetryCount; i++ {
		nodes, err := input.Clientset.ListNodes(ctx, ListNodesInput{})
		if err != nil {
			time.Sleep(1 * time.Second)
		}

		if input.NodeName == "" {
			matchingNodes = nodes
			break
		}

		for _, node := range nodes {
			if node.Name == input.NodeName {
				matchingNodes = append(matchingNodes, node)
				break
			}
		}
		if len(matchingNodes) > 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		return matchingNodes, fmt.Errorf("Error listing nodes: %w", err)
	}
	return matchingNodes, nil
}

func waitForPodToExist(ctx context.Context, input WaitForPodToExistInput) ([]v1.Pod, error) {
	var pods []v1.Pod
	var err error
	for i := 0; i < input.RetryCount; i++ {
		pods, err = input.Clientset.ListPods(ctx, ListPodsInput{
			Namespace:     input.Namespace,
			LabelSelector: input.LabelSelector,
		})
		if err != nil {
			time.Sleep(1 * time.Second)
		}

		if input.PodName == "" {
			break
		}

		for _, pod := range pods {
			if pod.Name == input.PodName {
				break
			}
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		return pods, fmt.Errorf("Error listing pods: %w", err)
	}
	return pods, nil
}
