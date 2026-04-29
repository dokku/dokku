package scheduler_k3s

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dokku/dokku/plugins/common"
)

// ImagePullSecretValues contains the values for the dokku-managed image pull secret helm chart
type ImagePullSecretValues struct {
	Global ImagePullSecretGlobalValues `yaml:"global"`
}

// ImagePullSecretGlobalValues contains the global values for the image pull secret chart
type ImagePullSecretGlobalValues struct {
	Annotations      map[string]string `yaml:"annotations,omitempty"`
	AppName          string            `yaml:"app_name"`
	Labels           map[string]string `yaml:"labels,omitempty"`
	Namespace        string            `yaml:"namespace"`
	PullSecretBase64 string            `yaml:"pull_secret_base64"`
}

// GetImagePullSecretReleaseName returns the helm release name for the image pull secret
func GetImagePullSecretReleaseName(appName string) string {
	return fmt.Sprintf("pull-secret-%s", appName)
}

// GetImagePullSecretName returns the kubernetes secret name for the dokku-managed image pull secret
func GetImagePullSecretName(appName string) string {
	return fmt.Sprintf("pull-secret-%s", appName)
}

// CreateOrUpdateImagePullSecretInput contains the inputs to CreateOrUpdateImagePullSecret
type CreateOrUpdateImagePullSecretInput struct {
	AppName          string
	DockerConfigJSON []byte
	Annotations      map[string]string
	Labels           map[string]string
}

// CreateOrUpdateImagePullSecret creates or updates the image pull secret helm chart for an app
func CreateOrUpdateImagePullSecret(ctx context.Context, input CreateOrUpdateImagePullSecretInput) error {
	appName := input.AppName
	if err := isKubernetesAvailable(); err != nil {
		common.LogDebug("kubernetes not available, skipping image pull secret creation")
		return nil
	}

	scheduler := common.PropertyGetDefault("scheduler", appName, "selected", "")
	globalScheduler := common.PropertyGetDefault("scheduler", "--global", "selected", "docker-local")
	if scheduler == "" {
		scheduler = globalScheduler
	}
	if scheduler != "k3s" {
		common.LogDebug("app does not use k3s scheduler, skipping image pull secret creation")
		return nil
	}

	chartDir, err := os.MkdirTemp("", "dokku-pull-secret-chart-")
	if err != nil {
		return fmt.Errorf("error creating chart directory: %w", err)
	}
	defer os.RemoveAll(chartDir)

	if err := os.MkdirAll(filepath.Join(chartDir, "templates"), os.FileMode(0755)); err != nil {
		return fmt.Errorf("error creating chart templates directory: %w", err)
	}

	namespace := getComputedNamespace(appName)
	releaseName := GetImagePullSecretReleaseName(appName)

	chart := &Chart{
		ApiVersion: "v2",
		AppVersion: "1.0.0",
		Name:       releaseName,
		Icon:       "https://dokku.com/assets/dokku-logo.svg",
		Version:    "0.0.1",
	}

	err = writeYaml(WriteYamlInput{
		Object: chart,
		Path:   filepath.Join(chartDir, "Chart.yaml"),
	})
	if err != nil {
		return fmt.Errorf("error writing chart: %w", err)
	}

	b, err := templates.ReadFile("templates/pull-secret-chart/templates/pull-secret.yaml")
	if err != nil {
		return fmt.Errorf("error reading pull-secret template: %w", err)
	}

	filename := filepath.Join(chartDir, "templates", "pull-secret.yaml")
	err = os.WriteFile(filename, b, os.FileMode(0644))
	if err != nil {
		return fmt.Errorf("error writing pull-secret template: %w", err)
	}

	if os.Getenv("DOKKU_TRACE") == "1" {
		common.CatFile(filename)
	}

	values := &ImagePullSecretValues{
		Global: ImagePullSecretGlobalValues{
			Annotations:      input.Annotations,
			AppName:          appName,
			Labels:           input.Labels,
			Namespace:        namespace,
			PullSecretBase64: base64.StdEncoding.EncodeToString(input.DockerConfigJSON),
		},
	}

	err = writeYaml(WriteYamlInput{
		Object: values,
		Path:   filepath.Join(chartDir, "values.yaml"),
	})
	if err != nil {
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

	common.LogVerboseQuiet(fmt.Sprintf("Installing image pull secret for %s", appName))
	err = helmAgent.InstallOrUpgradeChart(ctx, ChartInput{
		ChartPath:   chartPath,
		Namespace:   namespace,
		ReleaseName: releaseName,
		Wait:        false,
	})
	if err != nil {
		return fmt.Errorf("error installing image pull secret chart: %w", err)
	}

	return nil
}

// DeleteImagePullSecret deletes the image pull secret helm chart for an app
func DeleteImagePullSecret(ctx context.Context, appName string) error {
	if err := isKubernetesAvailable(); err != nil {
		common.LogDebug("kubernetes not available, skipping image pull secret deletion")
		return nil
	}

	namespace := getComputedNamespace(appName)
	releaseName := GetImagePullSecretReleaseName(appName)

	helmAgent, err := NewHelmAgent(namespace, DeployLogPrinter)
	if err != nil {
		return fmt.Errorf("error creating helm agent: %w", err)
	}

	exists, err := helmAgent.ChartExists(releaseName)
	if err != nil {
		return fmt.Errorf("error checking if image pull secret chart exists: %w", err)
	}

	if !exists {
		return nil
	}

	common.LogVerboseQuiet(fmt.Sprintf("Removing image pull secret for %s", appName))
	if err := helmAgent.UninstallChart(releaseName); err != nil {
		return fmt.Errorf("error uninstalling image pull secret chart: %w", err)
	}

	return nil
}
