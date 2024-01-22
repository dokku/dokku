package scheduler_k3s

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	appjson "github.com/dokku/dokku/plugins/app-json"
	"github.com/dokku/dokku/plugins/common"
	traefikv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/utils/ptr"
)

type Chart struct {
	ApiVersion string `yaml:"apiVersion"`
	AppVersion string `yaml:"appVersion"`
	Name       string `yaml:"name"`
	Version    string `yaml:"version"`
}

type Values struct {
	DeploymentID string                   `yaml:"deploment_id"`
	Secrets      map[string]string        `yaml:"secrets"`
	Processes    map[string]ValuesProcess `yaml:"processes"`
}

type ValuesProcess struct {
	Replicas int32 `yaml:"replicas"`
}

type CreateIngressRoutesInput struct {
	AppName    string
	ChartDir   string
	Deployment appsv1.Deployment
	Namespace  string
	PortMaps   []PortMap
	Service    v1.Service
}

func createIngressRoutesFiles(input CreateIngressRoutesInput) error {
	err := common.PlugnTrigger("domains-vhost-enabled", []string{input.AppName}...)
	isAppVhostEnabled := err == nil

	if isAppVhostEnabled {
		b, err := common.PlugnTriggerOutput("domains-list", []string{input.AppName}...)
		if err != nil {
			return fmt.Errorf("Error getting domains for deployment: %w", err)
		}
		domains := []string{}
		for _, domain := range strings.Split(string(b), "\n") {
			domain = strings.TrimSpace(domain)
			if domain != "" {
				domains = append(domains, domain)
			}
		}

		if len(domains) == 0 {
			return nil
		}

		for _, portMap := range input.PortMaps {
			ingressRoute := templateKubernetesIngressRoute(IngressRoute{
				AppName:     input.AppName,
				Domains:     domains,
				Namespace:   input.Namespace,
				PortMap:     portMap,
				ProcessType: "web",
				ServiceName: input.Service.Name,
			})

			err = writeResourceToFile(WriteResourceInput{
				Object: &ingressRoute,
				Path:   filepath.Join(input.ChartDir, fmt.Sprintf("templates/ingress-route-%s-%d-%d.yaml", portMap.Scheme, portMap.HostPort, portMap.ContainerPort)),
			})
			if err != nil {
				return fmt.Errorf("Error printing ingress route: %w", err)
			}
		}
	}
	return nil
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

func templateKubernetesCronJob(input Job) (batchv1.CronJob, error) {
	if input.Schedule == "" {
		return batchv1.CronJob{}, fmt.Errorf("Schedule cannot be empty")
	}
	labels := map[string]string{
		"app.kubernetes.io/instance": fmt.Sprintf("%s-%s", input.AppName, input.ProcessType),
		"app.kubernetes.io/name":     input.ProcessType,
		"app.kubernetes.io/part-of":  input.AppName,
		"dokku.com/cron-id":          input.ID,
	}
	annotations := map[string]string{
		"app.kubernetes.io/version": "DEPLOYMENT_ID_QUOTED",
		"dokku.com/builder-type":    input.ImageSourceType,
		"dokku.com/cron-id":         input.ID,
		"dokku.com/managed":         "true",
	}

	for key, value := range input.Labels {
		labels[key] = value
	}
	secretName := fmt.Sprintf("env-%s.DEPLOYMENT_ID", input.AppName)

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

	job := batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-%s-%s", input.AppName, input.ProcessType, suffix),
			Namespace:   input.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: batchv1.CronJobSpec{
			ConcurrencyPolicy:          batchv1.AllowConcurrent,
			FailedJobsHistoryLimit:     ptr.To(int32(10)),
			Schedule:                   input.Schedule,
			StartingDeadlineSeconds:    ptr.To(int64(60)),
			SuccessfulJobsHistoryLimit: ptr.To(int32(10)),
			Suspend:                    ptr.To(false),
			TimeZone:                   ptr.To("Etc/UTC"),
			JobTemplate: batchv1.JobTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: annotations,
				},
				Spec: batchv1.JobSpec{
					BackoffLimit:         ptr.To(int32(0)),
					PodReplacementPolicy: ptr.To(batchv1.Failed),
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
			},
		},
	}

	if input.Entrypoint != "" {
		job.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Command = []string{input.Entrypoint}
	}

	if input.RemoveContainer {
		job.Spec.JobTemplate.Spec.TTLSecondsAfterFinished = ptr.To(int32(60))
	}

	if input.ImagePullSecrets != "" {
		job.Spec.JobTemplate.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
			{
				Name: input.ImagePullSecrets,
			},
		}
	}

	cpuLimit, err := common.PlugnTriggerOutputAsString("resource-get-property", []string{input.AppName, input.ProcessType, "limit", "cpu"}...)
	if err != nil && cpuLimit != "" && cpuLimit != "0" {
		cpuQuantity, err := resource.ParseQuantity(cpuLimit)
		if err != nil {
			return job, fmt.Errorf("Error parsing cpu limit: %w", err)
		}
		job.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Resources.Limits["cpu"] = cpuQuantity
	} else {
		job.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Resources.Limits["cpu"] = resource.MustParse("500m")
	}
	nvidiaGpuLimit, err := common.PlugnTriggerOutputAsString("resource-get-property", []string{input.AppName, input.ProcessType, "limit", "nvidia-gpu"}...)
	if err != nil && nvidiaGpuLimit != "" && nvidiaGpuLimit != "0" {
		nvidiaGpuQuantity, err := resource.ParseQuantity(nvidiaGpuLimit)
		if err != nil {
			return job, fmt.Errorf("Error parsing nvidia-gpu limit: %w", err)
		}
		job.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Resources.Limits["nvidia.com/gpu"] = nvidiaGpuQuantity
	}
	memoryLimit, err := common.PlugnTriggerOutputAsString("resource-get-property", []string{input.AppName, input.ProcessType, "limit", "memory"}...)
	if err != nil && memoryLimit != "" && memoryLimit != "0" {
		memoryQuantity, err := resource.ParseQuantity(memoryLimit)
		if err != nil {
			return job, fmt.Errorf("Error parsing memory limit: %w", err)
		}
		job.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Resources.Limits["memory"] = memoryQuantity
	} else {
		job.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Resources.Limits["memory"] = resource.MustParse("512Mi")
	}

	cpuRequest, err := common.PlugnTriggerOutputAsString("resource-get-property", []string{input.AppName, input.ProcessType, "reserve", "cpu"}...)
	if err != nil && cpuRequest != "" && cpuRequest != "0" {
		cpuQuantity, err := resource.ParseQuantity(cpuRequest)
		if err != nil {
			return job, fmt.Errorf("Error parsing cpu request: %w", err)
		}
		job.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Resources.Requests["cpu"] = cpuQuantity
	} else {
		job.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Resources.Requests["cpu"] = resource.MustParse("500m")
	}
	memoryRequest, err := common.PlugnTriggerOutputAsString("resource-get-property", []string{input.AppName, input.ProcessType, "reserve", "memory"}...)
	if err != nil && memoryRequest != "" && memoryRequest != "0" {
		memoryQuantity, err := resource.ParseQuantity(memoryRequest)
		if err != nil {
			return job, fmt.Errorf("Error parsing memory request: %w", err)
		}
		job.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Resources.Requests["memory"] = memoryQuantity
	} else {
		job.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Resources.Requests["memory"] = resource.MustParse("512Mi")
	}

	return job, nil
}

type Deployment struct {
	AppName          string
	Command          []string
	Image            string
	ImagePullSecrets string
	ImageSourceType  string
	Healthchecks     []appjson.Healthcheck
	Namespace        string
	PrimaryPort      int32
	PortMaps         []PortMap
	ProcessType      string
	Replicas         int32
	WorkingDir       string
}

func templateKubernetesDeployment(input Deployment) (appsv1.Deployment, error) {
	labels := map[string]string{
		"app.kubernetes.io/instance": fmt.Sprintf("%s-%s", input.AppName, input.ProcessType),
		"app.kubernetes.io/name":     input.ProcessType,
		"app.kubernetes.io/part-of":  input.AppName,
	}
	annotations := map[string]string{
		"app.kubernetes.io/version": "DEPLOYMENT_ID_QUOTED",
		"dokku.com/builder-type":    input.ImageSourceType,
		"dokku.com/managed":         "true",
	}
	secretName := fmt.Sprintf("env-%s.DEPLOYMENT_ID", input.AppName)

	podAnnotations := annotations
	podAnnotations["kubectl.kubernetes.io/default-container"] = fmt.Sprintf("%s-%s", input.AppName, input.ProcessType)

	deployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-%s", input.AppName, input.ProcessType),
			Namespace:   input.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas:             ptr.To(input.Replicas),
			RevisionHistoryLimit: ptr.To(int32(5)),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: podAnnotations,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: fmt.Sprintf("%s-%s", input.AppName, input.ProcessType),
							Env:  []corev1.EnvVar{},
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
				},
			},
		},
	}

	if len(input.Command) > 0 {
		deployment.Spec.Template.Spec.Containers[0].Args = input.Command
	}

	if len(input.Healthchecks) > 0 {
		livenessChecks := []corev1.Probe{}
		readinessChecks := []corev1.Probe{}
		startupChecks := []corev1.Probe{}
		uptimeSeconds := []int32{}
		for _, healthcheck := range input.Healthchecks {
			probe := corev1.Probe{
				ProbeHandler:        corev1.ProbeHandler{},
				InitialDelaySeconds: healthcheck.InitialDelay,
				PeriodSeconds:       healthcheck.Wait,
				TimeoutSeconds:      healthcheck.Timeout,
				FailureThreshold:    healthcheck.Attempts,
				SuccessThreshold:    int32(1),
			}
			if len(healthcheck.Command) > 0 {
				probe.ProbeHandler.Exec = &corev1.ExecAction{
					Command: healthcheck.Command,
				}
			} else if healthcheck.Listening {
				probe.ProbeHandler.TCPSocket = &corev1.TCPSocketAction{
					Port: intstr.FromInt32(input.PrimaryPort),
				}
				for _, header := range healthcheck.HTTPHeaders {
					if header.Name == "Host" {
						probe.ProbeHandler.TCPSocket.Host = header.Value
					}
				}
			} else if healthcheck.Path != "" {
				probe.ProbeHandler.HTTPGet = &corev1.HTTPGetAction{
					Path:        healthcheck.Path,
					Port:        intstr.FromInt32(input.PrimaryPort),
					HTTPHeaders: []corev1.HTTPHeader{},
				}

				if healthcheck.Scheme != "" {
					probe.ProbeHandler.HTTPGet.Scheme = corev1.URIScheme(strings.ToUpper(healthcheck.Scheme))
				}

				for _, header := range healthcheck.HTTPHeaders {
					probe.ProbeHandler.HTTPGet.HTTPHeaders = append(probe.ProbeHandler.HTTPGet.HTTPHeaders, corev1.HTTPHeader{
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

		if len(livenessChecks) > 0 {
			deployment.Spec.Template.Spec.Containers[0].LivenessProbe = &livenessChecks[0]
		}
		if len(readinessChecks) > 0 {
			deployment.Spec.Template.Spec.Containers[0].ReadinessProbe = &readinessChecks[0]
		}
		if len(startupChecks) > 0 {
			deployment.Spec.Template.Spec.Containers[0].StartupProbe = &startupChecks[0]
		}
		if len(uptimeSeconds) > 0 {
			deployment.Spec.MinReadySeconds = uptimeSeconds[0]
		}
	}

	if input.ProcessType == "web" {
		for _, portMap := range input.PortMaps {
			protocol := "TCP"
			if portMap.Scheme == "udp" {
				protocol = "UDP"
			}
			deployment.Spec.Template.Spec.Containers[0].Ports = append(deployment.Spec.Template.Spec.Containers[0].Ports, corev1.ContainerPort{
				Name:          fmt.Sprintf("%s-%d-%d", portMap.Scheme, portMap.HostPort, portMap.ContainerPort),
				ContainerPort: portMap.ContainerPort,
				Protocol:      corev1.Protocol(protocol),
			})
		}

		deployment.Spec.Template.Spec.Containers[0].Env = append(deployment.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
			Name:  "PORT",
			Value: fmt.Sprint(input.PrimaryPort),
		})
	}

	if input.ImagePullSecrets != "" {
		deployment.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
			{
				Name: input.ImagePullSecrets,
			},
		}
	}

	cpuLimit, err := common.PlugnTriggerOutputAsString("resource-get-property", []string{input.AppName, input.ProcessType, "limit", "cpu"}...)
	if err != nil && cpuLimit != "" && cpuLimit != "0" {
		cpuQuantity, err := resource.ParseQuantity(cpuLimit)
		if err != nil {
			return deployment, fmt.Errorf("Error parsing cpu limit: %w", err)
		}
		deployment.Spec.Template.Spec.Containers[0].Resources.Limits["cpu"] = cpuQuantity
	}
	nvidiaGpuLimit, err := common.PlugnTriggerOutputAsString("resource-get-property", []string{input.AppName, input.ProcessType, "limit", "nvidia-gpu"}...)
	if err != nil && nvidiaGpuLimit != "" && nvidiaGpuLimit != "0" {
		nvidiaGpuQuantity, err := resource.ParseQuantity(nvidiaGpuLimit)
		if err != nil {
			return deployment, fmt.Errorf("Error parsing nvidia-gpu limit: %w", err)
		}
		deployment.Spec.Template.Spec.Containers[0].Resources.Limits["nvidia.com/gpu"] = nvidiaGpuQuantity
	}
	memoryLimit, err := common.PlugnTriggerOutputAsString("resource-get-property", []string{input.AppName, input.ProcessType, "limit", "memory"}...)
	if err != nil && memoryLimit != "" && memoryLimit != "0" {
		memoryQuantity, err := resource.ParseQuantity(memoryLimit)
		if err != nil {
			return deployment, fmt.Errorf("Error parsing memory limit: %w", err)
		}
		deployment.Spec.Template.Spec.Containers[0].Resources.Limits["memory"] = memoryQuantity
	}

	cpuRequest, err := common.PlugnTriggerOutputAsString("resource-get-property", []string{input.AppName, input.ProcessType, "reserve", "cpu"}...)
	if err != nil && cpuRequest != "" && cpuRequest != "0" {
		cpuQuantity, err := resource.ParseQuantity(cpuRequest)
		if err != nil {
			return deployment, fmt.Errorf("Error parsing cpu request: %w", err)
		}
		deployment.Spec.Template.Spec.Containers[0].Resources.Requests["cpu"] = cpuQuantity
	}
	memoryRequest, err := common.PlugnTriggerOutputAsString("resource-get-property", []string{input.AppName, input.ProcessType, "reserve", "memory"}...)
	if err != nil && memoryRequest != "" && memoryRequest != "0" {
		memoryQuantity, err := resource.ParseQuantity(memoryRequest)
		if err != nil {
			return deployment, fmt.Errorf("Error parsing memory request: %w", err)
		}
		deployment.Spec.Template.Spec.Containers[0].Resources.Requests["memory"] = memoryQuantity
	}

	return deployment, nil
}

type IngressRouteEntrypoint string

const (
	IngressRouteEntrypoint_HTTP  IngressRouteEntrypoint = "web"
	IngressRouteEntrypoint_HTTPS IngressRouteEntrypoint = "websecure"
)

type IngressRoute struct {
	AppName     string
	Entrypoints []IngressRouteEntrypoint
	Domains     []string
	Namespace   string
	PortMap     PortMap
	ProcessType string
	ServiceName string
}

func templateKubernetesIngressRoute(input IngressRoute) traefikv1alpha1.IngressRoute {
	entryPoint := IngressRouteEntrypoint_HTTP
	if input.PortMap.Scheme == "https" {
		entryPoint = IngressRouteEntrypoint_HTTPS
	}

	labels := map[string]string{
		"app.kubernetes.io/instance": fmt.Sprintf("%s-%s", input.AppName, input.ProcessType),
		"app.kubernetes.io/name":     input.ProcessType,
		"app.kubernetes.io/part-of":  input.AppName,
	}
	annotations := map[string]string{
		"dokku.com/managed": "true",
	}

	port := fmt.Sprintf("%s-%d-%d", input.PortMap.Scheme, input.PortMap.HostPort, input.PortMap.ContainerPort)
	ingressRoute := traefikv1alpha1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-%s", input.ServiceName, port),
			Namespace:   input.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: traefikv1alpha1.IngressRouteSpec{
			EntryPoints: []string{string(entryPoint)},
			Routes:      []traefikv1alpha1.Route{},
		},
	}

	sort.Strings(input.Domains)

	for _, hostname := range input.Domains {
		rule := traefikv1alpha1.Route{
			Kind:  "Rule",
			Match: "Host(`" + hostname + "`)",
			Services: []traefikv1alpha1.Service{
				{
					LoadBalancerSpec: traefikv1alpha1.LoadBalancerSpec{
						Name:           input.ServiceName,
						Namespace:      input.Namespace,
						PassHostHeader: ptr.To(true),
						Port:           intstr.FromString(port),
						Scheme:         input.PortMap.Scheme,
					},
				},
			},
		}

		ingressRoute.Spec.Routes = append(ingressRoute.Spec.Routes, rule)
	}

	return ingressRoute
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

	cpuLimit, err := common.PlugnTriggerOutputAsString("resource-get-property", []string{input.AppName, input.ProcessType, "limit", "cpu"}...)
	if err != nil && cpuLimit != "" && cpuLimit != "0" {
		cpuQuantity, err := resource.ParseQuantity(cpuLimit)
		if err != nil {
			return job, fmt.Errorf("Error parsing cpu limit: %w", err)
		}
		job.Spec.Template.Spec.Containers[0].Resources.Limits["cpu"] = cpuQuantity
	}
	nvidiaGpuLimit, err := common.PlugnTriggerOutputAsString("resource-get-property", []string{input.AppName, input.ProcessType, "limit", "nvidia-gpu"}...)
	if err != nil && nvidiaGpuLimit != "" && nvidiaGpuLimit != "0" {
		nvidiaGpuQuantity, err := resource.ParseQuantity(nvidiaGpuLimit)
		if err != nil {
			return job, fmt.Errorf("Error parsing nvidia-gpu limit: %w", err)
		}
		job.Spec.Template.Spec.Containers[0].Resources.Limits["nvidia.com/gpu"] = nvidiaGpuQuantity
	}
	memoryLimit, err := common.PlugnTriggerOutputAsString("resource-get-property", []string{input.AppName, input.ProcessType, "limit", "memory"}...)
	if err != nil && memoryLimit != "" && memoryLimit != "0" {
		memoryQuantity, err := resource.ParseQuantity(memoryLimit)
		if err != nil {
			return job, fmt.Errorf("Error parsing memory limit: %w", err)
		}
		job.Spec.Template.Spec.Containers[0].Resources.Limits["memory"] = memoryQuantity
	}

	cpuRequest, err := common.PlugnTriggerOutputAsString("resource-get-property", []string{input.AppName, input.ProcessType, "reserve", "cpu"}...)
	if err != nil && cpuRequest != "" && cpuRequest != "0" {
		cpuQuantity, err := resource.ParseQuantity(cpuRequest)
		if err != nil {
			return job, fmt.Errorf("Error parsing cpu request: %w", err)
		}
		job.Spec.Template.Spec.Containers[0].Resources.Requests["cpu"] = cpuQuantity
	}
	memoryRequest, err := common.PlugnTriggerOutputAsString("resource-get-property", []string{input.AppName, input.ProcessType, "reserve", "memory"}...)
	if err != nil && memoryRequest != "" && memoryRequest != "0" {
		memoryQuantity, err := resource.ParseQuantity(memoryRequest)
		if err != nil {
			return job, fmt.Errorf("Error parsing memory request: %w", err)
		}
		job.Spec.Template.Spec.Containers[0].Resources.Requests["memory"] = memoryQuantity
	}

	return job, nil
}

type Secret struct {
	AppName   string
	Env       map[string]string
	Namespace string
}

func templateKubernetesSecret(input Secret) corev1.Secret {
	secretName := fmt.Sprintf("env-%s.DEPLOYMENT_ID", input.AppName)
	labels := map[string]string{
		"app.kubernetes.io/instance": secretName,
		"app.kubernetes.io/name":     fmt.Sprintf("%s-env", input.AppName),
		"app.kubernetes.io/part-of":  input.AppName,
	}

	annotations := map[string]string{
		"app.kubernetes.io/version": "DEPLOYMENT_ID_QUOTED",
		"dokku.com/managed":         "true",
	}
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        secretName,
			Namespace:   input.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Data: map[string][]byte{},
	}

	return secret
}

type Service struct {
	AppName   string
	Namespace string
	PortMaps  []PortMap
}

func templateKubernetesService(input Service) corev1.Service {
	labels := map[string]string{
		"app.kubernetes.io/instance": fmt.Sprintf("%s-%s", input.AppName, "web"),
		"app.kubernetes.io/name":     "web",
		"app.kubernetes.io/part-of":  input.AppName,
	}
	annotations := map[string]string{
		"dokku.com/managed": "true",
	}
	service := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-%s", input.AppName, "web"),
			Namespace:   input.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
		},
	}

	for _, portMap := range input.PortMaps {
		protocol := "TCP"
		if portMap.Scheme == "udp" {
			protocol = "UDP"
		}
		service.Spec.Ports = append(service.Spec.Ports, corev1.ServicePort{
			Name:       fmt.Sprintf("%s-%d-%d", portMap.Scheme, portMap.HostPort, portMap.ContainerPort),
			Port:       portMap.HostPort,
			TargetPort: intstr.FromString(fmt.Sprintf("%s-%d-%d", portMap.Scheme, portMap.HostPort, portMap.ContainerPort)),
			Protocol:   corev1.Protocol(protocol),
		})
	}

	return service
}

type WriteResourceInput struct {
	AppendContents string
	Object         runtime.Object
	Path           string
	Replacements   *orderedmap.OrderedMap[string, string]
}

func writeResourceToFile(input WriteResourceInput) error {
	common.LogDebug(fmt.Sprintf("Printing resource: %s", input.Path))
	printr := printers.NewTypeSetter(runtimeScheme).ToPrinter(&printers.YAMLPrinter{})
	handle, err := os.Create(input.Path)
	if err != nil {
		return fmt.Errorf("Error creating template file: %w", err)
	}
	defer handle.Close()

	if err := printr.PrintObj(input.Object, handle); err != nil {
		return fmt.Errorf("Error writing template file: %w", err)
	}

	if input.Replacements != nil {
		b, err := os.ReadFile(input.Path)
		if err != nil {
			return fmt.Errorf("Error reading template file: %w", err)
		}

		contents := string(b)
		for pair := input.Replacements.Oldest(); pair != nil; pair = pair.Next() {
			contents = strings.ReplaceAll(string(contents), pair.Key, pair.Value)
		}

		err = os.WriteFile(input.Path, []byte(contents), os.FileMode(0644))
		if err != nil {
			return fmt.Errorf("Error updating template file with replacements: %w", err)
		}
	}

	if input.AppendContents != "" {
		b, err := os.ReadFile(input.Path)
		if err != nil {
			return fmt.Errorf("Error reading template file: %w", err)
		}

		contents := string(b) + "\n" + input.AppendContents

		err = os.WriteFile(input.Path, []byte(contents), os.FileMode(0644))
		if err != nil {
			return fmt.Errorf("Error updating template file with replacements: %w", err)
		}
	}

	if os.Getenv("DOKKU_TRACE") == "1" {
		common.CatFile(input.Path)
	}
	return nil
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
