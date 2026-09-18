package scheduler_k3s

import (
	"testing"

	"helm.sh/helm/v3/pkg/action"
)

func TestApplyChartPathOptions(t *testing.T) {
	cases := []struct {
		name        string
		options     action.ChartPathOptions
		input       ChartInput
		wantRepoURL string
		wantVersion string
	}{
		{
			name:        "pins both the repository and the version",
			input:       ChartInput{RepoURL: "https://helm.traefik.io/traefik", Version: "26.0.0"},
			wantRepoURL: "https://helm.traefik.io/traefik",
			wantVersion: "26.0.0",
		},
		{
			name:        "an empty version leaves the field unset",
			input:       ChartInput{RepoURL: "https://helm.traefik.io/traefik"},
			wantRepoURL: "https://helm.traefik.io/traefik",
			wantVersion: "",
		},
		{
			name:        "an empty repository leaves the field unset",
			input:       ChartInput{Version: "26.0.0"},
			wantRepoURL: "",
			wantVersion: "26.0.0",
		},
		{
			name:        "an empty input leaves both fields unset",
			input:       ChartInput{},
			wantRepoURL: "",
			wantVersion: "",
		},
		{
			name:        "existing values are preserved when the input is empty",
			options:     action.ChartPathOptions{RepoURL: "https://charts.jetstack.io", Version: "v1.13.3"},
			input:       ChartInput{},
			wantRepoURL: "https://charts.jetstack.io",
			wantVersion: "v1.13.3",
		},
		{
			name:        "existing values are overridden by the input",
			options:     action.ChartPathOptions{RepoURL: "https://charts.jetstack.io", Version: "v1.13.3"},
			input:       ChartInput{RepoURL: "https://helm.traefik.io/traefik", Version: "26.0.0"},
			wantRepoURL: "https://helm.traefik.io/traefik",
			wantVersion: "26.0.0",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			options := tc.options
			applyChartPathOptions(&options, tc.input)

			if options.RepoURL != tc.wantRepoURL {
				t.Errorf("RepoURL = %q, want %q", options.RepoURL, tc.wantRepoURL)
			}
			if options.Version != tc.wantVersion {
				t.Errorf("Version = %q, want %q", options.Version, tc.wantVersion)
			}
		})
	}
}

func makeRevisions(revisions ...int) []Release {
	out := make([]Release, 0, len(revisions))
	for _, revision := range revisions {
		out = append(out, Release{Name: "traefik", Revision: revision})
	}
	return out
}

func revisionNumbers(releases []Release) []int {
	out := make([]int, 0, len(releases))
	for _, release := range releases {
		out = append(out, release.Revision)
	}
	return out
}

func equalInts(a, b []int) bool {
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

func TestSelectRevisions(t *testing.T) {
	cases := []struct {
		name     string
		releases []Release
		max      int
		want     []int
	}{
		{
			name:     "sorts an unordered ledger ascending",
			releases: makeRevisions(3, 1, 2),
			max:      0,
			want:     []int{1, 2, 3},
		},
		{
			name:     "a max of one keeps the newest revision",
			releases: makeRevisions(3, 1, 2),
			max:      1,
			want:     []int{3},
		},
		{
			name:     "a max below the ledger size keeps the newest revisions",
			releases: makeRevisions(4, 1, 3, 2),
			max:      2,
			want:     []int{3, 4},
		},
		{
			name:     "a max above the ledger size keeps everything",
			releases: makeRevisions(2, 1),
			max:      10,
			want:     []int{1, 2},
		},
		{
			name:     "an empty ledger stays empty",
			releases: []Release{},
			max:      1,
			want:     []int{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := revisionNumbers(selectRevisions(tc.releases, tc.max))
			if !equalInts(got, tc.want) {
				t.Errorf("selectRevisions() = %v, want %v", got, tc.want)
			}
		})
	}
}
