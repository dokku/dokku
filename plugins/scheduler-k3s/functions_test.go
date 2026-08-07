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

func TestNodeLabels(t *testing.T) {
	cases := []struct {
		name        string
		role        string
		profileName string
		wantKey     string
		wantValue   string
		wantProfile bool
	}{
		{
			name:        "server without a profile",
			role:        "server",
			profileName: "",
			wantKey:     "svccontroller.k3s.cattle.io/enablelb",
			wantValue:   "true",
			wantProfile: false,
		},
		{
			name:        "worker without a profile",
			role:        "worker",
			profileName: "",
			wantKey:     "node-role.kubernetes.io/worker",
			wantValue:   "worker",
			wantProfile: false,
		},
		{
			name:        "worker with a profile",
			role:        "worker",
			profileName: "edge-workers",
			wantKey:     "node-role.kubernetes.io/worker",
			wantValue:   "worker",
			wantProfile: true,
		},
		{
			name:        "server with a profile",
			role:        "server",
			profileName: "control-plane",
			wantKey:     "svccontroller.k3s.cattle.io/enablelb",
			wantValue:   "true",
			wantProfile: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := nodeLabels(tc.role, tc.profileName)

			if got[tc.wantKey] != tc.wantValue {
				t.Errorf("nodeLabels() role label = %q, want %q", got[tc.wantKey], tc.wantValue)
			}

			profile, ok := got[NodeProfileLabel]
			if ok != tc.wantProfile {
				t.Errorf("nodeLabels() has %s = %v, want %v", NodeProfileLabel, ok, tc.wantProfile)
			}
			if tc.wantProfile && profile != tc.profileName {
				t.Errorf("nodeLabels() %s = %q, want %q", NodeProfileLabel, profile, tc.profileName)
			}
		})
	}
}

func TestNodeLabelsDoesNotMutatePackageLabels(t *testing.T) {
	serverBefore := len(ServerLabels)
	workerBefore := len(WorkerLabels)

	nodeLabels("server", "control-plane")
	nodeLabels("worker", "edge-workers")

	if len(ServerLabels) != serverBefore {
		t.Errorf("nodeLabels() mutated ServerLabels: len = %d, want %d", len(ServerLabels), serverBefore)
	}
	if len(WorkerLabels) != workerBefore {
		t.Errorf("nodeLabels() mutated WorkerLabels: len = %d, want %d", len(WorkerLabels), workerBefore)
	}
}

func TestIsNamespacedSysctl(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{name: "net.core.somaxconn", want: true},
		{name: "net.ipv4.tcp_rmem", want: true},
		{name: "kernel.shm_rmid_forced", want: true},
		{name: "kernel.shmmax", want: true},
		{name: "kernel.msgmax", want: true},
		{name: "kernel.sem", want: true},
		{name: "fs.mqueue.msg_max", want: true},
		{name: "vm.max_map_count", want: false},
		{name: "vm.swappiness", want: false},
		{name: "kernel.pid_max", want: false},
		{name: "kernel.semaphore", want: false},
		{name: "fs.file-max", want: false},
		{name: "", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isNamespacedSysctl(tc.name); got != tc.want {
				t.Errorf("isNamespacedSysctl(%q) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}

func TestParseSysctls(t *testing.T) {
	cases := []struct {
		name    string
		values  []string
		want    []Sysctl
		wantErr bool
	}{
		{
			name:   "empty input",
			values: []string{},
			want:   []Sysctl{},
		},
		{
			name:   "single namespaced sysctl",
			values: []string{"net.core.somaxconn=1024"},
			want:   []Sysctl{{Name: "net.core.somaxconn", Value: "1024"}},
		},
		{
			name:   "sorted by name regardless of input order",
			values: []string{"net.core.somaxconn=1024", "kernel.sem=250", "fs.mqueue.msg_max=20"},
			want: []Sysctl{
				{Name: "fs.mqueue.msg_max", Value: "20"},
				{Name: "kernel.sem", Value: "250"},
				{Name: "net.core.somaxconn", Value: "1024"},
			},
		},
		{
			name:   "value containing an equals sign is preserved",
			values: []string{"net.ipv4.tcp_rmem=4096 87380 6291456"},
			want:   []Sysctl{{Name: "net.ipv4.tcp_rmem", Value: "4096 87380 6291456"}},
		},
		{
			name:    "non-namespaced sysctl is rejected",
			values:  []string{"vm.max_map_count=262144"},
			wantErr: true,
		},
		{
			name:    "missing equals sign is rejected",
			values:  []string{"net.core.somaxconn"},
			wantErr: true,
		},
		{
			name:    "empty name is rejected",
			values:  []string{"=1024"},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseSysctls(tc.values)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("parseSysctls(%v) expected an error, got %v", tc.values, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseSysctls(%v) unexpected error: %v", tc.values, err)
			}

			if len(got) != len(tc.want) {
				t.Fatalf("parseSysctls(%v) = %v, want %v", tc.values, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("parseSysctls(%v)[%d] = %v, want %v", tc.values, i, got[i], tc.want[i])
				}
			}
		})
	}
}
