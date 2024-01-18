package scheduler_k3s

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"
	"unicode"

	"github.com/dokku/dokku/plugins/common"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/storage/driver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

type ChartInput struct {
	ChartPath         string
	Namespace         string
	ReleaseName       string
	RollbackOnFailure bool
	Timeout           time.Duration
	Values            map[string]interface{}
}

type Release struct {
	Name      string
	Namespace string
	Version   int
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

	kubeConfig := kube.GetConfig(KubeConfigPath, "", namespace)
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

func (h *HelmAgent) DeleteRevision(releaseName string, revision int) error {
	clientset, err := NewKubernetesClient()
	if err != nil {
		return fmt.Errorf("Error creating kubernetes client: %w", err)
	}

	secretName := fmt.Sprintf("sh.helm.release.v1.%s.v%d", releaseName, revision)
	err = clientset.Client.CoreV1().Secrets(h.Namespace).Delete(context.Background(), secretName, metav1.DeleteOptions{})
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

func (h *HelmAgent) InstallOrUpgradeChart(input ChartInput) error {
	chartExists, err := h.ChartExists(input.ReleaseName)
	if err != nil {
		return fmt.Errorf("Error checking if chart exists: %w", err)
	}

	if chartExists {
		common.LogExclaim(fmt.Sprintf("Upgrading %s", input.ReleaseName))
		return h.UpgradeChart(input)
	}

	common.LogExclaim(fmt.Sprintf("Installing %s", input.ReleaseName))
	return h.InstallChart(input)
}

func (h *HelmAgent) InstallChart(input ChartInput) error {
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

	client := action.NewInstall(h.Configuration)
	client.Atomic = false
	client.ChartPathOptions = action.ChartPathOptions{}
	client.CreateNamespace = true
	client.DryRun = false
	client.Namespace = namespace
	client.ReleaseName = input.ReleaseName
	client.Timeout = input.Timeout
	client.Wait = false

	chart, err := client.ChartPathOptions.LocateChart(input.ChartPath, nil)
	if err != nil {
		return fmt.Errorf("Error locating chart: %w", err)
	}

	chartRequested, err := loader.Load(chart)
	if err != nil {
		return fmt.Errorf("Error loading chart: %w", err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	cSignal := make(chan os.Signal, 2)
	signal.Notify(cSignal, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-cSignal
		common.LogWarn(fmt.Sprintf("Deployment of %s has been cancelled", input.ReleaseName))
		cancel()
	}()

	_, err = client.RunWithContext(ctx, chartRequested, input.Values)
	if err != nil {
		return fmt.Errorf("Error deploying: %w", err)
	}

	return nil
}

func (h *HelmAgent) ListRevisions(releaseName string) ([]Release, error) {
	client := action.NewHistory(h.Configuration)
	response, err := client.Run(releaseName)
	if err != nil {
		return nil, fmt.Errorf("Error getting revisions: %w", err)
	}

	releases := []Release{}
	for _, release := range response {
		releases = append(releases, Release{
			Name:      release.Name,
			Namespace: release.Namespace,
			Version:   release.Version,
		})
	}

	sort.Slice(releases, func(i, j int) bool {
		return releases[i].Version < releases[j].Version
	})

	return releases, nil
}

func (h *HelmAgent) UpgradeChart(input ChartInput) error {
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

	client := action.NewUpgrade(h.Configuration)
	client.Atomic = input.RollbackOnFailure
	client.ChartPathOptions = action.ChartPathOptions{}
	client.CleanupOnFail = true
	client.DryRun = false
	client.MaxHistory = 10
	client.Namespace = namespace
	client.Timeout = input.Timeout
	client.Wait = false

	chart, err := client.ChartPathOptions.LocateChart(input.ChartPath, nil)
	if err != nil {
		return fmt.Errorf("Error locating chart: %w", err)
	}

	chartRequested, err := loader.Load(chart)
	if err != nil {
		return fmt.Errorf("Error loading chart: %w", err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	cSignal := make(chan os.Signal, 2)
	signal.Notify(cSignal, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-cSignal
		common.LogWarn(fmt.Sprintf("Deployment of %s has been cancelled", input.ReleaseName))
		cancel()
	}()

	_, err = client.RunWithContext(ctx, input.ReleaseName, chartRequested, input.Values)
	if err != nil {
		return fmt.Errorf("Error deploying: %w", err)
	}

	return nil
}

func (h *HelmAgent) UninstallChart(releaseName string) error {
	uninstall := action.NewUninstall(h.Configuration)
	_, err := uninstall.Run(releaseName)
	if err != nil {
		return fmt.Errorf("Error uninstalling chart: %w", err)
	}

	return nil
}
