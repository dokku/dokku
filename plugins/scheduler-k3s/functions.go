package scheduler_k3s

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"mvdan.cc/sh/v3/shell"
)

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

func NewClient() (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", KubeConfigPath)
	if err != nil {
		return &kubernetes.Clientset{}, err
	}

	return kubernetes.NewForConfig(config)
}

func createKubernetesNamespace(ctx context.Context, namespaceName string) error {
	client, err := NewClient()
	if err != nil {
		return err
	}

	namespaces, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, namespace := range namespaces.Items {
		if namespace.Name == namespaceName {
			return nil
		}
	}

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}
	_, err = client.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func extractStartCommand(input StartCommandInput) string {
	command := ""
	if input.ImageSourceType == "herokuish" {
		command = "/start " + input.ProcessType
	}

	startCommandOverride, ok := config.Get(input.AppName, "DOKKU_START_CMD")
	if ok {
		command = startCommandOverride
	}

	if input.ImageSourceType == "herokuish" {
		return command
	}

	if input.ImageSourceType == "dockerfile" {
		startCommandOverride, ok := config.Get(input.AppName, "DOKKU_DOCKERFILE_START_CMD")
		if ok {
			command = startCommandOverride
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
