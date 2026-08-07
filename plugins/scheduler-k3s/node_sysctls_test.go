package scheduler_k3s

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
)

func TestGetNodeSysctlsProperty(t *testing.T) {
	if got := getNodeSysctlsProperty(""); got != "node-sysctls.global" {
		t.Errorf("getNodeSysctlsProperty(\"\") = %q, want node-sysctls.global", got)
	}
	if got := getNodeSysctlsProperty("edge-workers"); got != "node-sysctls.profile.edge-workers" {
		t.Errorf("getNodeSysctlsProperty(\"edge-workers\") = %q, want node-sysctls.profile.edge-workers", got)
	}
}

// TestGetNodeSysctlsPropertyIsReserved asserts the property prefix is excluded from
// the annotations scan. Without this a property like node-sysctls.profile.pod would
// be mistaken for a pod annotation.
func TestGetNodeSysctlsPropertyIsReserved(t *testing.T) {
	reserved := false
	for _, prefix := range reservedAnnotationPrefixes {
		if prefix == "node-sysctls." {
			reserved = true
		}
	}

	if !reserved {
		t.Error("node-sysctls. is not in reservedAnnotationPrefixes")
	}
}

func TestGetNodeSysctlsReleaseName(t *testing.T) {
	if got := getNodeSysctlsReleaseName(""); got != "dokku-node-sysctls-global" {
		t.Errorf("getNodeSysctlsReleaseName(\"\") = %q, want dokku-node-sysctls-global", got)
	}
	if got := getNodeSysctlsReleaseName("edge-workers"); got != "dokku-node-sysctls-profile-edge-workers" {
		t.Errorf("getNodeSysctlsReleaseName(\"edge-workers\") = %q, want dokku-node-sysctls-profile-edge-workers", got)
	}
}

func TestSortedSysctls(t *testing.T) {
	got := sortedSysctls(map[string]string{
		"vm.swappiness":    "10",
		"vm.max_map_count": "262144",
		"fs.file-max":      "100000",
	})

	want := []Sysctl{
		{Name: "fs.file-max", Value: "100000"},
		{Name: "vm.max_map_count", Value: "262144"},
		{Name: "vm.swappiness", Value: "10"},
	}

	if len(got) != len(want) {
		t.Fatalf("sortedSysctls() = %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("sortedSysctls()[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestSortedSysctlsEmpty(t *testing.T) {
	if got := sortedSysctls(map[string]string{}); len(got) != 0 {
		t.Errorf("sortedSysctls(empty) = %v, want empty", got)
	}
}

func TestNodeSysctlsScopeLabel(t *testing.T) {
	if got := nodeSysctlsScopeLabel(""); got != "unprofiled nodes" {
		t.Errorf("nodeSysctlsScopeLabel(\"\") = %q, want 'unprofiled nodes'", got)
	}
	if got := nodeSysctlsScopeLabel("edge"); got != "node profile edge" {
		t.Errorf("nodeSysctlsScopeLabel(\"edge\") = %q, want 'node profile edge'", got)
	}
}

func renderNodeSysctlsTemplate(t *testing.T, values NodeSysctlsGlobalValues) string {
	t.Helper()

	chartDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(chartDir, "templates"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	chartYAML := []byte("apiVersion: v2\nname: test\nversion: 0.0.1\n")
	if err := os.WriteFile(filepath.Join(chartDir, "Chart.yaml"), chartYAML, 0o644); err != nil {
		t.Fatalf("write Chart.yaml: %v", err)
	}

	tpl, err := templates.ReadFile("templates/node-sysctls-chart/templates/daemonset.yaml")
	if err != nil {
		t.Fatalf("read daemonset template: %v", err)
	}
	if err := os.WriteFile(filepath.Join(chartDir, "templates", "daemonset.yaml"), tpl, 0o644); err != nil {
		t.Fatalf("write daemonset template: %v", err)
	}

	loaded, err := loader.Load(chartDir)
	if err != nil {
		t.Fatalf("load chart: %v", err)
	}

	sysctls := []interface{}{}
	for _, sysctl := range values.Sysctls {
		sysctls = append(sysctls, map[string]interface{}{"name": sysctl.Name, "value": sysctl.Value})
	}

	renderValues, err := chartutil.ToRenderValues(loaded, map[string]interface{}{
		"global": map[string]interface{}{
			"image":        values.Image,
			"pause_image":  values.PauseImage,
			"profile_name": values.ProfileName,
			"release_name": values.ReleaseName,
			"sysctls":      sysctls,
		},
	}, chartutil.ReleaseOptions{Name: "test", Namespace: "kube-system"}, nil)
	if err != nil {
		t.Fatalf("ToRenderValues: %v", err)
	}

	rendered, err := engine.Render(loaded, renderValues)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	for name, content := range rendered {
		if filepath.Base(name) == "daemonset.yaml" {
			return content
		}
	}
	t.Fatalf("daemonset.yaml not rendered; got: %v", rendered)
	return ""
}

// TestNodeSysctlsDaemonSetRendering asserts each scope targets a disjoint set of
// nodes. If both the global and a profile DaemonSet landed on the same node they
// would race writing the same /proc/sys file, making the result depend on pod
// scheduling order.
func TestNodeSysctlsDaemonSetRendering(t *testing.T) {
	base := NodeSysctlsGlobalValues{
		Image:      DefaultNodeSysctlsImage,
		PauseImage: DefaultNodeSysctlsPauseImage,
		Sysctls:    []Sysctl{{Name: "vm.max_map_count", Value: "262144"}},
	}

	t.Run("global scope excludes profiled nodes", func(t *testing.T) {
		values := base
		values.ReleaseName = "dokku-node-sysctls-global"
		manifest := renderNodeSysctlsTemplate(t, values)

		if !strings.Contains(manifest, "operator: DoesNotExist") {
			t.Errorf("global daemonset missing DoesNotExist affinity:\n%s", manifest)
		}
		if !strings.Contains(manifest, "key: dokku.com/node-profile") {
			t.Errorf("global daemonset missing node profile affinity key:\n%s", manifest)
		}
		if strings.Contains(manifest, "nodeSelector:") {
			t.Errorf("global daemonset should not use a nodeSelector:\n%s", manifest)
		}
	})

	t.Run("profile scope targets only its own nodes", func(t *testing.T) {
		values := base
		values.ProfileName = "edge-workers"
		values.ReleaseName = "dokku-node-sysctls-profile-edge-workers"
		manifest := renderNodeSysctlsTemplate(t, values)

		if !strings.Contains(manifest, "dokku.com/node-profile: \"edge-workers\"") {
			t.Errorf("profile daemonset missing nodeSelector:\n%s", manifest)
		}
		if strings.Contains(manifest, "DoesNotExist") {
			t.Errorf("profile daemonset should not carry the global affinity:\n%s", manifest)
		}
	})

	t.Run("tolerates every taint", func(t *testing.T) {
		values := base
		values.ReleaseName = "dokku-node-sysctls-global"
		manifest := renderNodeSysctlsTemplate(t, values)

		if !strings.Contains(manifest, "- operator: Exists") {
			t.Errorf("daemonset does not tolerate all taints:\n%s", manifest)
		}
	})

	t.Run("applies each sysctl privileged", func(t *testing.T) {
		values := base
		values.ReleaseName = "dokku-node-sysctls-global"
		values.Sysctls = []Sysctl{
			{Name: "vm.max_map_count", Value: "262144"},
			{Name: "vm.swappiness", Value: "10"},
		}
		manifest := renderNodeSysctlsTemplate(t, values)

		if !strings.Contains(manifest, "privileged: true") {
			t.Errorf("daemonset init container is not privileged:\n%s", manifest)
		}
		if !strings.Contains(manifest, `sysctl -w "vm.max_map_count=262144"`) {
			t.Errorf("daemonset missing max_map_count write:\n%s", manifest)
		}
		if !strings.Contains(manifest, `sysctl -w "vm.swappiness=10"`) {
			t.Errorf("daemonset missing swappiness write:\n%s", manifest)
		}
	})

	t.Run("image overrides reach both containers", func(t *testing.T) {
		values := base
		values.ReleaseName = "dokku-node-sysctls-global"
		values.Image = "registry.internal/busybox:1.36"
		values.PauseImage = "registry.internal/pause:3.9"
		manifest := renderNodeSysctlsTemplate(t, values)

		if !strings.Contains(manifest, `image: "registry.internal/busybox:1.36"`) {
			t.Errorf("daemonset did not use the sysctl image override:\n%s", manifest)
		}
		if !strings.Contains(manifest, `image: "registry.internal/pause:3.9"`) {
			t.Errorf("daemonset did not use the pause image override:\n%s", manifest)
		}
	})
}

// TestMergeNodeSysctls asserts a profile scope inherits the global sysctls and
// overrides them on conflict. Profiled nodes are excluded from the global
// DaemonSet, so a global value omitted here would never reach them.
func TestMergeNodeSysctls(t *testing.T) {
	cases := []struct {
		name    string
		global  map[string]string
		profile map[string]string
		want    []Sysctl
	}{
		{
			name:    "no sysctls at all",
			global:  map[string]string{},
			profile: map[string]string{},
			want:    []Sysctl{},
		},
		{
			name:    "profile inherits global values",
			global:  map[string]string{"vm.max_map_count": "262144"},
			profile: map[string]string{},
			want:    []Sysctl{{Name: "vm.max_map_count", Value: "262144"}},
		},
		{
			name:    "profile wins on conflict",
			global:  map[string]string{"vm.max_map_count": "262144"},
			profile: map[string]string{"vm.max_map_count": "524288"},
			want:    []Sysctl{{Name: "vm.max_map_count", Value: "524288"}},
		},
		{
			name:    "profile adds to global",
			global:  map[string]string{"vm.max_map_count": "262144"},
			profile: map[string]string{"vm.swappiness": "10"},
			want: []Sysctl{
				{Name: "vm.max_map_count", Value: "262144"},
				{Name: "vm.swappiness", Value: "10"},
			},
		},
		{
			name:    "profile only",
			global:  map[string]string{},
			profile: map[string]string{"vm.swappiness": "10"},
			want:    []Sysctl{{Name: "vm.swappiness", Value: "10"}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeNodeSysctls(tc.global, tc.profile)
			if len(got) != len(tc.want) {
				t.Fatalf("mergeNodeSysctls() = %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("mergeNodeSysctls()[%d] = %v, want %v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

// TestMergeNodeSysctlsDoesNotMutateInputs asserts the merge leaves the caller's maps
// alone, since resolveNodeSysctlScopes reuses the global map across every profile.
func TestMergeNodeSysctlsDoesNotMutateInputs(t *testing.T) {
	global := map[string]string{"vm.max_map_count": "262144"}
	profile := map[string]string{"vm.max_map_count": "524288", "vm.swappiness": "10"}

	mergeNodeSysctls(global, profile)

	if len(global) != 1 || global["vm.max_map_count"] != "262144" {
		t.Errorf("mergeNodeSysctls() mutated the global map: %v", global)
	}
	if len(profile) != 2 {
		t.Errorf("mergeNodeSysctls() mutated the profile map: %v", profile)
	}
}
