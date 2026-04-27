package scheduler_k3s

import (
	"context"
	"fmt"
	"sort"

	"github.com/Masterminds/semver/v3"
)

// ChartHookFunc is the signature for a function that runs around a helm chart
// upgrade. It receives the active Kubernetes client, the chart definition, and
// the helm release that was installed before the upgrade started.
type ChartHookFunc func(ctx context.Context, clientset KubernetesClient, chart HelmChart, installedRevision Release) error

// ChartHook attaches version-specific behaviour to a helm chart upgrade. A
// hook fires when the currently installed release version is below
// TargetVersion and the chart's configured target version is at or above
// TargetVersion. PreUpgrade runs before the chart is upgraded to
// TargetVersion; PostUpgrade runs after that intermediate upgrade succeeds.
type ChartHook struct {
	TargetVersion string
	PreUpgrade    ChartHookFunc
	PostUpgrade   ChartHookFunc
}

// ChartHooks is keyed by HelmChart.ReleaseName. selectChartHooks resorts
// entries by semver, so authoring order is informative rather than load-bearing.
var ChartHooks = map[string][]ChartHook{
	"keda-add-ons-http": {
		{
			TargetVersion: "0.12.2",
			PreUpgrade:    deleteKedaAddOnsHttpDeployments,
		},
	},
}

// selectChartHooks returns the hooks for releaseName whose TargetVersion is
// strictly greater than installedVersion and less than or equal to
// targetVersion, sorted ascending by TargetVersion. An empty installedVersion
// signals a fresh install and yields no hooks.
func selectChartHooks(releaseName, installedVersion, targetVersion string) ([]ChartHook, error) {
	if installedVersion == "" {
		return nil, nil
	}

	hooks, ok := ChartHooks[releaseName]
	if !ok || len(hooks) == 0 {
		return nil, nil
	}

	installed, err := semver.NewVersion(installedVersion)
	if err != nil {
		return nil, fmt.Errorf("Invalid installed version %q for chart %s: %w", installedVersion, releaseName, err)
	}

	target, err := semver.NewVersion(targetVersion)
	if err != nil {
		return nil, fmt.Errorf("Invalid target version %q for chart %s: %w", targetVersion, releaseName, err)
	}

	type indexedHook struct {
		version *semver.Version
		hook    ChartHook
	}

	selected := []indexedHook{}
	for _, hook := range hooks {
		hookVersion, err := semver.NewVersion(hook.TargetVersion)
		if err != nil {
			return nil, fmt.Errorf("Invalid hook TargetVersion %q for chart %s: %w", hook.TargetVersion, releaseName, err)
		}

		if hookVersion.Compare(installed) <= 0 {
			continue
		}
		if hookVersion.Compare(target) > 0 {
			continue
		}

		selected = append(selected, indexedHook{version: hookVersion, hook: hook})
	}

	sort.Slice(selected, func(i, j int) bool {
		return selected[i].version.LessThan(selected[j].version)
	})

	out := make([]ChartHook, 0, len(selected))
	for _, entry := range selected {
		out = append(out, entry.hook)
	}
	return out, nil
}

// deleteKedaAddOnsHttpDeployments removes the deployments managed by the
// keda-add-ons-http helm release so the chart can recreate them with the
// updated selectors introduced in 0.12.x. Selectors on Deployments are
// immutable, so a straight helm upgrade fails without this step.
func deleteKedaAddOnsHttpDeployments(ctx context.Context, clientset KubernetesClient, chart HelmChart, installedRevision Release) error {
	return clientset.DeleteDeployment(ctx, DeleteDeploymentInput{
		Namespace:     chart.Namespace,
		LabelSelector: "app.kubernetes.io/instance=keda-add-ons-http",
	})
}
