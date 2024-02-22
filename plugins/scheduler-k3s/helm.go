package scheduler_k3s

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/dokku/dokku/plugins/common"
	"github.com/gofrs/flock"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/repo"
	"helm.sh/helm/v3/pkg/storage/driver"
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
	RepoURL           string
	RollbackOnFailure bool
	Timeout           time.Duration
	Version           string
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

	kubeconfigPath := getKubeconfigPath()
	kubeContext := getKubeContext()
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

type AddRepositoryInput struct {
	Name string
	URL  string
}

func (h *HelmAgent) AddRepository(ctx context.Context, helmRepo AddRepositoryInput) error {
	settings := cli.New()
	repoFile := settings.RepositoryConfig

	err := os.MkdirAll(filepath.Dir(repoFile), os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("Error creating repository directory: %w", err)
	}

	fileLock := flock.New(strings.Replace(repoFile, filepath.Ext(repoFile), ".lock", 1))
	locked, err := fileLock.TryLockContext(ctx, time.Second)
	if err != nil {
		return fmt.Errorf("Error acquiring file lock: %w", err)
	}

	if !locked {
		return fmt.Errorf("Could not acquire file lock")
	}

	defer fileLock.Unlock() // nolint: errcheck

	b, err := os.ReadFile(repoFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("Error reading repository file: %w", err)
	}

	var f repo.File
	if err := yaml.Unmarshal(b, &f); err != nil {
		return fmt.Errorf("Error unmarshalling yaml: %w", err)
	}

	if f.Has(helmRepo.Name) {
		return nil
	}

	c := repo.Entry{
		Name: helmRepo.Name,
		URL:  helmRepo.URL,
	}

	r, err := repo.NewChartRepository(&c, getter.All(settings))
	if err != nil {
		return fmt.Errorf("Error creating chart repository: %w", err)
	}

	if _, err := r.DownloadIndexFile(); err != nil {
		return fmt.Errorf("Specified repository '%q' is not a valid chart repository or cannot be reached: %w", helmRepo.URL, err)
	}

	f.Update(&c)

	if err := f.WriteFile(repoFile, 0644); err != nil {
		log.Fatal(err)
	}

	return nil
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

	client := action.NewInstall(h.Configuration)
	client.Atomic = false
	client.ChartPathOptions = action.ChartPathOptions{}
	client.CreateNamespace = true
	client.DryRun = false
	client.Namespace = namespace
	client.ReleaseName = input.ReleaseName
	client.Timeout = input.Timeout
	client.Wait = false

	settings := cli.New()
	if input.RepoURL != "" {
		client.ChartPathOptions.RepoURL = input.RepoURL
	}
	if input.Version != "" {
		client.ChartPathOptions.Version = input.Version
	}

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

func (h *HelmAgent) ListRevisions(ctx context.Context, releaseName string) ([]Release, error) {
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

	client := action.NewUpgrade(h.Configuration)
	client.Atomic = input.RollbackOnFailure
	client.ChartPathOptions = action.ChartPathOptions{}
	client.CleanupOnFail = true
	client.DryRun = false
	client.MaxHistory = 10
	client.Namespace = namespace
	client.Timeout = input.Timeout
	client.Wait = false
	if input.RepoURL != "" {
		client.RepoURL = input.RepoURL
	}

	chart, err := client.ChartPathOptions.LocateChart(input.ChartPath, nil)
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
