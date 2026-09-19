package scheduler_k3s

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/dokku/dokku/plugins/common"
	"github.com/fluxcd/pkg/kustomize/filesys"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/postrender"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
)

var DevNullPrinter = func(format string, v ...interface{}) {}

var DeployLogPrinter = func(format string, v ...interface{}) {
	message := strings.TrimSpace(fmt.Sprintf(format, v...))
	if message == "" {
		return
	}
	r := []rune(message)
	r[0] = unicode.ToUpper(r[0])
	s := string(r)

	if strings.HasPrefix(s, "Beginning wait") {
		common.LogExclaim(s)
	} else if strings.HasPrefix(s, "Warning:") {
		common.LogExclaim(s)
	} else {
		common.LogVerboseQuiet(s)
	}
}

// ChartInput is the input for the InstallOrUpgradeChart function
type ChartInput struct {
	// ChartPath is the path to the chart to install or upgrade
	ChartPath string

	// KustomizeRootPath is the path to the kustomize root path to use
	KustomizeRootPath string

	// Namespace is the namespace to install or upgrade the chart in
	Namespace string

	// ReleaseName is the name of the release to install or upgrade
	ReleaseName string

	// RepoURL is the URL of the chart repository to install or upgrade the chart from
	RepoURL string

	// RollbackOnFailure is whether to rollback on failure
	RollbackOnFailure bool

	// Timeout is the timeout for the install or upgrade
	Timeout time.Duration

	// Wait is whether to wait for the install or upgrade to complete
	Wait bool

	// Version is the version of the chart to install or upgrade
	Version string

	// Values is the values to pass to the chart
	Values map[string]interface{}
}

type Release struct {
	AppVersion string
	Name       string
	Namespace  string
	Revision   int
	Status     release.Status
	Version    string
}

// applyChartPathOptions pins the repository and the version from input onto the
// chart path options used to locate a chart. Both the install and the upgrade
// path must do this, otherwise helm resolves the newest chart published to the
// repository instead of the version dokku pins.
func applyChartPathOptions(options *action.ChartPathOptions, input ChartInput) {
	if input.RepoURL != "" {
		options.RepoURL = input.RepoURL
	}
	if input.Version != "" {
		options.Version = input.Version
	}
}

// helmHomeEnvVars are the environment variables helm resolves its cache, config
// and data homes from, mapped to the scratch subdirectory each one gets.
var helmHomeEnvVars = map[string]string{
	helmpath.CacheHomeEnvVar:  "cache",
	helmpath.ConfigHomeEnvVar: "config",
	helmpath.DataHomeEnvVar:   "data",
}

// newHelmSettings returns helm settings backed by a private scratch directory,
// along with a function that removes it and restores the environment.
//
// helm otherwise resolves its cache, config and data homes from $HOME. Dokku
// runs scheduler-k3s:initialize as root and every other subcommand as the dokku
// user, so $HOME is not always readable by the user a command ends up as, and
// helm treats a repositories file it cannot read as a fatal error rather than
// an absent one. Nothing dokku keeps there needs to outlive a command: every
// chart carries its own repository url, and helm re-fetches both the repository
// index and the chart archive on every call.
func newHelmSettings() (*cli.EnvSettings, func(), error) {
	directory, err := os.MkdirTemp("", "dokku-helm-")
	if err != nil {
		return nil, nil, fmt.Errorf("Error creating helm scratch directory: %w", err)
	}

	restore := map[string]*string{}
	cleanup := func() {
		for envVar, previous := range restore {
			if previous == nil {
				os.Unsetenv(envVar) // nolint: errcheck
				continue
			}

			os.Setenv(envVar, *previous) // nolint: errcheck
		}

		os.RemoveAll(directory) // nolint: errcheck
	}

	for envVar, subdirectory := range helmHomeEnvVars {
		if previous, ok := os.LookupEnv(envVar); ok {
			restore[envVar] = &previous
		} else {
			restore[envVar] = nil
		}

		if err := os.Setenv(envVar, filepath.Join(directory, subdirectory)); err != nil {
			cleanup()
			return nil, nil, fmt.Errorf("Error setting %s: %w", envVar, err)
		}
	}

	return cli.New(), cleanup, nil
}

type HelmAgent struct {
	Configuration *action.Configuration
	Namespace     string
	Logger        action.DebugLog
}

func NewHelmAgent(namespace string, logger action.DebugLog) (*HelmAgent, error) {
	actionConfig := new(action.Configuration)

	helmDriver := os.Getenv("HELM_DRIVER")
	if helmDriver == "" {
		helmDriver = "secrets"
	}

	kubeconfigPath := getComputedKubeconfigPath()
	kubeContext := getComputedKubeContext()
	kubeConfig := kube.GetConfig(kubeconfigPath, kubeContext, namespace)
	if err := actionConfig.Init(kubeConfig, namespace, helmDriver, logger); err != nil {
		return nil, err
	}

	return &HelmAgent{
		Configuration: actionConfig,
		Namespace:     namespace,
		Logger:        logger,
	}, nil
}

func (h *HelmAgent) ChartExists(releaseName string) (bool, error) {
	if releaseName == "" {
		return false, fmt.Errorf("Release name is required")
	}

	client := action.NewHistory(h.Configuration)
	client.Max = 1
	releases, err := client.Run(releaseName)
	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			return false, nil
		}
		return false, err
	}

	if len(releases) > 0 {
		return true, nil
	}

	return false, nil
}

func (h *HelmAgent) DeleteRevision(ctx context.Context, releaseName string, revision int) error {
	clientset, err := NewKubernetesClient()
	if err != nil {
		return fmt.Errorf("Error creating kubernetes client: %w", err)
	}

	secretName := fmt.Sprintf("sh.helm.release.v1.%s.v%d", releaseName, revision)
	err = clientset.DeleteSecret(ctx, DeleteSecretInput{
		Name:      secretName,
		Namespace: h.Namespace,
	})
	if err != nil {
		return fmt.Errorf("Error deleting secret: %w", err)
	}
	return nil
}

func (h *HelmAgent) GetValues(releaseName string) (map[string]interface{}, error) {
	client := action.NewGetValues(h.Configuration)
	client.AllValues = true
	values, err := client.Run(releaseName)
	if err != nil {
		return nil, fmt.Errorf("Error getting values: %w", err)
	}

	return values, nil
}

func (h *HelmAgent) InstallOrUpgradeChart(ctx context.Context, input ChartInput) error {
	chartExists, err := h.ChartExists(input.ReleaseName)
	if err != nil {
		return fmt.Errorf("Error checking if chart exists: %w", err)
	}

	if chartExists {
		return h.UpgradeChart(ctx, input)
	}

	return h.InstallChart(ctx, input)
}

func (h *HelmAgent) InstallChart(ctx context.Context, input ChartInput) error {
	namespace := input.Namespace
	if namespace == "" {
		namespace = h.Namespace
	}

	if input.ChartPath == "" {
		return fmt.Errorf("Chart path is required")
	}
	if input.ReleaseName == "" {
		return fmt.Errorf("Release name is required")
	}
	if input.Values == nil {
		input.Values = map[string]interface{}{}
	}

	kustomizeRenderer := KustomizeRenderer{
		ReleaseName:       input.ReleaseName,
		KustomizeRootPath: input.KustomizeRootPath,
	}

	client := action.NewInstall(h.Configuration)
	client.Atomic = false
	client.CreateNamespace = true
	client.DryRun = false
	if os.Getenv("DOKKU_TRACE") == "1" {
		client.PostRenderer = &DebugRenderer{
			Renderer: &kustomizeRenderer,
		}
	} else {
		client.PostRenderer = &kustomizeRenderer
	}
	client.Namespace = namespace
	client.ReleaseName = input.ReleaseName
	client.Timeout = input.Timeout
	client.Wait = input.Wait

	applyChartPathOptions(&client.ChartPathOptions, input)

	settings, cleanup, err := newHelmSettings()
	if err != nil {
		return err
	}
	defer cleanup()

	chart, err := client.ChartPathOptions.LocateChart(input.ChartPath, settings)
	if err != nil {
		return fmt.Errorf("Error locating chart: %w", err)
	}

	chartRequested, err := loader.Load(chart)
	if err != nil {
		return fmt.Errorf("Error loading chart: %w", err)
	}

	_, err = client.RunWithContext(ctx, chartRequested, input.Values)
	if err != nil {
		return fmt.Errorf("Error deploying: %w", err)
	}

	return nil
}

func (h *HelmAgent) InstalledRevision(releaseName string) (Release, error) {
	revisions, err := h.ListRevisions(ListRevisionsInput{
		ReleaseName: releaseName,
		Max:         1,
	})
	if err != nil {
		return Release{}, err
	}

	if len(revisions) == 0 {
		return Release{}, nil
	}

	return revisions[len(revisions)-1], nil
}

type ListRevisionsInput struct {
	ReleaseName string
	Max         int
}

// selectRevisions sorts releases ascending by revision and keeps at most the
// most recent max entries. A max of zero keeps every revision. helm's history
// action ignores its own Max field and hands back the entire ledger unsorted,
// so both the ordering and the truncation happen here.
func selectRevisions(releases []Release, max int) []Release {
	sort.Slice(releases, func(i, j int) bool {
		return releases[i].Revision < releases[j].Revision
	})

	if max > 0 && len(releases) > max {
		return releases[len(releases)-max:]
	}

	return releases
}

func (h *HelmAgent) ListRevisions(input ListRevisionsInput) ([]Release, error) {
	client := action.NewHistory(h.Configuration)

	releases := []Release{}
	response, err := client.Run(input.ReleaseName)
	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			return releases, nil
		}

		return nil, fmt.Errorf("Error getting revisions: %w", err)
	}

	for _, release := range response {
		appVersion := "MISSING"
		if release.Chart != nil && release.Chart.Metadata != nil {
			appVersion = release.Chart.AppVersion()
		}

		releases = append(releases, Release{
			AppVersion: appVersion,
			Name:       release.Name,
			Namespace:  release.Namespace,
			Revision:   release.Version,
			Status:     release.Info.Status,
			Version:    release.Chart.Metadata.Version,
		})
	}

	return selectRevisions(releases, input.Max), nil
}

func (h *HelmAgent) UpgradeChart(ctx context.Context, input ChartInput) error {
	namespace := input.Namespace
	if namespace == "" {
		namespace = h.Namespace
	}

	if input.ChartPath == "" {
		return fmt.Errorf("Chart path is required")
	}
	if input.ReleaseName == "" {
		return fmt.Errorf("Release name is required")
	}
	if input.Values == nil {
		input.Values = map[string]interface{}{}
	}

	kustomizeRenderer := KustomizeRenderer{
		ReleaseName:       input.ReleaseName,
		KustomizeRootPath: input.KustomizeRootPath,
	}

	client := action.NewUpgrade(h.Configuration)
	client.Atomic = input.RollbackOnFailure
	client.CleanupOnFail = true
	client.MaxHistory = 10
	if os.Getenv("DOKKU_TRACE") == "1" {
		client.PostRenderer = &DebugRenderer{
			Renderer: &kustomizeRenderer,
		}
	} else {
		client.PostRenderer = &kustomizeRenderer
	}
	client.Namespace = namespace
	client.Timeout = input.Timeout
	client.Wait = input.Wait

	applyChartPathOptions(&client.ChartPathOptions, input)

	settings, cleanup, err := newHelmSettings()
	if err != nil {
		return err
	}
	defer cleanup()

	chart, err := client.ChartPathOptions.LocateChart(input.ChartPath, settings)
	if err != nil {
		return fmt.Errorf("Error locating chart: %w", err)
	}

	chartRequested, err := loader.Load(chart)
	if err != nil {
		return fmt.Errorf("Error loading chart: %w", err)
	}

	_, err = client.RunWithContext(ctx, input.ReleaseName, chartRequested, input.Values)
	if err != nil {
		return fmt.Errorf("Error deploying: %w", err)
	}

	return nil
}

func (h *HelmAgent) UninstallChart(releaseName string) error {
	exists, err := h.ChartExists(releaseName)
	if err != nil {
		return fmt.Errorf("Error checking if chart exists: %w", err)
	}

	if !exists {
		return nil
	}

	uninstall := action.NewUninstall(h.Configuration)
	uninstall.DeletionPropagation = "foreground"
	_, err = uninstall.Run(releaseName)
	if err != nil {
		return fmt.Errorf("Error uninstalling chart: %w", err)
	}

	return nil
}

type DebugRenderer struct {
	Renderer postrender.PostRenderer
}

func (p *DebugRenderer) Run(renderedManifests *bytes.Buffer) (*bytes.Buffer, error) {
	renderedManifests, err := p.Renderer.Run(renderedManifests)
	if err != nil {
		return nil, err
	}

	for _, line := range strings.Split(renderedManifests.String(), "\n") {
		common.LogWarn(line)
	}
	return renderedManifests, nil
}

// KustomizeRenderer is a post renderer that kustomizes the rendered manifests
type KustomizeRenderer struct {
	// KustomizeRootPath is the path to the kustomize root path to use
	KustomizeRootPath string

	// ReleaseName is the name of the release to kustomize
	ReleaseName string
}

// Run kustomizes the rendered manifests
func (p *KustomizeRenderer) Run(renderedManifests *bytes.Buffer) (*bytes.Buffer, error) {
	if p.KustomizeRootPath == "" {
		return renderedManifests, nil
	}

	if !common.DirectoryExists(p.KustomizeRootPath) {
		return renderedManifests, nil
	}

	common.LogVerboseQuiet(fmt.Sprintf("Applying kustomization to %s", p.ReleaseName))
	fs, err := filesys.MakeFsOnDiskSecureBuild(p.KustomizeRootPath)
	if err != nil {
		return nil, fmt.Errorf("Error creating filesystem: %w", err)
	}

	var kfile string
	for _, f := range konfig.RecognizedKustomizationFileNames() {
		if kf := filepath.Join(p.KustomizeRootPath, f); fs.Exists(kf) {
			kfile = kf
			break
		}
	}
	if kfile == "" {
		return nil, fmt.Errorf("%s not found", konfig.DefaultKustomizationFileName())
	}

	if err := fs.WriteFile(filepath.Join(p.KustomizeRootPath, "rendered.yaml"), renderedManifests.Bytes()); err != nil {
		return nil, fmt.Errorf("Error writing rendered.yaml: %w", err)
	}

	buildOptions := &krusty.Options{
		LoadRestrictions: types.LoadRestrictionsNone,
		PluginConfig:     types.DisabledPluginConfig(),
	}

	k := krusty.MakeKustomizer(buildOptions)
	m, err := k.Run(fs, p.KustomizeRootPath)
	if err != nil {
		return nil, err
	}

	resources, err := m.AsYaml()
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(resources), nil
}
