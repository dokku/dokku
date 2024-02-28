package scheduler_k3s

import (
	"embed"
	"sync"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	traefikv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	kjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
)

var (
	// DefaultProperties is a map of all valid k3s properties with corresponding default property values
	DefaultProperties = map[string]string{
		"deploy-timeout":      "",
		"letsencrypt-server":  "",
		"image-pull-secrets":  "",
		"namespace":           "",
		"rollback-on-failure": "",
	}

	// GlobalProperties is a map of all valid global k3s properties
	GlobalProperties = map[string]bool{
		"deploy-timeout":         true,
		"image-pull-secrets":     true,
		"ingress-class":          true,
		"kube-context":           true,
		"kubeconfig-path":        true,
		"letsencrypt-server":     true,
		"letsencrypt-email-prod": true,
		"letsencrypt-email-stag": true,
		"namespace":              true,
		"network-interface":      true,
		"rollback-on-failure":    true,
		"token":                  true,
	}
)

const DefaultIngressClass = "traefik"
const GlobalProcessType = "--global"
const KubeConfigPath = "/etc/rancher/k3s/k3s.yaml"
const DefaultKubeContext = ""
const TriggerAuthPropertyPrefix = "trigger-auth."

var (
	runtimeScheme  = runtime.NewScheme()
	codecs         = serializer.NewCodecFactory(runtimeScheme)
	deserializer   = codecs.UniversalDeserializer()
	jsonSerializer = kjson.NewSerializerWithOptions(kjson.DefaultMetaFactory, runtimeScheme, runtimeScheme, kjson.SerializerOptions{})
)

var k8sNativeSchemeOnce sync.Once

type Manifest struct {
	Name    string
	Version string
	Path    string
}

var KubernetesManifests = []Manifest{
	{
		Name:    "system-upgrader",
		Version: "0.13.2",
		Path:    "https://github.com/rancher/system-upgrade-controller/releases/download/v0.13.2/system-upgrade-controller.yaml",
	},
}

type HelmChart struct {
	ChartPath       string
	CreateNamespace bool
	Namespace       string
	Path            string
	ReleaseName     string
	RepoURL         string
	Version         string
}

var HelmCharts = []HelmChart{
	{
		ChartPath:       "cert-manager",
		CreateNamespace: true,
		Namespace:       "cert-manager",
		ReleaseName:     "cert-manager",
		RepoURL:         "https://charts.jetstack.io",
		Version:         "v1.13.3",
	},
	{
		ChartPath:       "longhorn",
		CreateNamespace: true,
		Namespace:       "longhorn-system",
		ReleaseName:     "longhorn",
		RepoURL:         "https://charts.longhorn.io",
		Version:         "1.5.3",
	},
	{
		ChartPath:       "traefik",
		CreateNamespace: true,
		Namespace:       "traefik",
		ReleaseName:     "traefik",
		RepoURL:         "https://helm.traefik.io/traefik",
		Version:         "26.0.0",
	},
	{
		ChartPath:       "ingress-nginx",
		CreateNamespace: true,
		Namespace:       "ingress-nginx",
		ReleaseName:     "ingress-nginx",
		RepoURL:         "https://kubernetes.github.io/ingress-nginx",
		Version:         "4.7.5",
	},
	{
		ChartPath:       "keda",
		CreateNamespace: true,
		Namespace:       "keda",
		ReleaseName:     "keda",
		RepoURL:         "https://kedacore.github.io/charts",
		Version:         "2.13.1",
	},
}

type HelmRepository struct {
	Name string
	URL  string
}

var HelmRepositories = []HelmRepository{
	{
		Name: "jetstack",
		URL:  "https://charts.jetstack.io",
	},
	{
		Name: "longhorn",
		URL:  "https://charts.longhorn.io",
	},
	{
		Name: "traefik",
		URL:  "https://helm.traefik.io/traefik",
	},
}

var ServerLabels = map[string]string{
	"svccontroller.k3s.cattle.io/enablelb": "true",
}

var WorkerLabels = map[string]string{
	"node-role.kubernetes.io/worker": "worker",
}

//go:embed all:templates
var templates embed.FS

func init() {
	k8sNativeSchemeOnce.Do(func() {
		_ = appsv1.AddToScheme(runtimeScheme)
		_ = batchv1.AddToScheme(runtimeScheme)
		_ = certmanagerv1.AddToScheme(runtimeScheme)
		_ = corev1.AddToScheme(runtimeScheme)
		_ = traefikv1alpha1.AddToScheme(runtimeScheme)
	})
}
