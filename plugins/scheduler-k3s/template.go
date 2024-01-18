package scheduler_k3s

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	traefikv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
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

type CreateIngressRoutesInput struct {
	AppName    string
	ChartDir   string
	Deployment appsv1.Deployment
	Namespace  string
	PortMaps   []PortMap
	Service    v1.Service
}

type Deployment struct {
	AppName          string
	Command          []string
	Image            string
	ImagePullSecrets string
	ImageSourceType  string
	Namespace        string
	PrimaryPort      int32
	PortMaps         []PortMap
	ProcessType      string
	Replicas         int32
}

type IngressRouteEntrypoint string

const (
	IngressRouteEntrypoint_HTTP  IngressRouteEntrypoint = "web"
	IngressRouteEntrypoint_HTTPS IngressRouteEntrypoint = "websecure"
)

type IngressRoute struct {
	AppName     string
	Entrypoints []IngressRouteEntrypoint
	Hostnames   []string
	Namespace   string
	Port        int32
	PortMap     PortMap
	ProcessType string
	ServiceName string
}

type Secret struct {
	AppName   string
	Env       map[string]string
	Namespace string
}

type Service struct {
	AppName   string
	Namespace string
	Port      int32
}

type PrintInput struct {
	AppendContents string
	Object         runtime.Object
	Path           string
	Name           string
	Replacements   *orderedmap.OrderedMap[string, string]
}

type Values struct {
	DeploymentID string                   `yaml:"deploment_id"`
	Secrets      map[string]string        `yaml:"secrets"`
	Processes    map[string]ValuesProcess `yaml:"processes"`
}

type ValuesProcess struct {
	Replicas int32 `yaml:"replicas"`
}

type WriteYamlInput struct {
	Object interface{}
	Path   string
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
				Hostnames:   domains,
				Namespace:   input.Namespace,
				Port:        input.Deployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort,
				PortMap:     portMap,
				ProcessType: "web",
				ServiceName: input.Service.Name,
			})

			err = printResource(PrintInput{
				Object: &ingressRoute,
				Path:   filepath.Join(input.ChartDir, fmt.Sprintf("templates/ingress-route-%s-%d-%d.yaml", portMap.Scheme, portMap.HostPort, portMap.ContainerPort)),
				Name:   ingressRoute.Name,
			})
			if err != nil {
				return fmt.Errorf("Error printing ingress route: %w", err)
			}
		}
	}
	return nil
}

func printResource(input PrintInput) error {
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

func templateKubernetesDeployment(input Deployment) (appsv1.Deployment, error) {
	labels := map[string]string{
		"dokku.com/app-name":         input.AppName,
		"dokku.com/app-process-type": fmt.Sprintf("%s-%s", input.AppName, input.ProcessType),
		"dokku.com/process-type":     input.ProcessType,
	}
	annotations := map[string]string{
		"dokku.com/app-name":      input.AppName,
		"dokku.com/builder-type":  input.ImageSourceType,
		"dokku.com/deployment-id": "DEPLOYMENT_ID_QUOTED",
		"dokku.com/managed":       "true",
		"dokku.com/process-type":  input.ProcessType,
	}
	secretName := fmt.Sprintf("env-%s.DEPLOYMENT_ID", input.AppName)

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
					Annotations: annotations,
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
							ImagePullPolicy: corev1.PullPolicy("Always"),
							Resources: corev1.ResourceRequirements{
								Limits:   corev1.ResourceList{},
								Requests: corev1.ResourceList{},
							},
						},
					},
				},
			},
		},
	}

	if len(input.Command) > 0 {
		deployment.Spec.Template.Spec.Containers[0].Args = input.Command
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

func templateKubernetesIngressRoute(input IngressRoute) traefikv1alpha1.IngressRoute {
	entryPoint := IngressRouteEntrypoint_HTTP
	if input.PortMap.Scheme == "https" {
		entryPoint = IngressRouteEntrypoint_HTTPS
	}

	port := fmt.Sprintf("%s-%d-%d", input.PortMap.Scheme, input.PortMap.HostPort, input.PortMap.ContainerPort)
	ingressRoute := traefikv1alpha1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", input.ServiceName, port),
			Namespace: input.Namespace,
			Labels: map[string]string{
				"dokku.com/app-name":         input.AppName,
				"dokku.com/app-process-type": fmt.Sprintf("%s-%s", input.AppName, input.ProcessType),
				"dokku.com/process-type":     input.ProcessType,
			},
			Annotations: map[string]string{
				"dokku.com/managed": "true",
			},
		},
		Spec: traefikv1alpha1.IngressRouteSpec{
			EntryPoints: []string{string(entryPoint)},
			Routes:      []traefikv1alpha1.Route{},
		},
	}

	for _, hostname := range input.Hostnames {
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

func templateKubernetesSecret(input Secret) corev1.Secret {
	secretName := fmt.Sprintf("env-%s.DEPLOYMENT_ID", input.AppName)
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: input.Namespace,
			Labels: map[string]string{
				"dokku.com/app-name":      input.AppName,
				"dokku.com/deployment-id": "DEPLOYMENT_ID_QUOTED",
			},
			Annotations: map[string]string{
				"dokku.com/managed": "true",
			},
		},
		Data: map[string][]byte{},
	}

	return secret
}

func templateKubernetesService(input Service) corev1.Service {
	service := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", input.AppName, "web"),
			Namespace: input.Namespace,
			Labels: map[string]string{
				"dokku.com/app-name":         input.AppName,
				"dokku.com/app-process-type": fmt.Sprintf("%s-%s", input.AppName, "web"),
				"dokku.com/process-type":     "web",
			},
			Annotations: map[string]string{
				"dokku.com/managed": "true",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"dokku.com/app-name":         input.AppName,
				"dokku.com/app-process-type": fmt.Sprintf("%s-%s", input.AppName, "web"),
				"dokku.com/process-type":     "web",
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "web",
					Port:       input.Port,
					TargetPort: intstr.FromInt32(input.Port),
				},
			},
		},
	}
	return service
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
