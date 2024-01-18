package scheduler_k3s

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/dokku/dokku/plugins/common"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/util/term"
	"k8s.io/kubernetes/pkg/client/conditions"
	"mvdan.cc/sh/v3/shell"
)

type EnterPodInput struct {
	AppName     string
	Clientset   KubernetesClient
	Command     []string
	Entrypoint  string
	ProcessType string
	SelectedPod v1.Pod
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

type WaitForPodBySelectorRunningInput struct {
	Clientset KubernetesClient
	Namespace string
	Selector  string
	Timeout   int
	Waiter    func(ctx context.Context, clientset KubernetesClient, podName, namespace string) wait.ConditionWithContextFunc
}

type WaitForPodToExistInput struct {
	Clientset  KubernetesClient
	Namespace  string
	RetryCount int
	Selector   string
}

func createKubernetesNamespace(ctx context.Context, namespaceName string) error {
	clientset, err := NewKubernetesClient()
	if err != nil {
		return err
	}

	namespaces, err := clientset.ListNamespaces(ctx)
	if err != nil {
		return err
	}

	for _, namespace := range namespaces {
		if namespace.Name == namespaceName {
			return nil
		}
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
		Namespace: namespace,
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

	req := coreclient.RESTClient().Post().
		Resource("pods").
		Namespace(input.SelectedPod.Namespace).
		Name(input.SelectedPod.Name).
		SubResource("exec")

	req.Param("container", fmt.Sprintf("%s-%s", input.AppName, input.ProcessType))
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
		command = "/start " + input.ProcessType
	}

	startCommandResp, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:       "config-get",
		Args:          []string{input.AppName, "DOKKU_START_CMD"},
		CaptureOutput: true,
		StreamStdio:   false,
	})
	if err == nil && startCommandResp.ExitCode == 0 {
		command = startCommandResp.Stdout
	}

	if input.ImageSourceType == "herokuish" {
		return command
	}

	if input.ImageSourceType == "dockerfile" {
		startCommandDockerfileResp, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
			Trigger:       "config-get",
			Args:          []string{input.AppName, "DOKKU_DOCKERFILE_START_CMD"},
			CaptureOutput: true,
			StreamStdio:   false,
		})
		if err == nil && startCommandDockerfileResp.ExitCode == 0 {
			command = startCommandDockerfileResp.Stdout
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

type PortMap struct {
	ContainerPort int32  `json:"container_port"`
	HostPort      int32  `json:"host_port"`
	Scheme        string `json:"scheme"`
}

func (p PortMap) IsAllowedHttp() bool {
	return p.Scheme == "http" || p.ContainerPort == 80
}

func (p PortMap) IsAllowedHttps() bool {
	return p.Scheme == "https" || p.ContainerPort == 443
}

func getPortMaps(appName string) ([]PortMap, error) {
	portMaps := []PortMap{}

	output, err := common.PlugnTriggerOutputAsString("ports-get", []string{appName, "json"}...)
	if err != nil {
		return portMaps, err
	}

	err = json.Unmarshal([]byte(output), &portMaps)
	if err != nil {
		return portMaps, err
	}

	allowedMappings := []PortMap{}
	for _, portMap := range portMaps {
		if !portMap.IsAllowedHttp() && !portMap.IsAllowedHttps() {
			// todo: log warning
			continue
		}

		allowedMappings = append(allowedMappings, portMap)
	}

	return allowedMappings, nil
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

func isPodReady(ctx context.Context, clientset KubernetesClient, podName, namespace string) wait.ConditionWithContextFunc {
	return func(ctx context.Context) (bool, error) {
		fmt.Printf(".") // progress bar!

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

func waitForPodBySelectorRunning(ctx context.Context, input WaitForPodBySelectorRunningInput) error {
	pods, err := waitForPodToExist(ctx, WaitForPodToExistInput{
		Clientset:  input.Clientset,
		Namespace:  input.Namespace,
		RetryCount: 3,
		Selector:   input.Selector,
	})
	if err != nil {
		return fmt.Errorf("Error waiting for pod to exist: %w", err)
	}

	if len(pods) == 0 {
		return fmt.Errorf("no pods in %s with selector %s", input.Namespace, input.Selector)
	}

	timeout := time.Duration(input.Timeout) * time.Second
	for _, pod := range pods {
		if err := wait.PollUntilContextTimeout(ctx, time.Second, timeout, false, input.Waiter(ctx, input.Clientset, pod.Name, pod.Namespace)); err != nil {
			print("\n")
			return fmt.Errorf("Error waiting for pod to be ready: %w", err)
		}
	}
	print("\n")
	return nil
}

func waitForPodToExist(ctx context.Context, input WaitForPodToExistInput) ([]v1.Pod, error) {
	var pods []v1.Pod
	var err error
	for i := 0; i < input.RetryCount; i++ {
		pods, err = input.Clientset.ListPods(ctx, ListPodsInput{
			Namespace:     input.Namespace,
			LabelSelector: input.Selector,
		})
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		return pods, fmt.Errorf("Error listing pods: %w", err)
	}
	return pods, nil
}
