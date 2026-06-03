package scheduler_k3s

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/scheduler-k3s/internal/helmdiff"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
)

// CommandPreview prints a unified diff between the manifests currently stored
// in the live Helm release for an app and the manifests that the next deploy
// would roll out. Only the main app helm release is compared; the auxiliary
// config-secret, image-pull-secret, and TLS-secret releases are out of scope.
//
// diffContext controls the number of unchanged lines shown around each change;
// pass -1 to print the full resource.
func CommandPreview(appName string, diffContext int, showSecrets bool, showSecretsDecoded bool) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	scheduler := common.PropertyGetDefault("scheduler", appName, "selected", "")
	globalScheduler := common.PropertyGetDefault("scheduler", "--global", "selected", "docker-local")
	if scheduler == "" {
		scheduler = globalScheduler
	}
	if scheduler != "k3s" {
		return fmt.Errorf("App %s is not configured to use the k3s scheduler", appName)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM)
	go func() {
		<-signals
		cancel()
	}()

	namespace := getComputedNamespace(appName)
	helmAgent, err := NewHelmAgent(namespace, DevNullPrinter)
	if err != nil {
		return fmt.Errorf("Error creating helm agent: %w", err)
	}

	currentRelease, err := fetchCurrentRelease(helmAgent, appName)
	if err != nil {
		return err
	}

	buildOpts := BuildOptions{AllowMissingImage: true}
	if currentRelease != nil {
		buildOpts.OverrideDeploymentID = deploymentIDFromRelease(currentRelease)
	}

	chartResult, err := BuildAppChart(ctx, appName, "", buildOpts)
	if err != nil {
		return err
	}
	defer os.RemoveAll(chartResult.ChartDir)

	chartRequested, err := loadChartForPreview(chartResult.ChartPath)
	if err != nil {
		return err
	}

	var currentManifest, proposedManifest string
	if currentRelease != nil {
		currentManifest = currentRelease.Manifest
		proposedManifest, err = renderUpgradeDryRun(ctx, helmAgent, chartResult, chartRequested)
	} else {
		proposedManifest, err = renderInstallDryRun(ctx, helmAgent, chartResult, chartRequested)
	}
	if err != nil {
		return err
	}

	options := &helmdiff.Options{
		OutputFormat:       "diff",
		OutputContext:      diffContext,
		ShowSecrets:        showSecrets,
		ShowSecretsDecoded: showSecretsDecoded,
	}

	current := helmdiff.Parse([]byte(currentManifest), chartResult.Namespace, true)
	proposed := helmdiff.Parse([]byte(proposedManifest), chartResult.Namespace, true)
	helmdiff.Manifests(current, proposed, options, os.Stdout)
	return nil
}

func loadChartForPreview(chartPath string) (*chart.Chart, error) {
	settings := cli.New()
	opts := action.ChartPathOptions{}
	located, err := opts.LocateChart(chartPath, settings)
	if err != nil {
		return nil, fmt.Errorf("Error locating chart: %w", err)
	}
	chartRequested, err := loader.Load(located)
	if err != nil {
		return nil, fmt.Errorf("Error loading chart: %w", err)
	}
	return chartRequested, nil
}

func fetchCurrentRelease(helmAgent *HelmAgent, releaseName string) (*release.Release, error) {
	getAction := action.NewGet(helmAgent.Configuration)
	rel, err := getAction.Run(releaseName)
	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("Error fetching current release: %w", err)
	}
	return rel, nil
}

// deploymentIDFromRelease extracts the deployment id encoded in a release's
// chart version. BuildAppChart writes the chart version as "0.0.<id>"
// (chart.go); the parsed id lets CommandPreview pin the previewed manifests
// to the currently-deployed value so the `app.kubernetes.io/version`
// annotation does not churn in the diff. Returns 0 when the version does not
// match the expected shape - the caller then falls back to the timestamp.
func deploymentIDFromRelease(rel *release.Release) int64 {
	if rel == nil || rel.Chart == nil || rel.Chart.Metadata == nil {
		return 0
	}
	const prefix = "0.0."
	v := rel.Chart.Metadata.Version
	if !strings.HasPrefix(v, prefix) {
		return 0
	}
	id, err := strconv.ParseInt(strings.TrimPrefix(v, prefix), 10, 64)
	if err != nil {
		return 0
	}
	return id
}

func renderUpgradeDryRun(ctx context.Context, helmAgent *HelmAgent, chartResult BuildResult, chartRequested *chart.Chart) (string, error) {
	kustomizeRenderer := KustomizeRenderer{
		ReleaseName:       chartResult.ReleaseName,
		KustomizeRootPath: chartResult.KustomizeRootPath,
	}

	client := action.NewUpgrade(helmAgent.Configuration)
	client.DryRun = true
	client.DryRunOption = "client"
	client.Namespace = chartResult.Namespace
	client.PostRenderer = &kustomizeRenderer

	rel, err := client.RunWithContext(ctx, chartResult.ReleaseName, chartRequested, map[string]interface{}{})
	if err != nil {
		return "", fmt.Errorf("Error rendering upgrade dry-run: %w", err)
	}
	return rel.Manifest, nil
}

func renderInstallDryRun(ctx context.Context, helmAgent *HelmAgent, chartResult BuildResult, chartRequested *chart.Chart) (string, error) {
	kustomizeRenderer := KustomizeRenderer{
		ReleaseName:       chartResult.ReleaseName,
		KustomizeRootPath: chartResult.KustomizeRootPath,
	}

	client := action.NewInstall(helmAgent.Configuration)
	client.DryRun = true
	client.ClientOnly = true
	client.Namespace = chartResult.Namespace
	client.ReleaseName = chartResult.ReleaseName
	client.PostRenderer = &kustomizeRenderer

	rel, err := client.RunWithContext(ctx, chartRequested, map[string]interface{}{})
	if err != nil {
		return "", fmt.Errorf("Error rendering install dry-run: %w", err)
	}
	return rel.Manifest, nil
}
