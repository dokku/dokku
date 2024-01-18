package scheduler_k3s

import (
	"context"
	"errors"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	// Namespace is the Kubernetes namespace
	Namespace v1.Namespace
}

// CreateNamespace creates a Kubernetes namespace
func (k KubernetesClient) CreateNamespace(ctx context.Context, input CreateNamespaceInput) (v1.Namespace, error) {
	namespace, err := k.Client.CoreV1().Namespaces().Create(ctx, &input.Namespace, metav1.CreateOptions{})
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
