package scheduler_k3s

import (
	"crypto/rand"
	"fmt"
	"os"
	"strings"

	acmev1 "github.com/cert-manager/cert-manager/pkg/apis/acme/v1"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	certmanagermetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/dokku/dokku/plugins/common"
	"gopkg.in/yaml.v3"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

type Chart struct {
	ApiVersion string `yaml:"apiVersion"`
	AppVersion string `yaml:"appVersion"`
	Name       string `yaml:"name"`
	Version    string `yaml:"version"`
}

type Values struct {
	Global    GlobalValues             `yaml:"global"`
	Processes map[string]ProcessValues `yaml:"processes"`
}

type GlobalValues struct {
	AppName      string            `yaml:"app_name"`
	DeploymentID string            `yaml:"deploment_id"`
	Image        GlobalImage       `yaml:"image"`
	Namespace    string            `yaml:"namespace"`
	PrimaryPort  int32             `yaml:"primary_port"`
	Secrets      map[string]string `yaml:"secrets,omitempty"`
}

type GlobalImage struct {
	ImagePullSecrets string `yaml:"image_pull_secrets"`
	Name             string `yaml:"name"`
	Type             string `yaml:"type"`
	WorkingDir       string `yaml:"working_dir"`
}

type ProcessValues struct {
	Args         []string            `yaml:"args,omitempty"`
	Cron         ProcessCron         `yaml:"cron,omitempty"`
	Healthchecks ProcessHealthchecks `yaml:"healthchecks,omitempty"`
	ProcessType  ProcessType         `yaml:"process_type"`
	Replicas     int32               `yaml:"replicas"`
	Resources    ProcessResourcesMap `yaml:"resources,omitempty"`
	Web          ProcessWeb          `yaml:"web,omitempty"`
}

type ProcessHealthchecks struct {
	Liveness        ProcessHealthcheck `yaml:"liveness,omitempty"`
	Readiness       ProcessHealthcheck `yaml:"readiness,omitempty"`
	Startup         ProcessHealthcheck `yaml:"startup,omitempty"`
	MinReadySeconds int32              `yaml:"min_ready_seconds,omitempty"`
}

type ProcessHealthcheck struct {
	Exec      *ExecHealthcheck `yaml:"exec,omitempty"`
	HTTPGet   *HTTPHealthcheck `yaml:"httpGet,omitempty"`
	TCPSocket *TCPHealthcheck  `yaml:"tcpSocket,omitempty"`

	InitialDelaySeconds           int32  `yaml:"initialDelaySeconds,omitempty"`
	TimeoutSeconds                int32  `yaml:"timeoutSeconds,omitempty"`
	PeriodSeconds                 int32  `yaml:"periodSeconds,omitempty"`
	SuccessThreshold              int32  `yaml:"successThreshold,omitempty"`
	FailureThreshold              int32  `yaml:"failureThreshold,omitempty"`
	TerminationGracePeriodSeconds *int64 `yaml:"terminationGracePeriodSeconds,omitempty"`
}

type ExecHealthcheck struct {
	Command []string `yaml:"command,omitempty"`
}

type HTTPHealthcheck struct {
	Path        string       `yaml:"path,omitempty"`
	Port        int32        `yaml:"port,omitempty"`
	Host        string       `yaml:"host,omitempty"`
	Scheme      URIScheme    `yaml:"scheme,omitempty"`
	HTTPHeaders []HTTPHeader `yaml:"httpHeaders,omitempty"`
}

type TCPHealthcheck struct {
	Port int32  `yaml:"port,omitempty"`
	Host string `yaml:"host,omitempty"`
}

type HTTPHeader struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type URIScheme string

const (
	URISchemeHTTP  URIScheme = "HTTP"
	URISchemeHTTPS URIScheme = "HTTPS"
)

type ProcessWeb struct {
	Domains  []string         `yaml:"domains,omitempty"`
	PortMaps []ProcessPortMap `yaml:"port_maps,omitempty"`
	TLS      ProcessTls       `yaml:"tls"`
}

type ProcessResourcesMap struct {
	Limits   ProcessResources `yaml:"limits,omitempty"`
	Requests ProcessResources `yaml:"requests,omitempty"`
}

type ProcessResources struct {
	NvidiaGPU string `yaml:"nvidia.com/gpu,omitempty"`
	CPU       string `yaml:"cpu,omitempty"`
	Memory    string `yaml:"memory,omitempty"`
}

type ProcessType string

const (
	ProcessType_Cron   ProcessType = "cron"
	ProcessType_Job    ProcessType = "job"
	ProcessType_Web    ProcessType = "web"
	ProcessType_Worker ProcessType = "worker"
)

type ProcessCron struct {
	ID       string `yaml:"id"`
	Schedule string `yaml:"schedule"`
	Suffix   string `yaml:"suffix"`
}

type ProcessPortMap struct {
	ContainerPort int32           `yaml:"container_port"`
	HostPort      int32           `yaml:"host_port"`
	Scheme        string          `yaml:"scheme"`
	Protocol      PortmapProtocol `yaml:"protocol"`
	Name          string          `yaml:"name"`
}

type PortmapProtocol string

const (
	PortmapProtocol_TCP PortmapProtocol = "TCP"
	PortmapProtocol_UDP PortmapProtocol = "UDP"
)

type NameSorter []ProcessPortMap

func (a NameSorter) Len() int           { return len(a) }
func (a NameSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a NameSorter) Less(i, j int) bool { return a[i].Name < a[j].Name }

type ProcessTls struct {
	Enabled    bool   `yaml:"enabled"`
	IssuerName string `yaml:"issuer_name"`
}

type ClusterIssuer struct {
	Email        string
	IngressClass string
	Name         string
	Namespace    string
	Server       string
}

func templateKubernetesClusterIssuer(input ClusterIssuer) (certmanagerv1.ClusterIssuer, error) {
	if input.Email == "" {
		return certmanagerv1.ClusterIssuer{}, fmt.Errorf("Email cannot be empty")
	}
	if input.Server == "" {
		return certmanagerv1.ClusterIssuer{}, fmt.Errorf("Server cannot be empty")
	}
	if input.Name == "" {
		return certmanagerv1.ClusterIssuer{}, fmt.Errorf("Name cannot be empty")
	}
	if input.Namespace == "" {
		return certmanagerv1.ClusterIssuer{}, fmt.Errorf("Namespace cannot be empty")
	}

	if input.Server == "staging" {
		input.Server = "https://acme-staging-v02.api.letsencrypt.org/directory"
	} else if input.Server == "production" {
		input.Server = "https://acme-v02.api.letsencrypt.org/directory"
	} else {
		return certmanagerv1.ClusterIssuer{}, fmt.Errorf("Server must be either staging or production")
	}

	clusterIssuer := certmanagerv1.ClusterIssuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      input.Name,
			Namespace: input.Namespace,
		},
		Spec: certmanagerv1.IssuerSpec{
			IssuerConfig: certmanagerv1.IssuerConfig{
				ACME: &acmev1.ACMEIssuer{
					Email:  input.Email,
					Server: input.Server,
					PrivateKey: certmanagermetav1.SecretKeySelector{
						LocalObjectReference: certmanagermetav1.LocalObjectReference{
							Name: input.Name,
						},
					},
					Solvers: []acmev1.ACMEChallengeSolver{
						{
							HTTP01: &acmev1.ACMEChallengeSolverHTTP01{
								Ingress: &acmev1.ACMEChallengeSolverHTTP01Ingress{
									Class: ptr.To(input.IngressClass),
								},
							},
						},
					},
				},
			},
		},
	}

	return clusterIssuer, nil
}

type Job struct {
	AppName          string
	Command          []string
	DeploymentID     int64
	Entrypoint       string
	Env              map[string]string
	ID               string
	Image            string
	ImagePullSecrets string
	ImageSourceType  string
	Interactive      bool
	Labels           map[string]string
	Namespace        string
	ProcessType      string
	Schedule         string
	Suffix           string
	RemoveContainer  bool
	WorkingDir       string
}

func templateKubernetesJob(input Job) (batchv1.Job, error) {
	labels := map[string]string{
		"app.kubernetes.io/instance": fmt.Sprintf("%s-%s", input.AppName, input.ProcessType),
		"app.kubernetes.io/name":     input.ProcessType,
		"app.kubernetes.io/part-of":  input.AppName,
	}
	annotations := map[string]string{
		"app.kubernetes.io/version": fmt.Sprint(input.DeploymentID),
		"dokku.com/builder-type":    input.ImageSourceType,
		"dokku.com/managed":         "true",
	}

	for key, value := range input.Labels {
		labels[key] = value
	}
	secretName := fmt.Sprintf("env-%s.%d", input.AppName, input.DeploymentID)

	env := []corev1.EnvVar{}
	for key, value := range input.Env {
		env = append(env, corev1.EnvVar{
			Name:  key,
			Value: value,
		})
	}

	suffix := input.Suffix
	if suffix == "" {
		n := 5
		b := make([]byte, n)
		if _, err := rand.Read(b); err != nil {
			panic(err)
		}
		suffix = strings.ToLower(fmt.Sprintf("%X", b))
	}
	annotations["dokku.com/job-suffix"] = suffix

	podAnnotations := annotations
	podAnnotations["kubectl.kubernetes.io/default-container"] = fmt.Sprintf("%s-%s", input.AppName, input.ProcessType)

	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-%s-%s", input.AppName, input.ProcessType, suffix),
			Namespace:   input.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: ptr.To(int32(0)),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: podAnnotations,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Args: input.Command,
							Name: fmt.Sprintf("%s-%s", input.AppName, input.ProcessType),
							Env:  env,
							EnvFrom: []corev1.EnvFromSource{
								{
									SecretRef: &corev1.SecretEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: secretName,
										},
										Optional: ptr.To(true),
									},
								},
							},
							Image:           input.Image,
							ImagePullPolicy: corev1.PullAlways,
							Resources: corev1.ResourceRequirements{
								Limits:   corev1.ResourceList{},
								Requests: corev1.ResourceList{},
							},
							WorkingDir: input.WorkingDir,
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}

	if input.Entrypoint != "" {
		job.Spec.Template.Spec.Containers[0].Command = []string{input.Entrypoint}
	}

	if input.Interactive {
		job.Spec.Template.Spec.Containers[0].Stdin = true
		job.Spec.Template.Spec.Containers[0].StdinOnce = true
		job.Spec.Template.Spec.Containers[0].TTY = true
	}

	if input.RemoveContainer {
		job.Spec.TTLSecondsAfterFinished = ptr.To(int32(60))
	}

	if input.ImagePullSecrets != "" {
		job.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
			{
				Name: input.ImagePullSecrets,
			},
		}
	}

	processResources, err := getProcessResources(input.AppName, input.ProcessType)
	if err != nil {
		return job, fmt.Errorf("Error getting process resources: %w", err)
	}
	if processResources.Limits.CPU != "" {
		cpuQuantity, _ := resource.ParseQuantity(processResources.Limits.CPU)
		job.Spec.Template.Spec.Containers[0].Resources.Limits["cpu"] = cpuQuantity
	}
	if processResources.Limits.Memory != "" {
		memoryQuantity, _ := resource.ParseQuantity(processResources.Limits.Memory)
		job.Spec.Template.Spec.Containers[0].Resources.Limits["memory"] = memoryQuantity
	}
	if processResources.Limits.NvidiaGPU != "" {
		nvidiaGpuQuantity, _ := resource.ParseQuantity(processResources.Limits.NvidiaGPU)
		job.Spec.Template.Spec.Containers[0].Resources.Limits["nvidia.com/gpu"] = nvidiaGpuQuantity
	}
	if processResources.Requests.CPU != "" {
		cpuQuantity, _ := resource.ParseQuantity(processResources.Requests.CPU)
		job.Spec.Template.Spec.Containers[0].Resources.Requests["cpu"] = cpuQuantity
	}
	if processResources.Requests.Memory != "" {
		memoryQuantity, _ := resource.ParseQuantity(processResources.Requests.Memory)
		job.Spec.Template.Spec.Containers[0].Resources.Requests["memory"] = memoryQuantity
	}

	return job, nil
}

type WriteYamlInput struct {
	Object interface{}
	Path   string
}

func writeYaml(input WriteYamlInput) error {
	common.LogDebug(fmt.Sprintf("Printing resource: %s", input.Path))
	data, err := yaml.Marshal(input.Object)
	if err != nil {
		return fmt.Errorf("Error marshalling chart: %w", err)
	}

	err = os.WriteFile(input.Path, data, os.FileMode(0644))
	if err != nil {
		return fmt.Errorf("Error writing chart: %w", err)
	}

	if os.Getenv("DOKKU_TRACE") == "1" {
		common.CatFile(input.Path)
	}

	return nil
}
