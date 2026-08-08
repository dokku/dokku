package scheduler_k3s

import (
	"strings"
	"testing"

	"github.com/dokku/dokku/plugins/common"
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

func TestNormalizeCertIssuerKind(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		want    string
		wantErr bool
	}{
		{name: "empty is left unset", value: "", want: ""},
		{name: "canonical issuer", value: "Issuer", want: CertIssuerKindIssuer},
		{name: "canonical cluster issuer", value: "ClusterIssuer", want: CertIssuerKindClusterIssuer},
		{name: "lowercase issuer is canonicalized", value: "issuer", want: CertIssuerKindIssuer},
		{name: "lowercase cluster issuer is canonicalized", value: "clusterissuer", want: CertIssuerKindClusterIssuer},
		{name: "uppercase issuer is canonicalized", value: "ISSUER", want: CertIssuerKindIssuer},
		{name: "unknown kind", value: "Certificate", wantErr: true},
		{name: "typo", value: "ClusterIssue", wantErr: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := normalizeCertIssuerKind(tc.value)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("normalizeCertIssuerKind(%q) expected an error, got %q", tc.value, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeCertIssuerKind(%q) unexpected error: %v", tc.value, err)
			}
			if got != tc.want {
				t.Errorf("normalizeCertIssuerKind(%q) = %q, want %q", tc.value, got, tc.want)
			}
		})
	}
}

func TestValidateCertIssuerName(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "empty unsets the property", value: ""},
		{name: "reserved disabled value", value: CertIssuerNameDisabled},
		{name: "simple name", value: "acme-dns"},
		{name: "dotted name", value: "my.issuer.example"},
		{name: "underscores are not legal kubernetes names", value: "my_own_issuer_name", wantErr: true},
		{name: "uppercase is not a legal kubernetes name", value: "AcmeDNS", wantErr: true},
		{name: "leading dash", value: "-acme-dns", wantErr: true},
		{name: "trailing dash", value: "acme-dns-", wantErr: true},
		{name: "too long", value: strings.Repeat("a", 254), wantErr: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validateCertIssuerName(tc.value)
			if tc.wantErr && err == nil {
				t.Fatalf("validateCertIssuerName(%q) expected an error", tc.value)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("validateCertIssuerName(%q) unexpected error: %v", tc.value, err)
			}
		})
	}
}

func TestValidateLetsencryptServer(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "empty unsets the property", value: ""},
		{name: "prod", value: "prod"},
		{name: "production", value: "production"},
		{name: "stag", value: "stag"},
		{name: "staging", value: "staging"},
		{name: "disabled", value: LetsencryptServerDisabled},
		{name: "typo", value: "prodd", wantErr: true},
		{name: "issuer name is not a server", value: "my-own-issuer", wantErr: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validateLetsencryptServer(tc.value)
			if tc.wantErr && err == nil {
				t.Fatalf("validateLetsencryptServer(%q) expected an error", tc.value)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("validateLetsencryptServer(%q) unexpected error: %v", tc.value, err)
			}
		})
	}
}

func TestDomainSlug(t *testing.T) {
	cases := []struct {
		name   string
		domain string
		want   string
	}{
		{name: "plain domain", domain: "example.com", want: "example-com"},
		{name: "wildcard domain keeps a distinct slug", domain: "*.example.com", want: "wildcard-example-com"},
		{name: "subdomain", domain: "app.example.com", want: "app-example-com"},
		{name: "wildcard subdomain", domain: "*.app.example.com", want: "wildcard-app-example-com"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := domainSlug(tc.domain); got != tc.want {
				t.Errorf("domainSlug(%q) = %q, want %q", tc.domain, got, tc.want)
			}
		})
	}
}

func TestDomainSlugDistinguishesWildcardFromApex(t *testing.T) {
	// A kubernetes wildcard host matches exactly one label, so *.example.com does not
	// cover example.com and both are commonly added to the same app. The generated
	// ingress name is derived from this slug, so a collision would render two objects
	// with the same name and fail the helm install.
	if domainSlug("*.example.com") == domainSlug("example.com") {
		t.Fatalf("domainSlug collides for *.example.com and example.com: %q", domainSlug("example.com"))
	}
}

func TestResolveAppTLSConfig(t *testing.T) {
	cases := []struct {
		name               string
		appProperties      map[string]string
		globalProperties   map[string]string
		importedCertExists bool
		want               AppTLSConfig
		wantErr            bool
	}{
		{
			name: "defaults to the shared production ClusterIssuer once a global email is set",
			globalProperties: map[string]string{
				"letsencrypt-email-prod": "ops@dokku.me",
			},
			want: AppTLSConfig{
				Enabled:    true,
				IssuerKind: CertIssuerKindClusterIssuer,
				IssuerName: "letsencrypt-prod",
			},
		},
		{
			name: "no letsencrypt email leaves tls disabled",
			want: AppTLSConfig{
				IssuerKind: CertIssuerKindClusterIssuer,
				IssuerName: "letsencrypt-prod",
			},
		},
		{
			name: "an app level email renders a namespaced Issuer",
			appProperties: map[string]string{
				"letsencrypt-server":     "staging",
				"letsencrypt-email-stag": "app@dokku.me",
			},
			want: AppTLSConfig{
				Enabled:    true,
				IssuerKind: CertIssuerKindIssuer,
				IssuerName: "myapp-letsencrypt-stag",
				Issuer: AppIssuer{
					Email:        "app@dokku.me",
					Enabled:      true,
					IngressClass: DefaultIngressClass,
					Name:         "myapp-letsencrypt-stag",
					Server:       LetsencryptServerStag,
				},
			},
		},
		{
			name:               "an imported certificate wins over a custom issuer",
			importedCertExists: true,
			appProperties: map[string]string{
				"cert-issuer-name": "acme-dns",
			},
			want: AppTLSConfig{
				Enabled:         true,
				UseImportedCert: true,
			},
		},
		{
			name: "letsencrypt-server false disables a globally configured custom issuer",
			appProperties: map[string]string{
				"letsencrypt-server": LetsencryptServerDisabled,
			},
			globalProperties: map[string]string{
				"cert-issuer-name": "acme-dns",
			},
			want: AppTLSConfig{},
		},
		{
			name: "a custom issuer enables tls with no letsencrypt email anywhere",
			appProperties: map[string]string{
				"cert-issuer-name": "acme-dns",
			},
			want: AppTLSConfig{
				Enabled:          true,
				IssuerKind:       CertIssuerKindClusterIssuer,
				IssuerName:       "acme-dns",
				UsesCustomIssuer: true,
			},
		},
		{
			name: "a namespaced custom issuer renders no Dokku managed issuer",
			appProperties: map[string]string{
				"cert-issuer-name": "acme-dns",
				"cert-issuer-kind": CertIssuerKindIssuer,
			},
			want: AppTLSConfig{
				Enabled:          true,
				IssuerKind:       CertIssuerKindIssuer,
				IssuerName:       "acme-dns",
				UsesCustomIssuer: true,
			},
		},
		{
			name: "a custom issuer beats a fully configured letsencrypt setup",
			appProperties: map[string]string{
				"cert-issuer-name":       "acme-dns",
				"letsencrypt-email-prod": "app@dokku.me",
			},
			want: AppTLSConfig{
				Enabled:          true,
				IssuerKind:       CertIssuerKindClusterIssuer,
				IssuerName:       "acme-dns",
				UsesCustomIssuer: true,
			},
		},
		{
			name: "an app level custom issuer overrides the global one",
			appProperties: map[string]string{
				"cert-issuer-name": "app-issuer",
			},
			globalProperties: map[string]string{
				"cert-issuer-name": "global-issuer",
			},
			want: AppTLSConfig{
				Enabled:          true,
				IssuerKind:       CertIssuerKindClusterIssuer,
				IssuerName:       "app-issuer",
				UsesCustomIssuer: true,
			},
		},
		{
			name: "the reserved disabled value falls back to letsencrypt",
			appProperties: map[string]string{
				"cert-issuer-name": CertIssuerNameDisabled,
			},
			globalProperties: map[string]string{
				"cert-issuer-name":       "global-issuer",
				"letsencrypt-email-prod": "ops@dokku.me",
			},
			want: AppTLSConfig{
				Enabled:    true,
				IssuerKind: CertIssuerKindClusterIssuer,
				IssuerName: "letsencrypt-prod",
			},
		},
		{
			name: "an invalid letsencrypt server still errors",
			appProperties: map[string]string{
				"letsencrypt-server": "prodd",
			},
			wantErr: true,
		},
		{
			name: "an invalid stored cert issuer kind errors",
			appProperties: map[string]string{
				"cert-issuer-name": "acme-dns",
				"cert-issuer-kind": "Certificate",
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			setupReportTest(t, "myapp")
			for key, value := range tc.appProperties {
				if err := common.PropertyWrite("scheduler-k3s", "myapp", key, value); err != nil {
					t.Fatalf("PropertyWrite(%q): %v", key, err)
				}
			}
			for key, value := range tc.globalProperties {
				if err := common.PropertyWrite("scheduler-k3s", "--global", key, value); err != nil {
					t.Fatalf("PropertyWrite(--global, %q): %v", key, err)
				}
			}

			got, err := resolveAppTLSConfig(ResolveAppTLSConfigInput{
				AppName:            "myapp",
				ImportedCertExists: tc.importedCertExists,
			})
			if tc.wantErr {
				if err == nil {
					t.Fatalf("resolveAppTLSConfig() expected an error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveAppTLSConfig() unexpected error: %v", err)
			}

			if got != tc.want {
				t.Errorf("resolveAppTLSConfig() = %+v, want %+v", got, tc.want)
			}
		})
	}
}
