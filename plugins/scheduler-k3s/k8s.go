package scheduler_k3s

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/dokku/dokku/plugins/common"
	"github.com/fatih/color"
	"github.com/go-openapi/jsonpointer"
	kedav1alpha1 "github.com/kedacore/keda/v2/apis/keda/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/util/term"
	"k8s.io/utils/ptr"
)

func getKubeconfigPath() string {
	return common.PropertyGetDefault("scheduler-k3s", "--global", "kubeconfig-path", KubeConfigPath)
}

func getKubeContext() string {
	return common.PropertyGetDefault("scheduler-k3s", "--global", "kube-context", DefaultKubeContext)
}

type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

type NilResponseError struct {
	Message string
}

func (e *NilResponseError) Error() string {
	return e.Message
}

// KubernetesClient is a wrapper around the Kubernetes client
type KubernetesClient struct {
	// Client is the Kubernetes client
	Client kubernetes.Clientset

	// DynamicClient is the Kubernetes dynamic client
	DynamicClient dynamic.Interface

	// KubeConfigPath is the path to the Kubernetes config
	KubeConfigPath string

	// RestClient is the Kubernetes REST client
	RestClient rest.Interface

	// RestConfig is the Kubernetes REST config
	RestConfig rest.Config
}

// NewKubernetesClient creates a new Kubernetes client
func NewKubernetesClient() (KubernetesClient, error) {
	kubeconfigPath := getKubeconfigPath()
	kubeContext := getKubeContext()
	clientConfig := KubernetesClientConfig(kubeconfigPath, kubeContext)
	restConf, err := clientConfig.ClientConfig()
	if err != nil {
		return KubernetesClient{}, err
	}

	restConf.GroupVersion = &schema.GroupVersion{
		Group:   "api",
		Version: "v1",
	}

	client, err := kubernetes.NewForConfig(restConf)
	if err != nil {
		return KubernetesClient{}, err
	}

	restConf.NegotiatedSerializer = runtime.NewSimpleNegotiatedSerializer(runtime.SerializerInfo{})

	restClient, err := rest.RESTClientFor(restConf)
	if err != nil {
		return KubernetesClient{}, err
	}

	dynamicClient, err := dynamic.NewForConfig(restConf)
	if err != nil {
		return KubernetesClient{}, err
	}

	return KubernetesClient{
		Client:         *client,
		DynamicClient:  dynamicClient,
		KubeConfigPath: kubeconfigPath,
		RestConfig:     *restConf,
		RestClient:     restClient,
	}, nil
}

// KubernetesClientConfig returns a Kubernetes client config
func KubernetesClientConfig(kubeconfigPath string, kubecontext string) clientcmd.ClientConfig {
	configOverrides := clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: ""}}
	if kubecontext != "" {
		configOverrides.CurrentContext = kubecontext
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&configOverrides,
	)
}

func (k KubernetesClient) Ping() error {
	_, err := k.Client.Discovery().ServerVersion()
	return err
}

func (k KubernetesClient) GetLowestNodeVersion(ctx context.Context, input ListNodesInput) (string, error) {
	nodes, err := k.ListNodes(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to list nodes: %w", err)
	}

	if len(nodes) == 0 {
		return "", fmt.Errorf("no nodes found in the cluster")
	}

	var lowestVersion *semver.Version
	for _, node := range nodes {
		kubeletVersion := node.Status.NodeInfo.KubeletVersion
		if kubeletVersion == "" {
			continue
		}

		versionStr := strings.TrimPrefix(kubeletVersion, "v")
		version, err := semver.NewVersion(versionStr)
		if err != nil {
			common.LogWarn(fmt.Sprintf("Failed to parse version %s for node %s: %v", kubeletVersion, node.Name, err))
			continue
		}

		if lowestVersion == nil || version.LessThan(lowestVersion) {
			lowestVersion = version
		}
	}

	if lowestVersion == nil {
		return "", fmt.Errorf("no valid kubelet versions found")
	}

	return "v" + lowestVersion.String(), nil
}

// AnnotateNodeInput contains all the information needed to annotates a Kubernetes node
type AnnotateNodeInput struct {
	// Name is the Kubernetes node name
	Name string
	// Key is the annotation key
	Key string
	// Value is the annotation value
	Value string
}

// AnnotateNode annotates a Kubernetes node
func (k KubernetesClient) AnnotateNode(ctx context.Context, input AnnotateNodeInput) error {
	node, err := k.Client.CoreV1().Nodes().Get(ctx, input.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if node == nil {
		return &NotFoundError{"node is nil"}
	}

	keyPath := fmt.Sprintf("/metadata/annotations/%s", jsonpointer.Escape(input.Key))
	patch := fmt.Sprintf(`[{"op":"add", "path":"%s", "value":"%s" }]`, keyPath, input.Value)
	_, err = k.Client.CoreV1().Nodes().Patch(ctx, node.Name, types.JSONPatchType, []byte(patch), metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("failed to annotate node: %w", err)
	}

	return nil
}

type ApplyKubernetesManifestInput struct {
	// Manifest is the path to the Kubernetes manifest
	Manifest string
}

func (k KubernetesClient) ApplyKubernetesManifest(ctx context.Context, input ApplyKubernetesManifestInput) error {
	args := []string{
		"apply",
		"-f",
		input.Manifest,
	}

	if kubeContext := getKubeContext(); kubeContext != "" {
		args = append([]string{"--context", kubeContext}, args...)
	}

	if kubeconfigPath := getKubeconfigPath(); kubeconfigPath != "" {
		args = append([]string{"--kubeconfig", kubeconfigPath}, args...)
	}

	upgradeCmd, err := common.CallExecCommand(common.ExecCommandInput{
		Command:     "kubectl",
		Args:        args,
		StreamStdio: true,
	})
	if err != nil {
		return fmt.Errorf("Unable to call kubectl command: %w", err)
	}
	if upgradeCmd.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from kubectl command: %d", upgradeCmd.ExitCode)
	}

	return nil
}

// CreateJobInput contains all the information needed to create a Kubernetes job
type CreateJobInput struct {
	// Job is the Kubernetes job
	Job batchv1.Job

	// Namespace is the Kubernetes namespace
	Namespace string
}

// CreateJob creates a Kubernetes job
func (k KubernetesClient) CreateJob(ctx context.Context, input CreateJobInput) (batchv1.Job, error) {
	job, err := k.Client.BatchV1().Jobs(input.Namespace).Create(ctx, &input.Job, metav1.CreateOptions{})
	if err != nil {
		return batchv1.Job{}, err
	}

	if job == nil {
		return batchv1.Job{}, &NotFoundError{"job is nil"}
	}

	return *job, err
}

// CreateNamespaceInput contains all the information needed to create a Kubernetes namespace
type CreateNamespaceInput struct {
	// Name is the name of the Kubernetes namespace
	Name corev1.Namespace
}

// CreateNamespace creates a Kubernetes namespace
func (k KubernetesClient) CreateNamespace(ctx context.Context, input CreateNamespaceInput) (corev1.Namespace, error) {
	namespaces, err := k.ListNamespaces(ctx)
	if err != nil {
		return corev1.Namespace{}, err
	}

	for _, namespace := range namespaces {
		if namespace.Name == input.Name.Name {
			return namespace, nil
		}
	}

	namespace, err := k.Client.CoreV1().Namespaces().Create(ctx, &input.Name, metav1.CreateOptions{})
	if err != nil {
		return corev1.Namespace{}, err
	}

	if namespace == nil {
		return corev1.Namespace{}, &NotFoundError{"namespace is nil"}
	}

	return *namespace, err
}

// DeleteIngressInput contains all the information needed to delete a Kubernetes ingress
type DeleteIngressInput struct {
	// Name is the Kubernetes ingress name
	Name string

	// Namespace is the Kubernetes namespace
	Namespace string
}

// DeleteIngress deletes a Kubernetes ingress
func (k KubernetesClient) DeleteIngress(ctx context.Context, input DeleteIngressInput) error {
	return k.Client.NetworkingV1().Ingresses(input.Namespace).Delete(ctx, input.Name, metav1.DeleteOptions{})
}

// DeleteJobInput contains all the information needed to delete a Kubernetes job
type DeleteJobInput struct {
	// Name is the Kubernetes job name
	Name string

	// Namespace is the Kubernetes namespace
	Namespace string
}

// DeleteJob deletes a Kubernetes job
func (k KubernetesClient) DeleteJob(ctx context.Context, input DeleteJobInput) error {
	return k.Client.BatchV1().Jobs(input.Namespace).Delete(ctx, input.Name, metav1.DeleteOptions{
		PropagationPolicy: ptr.To(metav1.DeletePropagationForeground),
	})
}

// DeleteNodeInput contains all the information needed to delete a Kubernetes node
type DeleteNodeInput struct {
	// Name is the Kubernetes node name
	Name string
}

// DeleteNode deletes a Kubernetes node
func (k KubernetesClient) DeleteNode(ctx context.Context, input DeleteNodeInput) error {
	return k.Client.CoreV1().Nodes().Delete(ctx, input.Name, metav1.DeleteOptions{})
}

// DeleteSecretInput contains all the information needed to delete a Kubernetes secret
type DeleteSecretInput struct {
	// Name is the Kubernetes secret name
	Name string

	// Namespace is the Kubernetes namespace
	Namespace string
}

// DeleteSecret deletes a Kubernetes secret
func (k KubernetesClient) DeleteSecret(ctx context.Context, input DeleteSecretInput) error {
	return k.Client.CoreV1().Secrets(input.Namespace).Delete(ctx, input.Name, metav1.DeleteOptions{})
}

// ExecCommandInput contains all the information needed to execute a command in a Kubernetes pod
type ExecCommandInput struct {
	// Command is the command to execute
	Command []string

	// ContainerName is the Kubernetes container name
	ContainerName string

	// Entrypoint is the command entrypoint
	Entrypoint string

	// Name is the Kubernetes pod name
	Name string

	// Namespace is the Kubernetes namespace
	Namespace string

	// Stderr is the error writer
	Stderr io.Writer

	// Stdout is the output writer
	Stdout io.Writer
}

// ExecCommand executes a command in a Kubernetes pod
func (k KubernetesClient) ExecCommand(ctx context.Context, input ExecCommandInput) error {
	coreclient, err := corev1client.NewForConfig(&k.RestConfig)
	if err != nil {
		return fmt.Errorf("Error creating corev1 client: %w", err)
	}

	req := coreclient.RESTClient().Post().
		Resource("pods").
		Namespace(input.Namespace).
		Name(input.Name).
		SubResource("exec")

	req.Param("container", input.ContainerName)
	req.Param("stdin", "true")
	req.Param("stdout", "true")
	req.Param("stderr", "true")

	if input.Entrypoint != "" {
		req.Param("command", input.Entrypoint)
	}
	for _, cmd := range input.Command {
		req.Param("command", cmd)
	}

	var stdout io.Writer
	if input.Stdout == nil {
		stdout = os.Stdout
	} else {
		stdout = input.Stdout
	}

	var stderr io.Writer
	if input.Stderr == nil {
		stderr = os.Stderr
	} else {
		stderr = input.Stderr
	}

	t := term.TTY{
		In:  os.Stdin,
		Out: stdout,
		Raw: true,
	}

	size := t.GetSize()
	sizeQueue := t.MonitorSize(size)
	actuallyTty := (sizeQueue != nil) || common.ToBool(os.Getenv("DOKKU_FORCE_TTY"))

	if actuallyTty {
		req.Param("tty", "true")
	} else {
		req.Param("tty", "false")
		t = term.TTY{
			In:  os.Stdin,
			Out: stdout,
			Raw: false,
		}
		size = t.GetSize()
		sizeQueue = t.MonitorSize(size)
	}

	return t.Safe(func() error {
		exec, err := remotecommand.NewSPDYExecutor(&k.RestConfig, "POST", req.URL())
		if err != nil {
			return fmt.Errorf("Error creating executor: %w", err)
		}

		return exec.StreamWithContext(ctx, remotecommand.StreamOptions{
			Stdin:             os.Stdin,
			Stdout:            stdout,
			Stderr:            stderr,
			Tty:               actuallyTty,
			TerminalSizeQueue: sizeQueue,
		})
	})
}

// GetPodLogsInput contains all the information needed to get the logs for a Kubernetes pod
type GetLogsInput struct {
	// Name is the Kubernetes pod name
	Name string

	// Namespace is the Kubernetes namespace
	Namespace string
}

// GetLogs gets the logs for a Kubernetes pod
func (k KubernetesClient) GetLogs(ctx context.Context, input GetLogsInput) ([]byte, error) {
	logOptions := corev1.PodLogOptions{}

	request := k.Client.CoreV1().Pods(input.Namespace).GetLogs(input.Name, &logOptions)

	readCloser, err := request.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}
	defer readCloser.Close()

	bytes, err := io.ReadAll(readCloser)
	if err != nil {
		return nil, fmt.Errorf("failed to read logs: %w", err)
	}

	return bytes, nil
}

// GetNodeInput contains all the information needed to get a Kubernetes node
type GetNodeInput struct {
	// Name is the Kubernetes node name
	Name string
}

// GetNode gets a Kubernetes node
func (k KubernetesClient) GetNode(ctx context.Context, input GetNodeInput) (Node, error) {
	if input.Name == "" {
		return Node{}, errors.New("node name is required")
	}

	node, err := k.Client.CoreV1().Nodes().Get(ctx, input.Name, metav1.GetOptions{})
	if err != nil {
		return Node{}, err
	}

	if node == nil {
		return Node{}, &NotFoundError{"node is nil"}
	}

	return kubernetesNodeToNode(*node), err
}

// GetJobInput contains all the information needed to get a Kubernetes job
type GetPodInput struct {
	// Name is the Kubernetes pod name
	Name string

	// Namespace is the Kubernetes namespace
	Namespace string
}

// GetPod gets a Kubernetes pod
func (k KubernetesClient) GetPod(ctx context.Context, input GetPodInput) (corev1.Pod, error) {
	pod, err := k.Client.CoreV1().Pods(input.Namespace).Get(ctx, input.Name, metav1.GetOptions{})
	if err != nil {
		return corev1.Pod{}, err
	}

	if pod == nil {
		return corev1.Pod{}, &NotFoundError{"pod is nil"}
	}

	return *pod, err
}

// GetSecretInput contains all the information needed to get a Kubernetes secret
type GetSecretInput struct {
	// Name is the Kubernetes secret name
	Name string

	// Namespace is the Kubernetes namespace
	Namespace string
}

// GetSecret gets a Kubernetes secret
func (k KubernetesClient) GetSecret(ctx context.Context, input GetSecretInput) (corev1.Secret, error) {
	secret, err := k.Client.CoreV1().Secrets(input.Namespace).Get(ctx, input.Name, metav1.GetOptions{})
	if err != nil {
		return corev1.Secret{}, err
	}

	if secret == nil {
		return corev1.Secret{}, &NotFoundError{"secret is nil"}
	}

	return *secret, err
}

// LabelNodeInput contains all the information needed to label a Kubernetes node
type LabelNodeInput struct {
	// Name is the Kubernetes node name
	Name string
	// Key is the label key
	Key string
	// Value is the label value
	Value string
}

// LabelNode labels a Kubernetes node
func (k KubernetesClient) LabelNode(ctx context.Context, input LabelNodeInput) error {
	node, err := k.Client.CoreV1().Nodes().Get(ctx, input.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if node == nil {
		return &NotFoundError{"node is nil"}
	}

	keyPath := fmt.Sprintf("/metadata/labels/%s", jsonpointer.Escape(input.Key))
	patch := fmt.Sprintf(`[{"op":"add", "path":"%s", "value":"%s" }]`, keyPath, input.Value)
	_, err = k.Client.CoreV1().Nodes().Patch(ctx, node.Name, types.JSONPatchType, []byte(patch), metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("failed to label node: %w", err)
	}

	return nil
}

// ListClusterTriggerAuthenticationsInput contains all the information needed to list Kubernetes trigger authentications
type ListClusterTriggerAuthenticationsInput struct {
	// Namespace is the Kubernetes namespace
	Namespace string

	// LabelSelector is the Kubernetes label selector
	LabelSelector string
}

// ListClusterTriggerAuthentications lists Kubernetes trigger authentications
func (k KubernetesClient) ListClusterTriggerAuthentications(ctx context.Context, input ListClusterTriggerAuthenticationsInput) ([]kedav1alpha1.ClusterTriggerAuthentication, error) {
	listOptions := metav1.ListOptions{LabelSelector: input.LabelSelector}

	gvr := schema.GroupVersionResource{
		Group:    "keda.sh",
		Version:  "v1alpha1",
		Resource: "clustertriggerauthentications",
	}

	response, err := k.DynamicClient.Resource(gvr).Namespace(input.Namespace).List(ctx, listOptions)
	if err != nil {
		return []kedav1alpha1.ClusterTriggerAuthentication{}, err
	}

	if response == nil {
		return []kedav1alpha1.ClusterTriggerAuthentication{}, &NilResponseError{"cluster trigger authentications is nil"}
	}

	triggerAuthentications := []kedav1alpha1.ClusterTriggerAuthentication{}
	for _, triggerAuthentication := range response.Items {
		var ta kedav1alpha1.ClusterTriggerAuthentication
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(triggerAuthentication.Object, &ta)
		if err != nil {
			return []kedav1alpha1.ClusterTriggerAuthentication{}, err
		}

		triggerAuthentications = append(triggerAuthentications, ta)
	}

	return triggerAuthentications, nil
}

// ListCronJobsInput contains all the information needed to list Kubernetes cron jobs
type ListCronJobsInput struct {
	// LabelSelector is the Kubernetes label selector
	LabelSelector string

	// Namespace is the Kubernetes namespace
	Namespace string
}

// ListCronJobs lists Kubernetes cron jobs
func (k KubernetesClient) ListCronJobs(ctx context.Context, input ListCronJobsInput) ([]batchv1.CronJob, error) {
	listOptions := metav1.ListOptions{}
	if input.LabelSelector != "" {
		listOptions.LabelSelector = input.LabelSelector
	}

	cronJobs, err := k.Client.BatchV1().CronJobs(input.Namespace).List(ctx, listOptions)
	if err != nil {
		return []batchv1.CronJob{}, err
	}

	if cronJobs == nil {
		return []batchv1.CronJob{}, &NilResponseError{"cron jobs is nil"}
	}

	return cronJobs.Items, err
}

// ListDeploymentsInput contains all the information needed to list Kubernetes deployments
type ListDeploymentsInput struct {
	// Namespace is the Kubernetes namespace
	Namespace string

	// LabelSelector is the Kubernetes label selector
	LabelSelector string
}

// ListDeployments lists Kubernetes deployments
func (k KubernetesClient) ListDeployments(ctx context.Context, input ListDeploymentsInput) ([]appsv1.Deployment, error) {
	listOptions := metav1.ListOptions{LabelSelector: input.LabelSelector}
	deployments, err := k.Client.AppsV1().Deployments(input.Namespace).List(ctx, listOptions)
	if err != nil {
		return []appsv1.Deployment{}, err
	}

	if deployments == nil {
		return []appsv1.Deployment{}, &NilResponseError{"deployments list is nil"}
	}

	return deployments.Items, nil
}

// ListIngressesInput contains all the information needed to list Kubernetes ingresses
type ListIngressesInput struct {
	// Namespace is the Kubernetes namespace
	Namespace string

	// LabelSelector is the Kubernetes label selector
	LabelSelector string
}

// ListIngresses lists Kubernetes ingresses
func (k KubernetesClient) ListIngresses(ctx context.Context, input ListIngressesInput) ([]networkingv1.Ingress, error) {
	listOptions := metav1.ListOptions{LabelSelector: input.LabelSelector}
	ingresses, err := k.Client.NetworkingV1().Ingresses(input.Namespace).List(ctx, listOptions)
	if err != nil {
		return []networkingv1.Ingress{}, err
	}

	if ingresses == nil {
		return []networkingv1.Ingress{}, &NilResponseError{"ingresses is nil"}
	}

	return ingresses.Items, nil
}

// ListNamespaces lists Kubernetes namespaces
func (k KubernetesClient) ListNamespaces(ctx context.Context) ([]corev1.Namespace, error) {
	namespaces, err := k.Client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return []corev1.Namespace{}, err
	}
	if namespaces == nil {
		return []corev1.Namespace{}, &NilResponseError{"namespaces list is nil"}
	}

	return namespaces.Items, nil
}

// ListNodesInput contains all the information needed to list Kubernetes nodes
type ListNodesInput struct {
	// LabelSelector is the Kubernetes label selector
	LabelSelector string
}

// ListNodes lists Kubernetes nodes
func (k KubernetesClient) ListNodes(ctx context.Context, input ListNodesInput) ([]corev1.Node, error) {
	listOptions := metav1.ListOptions{}
	if input.LabelSelector != "" {
		common.LogDebug(fmt.Sprintf("Using label selector: %s", input.LabelSelector))
		listOptions.LabelSelector = input.LabelSelector
	}
	nodeList, err := k.Client.CoreV1().Nodes().List(ctx, listOptions)
	if err != nil {
		return []corev1.Node{}, err
	}

	if nodeList == nil {
		return []corev1.Node{}, &NilResponseError{"pod list is nil"}
	}

	return nodeList.Items, err
}

// ListPodsInput contains all the information needed to list Kubernetes pods
type ListPodsInput struct {
	// Namespace is the Kubernetes namespace
	Namespace string

	// LabelSelector is the Kubernetes label selector
	LabelSelector string
}

// ListPods lists Kubernetes pods
func (k KubernetesClient) ListPods(ctx context.Context, input ListPodsInput) ([]corev1.Pod, error) {
	listOptions := metav1.ListOptions{LabelSelector: input.LabelSelector}
	podList, err := k.Client.CoreV1().Pods(input.Namespace).List(ctx, listOptions)
	if err != nil {
		return []corev1.Pod{}, err
	}

	if podList == nil {
		return []corev1.Pod{}, &NilResponseError{"pod list is nil"}
	}

	return podList.Items, err
}

// ListTriggerAuthenticationsInput contains all the information needed to list Kubernetes trigger authentications
type ListTriggerAuthenticationsInput struct {
	// Namespace is the Kubernetes namespace
	Namespace string

	// LabelSelector is the Kubernetes label selector
	LabelSelector string
}

// ListTriggerAuthentications lists Kubernetes trigger authentications
func (k KubernetesClient) ListTriggerAuthentications(ctx context.Context, input ListTriggerAuthenticationsInput) ([]kedav1alpha1.TriggerAuthentication, error) {
	listOptions := metav1.ListOptions{LabelSelector: input.LabelSelector}

	gvr := schema.GroupVersionResource{
		Group:    "keda.sh",
		Version:  "v1alpha1",
		Resource: "triggerauthentications",
	}

	response, err := k.DynamicClient.Resource(gvr).Namespace(input.Namespace).List(ctx, listOptions)
	if err != nil {
		return []kedav1alpha1.TriggerAuthentication{}, err
	}

	if response == nil {
		return []kedav1alpha1.TriggerAuthentication{}, &NilResponseError{"trigger authentications is nil"}
	}

	triggerAuthentications := []kedav1alpha1.TriggerAuthentication{}
	for _, triggerAuthentication := range response.Items {
		var ta kedav1alpha1.TriggerAuthentication
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(triggerAuthentication.Object, &ta)
		if err != nil {
			return []kedav1alpha1.TriggerAuthentication{}, err
		}

		triggerAuthentications = append(triggerAuthentications, ta)
	}

	return triggerAuthentications, nil
}

// ResumeCronJobsInput contains all the information needed to resume a Kubernetes cron job
type ResumeCronJobsInput struct {
	// LabelSelector is the Kubernetes label selector
	LabelSelector string

	// Namespace is the Kubernetes namespace
	Namespace string
}

// ResumeCronJobs resumes a Kubernetes cron job
func (k KubernetesClient) ResumeCronJobs(ctx context.Context, input ResumeCronJobsInput) error {
	cronJobs, err := k.Client.BatchV1().CronJobs(input.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: input.LabelSelector,
	})
	if err != nil {
		return err
	}

	for _, cronJob := range cronJobs.Items {
		cronJob.Spec.Suspend = ptr.To(false)
		_, err := k.Client.BatchV1().CronJobs(input.Namespace).Update(ctx, &cronJob, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

// ScaleDeploymentInput contains all the information needed to scale a Kubernetes deployment
type ScaleDeploymentInput struct {
	// Name is the Kubernetes deployment name
	Name string

	// Namespace is the Kubernetes namespace
	Namespace string

	// Replicas is the number of replicas to scale to
	Replicas int32
}

// ScaleDeployment scales a Kubernetes deployment
func (k KubernetesClient) ScaleDeployment(ctx context.Context, input ScaleDeploymentInput) error {
	_, err := k.Client.AppsV1().Deployments(input.Namespace).UpdateScale(ctx, input.Name, &autoscalingv1.Scale{
		ObjectMeta: metav1.ObjectMeta{
			Name:      input.Name,
			Namespace: input.Namespace,
		},
		Spec: autoscalingv1.ScaleSpec{
			Replicas: input.Replicas,
		},
	}, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

type StreamLogsInput struct {
	// ContainerName is the Kubernetes container name
	ContainerName string

	// Follow is whether to follow the logs
	Follow bool

	// LabelSelector is the Kubernetes label selector
	LabelSelector []string

	// Namespace is the Kubernetes namespace
	Namespace string

	// Quiet is whether to suppress output
	Quiet bool

	// SinceSeconds is the number of seconds to go back
	SinceSeconds int64

	// TailLines is the number of lines to tail
	TailLines int64
}

func (k KubernetesClient) StreamLogs(ctx context.Context, input StreamLogsInput) error {
	ctx, cancel := context.WithCancel(ctx)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM)
	go func() {
		<-signals
		cancel()
	}()

	if err := k.Ping(); err != nil {
		return fmt.Errorf("kubernetes api not available: %w", err)
	}

	labelSelector := input.LabelSelector
	processIndex := 0
	if input.ContainerName != "" {
		parts := strings.SplitN(input.ContainerName, ".", 2)
		if len(parts) == 2 {
			var err error
			input.ContainerName = parts[0]
			processIndex, err = strconv.Atoi(parts[1])
			if err != nil {
				return fmt.Errorf("Error parsing process index: %w", err)
			}
		}
		labelSelector = append(labelSelector, fmt.Sprintf("app.kubernetes.io/name=%s", input.ContainerName))
	}

	pods, err := k.ListPods(ctx, ListPodsInput{
		Namespace:     input.Namespace,
		LabelSelector: strings.Join(labelSelector, ","),
	})
	if err != nil {
		return fmt.Errorf("Error listing pods: %w", err)
	}
	if len(pods) == 0 {
		return fmt.Errorf("No pods found matching specified labels")
	}

	if os.Getenv("FORCE_TTY") == "1" {
		color.NoColor = false
	}

	logOptions := corev1.PodLogOptions{
		Follow: input.Follow,
	}
	if input.TailLines > 0 {
		logOptions.TailLines = ptr.To(input.TailLines)
	}
	if input.SinceSeconds != 0 {
		// round up to the nearest second
		sec := int64(time.Duration(input.SinceSeconds * int64(time.Second)).Seconds())
		logOptions.SinceSeconds = &sec
	}

	requests := make([]rest.ResponseWrapper, len(pods))
	for i := 0; i < len(pods); i++ {
		if processIndex > 0 && i != (processIndex-1) {
			continue
		}

		podName := pods[i].Name
		requests[i] = k.Client.CoreV1().Pods(input.Namespace).GetLogs(podName, &logOptions)
	}

	reader, writer := io.Pipe()
	wg := &sync.WaitGroup{}
	wg.Add(len(requests))

	colors := []color.Attribute{
		color.FgRed,
		color.FgYellow,
		color.FgGreen,
		color.FgCyan,
		color.FgBlue,
		color.FgMagenta,
	}

	for i := 0; i < len(requests); i++ {
		request := requests[i]
		podName := pods[i].Name
		podColor := colors[i%len(colors)]
		dynoText := color.New(podColor).SprintFunc()
		prefix := dynoText(fmt.Sprintf("app[%s]: ", podName))

		go func(ctx context.Context, request rest.ResponseWrapper, prefix string) {
			defer wg.Done()

			out := k.addPrefixingWriter(writer, prefix, input.Quiet)
			if err := streamLogsFromRequest(ctx, request, out); err != nil {
				// check if error is context canceled
				if errors.Is(err, context.Canceled) {
					writer.Close()
					return
				}

				writer.CloseWithError(err)
				return
			}
		}(ctx, request, prefix)
	}

	go func() {
		wg.Wait()
		writer.Close()
	}()

	_, err = io.Copy(os.Stdout, reader)
	return err
}

func (k KubernetesClient) addPrefixingWriter(writer *io.PipeWriter, prefix string, quiet bool) io.Writer {
	if quiet {
		return writer
	}

	return &common.PrefixingWriter{
		Prefix: []byte(prefix),
		Writer: writer,
	}
}

func streamLogsFromRequest(ctx context.Context, request rest.ResponseWrapper, out io.Writer) error {
	readCloser, err := request.Stream(ctx)
	if err != nil {
		return err
	}
	defer readCloser.Close()

	r := bufio.NewReader(readCloser)
	for {
		bytes, err := r.ReadBytes('\n')

		if _, err := out.Write(bytes); err != nil {
			return err
		}

		if err != nil {
			if err != io.EOF {
				return err
			}
			return nil
		}
	}
}

// SuspendCronJobsInput contains all the information needed to suspend a Kubernetes cron job
type SuspendCronJobsInput struct {
	// LabelSelector is the Kubernetes label selector
	LabelSelector string

	// Namespace is the Kubernetes namespace
	Namespace string
}

// SuspendCronJobs suspends a Kubernetes cron job
func (k KubernetesClient) SuspendCronJobs(ctx context.Context, input SuspendCronJobsInput) error {
	cronJobs, err := k.Client.BatchV1().CronJobs(input.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: input.LabelSelector,
	})
	if err != nil {
		return err
	}

	for _, cronJob := range cronJobs.Items {
		cronJob.Spec.Suspend = ptr.To(true)
		_, err := k.Client.BatchV1().CronJobs(input.Namespace).Update(ctx, &cronJob, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}
