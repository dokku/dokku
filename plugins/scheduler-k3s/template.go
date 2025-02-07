package scheduler_k3s

import (
	"crypto/rand"
	"fmt"
	"os"
	"strings"

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
	Icon       string `yaml:"icon"`
}

type ClusterIssuerValues struct {
	ClusterIssuers map[string]ClusterIssuer `yaml:"cluster_issuers"`
}

type ClusterKedaValues struct {
	Global struct {
		Annotations ProcessAnnotations `yaml:"annotations,omitempty"`
	} `yaml:"global"`
	Secrets map[string]string `yaml:"secrets"`
	Type    string            `yaml:"type"`
}

type AppValues struct {
	Global    GlobalValues             `yaml:"global"`
	Processes map[string]ProcessValues `yaml:"processes"`
}

type GlobalValues struct {
	Annotations  ProcessAnnotations `yaml:"annotations,omitempty"`
	AppName      string             `yaml:"app_name"`
	DeploymentID string             `yaml:"deployment_id"`
	Image        GlobalImage        `yaml:"image"`
	Labels       ProcessLabels      `yaml:"labels,omitempty"`
	Keda         GlobalKedaValues   `yaml:"keda"`
	Namespace    string             `yaml:"namespace"`
	Network      GlobalNetwork      `yaml:"network"`
	Secrets      map[string]string  `yaml:"secrets,omitempty"`
}

type GlobalImage struct {
	ImagePullSecrets string `yaml:"image_pull_secrets"`
	Name             string `yaml:"name"`
	PullSecretBase64 string `yaml:"pull_secret_base64"`
	Type             string `yaml:"type"`
	WorkingDir       string `yaml:"working_dir"`
}

type GlobalNetwork struct {
	// IngressClass is the default ingress class to use
	IngressClass string `yaml:"ingress_class"`

	// PrimaryPort is the primary port to use
	PrimaryPort int32 `yaml:"primary_port"`

	// PrimaryServicePort is the primary service port to use
	PrimaryServicePort int32 `yaml:"primary_service_port"`
}

// GlobalKedaValues contains the global keda configuration
type GlobalKedaValues struct {
	// Authentications is a map of authentication objects to use for keda
	Authentications map[string]KedaAuthentication `yaml:"authentications"`

	// GlobalAuthentications is a map of global authentication objects to use for keda
	GlobalAuthentications map[string]KedaAuthentication `yaml:"global_authentications"`
}

// KedaAuthentication contains the authentication configuration for keda
type KedaAuthentication struct {
	// Name is the name of the authentication object
	Name string `yaml:"name"`

	// Kind is the kind of authentication object
	Kind KedaAuthenticationKind `yaml:"kind,omitempty"`

	// Type is the type of authentication to use
	Type string `yaml:"type"`

	// Secrets is a map of secrets to use for authentication
	Secrets map[string]string `yaml:"secrets,omitempty"`
}

type KedaAuthenticationKind string

const (
	KedaAuthenticationKind_ClusterTriggerAuthentication KedaAuthenticationKind = "ClusterTriggerAuthentication"
	KedaAuthenticationKind_TriggerAuthentication        KedaAuthenticationKind = "TriggerAuthentication"
)

type ProcessValues struct {
	Annotations  ProcessAnnotations  `yaml:"annotations,omitempty"`
	Args         []string            `yaml:"args,omitempty"`
	Autoscaling  ProcessAutoscaling  `yaml:"autoscaling,omitempty"`
	Cron         ProcessCron         `yaml:"cron,omitempty"`
	Healthchecks ProcessHealthchecks `yaml:"healthchecks,omitempty"`
	Labels       ProcessLabels       `yaml:"labels,omitempty"`
	ProcessType  ProcessType         `yaml:"process_type"`
	Replicas     int32               `yaml:"replicas"`
	Resources    ProcessResourcesMap `yaml:"resources,omitempty"`
	Web          ProcessWeb          `yaml:"web,omitempty"`
	Volumes      []ProcessVolume     `yaml:"volumes,omitempty"`
}

type ProcessAnnotations struct {
	CertificateAnnotations               map[string]string `yaml:"certificate,omitempty"`
	CronJobAnnotations                   map[string]string `yaml:"cronjob,omitempty"`
	DeploymentAnnotations                map[string]string `yaml:"deployment,omitempty"`
	IngressAnnotations                   map[string]string `yaml:"ingress,omitempty"`
	JobAnnotations                       map[string]string `yaml:"job,omitempty"`
	KedaScalingObjectAnnotations         map[string]string `yaml:"keda_scaled_object,omitempty"`
	KedaHTTPScaledObjectAnnotations      map[string]string `yaml:"keda_http_scaled_object,omitempty"`
	KedaInterceptorProxyAnnotations      map[string]string `yaml:"keda_interceptor_proxy,omitempty"`
	KedaSecretAnnotations                map[string]string `yaml:"keda_secret,omitempty"`
	KedaTriggerAuthenticationAnnotations map[string]string `yaml:"keda_trigger_authentication,omitempty"`
	PodAnnotations                       map[string]string `yaml:"pod,omitempty"`
	SecretAnnotations                    map[string]string `yaml:"secret,omitempty"`
	ServiceAccountAnnotations            map[string]string `yaml:"serviceaccount,omitempty"`
	ServiceAnnotations                   map[string]string `yaml:"service,omitempty"`
	TraefikIngressRouteAnnotations       map[string]string `yaml:"traefik_ingressroute,omitempty"`
	TraefikMiddlewareAnnotations         map[string]string `yaml:"traefik_middleware,omitempty"`
	PvcAnnotations                       map[string]string `yaml:"pvc,omitempty"`
}

// ProcessAutoscaling contains the autoscaling configuration for a process
type ProcessAutoscaling struct {
	// CooldownPeriodSeconds is the number of seconds after a scaling event before another can be triggered
	CooldownPeriodSeconds int `yaml:"cooldown_period_seconds,omitempty"`

	// Enabled is a flag to enable autoscaling
	Enabled bool `yaml:"enabled"`

	// MaxReplicas is the maximum number of replicas to scale to
	MaxReplicas int `yaml:"max_replicas,omitempty"`

	// MinReplicas is the minimum number of replicas to scale to
	MinReplicas int `yaml:"min_replicas"`

	// PollingIntervalSeconds is the number of seconds between polling for new metrics
	PollingIntervalSeconds int `yaml:"polling_interval_seconds,omitempty"`

	// HttpTrigger is the http trigger config to use for autoscaling
	HttpTrigger ProcessAutoscalingTrigger `yaml:"http_trigger,omitempty"`

	// Triggers is a list of triggers to use for autoscaling
	Triggers []ProcessAutoscalingTrigger `yaml:"triggers,omitempty"`

	// Type is the type of autoscaling to use
	Type string `yaml:"type"`
}

// ProcessAutoscalingTrigger is a trigger to use for autoscaling
type ProcessAutoscalingTrigger struct {
	// Name is the name of the trigger
	Name string `yaml:"name"`

	// Type is the type of trigger to use
	Type string `yaml:"type"`

	// Metadata is a map of key-value pairs that can be used to store arbitrary trigger data
	Metadata map[string]string `yaml:"metadata,omitempty"`

	// AuthenticationRef is a reference to an authentication object
	AuthenticationRef *ProcessAutoscalingTriggerAuthenticationRef `yaml:"authenticationRef,omitempty"`
}

// ProcessAutoscalingTriggerAuthenticationRef is a reference to an authentication object
type ProcessAutoscalingTriggerAuthenticationRef struct {
	// Name is the name of the authentication object
	Name string `yaml:"name"`

	// Kind is the kind of authentication object
	Kind string `yaml:"kind,omitempty"`
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

type ProcessLabels struct {
	CertificateLabels               map[string]string `yaml:"certificate,omitempty"`
	CronJobLabels                   map[string]string `yaml:"cronjob,omitempty"`
	DeploymentLabels                map[string]string `yaml:"deployment,omitempty"`
	IngressLabels                   map[string]string `yaml:"ingress,omitempty"`
	JobLabels                       map[string]string `yaml:"job,omitempty"`
	KedaScalingObjectLabels         map[string]string `yaml:"keda_scaled_object,omitempty"`
	KedaHTTPScaledObjectLabels      map[string]string `yaml:"keda_http_scaled_object,omitempty"`
	KedaInterceptorProxyLabels      map[string]string `yaml:"keda_interceptor_proxy,omitempty"`
	KedaSecretLabels                map[string]string `yaml:"keda_secret,omitempty"`
	KedaTriggerAuthenticationLabels map[string]string `yaml:"keda_trigger_authentication,omitempty"`
	PodLabels                       map[string]string `yaml:"pod,omitempty"`
	SecretLabels                    map[string]string `yaml:"secret,omitempty"`
	ServiceAccountLabels            map[string]string `yaml:"serviceaccount,omitempty"`
	ServiceLabels                   map[string]string `yaml:"service,omitempty"`
	TraefikIngressRouteLabels       map[string]string `yaml:"traefik_ingressroute,omitempty"`
	TraefikMiddlewareLabels         map[string]string `yaml:"traefik_middleware,omitempty"`
	PvcLabels                       map[string]string `yaml:"pvc,omitempty"`
}

type ProcessWeb struct {
	Domains  []ProcessDomains `yaml:"domains,omitempty"`
	PortMaps []ProcessPortMap `yaml:"port_maps,omitempty"`
	TLS      ProcessTls       `yaml:"tls"`
}

type ProcessDomains struct {
	Name string `yaml:"name"`
	Slug string `yaml:"slug"`
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

type ProcessVolume struct {
	Name          string `yaml:"name"`
	Type          string `yaml:"type"`
	MountPath     string `yaml:"mountPath"`
	SubPath       string `yaml:"subPath,omitempty"`
	ReadOnly      bool   `yaml:"readOnly,omitempty"`
	ClaimName     string `yaml:"claimName,omitempty"`
	ConfigMapName string `yaml:"configMapName,omitempty"`
	SecretName    string `yaml:"secretName,omitempty"`
	Chown         string `yaml:"chown,omitempty"`
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
	Email        string `yaml:"email"`
	Enabled      bool   `yaml:"enabled"`
	IngressClass string `yaml:"ingress_class"`
	Name         string `yaml:"name"`
	Server       string `yaml:"server"`
}

type PersistentVolumeClaim struct {
	Global struct {
		Annotations ProcessAnnotations `yaml:"annotations,omitempty"`
		Labels      ProcessLabels      `yaml:"labels,omitempty"`
	} `yaml:"global"`
	Name         string `yaml:"name"`
	AccessMode   string `yaml:"accessMode"`
	Storage      string `yaml:"storage"`
	StorageClass string `yaml:"storageClass"`
	Namespace    string `yaml:"namespace"`
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
			return batchv1.Job{}, fmt.Errorf("Error generating random suffix: %w", err)
		}
		suffix = strings.ToLower(fmt.Sprintf("%X", b))
	}
	annotations["dokku.com/job-suffix"] = suffix

	podAnnotations := annotations
	podAnnotations["kubectl.kubernetes.io/default-container"] = fmt.Sprintf("%s-%s", input.AppName, input.ProcessType)

	globalAnnotations, err := getGlobalAnnotations(input.AppName)
	if err != nil {
		return batchv1.Job{}, fmt.Errorf("Error getting global annotations: %w", err)
	}
	for key, value := range globalAnnotations.JobAnnotations {
		annotations[key] = value
	}
	for key, value := range globalAnnotations.PodAnnotations {
		podAnnotations[key] = value
	}

	processAnnotations, err := getAnnotations(input.AppName, input.ProcessType)
	if err != nil {
		return batchv1.Job{}, fmt.Errorf("Error getting process annotations: %w", err)
	}
	for key, value := range processAnnotations.JobAnnotations {
		annotations[key] = value
	}
	for key, value := range processAnnotations.PodAnnotations {
		podAnnotations[key] = value
	}

	// get volumes
	processVolumes, err := getVolumes(input.AppName, input.ProcessType)
	if err != nil {
		return batchv1.Job{}, fmt.Errorf("Error getting process volumes: %w", err)
	}
	var volumeMounts []corev1.VolumeMount
	var podVolumes []corev1.Volume
	var initContainerVolumeMounts []corev1.VolumeMount
	var chownCommands []string
	hasChown := false
	for _, volume := range processVolumes {
		// create volumeMounts
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name: volume.Name,
			// ReadOnly:  volume.ReadOnly,
			MountPath: volume.MountPath,
			// SubPath:   volume.SubPath,
		})
		// create podVolumes
		volumeSource := corev1.VolumeSource{}
		switch volume.Type {
		case "persistentVolumeClaim":
			volumeSource.PersistentVolumeClaim = &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: volume.ClaimName,
				ReadOnly:  volume.ReadOnly,
			}
		case "configMap":
			volumeSource.ConfigMap = &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: volume.ClaimName, // Assuming ClaimName is used as ConfigMap name
				},
			}
		case "secret":
			volumeSource.Secret = &corev1.SecretVolumeSource{
				SecretName: volume.ClaimName, // Assuming ClaimName is used as Secret name
			}
		default:
			return batchv1.Job{}, fmt.Errorf("unknown volume type: %s", volume.Type)
		}
		podVolumes = append(podVolumes, corev1.Volume{
			Name:         volume.Name,
			VolumeSource: volumeSource,
		})
		// Handle ownership fix if chown is specified
		if volume.Chown != "" {
			hasChown = true
			initContainerVolumeMounts = append(initContainerVolumeMounts, corev1.VolumeMount{
				Name:      volume.Name,
				MountPath: volume.MountPath,
			})
			chownCommands = append(chownCommands, fmt.Sprintf(`
if [ -d "%s" ]; then
  echo "Setting ownership for %s";
  chown -R %s %s;
else
  echo "Warning: Directory %s not found, skipping chown";
fi;
`, volume.MountPath, volume.MountPath, volume.Chown, volume.MountPath, volume.MountPath))
		}
	}
	// Create the initContainer only if there are chown operations
	var initContainers []corev1.Container
	if hasChown {
		initContainers = []corev1.Container{
			{
				Name:         "fix-permissions",
				Image:        "busybox",
				Command:      []string{"sh", "-c", strings.Join(chownCommands, "\n")},
				VolumeMounts: initContainerVolumeMounts,
			},
		}
	}

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
					InitContainers: initContainers,
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
							WorkingDir:   input.WorkingDir,
							VolumeMounts: volumeMounts,
						},
					},
					RestartPolicy:      corev1.RestartPolicyNever,
					ServiceAccountName: input.AppName,
					Volumes:            podVolumes,
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
