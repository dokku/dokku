package scheduler_k3s

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	storage "github.com/dokku/dokku/plugins/storage"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StorageChartValues are the helm chart inputs for the per-entry storage chart.
type StorageChartValues struct {
	Global StorageChartGlobal `yaml:"global"`
}

// StorageChartGlobal contains global values for the storage chart.
type StorageChartGlobal struct {
	EntryName     string            `yaml:"entry_name"`
	Namespace     string            `yaml:"namespace"`
	Size          string            `yaml:"size"`
	AccessMode    string            `yaml:"access_mode,omitempty"`
	StorageClass  string            `yaml:"storage_class,omitempty"`
	HostPath      string            `yaml:"host_path,omitempty"`
	ReclaimPolicy string            `yaml:"reclaim_policy,omitempty"`
	Annotations   map[string]string `yaml:"annotations,omitempty"`
	Labels        map[string]string `yaml:"labels,omitempty"`
}

// GetStorageReleaseName returns the helm release name for a storage entry.
func GetStorageReleaseName(entryName string) string {
	return fmt.Sprintf("storage-%s", entryName)
}

// readEntryFromStdin decodes a storage.Entry from stdin. The storage
// plugin pipes the JSON-encoded entry to scheduler triggers.
func readEntryFromStdin() (*storage.Entry, error) {
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("unable to read storage entry from stdin: %w", err)
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("no storage entry payload received on stdin")
	}
	entry := &storage.Entry{}
	if err := json.Unmarshal(raw, entry); err != nil {
		return nil, fmt.Errorf("unable to parse storage entry payload: %w", err)
	}
	return entry, nil
}

// TriggerStorageCreate provisions or updates the PVC/PV for a k3s storage
// entry. Idempotent: re-runs against the same entry are a helm upgrade.
func TriggerStorageCreate(ctx context.Context, entryName string) error {
	if err := isKubernetesAvailable(); err != nil {
		return fmt.Errorf("kubernetes not available: %w", err)
	}

	entry, err := readEntryFromStdin()
	if err != nil {
		return err
	}
	if entry.Name != entryName {
		return fmt.Errorf("entry name mismatch: stdin=%q, arg=%q", entry.Name, entryName)
	}
	if entry.Scheduler != storage.SchedulerK3s {
		return nil
	}

	if err := validateStorageClass(ctx, entry.StorageClass); err != nil {
		return err
	}

	chartDir, err := os.MkdirTemp("", "dokku-storage-chart-")
	if err != nil {
		return fmt.Errorf("error creating chart directory: %w", err)
	}
	defer os.RemoveAll(chartDir)

	releaseName := GetStorageReleaseName(entry.Name)
	namespace := entry.Namespace
	if namespace == "" {
		namespace = "default"
	}

	if err := os.MkdirAll(filepath.Join(chartDir, "templates"), 0755); err != nil {
		return fmt.Errorf("error creating chart templates directory: %w", err)
	}

	chart := &Chart{
		ApiVersion: "v2",
		AppVersion: "1.0.0",
		Name:       releaseName,
		Icon:       "https://dokku.com/assets/dokku-logo.svg",
		Version:    "0.0.1",
	}
	if err := writeYaml(WriteYamlInput{Object: chart, Path: filepath.Join(chartDir, "Chart.yaml")}); err != nil {
		return fmt.Errorf("error writing chart: %w", err)
	}

	for _, name := range []string{"persistent-volume-claim.yaml", "persistent-volume.yaml"} {
		b, err := templates.ReadFile(filepath.Join("templates", "storage-chart", "templates", name))
		if err != nil {
			return fmt.Errorf("error reading %s: %w", name, err)
		}
		if err := os.WriteFile(filepath.Join(chartDir, "templates", name), b, 0644); err != nil {
			return fmt.Errorf("error writing %s: %w", name, err)
		}
	}

	values := &StorageChartValues{
		Global: StorageChartGlobal{
			EntryName:     entry.Name,
			Namespace:     namespace,
			Size:          entry.Size,
			AccessMode:    entry.AccessMode,
			StorageClass:  entry.StorageClass,
			HostPath:      entry.HostPath,
			ReclaimPolicy: entry.ReclaimPolicy,
			Annotations:   entry.Annotations,
			Labels:        entry.Labels,
		},
	}
	if err := writeYaml(WriteYamlInput{Object: values, Path: filepath.Join(chartDir, "values.yaml")}); err != nil {
		return fmt.Errorf("error writing values: %w", err)
	}

	if err := createKubernetesNamespace(ctx, namespace); err != nil {
		return fmt.Errorf("error creating namespace: %w", err)
	}

	helmAgent, err := NewHelmAgent(namespace, DeployLogPrinter)
	if err != nil {
		return fmt.Errorf("error creating helm agent: %w", err)
	}

	chartPath, err := filepath.Abs(chartDir)
	if err != nil {
		return fmt.Errorf("error getting chart path: %w", err)
	}

	common.LogVerboseQuiet(fmt.Sprintf("Installing storage chart for %s", entry.Name))
	if err := helmAgent.InstallOrUpgradeChart(ctx, ChartInput{
		ChartPath:   chartPath,
		Namespace:   namespace,
		ReleaseName: releaseName,
		Wait:        false,
	}); err != nil {
		return fmt.Errorf("error installing storage chart: %w", err)
	}
	return nil
}

// TriggerStorageDestroy removes the helm release backing a storage entry.
func TriggerStorageDestroy(ctx context.Context, entryName string) error {
	if err := isKubernetesAvailable(); err != nil {
		return fmt.Errorf("kubernetes not available: %w", err)
	}

	entry, err := readEntryFromStdin()
	if err != nil {
		return err
	}
	if entry.Scheduler != storage.SchedulerK3s {
		return nil
	}

	namespace := entry.Namespace
	if namespace == "" {
		namespace = "default"
	}
	releaseName := GetStorageReleaseName(entryName)

	helmAgent, err := NewHelmAgent(namespace, DeployLogPrinter)
	if err != nil {
		return fmt.Errorf("error creating helm agent: %w", err)
	}
	exists, err := helmAgent.ChartExists(releaseName)
	if err != nil {
		return fmt.Errorf("error checking if storage chart exists: %w", err)
	}
	if !exists {
		return nil
	}

	common.LogVerboseQuiet(fmt.Sprintf("Removing storage chart for %s", entryName))
	if err := helmAgent.UninstallChart(releaseName); err != nil {
		return fmt.Errorf("error uninstalling storage chart: %w", err)
	}
	return nil
}

// TriggerStorageStatus prints "Bound" if the entry's PVC is bound,
// "Pending" or "Lost" otherwise. Used by the storage:wait command.
func TriggerStorageStatus(ctx context.Context, entryName string) error {
	if err := isKubernetesAvailable(); err != nil {
		return fmt.Errorf("kubernetes not available: %w", err)
	}
	entry, err := readEntryFromStdin()
	if err != nil {
		return err
	}
	namespace := entry.Namespace
	if namespace == "" {
		namespace = "default"
	}

	clientset, err := NewKubernetesClient()
	if err != nil {
		return err
	}
	pvc, err := clientset.Client.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, entryName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error fetching PVC %s/%s: %w", namespace, entryName, err)
	}
	fmt.Println(string(pvc.Status.Phase))
	return nil
}

// validateStorageClass surfaces a clear error when the requested storage
// class doesn't exist in the cluster.
func validateStorageClass(ctx context.Context, name string) error {
	if name == "" {
		return nil
	}
	clientset, err := NewKubernetesClient()
	if err != nil {
		return err
	}
	classes, err := clientset.Client.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error listing storage classes: %w", err)
	}
	available := []string{}
	for _, cls := range classes.Items {
		if cls.Name == name {
			return nil
		}
		available = append(available, cls.Name)
	}
	return fmt.Errorf("storage class %q does not exist; available: %s", name, strings.Join(available, ", "))
}

// AppMountPair mirrors the storage plugin's wire format so we can decode
// the storage-app-mounts trigger output here.
type AppMountPair struct {
	Entry      *storage.Entry      `json:"entry"`
	Attachment *storage.Attachment `json:"attachment"`
}

// LoadAppMounts asks the storage plugin for an app's mount pairs in the
// given phase. Returns an empty slice when there are none.
func LoadAppMounts(appName string, phase string) ([]AppMountPair, error) {
	results, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "storage-app-mounts",
		Args:    []string{appName, phase},
	})
	if err != nil {
		return nil, err
	}
	output := strings.TrimSpace(results.StdoutContents())
	if output == "" {
		return nil, nil
	}
	pairs := []AppMountPair{}
	if err := json.Unmarshal([]byte(output), &pairs); err != nil {
		return nil, fmt.Errorf("unable to parse storage-app-mounts output: %w", err)
	}
	return pairs, nil
}

// ToProcessVolumes converts each AppMountPair into a ProcessVolume. K3s
// app deployments reference the PVC by name; the PVC itself is owned by
// the storage entry's separate helm release. Any docker-local entries
// found here are an error - they cannot be mounted on a k3s app.
func ToProcessVolumes(pairs []AppMountPair) ([]ProcessVolume, error) {
	volumes := []ProcessVolume{}
	for _, pair := range pairs {
		if pair.Entry == nil || pair.Attachment == nil {
			continue
		}
		if pair.Entry.Scheduler == storage.SchedulerDockerLocal {
			return nil, fmt.Errorf("storage entry %q is scheduler=docker-local but is mounted on a k3s app; recreate it with --scheduler k3s", pair.Entry.Name)
		}
		volumes = append(volumes, ProcessVolume{
			Name:      pair.Entry.Name,
			MountPath: pair.Attachment.ContainerPath,
			SubPath:   pair.Attachment.Subpath,
			ReadOnly:  pair.Attachment.Readonly,
			PersistentClaim: &ProcessVolumePersistentClaim{
				ClaimName: pair.Entry.Name,
			},
		})
	}
	return volumes, nil
}

// asK8sVolumeMount is a small helper used by tests / other callers that
// want a corev1.VolumeMount from a ProcessVolume.
func asK8sVolumeMount(v ProcessVolume) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      v.Name,
		MountPath: v.MountPath,
		SubPath:   v.SubPath,
		ReadOnly:  v.ReadOnly,
	}
}
