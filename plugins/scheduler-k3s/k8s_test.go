package scheduler_k3s

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func newNode(name string, labels map[string]string, kubeletVersion string) *corev1.Node {
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Status: corev1.NodeStatus{
			NodeInfo: corev1.NodeSystemInfo{
				KubeletVersion: kubeletVersion,
			},
		},
	}
}

func TestGetLowestNodeVersion(t *testing.T) {
	controlPlaneLabels := map[string]string{
		"node-role.kubernetes.io/control-plane": "true",
	}
	dualLabels := map[string]string{
		"node-role.kubernetes.io/control-plane": "true",
		"node-role.kubernetes.io/master":        "true",
	}
	workerLabels := map[string]string{
		"node-role.kubernetes.io/worker": "true",
	}
	legacyMasterLabels := map[string]string{
		"node-role.kubernetes.io/master": "true",
	}

	cases := []struct {
		name        string
		nodes       []*corev1.Node
		selector    string
		want        string
		wantErr     string
	}{
		{
			name: "single control-plane node returns its kubelet version",
			nodes: []*corev1.Node{
				newNode("cp-0", controlPlaneLabels, "v1.35.5+k3s1"),
			},
			selector: "node-role.kubernetes.io/control-plane=true",
			want:     "v1.35.5+k3s1",
		},
		{
			name: "mixed cluster returns lowest control-plane version, ignoring workers",
			nodes: []*corev1.Node{
				newNode("cp-0", controlPlaneLabels, "v1.36.0+k3s1"),
				newNode("cp-1", controlPlaneLabels, "v1.35.5+k3s1"),
				newNode("worker-0", workerLabels, "v1.30.0+k3s1"),
			},
			selector: "node-role.kubernetes.io/control-plane=true",
			want:     "v1.35.5+k3s1",
		},
		{
			name: "dual-labeled node (k8s 1.20-1.23 era) is matched by control-plane selector",
			nodes: []*corev1.Node{
				newNode("cp-0", dualLabels, "v1.23.0+k3s1"),
			},
			selector: "node-role.kubernetes.io/control-plane=true",
			want:     "v1.23.0+k3s1",
		},
		{
			name: "cluster with no control-plane label returns no nodes found",
			nodes: []*corev1.Node{
				newNode("legacy-cp", legacyMasterLabels, "v1.19.0+k3s1"),
				newNode("worker-0", workerLabels, "v1.19.0+k3s1"),
			},
			selector: "node-role.kubernetes.io/control-plane=true",
			wantErr:  "no nodes found in the cluster",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			objs := make([]runtime.Object, 0, len(tc.nodes))
			for _, n := range tc.nodes {
				objs = append(objs, n)
			}
			clientset := fake.NewSimpleClientset(objs...)
			k := KubernetesClient{Client: clientset}

			got, err := k.GetLowestNodeVersion(context.Background(), ListNodesInput{
				LabelSelector: tc.selector,
			})

			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil (returned version %q)", tc.wantErr, got)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tc.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("expected version %q, got %q", tc.want, got)
			}
		})
	}
}
