package scheduler_k3s

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestNeedsImagePullSecretsPrune(t *testing.T) {
	cases := []struct {
		name    string
		live    []corev1.LocalObjectReference
		keepSet map[string]struct{}
		want    bool
	}{
		{
			name:    "empty live, empty keep",
			live:    []corev1.LocalObjectReference{},
			keepSet: map[string]struct{}{},
			want:    false,
		},
		{
			name: "matches keep set exactly",
			live: []corev1.LocalObjectReference{{Name: "pull-secret-foo"}},
			keepSet: map[string]struct{}{
				"pull-secret-foo": {},
			},
			want: false,
		},
		{
			name: "leaked entries beyond keep set",
			live: []corev1.LocalObjectReference{
				{Name: "ims-foo.111"},
				{Name: "ims-foo.222"},
				{Name: "pull-secret-foo"},
			},
			keepSet: map[string]struct{}{
				"pull-secret-foo": {},
			},
			want: true,
		},
		{
			name: "live missing the keep entry",
			live: []corev1.LocalObjectReference{
				{Name: "ims-foo.111"},
			},
			keepSet: map[string]struct{}{
				"pull-secret-foo": {},
			},
			want: true,
		},
		{
			name:    "live populated, keep set empty",
			live:    []corev1.LocalObjectReference{{Name: "ims-foo.111"}},
			keepSet: map[string]struct{}{},
			want:    true,
		},
		{
			name:    "empty live, keep set populated",
			live:    []corev1.LocalObjectReference{},
			keepSet: map[string]struct{}{"pull-secret-foo": {}},
			want:    true,
		},
		{
			name: "user override only, no leaked entries",
			live: []corev1.LocalObjectReference{{Name: "my-custom-secret"}},
			keepSet: map[string]struct{}{
				"my-custom-secret": {},
			},
			want: false,
		},
		{
			name: "user override with leaked dokku-managed entries",
			live: []corev1.LocalObjectReference{
				{Name: "ims-foo.111"},
				{Name: "my-custom-secret"},
			},
			keepSet: map[string]struct{}{
				"my-custom-secret": {},
			},
			want: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := needsImagePullSecretsPrune(tc.live, tc.keepSet)
			if got != tc.want {
				t.Errorf("needsImagePullSecretsPrune() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestInitializeInstallerArgs(t *testing.T) {
	cases := []struct {
		name       string
		input      InitializeInstallerArgsInput
		wantPairs  [][2]string
		wantAbsent []string
	}{
		{
			name: "no kubelet args",
			input: InitializeInstallerArgsInput{
				IngressClass: "traefik",
				NodeName:     "ip-10-0-0-1-abc",
				Token:        "sometoken",
			},
			wantAbsent: []string{"--kubelet-arg", "--node-taint"},
		},
		{
			name: "single kubelet arg",
			input: InitializeInstallerArgsInput{
				IngressClass: "traefik",
				KubeletArgs:  []string{"allowed-unsafe-sysctls=net.ipv4.tcp_rmem"},
				NodeName:     "ip-10-0-0-1-abc",
				Token:        "sometoken",
			},
			wantPairs: [][2]string{
				{"--kubelet-arg", "allowed-unsafe-sysctls=net.ipv4.tcp_rmem"},
			},
		},
		{
			name: "multiple kubelet args each get their own flag",
			input: InitializeInstallerArgsInput{
				IngressClass: "traefik",
				KubeletArgs:  []string{"allowed-unsafe-sysctls=net.ipv4.tcp_rmem", "max-pods=150"},
				NodeName:     "ip-10-0-0-1-abc",
				Token:        "sometoken",
			},
			wantPairs: [][2]string{
				{"--kubelet-arg", "allowed-unsafe-sysctls=net.ipv4.tcp_rmem"},
				{"--kubelet-arg", "max-pods=150"},
			},
		},
		{
			name: "kubelet args coexist with taint scheduling",
			input: InitializeInstallerArgsInput{
				IngressClass:    "nginx",
				KubeletArgs:     []string{"max-pods=150"},
				NodeName:        "ip-10-0-0-1-abc",
				TaintScheduling: true,
				Token:           "sometoken",
			},
			wantPairs: [][2]string{
				{"--node-taint", "CriticalAddonsOnly=true:NoSchedule"},
				{"--kubelet-arg", "max-pods=150"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := initializeInstallerArgs(tc.input)

			for _, pair := range tc.wantPairs {
				if !containsArgPair(got, pair[0], pair[1]) {
					t.Errorf("initializeInstallerArgs() missing %q %q, got %v", pair[0], pair[1], got)
				}
			}

			for _, absent := range tc.wantAbsent {
				for _, arg := range got {
					if arg == absent {
						t.Errorf("initializeInstallerArgs() unexpectedly contains %q, got %v", absent, got)
					}
				}
			}
		})
	}
}

func containsArgPair(args []string, flag string, value string) bool {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == flag && args[i+1] == value {
			return true
		}
	}
	return false
}
