package scheduler_k3s

import (
	"sync"

	traefikv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	kjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
)

var (
	// DefaultProperties is a map of all valid k3s properties with corresponding default property values
	DefaultProperties = map[string]string{
		"image-pull-secrets":  "",
		"rollback-on-failure": "",
		"namespace":           "default",
		"timeout":             "300",
	}

	// GlobalProperties is a map of all valid global k3s properties
	GlobalProperties = map[string]bool{
		"network-interface":   true,
		"rollback-on-failure": true,
		"token":               true,
		"timeout":             true,
	}
)

const KubeConfigPath = "/etc/rancher/k3s/k3s.yaml"

var (
	runtimeScheme  = runtime.NewScheme()
	codecs         = serializer.NewCodecFactory(runtimeScheme)
	deserializer   = codecs.UniversalDeserializer()
	jsonSerializer = kjson.NewSerializerWithOptions(kjson.DefaultMetaFactory, runtimeScheme, runtimeScheme, kjson.SerializerOptions{})
)

var k8sNativeSchemeOnce sync.Once

func init() {
	k8sNativeSchemeOnce.Do(func() {
		_ = appsv1.AddToScheme(runtimeScheme)
		_ = corev1.AddToScheme(runtimeScheme)
		_ = traefikv1alpha1.AddToScheme(runtimeScheme)
	})
}
