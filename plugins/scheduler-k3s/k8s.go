package scheduler_k3s

import (
	"context"
	"errors"
	"fmt"

	"github.com/dokku/dokku/plugins/common"
	"github.com/go-openapi/jsonpointer"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/utils/ptr"
)

// KubernetesClient is a wrapper around the Kubernetes client
type KubernetesClient struct {
	// Client is the Kubernetes client
	Client kubernetes.Clientset

	// RestClient is the Kubernetes REST client
	RestClient rest.Interface

	// RestConfig is the Kubernetes REST config
	RestConfig rest.Config
}

// NewKubernetesClient creates a new Kubernetes client
func NewKubernetesClient() (KubernetesClient, error) {
	clientConfig := KubernetesClientConfig()
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

	return KubernetesClient{
		Client:     *client,
		RestConfig: *restConf,
		RestClient: restClient,
	}, nil
}

// KubernetesClientConfig returns a Kubernetes client config
func KubernetesClientConfig() clientcmd.ClientConfig {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: KubeConfigPath},
		&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: ""}})
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
		return errors.New("node is nil")
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
	upgradeCmd, err := common.CallExecCommand(common.ExecCommandInput{
		Command: "kubectl",
		Args: []string{
			"apply",
			"-f",
			input.Manifest,
		},
		Env: map[string]string{
			"KUBECONFIG": KubeConfigPath,
		},
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
		return batchv1.Job{}, errors.New("job is nil")
	}

	return *job, err
}

// CreateNamespaceInput contains all the information needed to create a Kubernetes namespace
type CreateNamespaceInput struct {
	// Name is the name of the Kubernetes namespace
	Name v1.Namespace
}

// CreateNamespace creates a Kubernetes namespace
func (k KubernetesClient) CreateNamespace(ctx context.Context, input CreateNamespaceInput) (v1.Namespace, error) {
	namespaces, err := k.ListNamespaces(ctx)
	if err != nil {
		return v1.Namespace{}, err
	}

	for _, namespace := range namespaces {
		if namespace.Name == input.Name.Name {
			return namespace, nil
		}
	}

	namespace, err := k.Client.CoreV1().Namespaces().Create(ctx, &input.Name, metav1.CreateOptions{})
	if err != nil {
		return v1.Namespace{}, err
	}

	if namespace == nil {
		return v1.Namespace{}, errors.New("namespace is nil")
	}

	return *namespace, err
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
		return Node{}, errors.New("node is nil")
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

// GetJob gets a Kubernetes job
func (k KubernetesClient) GetPod(ctx context.Context, input GetPodInput) (v1.Pod, error) {
	pod, err := k.Client.CoreV1().Pods(input.Namespace).Get(ctx, input.Name, metav1.GetOptions{})
	if err != nil {
		return v1.Pod{}, err
	}

	if pod == nil {
		return v1.Pod{}, errors.New("pod is nil")
	}

	return *pod, err
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
		return errors.New("node is nil")
	}

	keyPath := fmt.Sprintf("/metadata/labels/%s", jsonpointer.Escape(input.Key))
	patch := fmt.Sprintf(`[{"op":"add", "path":"%s", "value":"%s" }]`, keyPath, input.Value)
	_, err = k.Client.CoreV1().Nodes().Patch(ctx, node.Name, types.JSONPatchType, []byte(patch), metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("failed to label node: %w", err)
	}

	return nil
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
		return []batchv1.CronJob{}, errors.New("cron jobs is nil")
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
		return []appsv1.Deployment{}, errors.New("deployments is nil")
	}

	return deployments.Items, nil
}

// ListNamespaces lists Kubernetes namespaces
func (k KubernetesClient) ListNamespaces(ctx context.Context) ([]v1.Namespace, error) {
	namespaces, err := k.Client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return []v1.Namespace{}, err
	}
	if namespaces == nil {
		return []v1.Namespace{}, errors.New("namespaces is nil")
	}

	return namespaces.Items, nil
}

// ListNodesInput contains all the information needed to list Kubernetes nodes
type ListNodesInput struct {
	// LabelSelector is the Kubernetes label selector
	LabelSelector string
}

// ListNodes lists Kubernetes nodes
func (k KubernetesClient) ListNodes(ctx context.Context, input ListNodesInput) ([]v1.Node, error) {
	listOptions := metav1.ListOptions{}
	if input.LabelSelector != "" {
		common.LogDebug(fmt.Sprintf("Using label selector: %s", input.LabelSelector))
		listOptions.LabelSelector = input.LabelSelector
	}
	nodeList, err := k.Client.CoreV1().Nodes().List(ctx, listOptions)
	if err != nil {
		return []v1.Node{}, err
	}

	if nodeList == nil {
		return []v1.Node{}, errors.New("pod list is nil")
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
func (k KubernetesClient) ListPods(ctx context.Context, input ListPodsInput) ([]v1.Pod, error) {
	listOptions := metav1.ListOptions{LabelSelector: input.LabelSelector}
	podList, err := k.Client.CoreV1().Pods(input.Namespace).List(ctx, listOptions)
	if err != nil {
		return []v1.Pod{}, err
	}

	if podList == nil {
		return []v1.Pod{}, errors.New("pod list is nil")
	}

	return podList.Items, err
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
