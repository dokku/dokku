package scheduler_k3s

import (
	"fmt"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// ReportSingleApp is an internal function that displays the scheduler-k3s report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	flags := map[string]common.ReportFunc{
		"--scheduler-k3s-computed-deploy-timeout":       reportComputedDeployTimeout,
		"--scheduler-k3s-deploy-timeout":                reportDeployTimeout,
		"--scheduler-k3s-global-deploy-timeout":         reportGlobalDeployTimeout,
		"--scheduler-k3s-computed-image-pull-secrets":   reportComputedImagePullSecrets,
		"--scheduler-k3s-image-pull-secrets":            reportImagePullSecrets,
		"--scheduler-k3s-global-image-pull-secrets":     reportGlobalImagePullSecrets,
		"--scheduler-k3s-global-kubeconfig-path":        reportGlobalKubeconfigPath,
		"--scheduler-k3s-global-kube-context":           reportGlobalKubeContext,
		"--scheduler-k3s-computed-letsencrypt-server":   reportComputedLetsencryptServer,
		"--scheduler-k3s-letsencrypt-server":            reportLetsencryptServer,
		"--scheduler-k3s-global-letsencrypt-server":     reportGlobalLetsencryptServer,
		"--scheduler-k3s-global-ingress-class":          reportGlobalIngressClass,
		"--scheduler-k3s-global-letsencrypt-email-prod": reportGlobalLetsencryptEmailProd,
		"--scheduler-k3s-global-letsencrypt-email-stag": reportGlobalLetsencryptEmailStag,
		"--scheduler-k3s-computed-namespace":            reportComputedNamespace,
		"--scheduler-k3s-namespace":                     reportNamespace,
		"--scheduler-k3s-global-namespace":              reportGlobalNamespace,
		"--scheduler-k3s-global-network-interface":      reportGlobalNetworkInterface,
		"--scheduler-k3s-computed-rollback-on-failure":  reportComputedRollbackOnFailure,
		"--scheduler-k3s-rollback-on-failure":           reportRollbackOnFailure,
		"--scheduler-k3s-global-rollback-on-failure":    reportGlobalRollbackOnFailure,
		"--scheduler-k3s-computed-shm-size":             reportComputedShmSize,
		"--scheduler-k3s-global-shm-size":               reportGlobalShmSize,
		"--scheduler-k3s-shm-size":                      reportShmSize,
	}

	chartProperties, err := common.PropertyGetAllByPrefix("scheduler-k3s", "--global", "chart.")
	if err != nil {
		return fmt.Errorf("Unable to get property list: %w", err)
	}
	for name, value := range chartProperties {
		flagName := "--scheduler-k3s-global-" + name
		flags[flagName] = func(appName string) string {
			return value
		}
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	trimPrefix := false
	uppercaseFirstCharacter := true
	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp("scheduler-k3s", appName, infoFlag, infoFlags, flagKeys, format, trimPrefix, uppercaseFirstCharacter)
}

// ReportAutoscalingAuthSingleApp is an internal function that displays the scheduler-k3s autoscaling-auth report for one app
func ReportAutoscalingAuthSingleApp(appName string, format string, includeMetadata bool) error {
	properties, err := common.PropertyGetAllByPrefix("scheduler-k3s", appName, TriggerAuthPropertyPrefix)
	if err != nil {
		return fmt.Errorf("Unable to get property list: %w", err)
	}

	infoFlags := map[string]string{}
	for key, value := range properties {
		metadataKey := strings.TrimPrefix(key, TriggerAuthPropertyPrefix)
		authType := strings.SplitN(metadataKey, ".", 2)[0]
		if authType == "" {
			continue
		}

		infoFlags[authType] = "configured"
		if includeMetadata {
			infoFlags[strings.ReplaceAll(metadataKey, ".", "-")] = value
		}
	}

	flagKeys := []string{}
	for flagKey := range infoFlags {
		flagKeys = append(flagKeys, flagKey)
	}

	trimPrefix := false
	uppercaseFirstCharacter := true
	return common.ReportSingleApp("scheduler-k3s", appName, "", infoFlags, flagKeys, format, trimPrefix, uppercaseFirstCharacter)
}

func reportComputedDeployTimeout(appName string) string {
	return getComputedDeployTimeout(appName)
}

func reportDeployTimeout(appName string) string {
	return getDeployTimeout(appName)
}

func reportGlobalDeployTimeout(appName string) string {
	return getGlobalDeployTimeout()
}

func reportComputedImagePullSecrets(appName string) string {
	return getComputedImagePullSecrets(appName)
}

func reportImagePullSecrets(appName string) string {
	return getImagePullSecrets(appName)
}

func reportGlobalImagePullSecrets(appName string) string {
	return getGlobalImagePullSecrets()
}

func reportGlobalIngressClass(appName string) string {
	return getGlobalIngressClass()
}

func reportGlobalKubeconfigPath(appName string) string {
	return getKubeconfigPath()
}

func reportGlobalKubeContext(appName string) string {
	return getKubeContext()
}
func reportComputedLetsencryptServer(appName string) string {
	return getComputedLetsencryptServer(appName)
}

func reportLetsencryptServer(appName string) string {
	return getLetsencryptServer(appName)
}

func reportGlobalLetsencryptServer(appName string) string {
	return getGlobalLetsencryptServer()
}

func reportGlobalLetsencryptEmailProd(appName string) string {
	return getGlobalLetsencryptEmailProd()
}

func reportGlobalLetsencryptEmailStag(appName string) string {
	return getGlobalLetsencryptEmailStag()
}

func reportComputedNamespace(appName string) string {
	return getComputedNamespace(appName)
}

func reportNamespace(appName string) string {
	return getNamespace(appName)
}

func reportGlobalNamespace(appName string) string {
	return getGlobalNamespace()
}

func reportGlobalNetworkInterface(appName string) string {
	return getGlobalNetworkInterface()
}

func reportComputedRollbackOnFailure(appName string) string {
	return getComputedRollbackOnFailure(appName)
}

func reportRollbackOnFailure(appName string) string {
	return getRollbackOnFailure(appName)
}

func reportGlobalRollbackOnFailure(appName string) string {
	return getGlobalRollbackOnFailure()
}

func reportComputedShmSize(appName string) string {
	return getComputedShmSize(appName)
}

func reportGlobalShmSize(appName string) string {
	return getGlobalShmSize()
}

func reportShmSize(appName string) string {
	return getShmSize(appName)
}
