package scheduler_k3s

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	acmev1 "github.com/cert-manager/cert-manager/pkg/apis/acme/v1"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	certmanagermetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
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
	Secrets      map[string]string        `yaml:"secrets,omitempty"`
	Processes    map[string]ProcessValues `yaml:"processes"`
}

type ProcessValues struct {
	Domains  []string `yaml:"domains,omitempty"`
	Replicas int32    `yaml:"replicas"`
	TLS      bool     `yaml:"tls"`
}

type CreateIngressRoutesInput struct {
	AppName     string
	ChartDir    string
	Deployment  appsv1.Deployment
	Namespace   string
	ProcessType string
	PortMaps    []PortMap
	Service     v1.Service
}

func createIngressRoutesFiles(input CreateIngressRoutesInput) error {
	for _, portMap := range input.PortMaps {
		ingressRoute := templateKubernetesIngressRoute(IngressRoute{
			AppName:     input.AppName,
			Namespace:   input.Namespace,
			PortMap:     portMap,
			ProcessType: input.ProcessType,
			ServiceName: input.Service.Name,
		})

		ingressRouteFile := filepath.Join(input.ChartDir, fmt.Sprintf("templates/ingress-route-%s.yaml", portMap.String()))
		err := writeResourceToFile(WriteResourceInput{
			Object: &ingressRoute,
			Path:   ingressRouteFile,
		})
		if err != nil {
			return fmt.Errorf("Error printing ingress route: %w", err)
		}

		b, err := os.ReadFile(ingressRouteFile)
		if err != nil {
			return fmt.Errorf("Error reading ingress route file: %w", err)
		}

		append, err := templates.ReadFile("templates/ingress-routes-append.yaml")
		if err != nil {
			return fmt.Errorf("Error reading ingress route append file: %w", err)
		}

		contents := strings.Join([]string{strings.TrimSpace(string(b)), string(append)}, "")
		contents = strings.ReplaceAll(contents, "  routes: null", "")
		contents = strings.Join([]string{
			"{{- if .Values.processes.PROCESS_TYPE.domains }}",
			contents,
			"{{- end }}",
		}, "\n")

		replacements := orderedmap.New[string, string]()
		replacements.Set("PROCESS_TYPE", input.ProcessType)
		replacements.Set("APP_NAME", input.AppName)
		replacements.Set("NAMESPACE", input.Namespace)
		replacements.Set("PORT_MAPPING", portMap.String())
		replacements.Set("PORT_SCHEME", portMap.Scheme)
		for pair := replacements.Oldest(); pair != nil; pair = pair.Next() {
			contents = strings.ReplaceAll(contents, pair.Key, pair.Value)
		}

		err = os.WriteFile(ingressRouteFile, []byte(contents), os.FileMode(0644))
		if err != nil {
			return fmt.Errorf("Error writing ingress route file: %w", err)
		}

		if os.Getenv("DOKKU_TRACE") == "1" {
			common.CatFile(ingressRouteFile)
		}
	}

	return nil
}

type CreateCertificateFileInput struct {
	Certificate Certificate
	ChartDir    string
	IssuerName  string
	ProcessType string
}

type Certificate struct {
	AppName   string
	Name      string
	Namespace string
	TLS       bool
}

func createCertificateFile(input CreateCertificateFileInput) error {
	certificate, err := templateKubernetesCertificate(input.Certificate)
	if err != nil {
		return fmt.Errorf("Error templating certificate: %w", err)
	}

	certificateFile := filepath.Join(input.ChartDir, fmt.Sprintf("templates/certificate-%s.yaml", input.ProcessType))
	err = writeResourceToFile(WriteResourceInput{
		Object: &certificate,
		Path:   certificateFile,
	})
	if err != nil {
		return fmt.Errorf("Error printing ingress route: %w", err)
	}

	b, err := os.ReadFile(certificateFile)
	if err != nil {
		return fmt.Errorf("Error reading ingress route file: %w", err)
	}

	append, err := templates.ReadFile("templates/certificate-append.yaml")
	if err != nil {
		return fmt.Errorf("Error reading ingress route append file: %w", err)
	}

	contents := string(bytes.Join([][]byte{b, append}, []byte("")))
	contents = strings.Join([]string{
		"{{- if and .Values.processes.PROCESS_TYPE.tls .Values.processes.PROCESS_TYPE.domains }}",
		contents,
		"{{- end }}",
	}, "\n")

	replacements := orderedmap.New[string, string]()
	replacements.Set("PROCESS_TYPE", input.ProcessType)
	replacements.Set("ISSUER_NAME", input.IssuerName)
	for pair := replacements.Oldest(); pair != nil; pair = pair.Next() {
		contents = strings.ReplaceAll(contents, pair.Key, pair.Value)
	}

	err = os.WriteFile(certificateFile, []byte(contents), os.FileMode(0644))
	if err != nil {
		return fmt.Errorf("Error writing ingress route file: %w", err)
	}

	if os.Getenv("DOKKU_TRACE") == "1" {
		common.CatFile(certificateFile)
	}

	return nil
}

func templateKubernetesCertificate(input Certificate) (certmanagerv1.Certificate, error) {
	if input.Name == "" {
		return certmanagerv1.Certificate{}, fmt.Errorf("Name cannot be empty")
	}
	if input.Namespace == "" {
		return certmanagerv1.Certificate{}, fmt.Errorf("Namespace cannot be empty")
	}

	labels := map[string]string{
		"app.kubernetes.io/name":    input.Name,
		"app.kubernetes.io/part-of": input.AppName,
	}
	annotations := map[string]string{
		"dokku.com/managed": "true",
	}

	certificate := certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:        input.Name,
			Namespace:   input.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: certmanagerv1.CertificateSpec{
			SecretName: "tls-" + input.Name,
			SecretTemplate: &certmanagerv1.CertificateSecretTemplate{
				Annotations: annotations,
				Labels:      labels,
			},
			IssuerRef: certmanagermetav1.ObjectReference{
				Name: "ISSUER_NAME",
				Kind: "ClusterIssuer",
			},
		},
	}

	return certificate, nil
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
				Name:          portMap.String(),
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
	Namespace   string
	PortMap     PortMap
	ProcessType string
	ServiceName string
}

func templateKubernetesIngressRoute(input IngressRoute) traefikv1alpha1.IngressRoute {
	labels := map[string]string{
		"app.kubernetes.io/instance": fmt.Sprintf("%s-%s", input.AppName, input.ProcessType),
		"app.kubernetes.io/name":     input.ProcessType,
		"app.kubernetes.io/part-of":  input.AppName,
	}
	annotations := map[string]string{
		"dokku.com/managed": "true",
	}

	port := input.PortMap.String()
	ingressRoute := traefikv1alpha1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-%s", input.ServiceName, port),
			Namespace:   input.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: traefikv1alpha1.IngressRouteSpec{
			EntryPoints: []string{string(IngressRouteEntrypoint_HTTP)},
		},
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
			Name:       portMap.String(),
			Port:       portMap.HostPort,
			TargetPort: intstr.FromString(portMap.String()),
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
