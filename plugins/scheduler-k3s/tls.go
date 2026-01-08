package scheduler_k3s

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// TLSSecretValues contains the values for a TLS secret helm chart
type TLSSecretValues struct {
	Global TLSSecretGlobalValues `yaml:"global"`
}

// TLSSecretGlobalValues contains the global values for TLS secret
type TLSSecretGlobalValues struct {
	AppName      string `yaml:"app_name"`
	Namespace    string `yaml:"namespace"`
	CertChecksum string `yaml:"cert_checksum"`
	TLSCrt       string `yaml:"tls_crt"`
	TLSKey       string `yaml:"tls_key"`
}

// GetCertContent retrieves certificate content via the certs-get trigger
func GetCertContent(appName string, keyType string) (string, error) {
	result, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "certs-get",
		Args:    []string{appName, keyType},
	})
	if err != nil {
		return "", fmt.Errorf("failed to get cert content: %w", err)
	}
	if result.ExitCode != 0 {
		return "", fmt.Errorf("certs-get trigger returned non-zero exit code")
	}
	return result.StdoutContents(), nil
}

// CertsExist checks if certificates exist for an app
func CertsExist(appName string) bool {
	result, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "certs-exists",
		Args:    []string{appName},
	})
	if err != nil {
		return false
	}
	return strings.TrimSpace(result.StdoutContents()) == "true"
}

// ComputeCertChecksum computes sha224 of combined cert+key content
// SHA224 produces a 56-character hex string which fits within Kubernetes label limit of 63 chars
func ComputeCertChecksum(certContent, keyContent string) string {
	combined := certContent + keyContent
	hash := sha256.Sum224([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// GetTLSSecretReleaseName returns the helm release name for TLS secret
func GetTLSSecretReleaseName(appName string) string {
	return fmt.Sprintf("tls-%s", appName)
}

// CreateOrUpdateTLSSecret creates or updates a TLS secret helm chart for an app
func CreateOrUpdateTLSSecret(ctx context.Context, appName string) error {
	if err := isKubernetesAvailable(); err != nil {
		common.LogDebug("kubernetes not available, skipping TLS secret creation")
		return nil
	}

	scheduler := common.PropertyGetDefault("scheduler", appName, "selected", "")
	globalScheduler := common.PropertyGetDefault("scheduler", "--global", "selected", "docker-local")
	if scheduler == "" {
		scheduler = globalScheduler
	}
	if scheduler != "k3s" {
		common.LogDebug("app does not use k3s scheduler, skipping TLS secret creation")
		return nil
	}

	certContent, err := GetCertContent(appName, "crt")
	if err != nil {
		return fmt.Errorf("failed to get certificate: %w", err)
	}

	keyContent, err := GetCertContent(appName, "key")
	if err != nil {
		return fmt.Errorf("failed to get key: %w", err)
	}

	checksum := ComputeCertChecksum(certContent, keyContent)

	chartDir, err := os.MkdirTemp("", "dokku-tls-chart-")
	if err != nil {
		return fmt.Errorf("error creating chart directory: %w", err)
	}
	defer os.RemoveAll(chartDir)

	if err := os.MkdirAll(filepath.Join(chartDir, "templates"), os.FileMode(0755)); err != nil {
		return fmt.Errorf("error creating chart templates directory: %w", err)
	}

	namespace := getComputedNamespace(appName)
	releaseName := GetTLSSecretReleaseName(appName)

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

	b, err := templates.ReadFile("templates/tls-secret-chart/templates/tls-secret.yaml")
	if err != nil {
		return fmt.Errorf("error reading tls-secret template: %w", err)
	}

	filename := filepath.Join(chartDir, "templates", "tls-secret.yaml")
	err = os.WriteFile(filename, b, os.FileMode(0644))
	if err != nil {
		return fmt.Errorf("error writing tls-secret template: %w", err)
	}

	if os.Getenv("DOKKU_TRACE") == "1" {
		common.CatFile(filename)
	}

	values := &TLSSecretValues{
		Global: TLSSecretGlobalValues{
			AppName:      appName,
			Namespace:    namespace,
			CertChecksum: checksum,
			TLSCrt:       base64.StdEncoding.EncodeToString([]byte(certContent)),
			TLSKey:       base64.StdEncoding.EncodeToString([]byte(keyContent)),
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

	common.LogInfo1(fmt.Sprintf("Installing TLS certificate for %s", appName))
	err = helmAgent.InstallOrUpgradeChart(ctx, ChartInput{
		ChartPath:   chartPath,
		Namespace:   namespace,
		ReleaseName: releaseName,
		Wait:        true,
	})
	if err != nil {
		return fmt.Errorf("error installing TLS secret chart: %w", err)
	}

	if err := common.PropertyWrite("scheduler-k3s", appName, "tls-cert-imported", "true"); err != nil {
		return fmt.Errorf("error setting tls-cert-imported property: %w", err)
	}

	return nil
}

// DeleteTLSSecret deletes the TLS secret helm chart for an app
func DeleteTLSSecret(ctx context.Context, appName string) (bool, error) {
	if err := isKubernetesAvailable(); err != nil {
		common.LogDebug("kubernetes not available, skipping TLS secret deletion")
		return false, nil
	}

	namespace := getComputedNamespace(appName)
	releaseName := GetTLSSecretReleaseName(appName)

	helmAgent, err := NewHelmAgent(namespace, DeployLogPrinter)
	if err != nil {
		return false, fmt.Errorf("error creating helm agent: %w", err)
	}

	exists, err := helmAgent.ChartExists(releaseName)
	if err != nil {
		return false, fmt.Errorf("error checking if TLS secret chart exists: %w", err)
	}

	if !exists {
		if err := common.PropertyDelete("scheduler-k3s", appName, "tls-cert-imported"); err != nil {
			common.LogDebug(fmt.Sprintf("error clearing tls-cert-imported property: %v", err))
		}
		return false, nil
	}

	common.LogInfo1(fmt.Sprintf("Removing TLS certificate for %s", appName))
	err = helmAgent.UninstallChart(releaseName)
	if err != nil {
		return false, fmt.Errorf("error uninstalling TLS secret chart: %w", err)
	}

	if err := common.PropertyDelete("scheduler-k3s", appName, "tls-cert-imported"); err != nil {
		common.LogDebug(fmt.Sprintf("error clearing tls-cert-imported property: %v", err))
	}

	return true, nil
}

// HasImportedTLSCert checks if an app has an imported TLS certificate
func HasImportedTLSCert(appName string) bool {
	value := common.PropertyGetDefault("scheduler-k3s", appName, "tls-cert-imported", "")
	return value == "true"
}

// TLSSecretExists checks if the TLS secret helm release exists in k8s
func TLSSecretExists(ctx context.Context, appName string) (bool, error) {
	namespace := getComputedNamespace(appName)
	releaseName := GetTLSSecretReleaseName(appName)

	helmAgent, err := NewHelmAgent(namespace, DevNullPrinter)
	if err != nil {
		return false, fmt.Errorf("error creating helm agent: %w", err)
	}

	return helmAgent.ChartExists(releaseName)
}

// TLSSecretNeedsUpdate checks if the TLS secret needs updating by comparing checksums
func TLSSecretNeedsUpdate(ctx context.Context, appName string) (bool, error) {
	clientset, err := NewKubernetesClient()
	if err != nil {
		return false, fmt.Errorf("error creating kubernetes client: %w", err)
	}

	namespace := getComputedNamespace(appName)
	secretName := fmt.Sprintf("tls-%s", appName)

	secret, err := clientset.GetSecret(ctx, GetSecretInput{
		Name:      secretName,
		Namespace: namespace,
	})
	if err != nil {
		return true, nil
	}

	existingChecksum, ok := secret.Labels["dokku.com/cert-checksum"]
	if !ok {
		return true, nil
	}

	certContent, err := GetCertContent(appName, "crt")
	if err != nil {
		return false, err
	}

	keyContent, err := GetCertContent(appName, "key")
	if err != nil {
		return false, err
	}

	currentChecksum := ComputeCertChecksum(certContent, keyContent)
	return existingChecksum != currentChecksum, nil
}

// syncExistingCertificates syncs certificates for all apps using k3s scheduler
func syncExistingCertificates() error {
	if err := isKubernetesAvailable(); err != nil {
		common.LogDebug("kubernetes not available, skipping certificate sync")
		return nil
	}

	ctx := context.Background()

	apps, err := common.DokkuApps()
	if err != nil {
		return fmt.Errorf("failed to list apps: %w", err)
	}

	for _, appName := range apps {
		scheduler := common.PropertyGetDefault("scheduler", appName, "selected", "")
		globalScheduler := common.PropertyGetDefault("scheduler", "--global", "selected", "docker-local")
		if scheduler == "" {
			scheduler = globalScheduler
		}
		if scheduler != "k3s" {
			continue
		}

		if !CertsExist(appName) {
			continue
		}

		needsUpdate, err := TLSSecretNeedsUpdate(ctx, appName)
		if err != nil {
			common.LogDebug(fmt.Sprintf("Error checking TLS secret for %s: %v", appName, err))
			continue
		}

		if needsUpdate {
			common.LogInfo1(fmt.Sprintf("Syncing TLS certificate for %s", appName))
			if err := CreateOrUpdateTLSSecret(ctx, appName); err != nil {
				common.LogWarn(fmt.Sprintf("Failed to sync TLS certificate for %s: %v", appName, err))
			}
		}
	}

	return nil
}
