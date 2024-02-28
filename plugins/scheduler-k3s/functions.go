package scheduler_k3s

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	appjson "github.com/dokku/dokku/plugins/app-json"
	"github.com/dokku/dokku/plugins/common"
	nginxvhosts "github.com/dokku/dokku/plugins/nginx-vhosts"
	resty "github.com/go-resty/resty/v2"
	"golang.org/x/sync/errgroup"
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
	"k8s.io/utils/ptr"
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

func applyClusterIssuers(ctx context.Context) error {
	chartDir, err := os.MkdirTemp("", "cluster-issuer-chart-")
	if err != nil {
		return fmt.Errorf("Error creating cluster-issuer chart directory: %w", err)
	}
	defer os.RemoveAll(chartDir)

	// create the chart.yaml
	chart := &Chart{
		ApiVersion: "v2",
		AppVersion: "1.0.0",
		Icon:       "https://dokku.com/assets/dokku-logo.svg",
		Name:       "cluster-issuers",
		Version:    "0.0.1",
	}

	err = writeYaml(WriteYamlInput{
		Object: chart,
		Path:   filepath.Join(chartDir, "Chart.yaml"),
	})
	if err != nil {
		return fmt.Errorf("Error writing cluster-issuer chart: %w", err)
	}

	// create the values.yaml
	letsencryptEmailStag := getGlobalLetsencryptEmailStag()
	letsencryptEmailProd := getGlobalLetsencryptEmailProd()

	clusterIssuerValues := ClusterIssuerValues{
		ClusterIssuers: map[string]ClusterIssuer{
			"letsencrypt-stag": {
				Email:        letsencryptEmailStag,
				Enabled:      letsencryptEmailStag != "",
				IngressClass: getGlobalIngressClass(),
				Name:         "letsencrypt-stag",
				Server:       "https://acme-staging-v02.api.letsencrypt.org/directory",
			},
			"letsencrypt-prod": {
				Email:        letsencryptEmailProd,
				Enabled:      letsencryptEmailProd != "",
				IngressClass: getGlobalIngressClass(),
				Name:         "letsencrypt-prod",
				Server:       "https://acme-v02.api.letsencrypt.org/directory",
			},
		},
	}

	if err := os.MkdirAll(filepath.Join(chartDir, "templates"), os.FileMode(0755)); err != nil {
		return fmt.Errorf("Error creating cluster-issuer chart templates directory: %w", err)
	}

	err = writeYaml(WriteYamlInput{
		Object: clusterIssuerValues,
		Path:   filepath.Join(chartDir, "values.yaml"),
	})
	if err != nil {
		return fmt.Errorf("Error writing chart: %w", err)
	}

	// create the templates/cluster-issuer.yaml
	b, err := templates.ReadFile("templates/chart/cluster-issuer.yaml")
	if err != nil {
		return fmt.Errorf("Error reading cluster-issuer template: %w", err)
	}

	filename := filepath.Join(chartDir, "templates", "cluster-issuer.yaml")
	err = os.WriteFile(filename, b, os.FileMode(0644))
	if err != nil {
		return fmt.Errorf("Error writing cluster-issuer template: %w", err)
	}

	if os.Getenv("DOKKU_TRACE") == "1" {
		common.CatFile(filename)
	}

	// install the chart
	helmAgent, err := NewHelmAgent("cert-manager", DevNullPrinter)
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

	err = helmAgent.InstallOrUpgradeChart(ctx, ChartInput{
		ChartPath:         chartPath,
		Namespace:         "cert-manager",
		ReleaseName:       "cluster-issuers",
		RollbackOnFailure: true,
		Timeout:           timeoutDuration,
	})
	if err != nil {
		return fmt.Errorf("Error installing cluster-issuer chart: %w", err)
	}

	return nil
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
		Trigger: "config-get",
		Args:    []string{input.AppName, "DOKKU_START_CMD"},
	})
	if err == nil && resp.ExitCode == 0 && len(resp.Stdout) > 0 {
		command = strings.TrimSpace(resp.Stdout)
	}

	if input.ImageSourceType == "dockerfile" {
		resp, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
			Trigger: "config-get",
			Args:    []string{input.AppName, "DOKKU_DOCKERFILE_START_CMD"},
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

// getAnnotations retrieves annotations for a given app and process type
func getAnnotations(appName string, processType string) (ProcessAnnotations, error) {
	annotations := ProcessAnnotations{}
	certificateAnnotations, err := getAnnotation(appName, processType, "certificate")
	if err != nil {
		return annotations, err
	}
	annotations.CertificateAnnotations = certificateAnnotations

	cronJobAnnotations, err := getAnnotation(appName, processType, "cronjob")
	if err != nil {
		return annotations, err
	}
	annotations.CronJobAnnotations = cronJobAnnotations

	deploymentAnnotations, err := getAnnotation(appName, processType, "deployment")
	if err != nil {
		return annotations, err
	}
	annotations.DeploymentAnnotations = deploymentAnnotations

	ingressAnnotations, err := getIngressAnnotations(appName, processType)
	if err != nil {
		return annotations, err
	}
	annotations.IngressAnnotations = ingressAnnotations

	jobAnnotations, err := getAnnotation(appName, processType, "job")
	if err != nil {
		return annotations, err
	}
	annotations.JobAnnotations = jobAnnotations

	kedaScalingObjectAnnotations, err := getAnnotation(appName, processType, "keda_scaled_object")
	if err != nil {
		return annotations, err
	}
	annotations.KedaScalingObjectAnnotations = kedaScalingObjectAnnotations

	kedaSecretAnnotations, err := getAnnotation(appName, processType, "keda_secret")
	if err != nil {
		return annotations, err
	}
	annotations.KedaSecretAnnotations = kedaSecretAnnotations

	kedaTriggerAuthenticationAnnotations, err := getAnnotation(appName, processType, "keda_trigger_authentication")
	if err != nil {
		return annotations, err
	}
	annotations.KedaTriggerAuthenticationAnnotations = kedaTriggerAuthenticationAnnotations

	podAnnotations, err := getAnnotation(appName, processType, "pod")
	if err != nil {
		return annotations, err
	}
	annotations.PodAnnotations = podAnnotations

	secretAnnotations, err := getAnnotation(appName, processType, "secret")
	if err != nil {
		return annotations, err
	}
	annotations.SecretAnnotations = secretAnnotations

	serviceAnnotations, err := getAnnotation(appName, processType, "service")
	if err != nil {
		return annotations, err
	}
	annotations.ServiceAnnotations = serviceAnnotations

	serviceAccountAnnotations, err := getAnnotation(appName, processType, "serviceaccount")
	if err != nil {
		return annotations, err
	}
	annotations.ServiceAccountAnnotations = serviceAccountAnnotations

	traefikIngressRouteAnnotations, err := getAnnotation(appName, processType, "traefik_ingressroute")
	if err != nil {
		return annotations, err
	}
	annotations.TraefikIngressRouteAnnotations = traefikIngressRouteAnnotations

	traefikMiddlewareAnnotations, err := getAnnotation(appName, processType, "traefik_middleware")
	if err != nil {
		return annotations, err
	}
	annotations.TraefikMiddlewareAnnotations = traefikMiddlewareAnnotations

	return annotations, nil
}

// GetAutoscalingInput contains all the information needed to get autoscaling config
type GetAutoscalingInput struct {
	// AppName is the name of the app
	AppName string

	// ProcessType is the process type
	ProcessType string

	// Replicas is the number of replicas
	Replicas int

	// KedaValues is the keda values
	KedaValues GlobalKedaValues
}

// getAutoscaling retrieves autoscaling config for a given app and process type
func getAutoscaling(input GetAutoscalingInput) (ProcessAutoscaling, error) {
	config, ok, err := appjson.GetAutoscalingConfig(input.AppName, input.ProcessType, input.Replicas)
	if err != nil {
		common.LogWarn(fmt.Sprintf("Error getting autoscaling config for %s: %v", input.AppName, err))
		return ProcessAutoscaling{}, err
	}

	if !ok {
		common.LogWarn(fmt.Sprintf("No autoscaling config found for %s", input.AppName))
		return ProcessAutoscaling{}, nil
	}

	replacements := map[string]string{
		"APP_NAME":        input.AppName,
		"PROCESS_TYPE":    input.ProcessType,
		"DEPLOYMENT_NAME": fmt.Sprintf("%s-%s", input.AppName, input.ProcessType),
	}

	triggers := []ProcessAutoscalingTrigger{}
	for _, trigger := range config.Triggers {
		metadata := map[string]string{}
		for key, value := range trigger.Metadata {
			tmpl, err := template.New("").Delims("[[", "]]").Parse(value)
			if err != nil {
				return ProcessAutoscaling{}, fmt.Errorf("Error parsing autoscaling trigger metadata: %w", err)
			}

			var output bytes.Buffer
			if err := tmpl.Execute(&output, replacements); err != nil {
				return ProcessAutoscaling{}, fmt.Errorf("Error executing autoscaling trigger metadata template: %w", err)
			}
			metadata[key] = output.String()
		}

		trigger := ProcessAutoscalingTrigger{
			Name:     trigger.Name,
			Type:     trigger.Type,
			Metadata: metadata,
		}

		if auth, ok := input.KedaValues.Authentications[trigger.Type]; ok {
			trigger.AuthenticationRef = &ProcessAutoscalingTriggerAuthenticationRef{
				Name: auth.Name,
			}
		} else if auth, ok := input.KedaValues.GlobalAuthentications[trigger.Type]; ok {
			trigger.AuthenticationRef = &ProcessAutoscalingTriggerAuthenticationRef{
				Name: auth.Name,
				Kind: "ClusterTriggerAuthentication",
			}
		}

		triggers = append(triggers, trigger)
	}

	autoscaling := ProcessAutoscaling{
		CooldownPeriodSeconds:  ptr.Deref(config.CooldownPeriodSeconds, 300),
		Enabled:                len(triggers) > 0,
		MaxReplicas:            ptr.Deref(config.MaxQuantity, 0),
		MinReplicas:            ptr.Deref(config.MinQuantity, 0),
		PollingIntervalSeconds: ptr.Deref(config.PollingIntervalSeconds, 30),
		Triggers:               triggers,
		Type:                   "keda",
	}

	return autoscaling, nil
}

// getKedaValues retrieves keda values for a given app and process type
func getKedaValues(ctx context.Context, clientset KubernetesClient, appName string) (GlobalKedaValues, error) {
	properties, err := common.PropertyGetAllByPrefix("scheduler-k3s", appName, TriggerAuthPropertyPrefix)
	if err != nil {
		return GlobalKedaValues{}, fmt.Errorf("Error getting trigger-auth properties: %w", err)
	}

	auths := map[string]KedaAuthentication{}
	for key, value := range properties {
		parts := strings.SplitN(strings.TrimPrefix(key, TriggerAuthPropertyPrefix), ".", 2)
		if len(parts) != 2 {
			return GlobalKedaValues{}, fmt.Errorf("Invalid trigger-auth property format: %s", key)
		}

		authType := parts[0]
		secretKey := parts[1]
		if len(secretKey) == 0 {
			return GlobalKedaValues{}, fmt.Errorf("Invalid trigger-auth property format: %s", key)
		}

		if _, ok := auths[authType]; !ok {
			auths[authType] = KedaAuthentication{
				Name:    fmt.Sprintf("%s-%s", appName, authType),
				Type:    authType,
				Secrets: make(map[string]string),
			}
		}

		auths[authType].Secrets[secretKey] = base64.StdEncoding.EncodeToString([]byte(value))
	}

	items, err := clientset.ListClusterTriggerAuthentications(ctx, ListClusterTriggerAuthenticationsInput{})
	if err != nil {
		return GlobalKedaValues{}, fmt.Errorf("Error listing cluster trigger authentications: %w", err)
	}

	globalAuths := map[string]KedaAuthentication{}
	for _, item := range items {
		if !strings.HasPrefix(item.Name, "global-auth-") {
			continue
		}

		authType := strings.TrimPrefix(item.Name, "global-auth-")
		globalAuths[authType] = KedaAuthentication{
			Name: item.Name,
			Type: authType,
		}
	}

	return GlobalKedaValues{
		Authentications:       auths,
		GlobalAuthentications: globalAuths,
	}, nil
}

// getGlobalAnnotations retrieves global annotations for a given app
func getGlobalAnnotations(appName string) (ProcessAnnotations, error) {
	return getAnnotations(appName, GlobalProcessType)
}

// getAnnotation retrieves an annotation for a given app, process type, and resource type
func getAnnotation(appName string, processType string, resourceType string) (map[string]string, error) {
	annotations := map[string]string{}
	annotationsList, err := common.PropertyListGet("scheduler-k3s", appName, fmt.Sprintf("%s.%s", processType, resourceType))
	if err != nil {
		return annotations, err
	}

	for _, annotation := range annotationsList {
		parts := strings.SplitN(annotation, ": ", 2)
		if len(parts) != 2 {
			return annotations, fmt.Errorf("Invalid annotation format: %s", annotation)
		}

		annotations[parts[0]] = parts[1]
	}

	return annotations, nil
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

func getGlobalIngressClass() string {
	return common.PropertyGetDefault("scheduler-k3s", "--global", "ingress-class", DefaultIngressClass)
}

func getIngressAnnotations(appName string, processType string) (map[string]string, error) {
	type annotation struct {
		annotation      string
		getter          func(appName string) string
		locationSnippet func(value string) string
		serverSnippet   func(value string) string
	}

	locationLines := []string{}
	serverLines := []string{}

	properties := map[string]annotation{
		"access-log-path": {
			getter: nginxvhosts.ComputedAccessLogPath,
			serverSnippet: func(value string) string {
				if value == "" {
					return ""
				}
				return fmt.Sprintf("access_log %s;", value)
			},
		},
		"bind-address-ipv4": {
			getter: nginxvhosts.ComputedBindAddressIPv4,
		},
		"bind-address-ipv6": {
			getter: nginxvhosts.ComputedBindAddressIPv6,
		},
		"client-max-body-size": {
			annotation: "nginx.ingress.kubernetes.io/proxy-body-size",
			getter:     nginxvhosts.ComputedClientMaxBodySize,
		},
		"disable-custom-config": {
			getter: nginxvhosts.ComputedDisableCustomConfig,
		},
		"error-log-path": {
			getter: nginxvhosts.ComputedErrorLogPath,
			serverSnippet: func(value string) string {
				if value == "" {
					return ""
				}
				return fmt.Sprintf("error_log %s;", value)
			},
		},
		// todo: handle hsts properly
		"hsts-include-subdomains": {
			getter: nginxvhosts.ComputedHSTSIncludeSubdomains,
		},
		"hsts-max-age": {
			getter: nginxvhosts.ComputedHSTSMaxAge,
		},
		"hsts-preload": {
			getter: nginxvhosts.ComputedHSTSPreload,
		},
		"hsts": {
			getter: nginxvhosts.ComputedHSTS,
		},
		"nginx-conf-sigil-path": {
			getter: nginxvhosts.ComputedNginxConfSigilPath,
		},
		"proxy-buffer-size": {
			annotation: "nginx.ingress.kubernetes.io/proxy-buffer-size",
			getter:     nginxvhosts.ComputedProxyBufferSize,
		},
		"proxy-buffering": {
			annotation: "nginx.ingress.kubernetes.io/proxy-buffering",
			getter:     nginxvhosts.ComputedProxyBuffering,
		},
		"proxy-buffers": {
			annotation: "nginx.ingress.kubernetes.io/proxy-buffers-number",
			getter:     nginxvhosts.ComputedProxyBuffers,
		},
		"proxy-busy-buffers-size": {
			getter: nginxvhosts.ComputedProxyBusyBuffersSize,
			locationSnippet: func(value string) string {
				if value == "" {
					return ""
				}
				return fmt.Sprintf("proxy_busy_buffers_size %s;", value)
			},
		},
		"proxy-read-timeout": {
			annotation: "nginx.ingress.kubernetes.io/proxy-read-timeout",
			getter:     nginxvhosts.ComputedProxyReadTimeout,
		},
		"underscore-in-headers": {
			getter: nginxvhosts.ComputedUnderscoreInHeaders,
			serverSnippet: func(value string) string {
				if value == "" {
					return ""
				}
				return fmt.Sprintf("underscores_in_headers %s;", value)
			},
		},
		"x-forwarded-for-value": {
			getter: nginxvhosts.ComputedXForwardedForValue,
			locationSnippet: func(value string) string {
				if value == "" {
					return ""
				}
				return fmt.Sprintf("proxy_set_header X-Forwarded-For %s;", value)
			},
		},
		"x-forwarded-port-value": {
			getter: nginxvhosts.ComputedXForwardedPortValue,
			locationSnippet: func(value string) string {
				if value == "" {
					return ""
				}
				return fmt.Sprintf("proxy_set_header X-Forwarded-Port %s;", value)
			},
		},
		"x-forwarded-proto-value": {
			getter: nginxvhosts.ComputedXForwardedProtoValue,
			locationSnippet: func(value string) string {
				if value == "" {
					return ""
				}
				return fmt.Sprintf("proxy_set_header X-Forwarded-Proto %s;", value)
			},
		},
		"x-forwarded-ssl": {
			getter: nginxvhosts.ComputedXForwardedSSL,
			locationSnippet: func(value string) string {
				if value == "" {
					return ""
				}
				return fmt.Sprintf("proxy_set_header X-Forwarded-SSL %s;", value)
			},
		},
	}

	annotations := map[string]string{}
	for _, newKey := range properties {
		if newKey.locationSnippet != nil {
			locationLines = append(locationLines, newKey.locationSnippet(newKey.getter(appName)))
		} else if newKey.serverSnippet != nil {
			serverLines = append(serverLines, newKey.serverSnippet(newKey.getter(appName)))
		} else if newKey.annotation != "" {
			annotations[newKey.annotation] = newKey.getter(appName)
		}
	}

	var locationSnippet string
	for _, line := range locationLines {
		if line != "" {
			locationSnippet += line + "\n"
		}
	}
	var serverSnippet string
	for _, line := range serverLines {
		if line != "" {
			serverSnippet += line + "\n"
		}
	}

	if locationSnippet != "" {
		annotations["nginx.ingress.kubernetes.io/configuration-snippet"] = locationSnippet
	}
	if serverSnippet != "" {
		annotations["nginx.ingress.kubernetes.io/server-snippet"] = serverSnippet
	}

	customAnnotations, err := getAnnotation(appName, processType, "deployment")
	if err != nil {
		return map[string]string{}, err
	}

	for key, value := range customAnnotations {
		if _, ok := annotations[key]; ok {
			common.LogWarn(fmt.Sprintf("Nginx-based annotation %s will be overwritten by custom annotation", key))
		}

		annotations[key] = value
	}

	return annotations, nil
}

// getLabels retrieves labels for a given app and process type
func getLabels(appName string, processType string) (ProcessLabels, error) {
	labels := ProcessLabels{}
	certificateLabels, err := getLabel(appName, processType, "certificate")
	if err != nil {
		return labels, err
	}
	labels.CertificateLabels = certificateLabels

	cronJobLabels, err := getLabel(appName, processType, "cronjob")
	if err != nil {
		return labels, err
	}
	labels.CronJobLabels = cronJobLabels

	deploymentLabels, err := getLabel(appName, processType, "deployment")
	if err != nil {
		return labels, err
	}
	labels.DeploymentLabels = deploymentLabels

	ingressLabels, err := getLabel(appName, processType, "ingress")
	if err != nil {
		return labels, err
	}
	labels.IngressLabels = ingressLabels

	jobLabels, err := getLabel(appName, processType, "job")
	if err != nil {
		return labels, err
	}
	labels.JobLabels = jobLabels

	podLabels, err := getLabel(appName, processType, "pod")
	if err != nil {
		return labels, err
	}
	labels.PodLabels = podLabels

	secretLabels, err := getLabel(appName, processType, "secret")
	if err != nil {
		return labels, err
	}
	labels.SecretLabels = secretLabels

	serviceLabels, err := getLabel(appName, processType, "service")
	if err != nil {
		return labels, err
	}
	labels.ServiceLabels = serviceLabels

	serviceAccountLabels, err := getLabel(appName, processType, "serviceaccount")
	if err != nil {
		return labels, err
	}
	labels.ServiceAccountLabels = serviceAccountLabels

	traefikIngressRouteLabels, err := getLabel(appName, processType, "traefik_ingressroute")
	if err != nil {
		return labels, err
	}
	labels.TraefikIngressRouteLabels = traefikIngressRouteLabels

	traefikMiddlewareLabels, err := getLabel(appName, processType, "traefik_middleware")
	if err != nil {
		return labels, err
	}
	labels.TraefikMiddlewareLabels = traefikMiddlewareLabels

	return labels, nil
}

// getGlobalLabel retrieves global labels for a given app
func getGlobalLabel(appName string) (ProcessLabels, error) {
	return getLabels(appName, GlobalProcessType)
}

// getLabel retrieves an label for a given app, process type, and resource type
func getLabel(appName string, processType string, resourceType string) (map[string]string, error) {
	labels := map[string]string{}
	labelsList, err := common.PropertyListGet("scheduler-k3s", appName, fmt.Sprintf("labels.%s.%s", processType, resourceType))
	if err != nil {
		return labels, err
	}

	for _, label := range labelsList {
		parts := strings.SplitN(label, ": ", 2)
		if len(parts) != 2 {
			return labels, fmt.Errorf("Invalid label format: %s", label)
		}

		labels[parts[0]] = parts[1]
	}

	return labels, nil
}

func getLetsencryptServer(appName string) string {
	return common.PropertyGetDefault("scheduler-k3s", appName, "letsencrypt-server", "")
}

func getGlobalLetsencryptServer() string {
	return common.PropertyGetDefault("scheduler-k3s", "--global", "letsencrypt-server", "prod")
}

func getComputedLetsencryptServer(appName string) string {
	letsencryptServer := getLetsencryptServer(appName)
	if letsencryptServer == "" {
		letsencryptServer = getGlobalLetsencryptServer()
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
		Limits: ProcessResources{},
		Requests: ProcessResources{
			CPU:    "100m",
			Memory: "128Mi",
		},
	}

	emptyValues := map[string]bool{
		"":  true,
		"0": true,
	}

	result, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "resource-get-property",
		Args:    []string{appName, processType, "limit", "cpu"},
	})
	if err == nil && !emptyValues[result.StdoutContents()] {
		quantity, err := resource.ParseQuantity(result.StdoutContents())
		if err != nil {
			return ProcessResourcesMap{}, fmt.Errorf("Error parsing cpu limit: %w", err)
		}
		if quantity.MilliValue() != 0 {
			processResources.Limits.CPU = quantity.String()
		} else {
			processResources.Limits.CPU = ""
		}
	}
	nvidiaGpuLimit, err := common.PlugnTriggerOutputAsString("resource-get-property", []string{appName, processType, "limit", "nvidia-gpu"}...)
	if err == nil && nvidiaGpuLimit != "" && nvidiaGpuLimit != "0" {
		_, err := resource.ParseQuantity(nvidiaGpuLimit)
		if err != nil {
			return ProcessResourcesMap{}, fmt.Errorf("Error parsing nvidia-gpu limit: %w", err)
		}
		processResources.Limits.NvidiaGPU = nvidiaGpuLimit
	}
	result, err = common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "resource-get-property",
		Args:    []string{appName, processType, "limit", "memory"},
	})
	if err == nil && !emptyValues[result.StdoutContents()] {
		quantity, err := parseMemoryQuantity(result.StdoutContents())
		if err != nil {
			return ProcessResourcesMap{}, fmt.Errorf("Error parsing memory limit: %w", err)
		}
		if quantity != "0Mi" {
			processResources.Limits.Memory = quantity
		} else {
			processResources.Limits.Memory = ""
		}
	}

	result, err = common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "resource-get-property",
		Args:    []string{appName, processType, "reserve", "cpu"},
	})
	if err == nil && !emptyValues[result.StdoutContents()] {
		quantity, err := resource.ParseQuantity(result.StdoutContents())
		if err != nil {
			return ProcessResourcesMap{}, fmt.Errorf("Error parsing cpu request: %w", err)
		}
		if quantity.MilliValue() != 0 {
			processResources.Requests.CPU = quantity.String()
		} else {
			processResources.Requests.CPU = ""
		}
	}
	result, err = common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "resource-get-property",
		Args:    []string{appName, processType, "reserve", "memory"},
	})
	if err == nil && !emptyValues[result.StdoutContents()] {
		quantity, err := parseMemoryQuantity(result.StdoutContents())
		if err != nil {
			return ProcessResourcesMap{}, fmt.Errorf("Error parsing memory request: %w", err)
		}
		if quantity != "0Mi" {
			processResources.Requests.Memory = quantity
		} else {
			processResources.Requests.Memory = ""
		}
	}

	return processResources, nil
}

func getServerIP() (string, error) {
	serverIP := ""
	networkInterface := getGlobalNetworkInterface()
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("Unable to get network interfaces: %w", err)
	}

	for _, iface := range ifaces {
		if iface.Name == networkInterface {
			addr, err := iface.Addrs()
			if err != nil {
				return "", fmt.Errorf("Unable to get network addresses for interface %s: %w", networkInterface, err)
			}
			for _, a := range addr {
				if ipnet, ok := a.(*net.IPNet); ok {
					if ipnet.IP.To4() != nil {
						serverIP = ipnet.IP.String()
					}
				}
			}
		}
	}

	if len(serverIP) == 0 {
		return "", fmt.Errorf(fmt.Sprintf("Unable to determine server ip address from network-interface %s", networkInterface))
	}
	return serverIP, nil
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

func installHelmCharts(ctx context.Context, clientset KubernetesClient, shouldInstall func(HelmChart) bool) error {
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
		if !shouldInstall(chart) {
			continue
		}

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

func installHelperCommands(ctx context.Context) error {
	urls := map[string]string{
		"kubectx": "https://github.com/ahmetb/kubectx/releases/latest/download/kubectx",
		"kubens":  "https://github.com/ahmetb/kubectx/releases/latest/download/kubens",
	}

	client := resty.New()
	for binaryName, url := range urls {
		resp, err := client.R().
			SetContext(ctx).
			Get(url)
		if err != nil {
			return fmt.Errorf("Unable to download %s: %w", binaryName, err)
		}
		if resp == nil {
			return fmt.Errorf("Missing response from %s download: %w", binaryName, err)
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("Invalid status code for %s: %d", binaryName, resp.StatusCode())
		}

		f, err := os.Create(filepath.Join("/usr/local/bin", binaryName))
		if err != nil {
			return fmt.Errorf("Unable to create %s: %w", binaryName, err)
		}

		if err := f.Close(); err != nil {
			return fmt.Errorf("Unable to close %s file: %w", binaryName, err)
		}

		err = common.WriteStringToFile(common.WriteStringToFileInput{
			Content:   resp.String(),
			Filename:  f.Name(),
			GroupName: "root",
			Mode:      os.FileMode(0755),
			Username:  "root",
		})
		if err != nil {
			return fmt.Errorf("Unable to write %s to file: %w", binaryName, err)
		}

		fi, err := os.Stat(f.Name())
		if err != nil {
			return fmt.Errorf("Unable to get %s file size: %w", binaryName, err)
		}

		if fi.Size() == 0 {
			return fmt.Errorf("Invalid %s filesize", binaryName)
		}
	}

	return installHelm(ctx)
}

func installHelm(ctx context.Context) error {
	client := resty.New()
	resp, err := client.R().
		SetContext(ctx).
		Get("https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3")
	if err != nil {
		return fmt.Errorf("Unable to download helm installer: %w", err)
	}
	if resp == nil {
		return fmt.Errorf("Missing response from helm installer download: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("Invalid status code for helm installer script: %d", resp.StatusCode())
	}

	f, err := os.CreateTemp("", "sample")
	if err != nil {
		return fmt.Errorf("Unable to create temporary file for helm installer: %w", err)
	}
	defer os.Remove(f.Name())

	if err := f.Close(); err != nil {
		return fmt.Errorf("Unable to close helm installer file: %w", err)
	}

	err = common.WriteStringToFile(common.WriteStringToFileInput{
		Content:  resp.String(),
		Filename: f.Name(),
		Mode:     os.FileMode(0755),
	})
	if err != nil {
		return fmt.Errorf("Unable to write helm installer to file: %w", err)
	}

	fi, err := os.Stat(f.Name())
	if err != nil {
		return fmt.Errorf("Unable to get helm installer file size: %w", err)
	}

	if fi.Size() == 0 {
		return fmt.Errorf("Invalid helm installer filesize")
	}

	common.LogInfo2Quiet("Running helm installer")
	installerCmd, err := common.CallExecCommand(common.ExecCommandInput{
		Command:     f.Name(),
		StreamStdio: true,
	})
	if err != nil {
		return fmt.Errorf("Unable to call helm installer command: %w", err)
	}
	if installerCmd.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from helm installer command: %d", installerCmd.ExitCode)
	}

	return nil
}

// isKubernetesAvailable returns an error if kubernetes api is not available
func isKubernetesAvailable() error {
	client, err := NewKubernetesClient()
	if err != nil {
		return fmt.Errorf("Error creating kubernetes client: %w", err)
	}

	if err := client.Ping(); err != nil {
		return fmt.Errorf("Error pinging kubernetes: %w", err)
	}

	return nil
}

// isK3sInstalled returns an error if k3s is not installed
func isK3sInstalled() error {
	if !common.FileExists("/usr/local/bin/k3s") {
		return fmt.Errorf("k3s binary is not available")
	}

	if !common.FileExists(getKubeconfigPath()) {
		return fmt.Errorf("k3s kubeconfig is not available")
	}

	return nil
}

// isK3sKubernetes returns true if the current kubernetes cluster is configured to be k3s
func isK3sKubernetes() bool {
	return getKubeconfigPath() == KubeConfigPath
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

// parseMemoryQuantity parses a string into a valid memory quantity
func parseMemoryQuantity(input string) (string, error) {
	if _, err := strconv.ParseInt(input, 10, 64); err == nil {
		input = fmt.Sprintf("%sMi", input)
	}
	quantity, err := resource.ParseQuantity(input)
	if err != nil {
		return "", err
	}

	return quantity.String(), nil
}

func uninstallHelperCommands(ctx context.Context) error {
	errs, _ := errgroup.WithContext(ctx)
	errs.Go(func() error {
		return os.RemoveAll("/usr/local/bin/kubectx")
	})
	errs.Go(func() error {
		return os.RemoveAll("/usr/local/bin/kubens")
	})
	return errs.Wait()
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
