package scheduler_k3s

import (
	"context"
	"testing"
)

func makeHook(version string) ChartHook {
	noop := func(context.Context, KubernetesClient, HelmChart, Release) error { return nil }
	return ChartHook{
		TargetVersion: version,
		PreUpgrade:    noop,
	}
}

func versionsOf(hooks []ChartHook) []string {
	out := make([]string, 0, len(hooks))
	for _, hook := range hooks {
		out = append(out, hook.TargetVersion)
	}
	return out
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestSelectChartHooks(t *testing.T) {
	previous := ChartHooks
	t.Cleanup(func() { ChartHooks = previous })

	ChartHooks = map[string][]ChartHook{
		"chart-with-hooks": {
			makeHook("0.12.0"),
			makeHook("0.10.0"),
			makeHook("0.11.0"),
		},
	}

	cases := []struct {
		name             string
		releaseName      string
		installedVersion string
		targetVersion    string
		want             []string
		wantErr          bool
	}{
		{
			name:             "fresh install yields no hooks",
			releaseName:      "chart-with-hooks",
			installedVersion: "",
			targetVersion:    "0.12.0",
			want:             nil,
		},
		{
			name:             "no hooks registered for release",
			releaseName:      "chart-without-hooks",
			installedVersion: "0.5.0",
			targetVersion:    "0.6.0",
			want:             nil,
		},
		{
			name:             "single applicable hook returned",
			releaseName:      "chart-with-hooks",
			installedVersion: "0.10.0",
			targetVersion:    "0.11.0",
			want:             []string{"0.11.0"},
		},
		{
			name:             "all hooks below installed version excluded",
			releaseName:      "chart-with-hooks",
			installedVersion: "0.12.0",
			targetVersion:    "0.12.0",
			want:             nil,
		},
		{
			name:             "all applicable hooks returned in ascending semver order",
			releaseName:      "chart-with-hooks",
			installedVersion: "0.9.0",
			targetVersion:    "0.12.0",
			want:             []string{"0.10.0", "0.11.0", "0.12.0"},
		},
		{
			name:             "partial coverage returns only unsatisfied hooks",
			releaseName:      "chart-with-hooks",
			installedVersion: "0.10.0",
			targetVersion:    "0.12.0",
			want:             []string{"0.11.0", "0.12.0"},
		},
		{
			name:             "target above all hooks still caps at chart target",
			releaseName:      "chart-with-hooks",
			installedVersion: "0.9.0",
			targetVersion:    "0.11.0",
			want:             []string{"0.10.0", "0.11.0"},
		},
		{
			name:             "unparseable installed version errors",
			releaseName:      "chart-with-hooks",
			installedVersion: "not-a-version",
			targetVersion:    "0.12.0",
			wantErr:          true,
		},
		{
			name:             "unparseable target version errors",
			releaseName:      "chart-with-hooks",
			installedVersion: "0.9.0",
			targetVersion:    "not-a-version",
			wantErr:          true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := selectChartHooks(tc.releaseName, tc.installedVersion, tc.targetVersion)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			gotVersions := versionsOf(got)
			if !equalStrings(gotVersions, tc.want) {
				t.Fatalf("versions mismatch: got %v, want %v", gotVersions, tc.want)
			}
		})
	}
}

func TestSelectChartHooksTreatsLeadingVAsEqual(t *testing.T) {
	previous := ChartHooks
	t.Cleanup(func() { ChartHooks = previous })

	ChartHooks = map[string][]ChartHook{
		"chart-v-prefixed": {
			makeHook("v1.13.3"),
		},
	}

	got, err := selectChartHooks("chart-v-prefixed", "v1.12.0", "v1.13.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].TargetVersion != "v1.13.3" {
		t.Fatalf("expected v1.13.3 hook, got %v", versionsOf(got))
	}
}
