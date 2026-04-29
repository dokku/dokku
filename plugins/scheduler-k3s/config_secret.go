package scheduler_k3s

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dokku/dokku/plugins/common"
)

// ConfigSecretValues contains the values for the config secret helm chart
type ConfigSecretValues struct {
	Global ConfigSecretGlobalValues `yaml:"global"`
}

// ConfigSecretGlobalValues contains the global values for the config secret chart
type ConfigSecretGlobalValues struct {
	Annotations map[string]string `yaml:"annotations,omitempty"`
	AppName     string            `yaml:"app_name"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Namespace   string            `yaml:"namespace"`
	Secrets     map[string]string `yaml:"secrets,omitempty"`
}

// GetConfigSecretReleaseName returns the helm release name for the config secret
func GetConfigSecretReleaseName(appName string) string {
	return fmt.Sprintf("config-%s", appName)
}

// GetConfigSecretName returns the kubernetes secret name for the config env
func GetConfigSecretName(appName string) string {
	return fmt.Sprintf("config-%s", appName)
}

// CreateOrUpdateConfigSecretInput contains the inputs to CreateOrUpdateConfigSecret
type CreateOrUpdateConfigSecretInput struct {
	AppName     string
	Env         map[string]string
	Annotations map[string]string
	Labels      map[string]string
}

// CreateOrUpdateConfigSecret creates or updates the config env secret helm chart for an app
func CreateOrUpdateConfigSecret(ctx context.Context, input CreateOrUpdateConfigSecretInput) error {
	appName := input.AppName
	if err := isKubernetesAvailable(); err != nil {
		common.LogDebug("kubernetes not available, skipping config secret creation")
		return nil
	}

	scheduler := common.PropertyGetDefault("scheduler", appName, "selected", "")
	globalScheduler := common.PropertyGetDefault("scheduler", "--global", "selected", "docker-local")
	if scheduler == "" {
		scheduler = globalScheduler
	}
	if scheduler != "k3s" {
		common.LogDebug("app does not use k3s scheduler, skipping config secret creation")
		return nil
	}

	chartDir, err := os.MkdirTemp("", "dokku-config-secret-chart-")
	if err != nil {
		return fmt.Errorf("error creating chart directory: %w", err)
	}
	defer os.RemoveAll(chartDir)

	if err := os.MkdirAll(filepath.Join(chartDir, "templates"), os.FileMode(0755)); err != nil {
		return fmt.Errorf("error creating chart templates directory: %w", err)
	}

	namespace := getComputedNamespace(appName)
	releaseName := GetConfigSecretReleaseName(appName)

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

	b, err := templates.ReadFile("templates/config-secret-chart/templates/config-secret.yaml")
	if err != nil {
		return fmt.Errorf("error reading config-secret template: %w", err)
	}

	filename := filepath.Join(chartDir, "templates", "config-secret.yaml")
	err = os.WriteFile(filename, b, os.FileMode(0644))
	if err != nil {
		return fmt.Errorf("error writing config-secret template: %w", err)
	}

	if os.Getenv("DOKKU_TRACE") == "1" {
		common.CatFile(filename)
	}

	encodedSecrets := map[string]string{}
	for key, value := range input.Env {
		encodedSecrets[key] = base64.StdEncoding.EncodeToString([]byte(value))
	}

	values := &ConfigSecretValues{
		Global: ConfigSecretGlobalValues{
			Annotations: input.Annotations,
			AppName:     appName,
			Labels:      input.Labels,
			Namespace:   namespace,
			Secrets:     encodedSecrets,
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

	common.LogVerboseQuiet(fmt.Sprintf("Installing config secret for %s", appName))
	err = helmAgent.InstallOrUpgradeChart(ctx, ChartInput{
		ChartPath:   chartPath,
		Namespace:   namespace,
		ReleaseName: releaseName,
		Wait:        false,
	})
	if err != nil {
		return fmt.Errorf("error installing config secret chart: %w", err)
	}

	return nil
}

// DeleteConfigSecret deletes the config env secret helm chart for an app
func DeleteConfigSecret(ctx context.Context, appName string) error {
	if err := isKubernetesAvailable(); err != nil {
		common.LogDebug("kubernetes not available, skipping config secret deletion")
		return nil
	}

	namespace := getComputedNamespace(appName)
	releaseName := GetConfigSecretReleaseName(appName)

	helmAgent, err := NewHelmAgent(namespace, DeployLogPrinter)
	if err != nil {
		return fmt.Errorf("error creating helm agent: %w", err)
	}

	exists, err := helmAgent.ChartExists(releaseName)
	if err != nil {
		return fmt.Errorf("error checking if config secret chart exists: %w", err)
	}

	if !exists {
		return nil
	}

	common.LogVerboseQuiet(fmt.Sprintf("Removing config secret for %s", appName))
	if err := helmAgent.UninstallChart(releaseName); err != nil {
		return fmt.Errorf("error uninstalling config secret chart: %w", err)
	}

	return nil
}
