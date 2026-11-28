package scheduler_k3s

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/repo"
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

func TestNewHelmSettings(t *testing.T) {
	sentinel := map[string]string{
		helmpath.CacheHomeEnvVar:  "/sentinel/cache",
		helmpath.ConfigHomeEnvVar: "/sentinel/config",
		helmpath.DataHomeEnvVar:   "/sentinel/data",
	}
	for envVar, value := range sentinel {
		t.Setenv(envVar, value)
	}

	settings, cleanup, err := newHelmSettings()
	if err != nil {
		t.Fatalf("newHelmSettings() returned an error: %v", err)
	}

	directory := os.Getenv(helmpath.CacheHomeEnvVar)
	if directory == sentinel[helmpath.CacheHomeEnvVar] {
		t.Fatalf("%s was not overridden", helmpath.CacheHomeEnvVar)
	}
	scratch := filepath.Dir(directory)

	if _, err := os.Stat(scratch); err != nil {
		t.Fatalf("scratch directory %q is not usable: %v", scratch, err)
	}

	for envVar, subdirectory := range helmHomeEnvVars {
		want := filepath.Join(scratch, subdirectory)
		if got := os.Getenv(envVar); got != want {
			t.Errorf("%s = %q, want %q", envVar, got, want)
		}
	}

	paths := map[string]string{
		"RepositoryConfig": settings.RepositoryConfig,
		"RepositoryCache":  settings.RepositoryCache,
		"RegistryConfig":   settings.RegistryConfig,
	}
	for name, path := range paths {
		if !strings.HasPrefix(path, scratch+string(os.PathSeparator)) {
			t.Errorf("%s = %q, want a path under %q", name, path, scratch)
		}
	}

	cleanup()

	if _, err := os.Stat(scratch); !os.IsNotExist(err) {
		t.Errorf("scratch directory %q still exists after cleanup", scratch)
	}

	for envVar, want := range sentinel {
		if got := os.Getenv(envVar); got != want {
			t.Errorf("%s = %q after cleanup, want the prior value %q", envVar, got, want)
		}
	}
}

func TestNewHelmSettingsUnsetsVariablesThatWereNotSet(t *testing.T) {
	for envVar := range helmHomeEnvVars {
		t.Setenv(envVar, "")
		os.Unsetenv(envVar) // nolint: errcheck
	}

	_, cleanup, err := newHelmSettings()
	if err != nil {
		t.Fatalf("newHelmSettings() returned an error: %v", err)
	}
	cleanup()

	for envVar := range helmHomeEnvVars {
		if value, ok := os.LookupEnv(envVar); ok {
			t.Errorf("%s = %q after cleanup, want it to be unset", envVar, value)
		}
	}
}

// newChartRepositoryServer serves a chart repository containing a single
// chart, failing the first failures requests for the chart archive
func newChartRepositoryServer(t *testing.T, failures int32) (*httptest.Server, *int32) {
	t.Helper()

	directory := t.TempDir()
	_, err := chartutil.Save(&chart.Chart{
		Metadata: &chart.Metadata{
			APIVersion: chart.APIVersionV2,
			Name:       "retry-test",
			Version:    "0.1.0",
		},
	}, directory)
	if err != nil {
		t.Fatalf("unable to package chart: %v", err)
	}

	var archiveRequests int32
	server := httptest.NewUnstartedServer(nil)
	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".tgz") && atomic.AddInt32(&archiveRequests, 1) <= failures {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		http.FileServer(http.Dir(directory)).ServeHTTP(w, r)
	})
	server.Start()
	t.Cleanup(server.Close)

	index, err := repo.IndexDirectory(directory, server.URL)
	if err != nil {
		t.Fatalf("unable to index chart repository: %v", err)
	}
	if err := index.WriteFile(filepath.Join(directory, "index.yaml"), 0644); err != nil {
		t.Fatalf("unable to write chart repository index: %v", err)
	}

	return server, &archiveRequests
}

func TestLocateChart(t *testing.T) {
	useFastRetries(t)

	cases := []struct {
		name         string
		failures     int32
		wantErr      bool
		wantRequests int32
	}{
		{
			name:         "downloads the chart on the first attempt",
			failures:     0,
			wantErr:      false,
			wantRequests: 1,
		},
		{
			name:         "retries a transient chart download error",
			failures:     1,
			wantErr:      false,
			wantRequests: 2,
		},
		{
			name:         "fails once every attempt is used",
			failures:     DefaultRetryAttempts,
			wantErr:      true,
			wantRequests: DefaultRetryAttempts,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server, archiveRequests := newChartRepositoryServer(t, tc.failures)

			settings, cleanup, err := newHelmSettings()
			if err != nil {
				t.Fatalf("newHelmSettings returned an error: %v", err)
			}
			defer cleanup()

			options := action.ChartPathOptions{RepoURL: server.URL, Version: "0.1.0"}
			location, err := locateChart(context.Background(), &options, "retry-test", settings)

			if got := atomic.LoadInt32(archiveRequests); got != tc.wantRequests {
				t.Errorf("archive requests = %d, want %d", got, tc.wantRequests)
			}

			if tc.wantErr {
				if err == nil {
					t.Fatal("err = nil, want an error")
				}
				return
			}

			if err != nil {
				t.Fatalf("err = %v, want nil", err)
			}
			if _, err := os.Stat(location); err != nil {
				t.Errorf("located chart %q does not exist: %v", location, err)
			}
		})
	}
}

func TestLocateChartDoesNotRetryLocalCharts(t *testing.T) {
	previous := DefaultRetryInitialDelay
	DefaultRetryInitialDelay = time.Hour
	t.Cleanup(func() {
		DefaultRetryInitialDelay = previous
	})

	settings, cleanup, err := newHelmSettings()
	if err != nil {
		t.Fatalf("newHelmSettings returned an error: %v", err)
	}
	defer cleanup()

	options := action.ChartPathOptions{}
	_, err = locateChart(context.Background(), &options, filepath.Join(t.TempDir(), "missing-chart"), settings)
	if err == nil {
		t.Fatal("err = nil, want an error for a missing local chart")
	}
}
