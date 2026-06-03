package scheduler_k3s

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
	"github.com/dokku/dokku/plugins/cron"
	nginxvhosts "github.com/dokku/dokku/plugins/nginx-vhosts"
	"github.com/dokku/dokku/plugins/registry"
	"github.com/fatih/color"
	"github.com/kballard/go-shellquote"
	"github.com/ryanuber/columnize"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubectl/pkg/util/term"
	"k8s.io/kubernetes/pkg/client/conditions"
)

// TriggerCorePostDeploy moves a configured kustomize root path to be in the app root dir
func TriggerCorePostDeploy(appName string) error {
	return common.CorePostDeploy(common.CorePostDeployInput{
		AppName:     appName,
		Destination: common.GetAppDataDirectory("scheduler-k3s", appName),
		PluginName:  "scheduler-k3s",
		ExtractedPaths: []common.CorePostDeployPath{
			{Path: "kustomization", IsDirectory: true},
		},
	})
}

// TriggerCorePostExtract moves a configured kustomize root path to be in the app root dir
func TriggerCorePostExtract(appName string, sourceWorkDir string) error {
	destination := common.GetAppDataDirectory("scheduler-k3s", appName)
	kustomizeRootPath := getComputedKustomizeRootPath(appName)
	return common.CorePostExtract(common.CorePostExtractInput{
		AppName:       appName,
		Destination:   destination,
		PluginName:    "scheduler-k3s",
		SourceWorkDir: sourceWorkDir,
		ToExtract: []common.CorePostExtractToExtract{
			{
				Path:        kustomizeRootPath,
				IsDirectory: true,
				Name:        "config/kustomize",
				Destination: "kustomization",
			},
		},
	})
}

// TriggerInstall runs the install step for the scheduler-k3s plugin
func TriggerInstall() error {
	if err := common.PropertySetup("scheduler-k3s"); err != nil {
		return fmt.Errorf("Unable to install the scheduler-k3s plugin: %s", err.Error())
	}

	if err := common.SetupAppData("scheduler-k3s"); err != nil {
		return err
	}

	if err := migrateChartPropertiesToMapFormat(); err != nil {
		return fmt.Errorf("Unable to migrate chart properties: %w", err)
	}

	if err := migrateAnnotationsLabelsToMapFormat(); err != nil {
		return fmt.Errorf("Unable to migrate annotations and labels: %w", err)
	}

	if err := syncExistingCertificates(); err != nil {
		common.LogWarn(fmt.Sprintf("Warning: failed to sync existing certificates: %v", err))
	}

	return nil
}

// migrateChartPropertiesToMapFormat moves legacy per-key chart.<chart>.<key>
// property files into a single chart-overrides.<chart> JSON map per chart.
// Without this migration, a dokku upgrade would silently strand any existing
// chart overrides because the new render path only reads chart-overrides.*.
// Idempotent: once the legacy files are gone the helper short-circuits per
// chart, and legacy values take precedence over any partial pre-existing
// entries on re-runs.
func migrateChartPropertiesToMapFormat() error {
	for _, chart := range HelmCharts {
		prefix := "chart." + chart.ReleaseName + "."
		legacy, err := common.PropertyGetAllByPrefix("scheduler-k3s", "--global", prefix)
		if err != nil {
			return fmt.Errorf("Unable to read legacy chart properties for %s: %w", chart.ReleaseName, err)
		}
		if len(legacy) == 0 {
			continue
		}

		mapProperty := "chart-overrides." + chart.ReleaseName
		merged, err := common.PropertyMapGet("scheduler-k3s", "--global", mapProperty)
		if err != nil {
			return fmt.Errorf("Unable to read existing chart-overrides for %s: %w", chart.ReleaseName, err)
		}
		for fullKey, value := range legacy {
			merged[strings.TrimPrefix(fullKey, prefix)] = value
		}

		if err := common.PropertyMapWrite("scheduler-k3s", "--global", mapProperty, merged); err != nil {
			return fmt.Errorf("Unable to write chart-overrides for %s: %w", chart.ReleaseName, err)
		}

		for fullKey := range legacy {
			if err := common.PropertyDelete("scheduler-k3s", "--global", fullKey); err != nil {
				return fmt.Errorf("Unable to remove legacy chart property %s: %w", fullKey, err)
			}
		}
	}
	return nil
}

// migrateAnnotationsLabelsToMapFormat converts the legacy line-formatted
// annotation and label property files ("key: value" per line) into JSON maps
// written via PropertyMapWrite. The property name (e.g. "web.deployment",
// "labels.global.service") is unchanged - only the file content layout shifts.
// Without this migration, post-upgrade reads via PropertyMapGet would fail to
// parse the legacy line format as JSON. Idempotent: any property whose content
// is already JSON (or empty) is skipped by probing PropertyMapGet first.
func migrateAnnotationsLabelsToMapFormat() error {
	annotationResources := map[string]bool{}
	for _, rt := range AnnotationResourceTypes {
		annotationResources[rt] = true
	}
	labelResources := map[string]bool{}
	for _, rt := range LabelResourceTypes {
		labelResources[rt] = true
	}

	scopes := []string{"--global"}
	apps, err := common.UnfilteredDokkuApps()
	if err != nil && !errors.Is(err, common.NoAppsExist) {
		return fmt.Errorf("Unable to list apps for annotation/label migration: %w", err)
	}
	scopes = append(scopes, apps...)

	for _, scope := range scopes {
		properties, err := common.PropertyGetAllByPrefix("scheduler-k3s", scope, "")
		if err != nil {
			return fmt.Errorf("Unable to list properties for %s: %w", scope, err)
		}

		for propertyName := range properties {
			isLabel := strings.HasPrefix(propertyName, "labels.")
			checkName := propertyName
			if isLabel {
				checkName = strings.TrimPrefix(propertyName, "labels.")
			} else if isReservedAnnotationProperty(propertyName) {
				continue
			}

			dot := strings.LastIndex(checkName, ".")
			if dot <= 0 || dot == len(checkName)-1 {
				continue
			}

			processType := checkName[:dot]
			resourceType := checkName[dot+1:]
			if processType == "" {
				continue
			}

			if isLabel {
				if !labelResources[resourceType] {
					continue
				}
			} else {
				if !annotationResources[resourceType] {
					continue
				}
			}

			if err := migrateAnnotationLabelProperty(scope, propertyName); err != nil {
				return err
			}
		}
	}
	return nil
}

// migrateAnnotationLabelProperty rewrites one legacy line-formatted property
// file as a JSON map. It is a no-op when the file is already valid JSON or
// empty (idempotent).
func migrateAnnotationLabelProperty(scope string, propertyName string) error {
	if _, err := common.PropertyMapGet("scheduler-k3s", scope, propertyName); err == nil {
		return nil
	}

	lines, err := common.PropertyListGet("scheduler-k3s", scope, propertyName)
	if err != nil {
		return fmt.Errorf("Unable to read legacy property %s/%s: %w", scope, propertyName, err)
	}

	m := map[string]string{}
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) != 2 {
			return fmt.Errorf("Invalid legacy entry in %s/%s: %s", scope, propertyName, line)
		}
		m[parts[0]] = parts[1]
	}

	if err := common.PropertyMapWrite("scheduler-k3s", scope, propertyName, m); err != nil {
		return fmt.Errorf("Unable to write %s/%s as map: %w", scope, propertyName, err)
	}
	return nil
}

// TriggerPostCertsUpdate handles post-certs-update trigger
func TriggerPostCertsUpdate(appName string) error {
	scheduler := common.PropertyGetDefault("scheduler", appName, "selected", "")
	globalScheduler := common.PropertyGetDefault("scheduler", "--global", "selected", "docker-local")
	if scheduler == "" {
		scheduler = globalScheduler
	}
	if scheduler != "k3s" {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	common.LogInfo1(fmt.Sprintf("Syncing TLS certificate for %s to kubernetes", appName))
	if err := CreateOrUpdateTLSSecret(ctx, appName); err != nil {
		return err
	}

	if !isAppDeployed(appName) {
		return nil
	}

	imageTag, err := getDeployedAppImageTag(appName)
	if err != nil {
		return nil
	}

	common.LogInfo1(fmt.Sprintf("Triggering redeploy for %s to update ingress configuration", appName))
	return TriggerSchedulerDeploy("k3s", appName, imageTag)
}

// TriggerPostCertsRemove handles post-certs-remove trigger
func TriggerPostCertsRemove(appName string) error {
	scheduler := common.PropertyGetDefault("scheduler", appName, "selected", "")
	globalScheduler := common.PropertyGetDefault("scheduler", "--global", "selected", "docker-local")
	if scheduler == "" {
		scheduler = globalScheduler
	}
	if scheduler != "k3s" {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deleted, err := DeleteTLSSecret(ctx, appName)
	if err != nil {
		return err
	}

	if !deleted {
		return nil
	}

	if !isAppDeployed(appName) {
		return nil
	}

	imageTag, err := getDeployedAppImageTag(appName)
	if err != nil {
		return nil
	}

	common.LogInfo1(fmt.Sprintf("Triggering redeploy for %s to update ingress configuration", appName))
	return TriggerSchedulerDeploy("k3s", appName, imageTag)
}

// TriggerPostAppCloneSetup creates new scheduler-k3s files
func TriggerPostAppCloneSetup(oldAppName string, newAppName string) error {
	err := common.PropertyClone("scheduler-k3s", oldAppName, newAppName)
	if err != nil {
		return err
	}

	return nil
}

// TriggerPostAppRenameSetup renames scheduler-k3s files
func TriggerPostAppRenameSetup(oldAppName string, newAppName string) error {
	if err := common.PropertyClone("scheduler-k3s", oldAppName, newAppName); err != nil {
		return err
	}

	if err := common.PropertyDestroy("scheduler-k3s", oldAppName); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if _, err := DeleteTLSSecret(ctx, oldAppName); err != nil {
		common.LogWarn(fmt.Sprintf("Error deleting TLS secret for old app name %s: %v", oldAppName, err))
	}

	if err := DeleteConfigSecret(ctx, oldAppName); err != nil {
		common.LogWarn(fmt.Sprintf("Error deleting config secret for old app name %s: %v", oldAppName, err))
	}

	if err := DeleteImagePullSecret(ctx, oldAppName); err != nil {
		common.LogWarn(fmt.Sprintf("Error deleting image pull secret for old app name %s: %v", oldAppName, err))
	}

	return nil
}

// TriggerPostCreate creates the scheduler-k3s data directory
func TriggerPostCreate(appName string) error {
	return common.CreateAppDataDirectory("scheduler-k3s", appName)
}

// TriggerPostDelete destroys the scheduler-k3s data for a given app container
func TriggerPostDelete(appName string) error {
	dataErr := common.RemoveAppDataDirectory("scheduler-k3s", appName)
	propertyErr := common.PropertyDestroy("scheduler-k3s", appName)

	if dataErr != nil {
		return dataErr
	}

	return propertyErr
}

// TriggerSchedulerAppStatus returns the status of an app on the scheduler
func TriggerSchedulerAppStatus(scheduler string, appName string) error {
	if scheduler != "k3s" {
		return nil
	}

	clientset, err := NewKubernetesClient()
	if err != nil {
		return fmt.Errorf("Error creating kubernetes client: %w", err)
	}

	if err := clientset.Ping(); err != nil {
		return fmt.Errorf("kubernetes api not available: %w", err)
	}

	namespace := getComputedNamespace(appName)
	deployments, err := clientset.ListDeployments(context.Background(), ListDeploymentsInput{
		Namespace:     namespace,
		LabelSelector: fmt.Sprintf("app.kubernetes.io/part-of=%s", appName),
	})
	if err != nil {
		return fmt.Errorf("Error listing pods: %w", err)
	}

	processCount := 0
	expectedProcesses := 0
	runningProcesses := 0
	for _, deployment := range deployments {
		processCount += int(*deployment.Spec.Replicas)
		expectedProcesses += int(*deployment.Spec.Replicas)
		runningProcesses += int(deployment.Status.AvailableReplicas)
	}

	running := "true"
	if expectedProcesses == 0 || runningProcesses == 0 {
		running = "false"
	} else if runningProcesses < expectedProcesses {
		running = "mixed"
	}

	fmt.Printf("%d %s", processCount, running)
	return nil
}

// TriggerSchedulerCronWrite writes out cron tasks for a given application
func TriggerSchedulerCronWrite(scheduler string, appName string) error {
	if scheduler != "k3s" {
		return nil
	}

	if appName == "" {
		return nil
	}

	cronTasks, err := cron.FetchCronTasks(cron.FetchCronTasksInput{AppName: appName})
	if err != nil {
		return fmt.Errorf("Error fetching cron tasks: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM)
	go func() {
		<-signals
		common.LogWarn(fmt.Sprintf("Deployment of %s has been cancelled", appName))
		cancel()
	}()

	namespace := getComputedNamespace(appName)

	clientset, err := NewKubernetesClient()
	if err != nil {
		return fmt.Errorf("Error creating kubernetes client: %w", err)
	}

	for _, cronTask := range cronTasks {
		labelSelector := []string{
			fmt.Sprintf("app.kubernetes.io/part-of=%s", appName),
			fmt.Sprintf("dokku.com/cron-hash=%s", cronIDLabelValue(cronTask.ID)),
		}

		if cronTask.Maintenance {
			err = clientset.SuspendCronJobs(ctx, SuspendCronJobsInput{
				Namespace:     namespace,
				LabelSelector: strings.Join(labelSelector, ","),
			})
			if err != nil {
				return fmt.Errorf("Error suspending cron jobs: %w", err)
			}
		} else {
			err = clientset.ResumeCronJobs(ctx, ResumeCronJobsInput{
				Namespace:     namespace,
				LabelSelector: strings.Join(labelSelector, ","),
			})
			if err != nil {
				return fmt.Errorf("Error resuming cron jobs: %w", err)
			}
		}
	}

	return nil
}

// TriggerSchedulerDeploy deploys an image tag for a given application
func TriggerSchedulerDeploy(scheduler string, appName string, imageTag string) error {
	if scheduler != "k3s" {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM)
	go func() {
		<-signals
		common.LogWarn(fmt.Sprintf("Deployment of %s has been cancelled", appName))
		cancel()
	}()

	namespace := getComputedNamespace(appName)
	if err := createKubernetesNamespace(ctx, namespace); err != nil {
		return fmt.Errorf("Error creating kubernetes namespace for deployment: %w", err)
	}

	chartResult, err := BuildAppChart(ctx, appName, imageTag, BuildOptions{})
	if err != nil {
		return err
	}
	defer os.RemoveAll(chartResult.ChartDir)

	env, err := config.LoadMergedAppEnv(appName)
	if err != nil {
		return fmt.Errorf("Error loading environment for deployment: %w", err)
	}

	globalAnnotations, err := getGlobalAnnotations(appName)
	if err != nil {
		return fmt.Errorf("Error getting global annotations: %w", err)
	}

	globalLabels, err := getGlobalLabel(appName)
	if err != nil {
		return fmt.Errorf("Error getting global labels: %w", err)
	}

	imagePullSecrets := getComputedImagePullSecrets(appName)
	dokkuManagedPullSecret := false
	var dokkuPullSecretBytes []byte
	if imagePullSecrets == "" {
		dockerConfigPath := filepath.Join(registry.GetComputedAppRegistryConfigDir(appName), "config.json")
		if fi, err := os.Stat(dockerConfigPath); err == nil && !fi.IsDir() {
			b, err := os.ReadFile(dockerConfigPath)
			if err != nil {
				return fmt.Errorf("Error reading docker config: %w", err)
			}

			imagePullSecrets = GetImagePullSecretName(appName)
			dokkuManagedPullSecret = true
			dokkuPullSecretBytes = b
		}
	}

	clientset, err := NewKubernetesClient()
	if err != nil {
		return fmt.Errorf("Error creating kubernetes client: %w", err)
	}

	helmAgent, err := NewHelmAgent(chartResult.Namespace, DeployLogPrinter)
	if err != nil {
		return fmt.Errorf("Error creating helm agent: %w", err)
	}

	ingresses, err := clientset.ListIngresses(ctx, ListIngressesInput{
		Namespace:     chartResult.Namespace,
		LabelSelector: fmt.Sprintf("app.kubernetes.io/instance=%s-web", appName),
	})
	if err != nil {
		return fmt.Errorf("Error listing ingresses: %w", err)
	}

	ingressesToDelete := []string{}

	for _, ingress := range ingresses {
		ingressIngressMethod := ingress.Annotations["dokku.com/ingress-method"]
		if ingressIngressMethod != "domains" {
			ingressesToDelete = append(ingressesToDelete, ingress.Name)
		}
	}

	if len(ingressesToDelete) > 0 {
		common.LogWarn("Manually removing non-matching ingress resources")
	}
	for _, ingressName := range ingressesToDelete {
		common.LogVerboseQuiet(fmt.Sprintf("Removing non-matching ingress resource: %s", ingressName))
		err := clientset.DeleteIngress(ctx, DeleteIngressInput{
			Name:      ingressName,
			Namespace: chartResult.Namespace,
		})

		if err != nil {
			return fmt.Errorf("Error deleting ingress: %w", err)
		}
	}

	if err := CreateOrUpdateConfigSecret(ctx, CreateOrUpdateConfigSecretInput{
		AppName:     appName,
		Env:         env.Map(),
		Annotations: globalAnnotations.SecretAnnotations,
		Labels:      globalLabels.SecretLabels,
	}); err != nil {
		return fmt.Errorf("Error syncing config secret: %w", err)
	}

	if dokkuManagedPullSecret {
		if err := CreateOrUpdateImagePullSecret(ctx, CreateOrUpdateImagePullSecretInput{
			AppName:          appName,
			DockerConfigJSON: dokkuPullSecretBytes,
			Annotations:      globalAnnotations.SecretAnnotations,
			Labels:           globalLabels.SecretLabels,
		}); err != nil {
			return fmt.Errorf("Error syncing image pull secret: %w", err)
		}
	} else {
		if err := DeleteImagePullSecret(ctx, appName); err != nil {
			return fmt.Errorf("Error removing stale image pull secret: %w", err)
		}
	}

	keepImagePullSecrets := []string{}
	if imagePullSecrets != "" {
		keepImagePullSecrets = []string{imagePullSecrets}
	}
	if err := pruneStaleImagePullSecretsFromDeployments(ctx, clientset, chartResult.Namespace, appName, keepImagePullSecrets); err != nil {
		return fmt.Errorf("Error pruning stale imagePullSecrets entries: %w", err)
	}

	common.LogInfo2(fmt.Sprintf("Installing %s", appName))
	err = helmAgent.InstallOrUpgradeChart(ctx, ChartInput{
		ChartPath:         chartResult.ChartPath,
		KustomizeRootPath: chartResult.KustomizeRootPath,
		Namespace:         chartResult.Namespace,
		ReleaseName:       chartResult.ReleaseName,
		RollbackOnFailure: chartResult.RollbackOnFailure,
		Timeout:           chartResult.Timeout,
		Wait:              true,
	})
	if err != nil {
		return err
	}

	common.LogInfo1("Running post-deploy")
	_, err = common.CallPlugnTrigger(common.PlugnTriggerInput{
		Args:        []string{appName, "", "", imageTag},
		StreamStdio: true,
		Trigger:     "core-post-deploy",
	})
	if err != nil {
		return fmt.Errorf("Error running core-post-deploy: %w", err)

	}
	_, err = common.CallPlugnTrigger(common.PlugnTriggerInput{
		Args:        []string{appName, "", "", imageTag},
		StreamStdio: true,
		Trigger:     "post-deploy",
	})
	if err != nil {
		return fmt.Errorf("Error running post-deploy: %w", err)
	}

	return nil
}

// TriggerSchedulerIsDeployed returns true if given app has a running container
func TriggerSchedulerIsDeployed(scheduler string, appName string) error {
	if scheduler != "k3s" {
		return nil
	}

	if isAppDeployed(appName) {
		return nil
	}

	return fmt.Errorf("App %s is not deployed", appName)
}

// TriggerSchedulerEnter enters a container for a given application
func TriggerSchedulerEnter(scheduler string, appName string, processType string, podName string, args []string) error {
	if scheduler != "k3s" {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM)
	go func() {
		<-signals
		cancel()
	}()

	clientset, err := NewKubernetesClient()
	if err != nil {
		return fmt.Errorf("Error creating kubernetes client: %w", err)
	}

	if err := clientset.Ping(); err != nil {
		return fmt.Errorf("kubernetes api not available: %w", err)
	}

	namespace := getComputedNamespace(appName)
	labelSelector := []string{fmt.Sprintf("app.kubernetes.io/part-of=%s", appName)}
	processIndex := 1
	if processType != "" {
		parts := strings.SplitN(processType, ".", 2)
		if len(parts) == 2 {
			processType = parts[0]
			processIndex, err = strconv.Atoi(parts[1])
			if err != nil {
				return fmt.Errorf("Error parsing process index: %w", err)
			}
		}
		labelSelector = append(labelSelector, fmt.Sprintf("app.kubernetes.io/name=%s", processType))
	}

	pods, err := clientset.ListPods(ctx, ListPodsInput{
		Namespace:     namespace,
		LabelSelector: strings.Join(labelSelector, ","),
	})
	if err != nil {
		return fmt.Errorf("Error listing pods: %w", err)
	}

	if len(pods) == 0 {
		return fmt.Errorf("No pods found for app %s", appName)
	}

	processIndex--
	if processIndex > len(pods) {
		return fmt.Errorf("Process index %d out of range for app %s", processIndex, appName)
	}

	var selectedPod corev1.Pod
	if podName != "" {
		for _, pod := range pods {
			if pod.Name == podName {
				selectedPod = pod
				break
			}
		}
	} else {
		selectedPod = pods[processIndex]
	}

	command := args
	if len(args) == 0 {
		command = []string{common.GetDokkuAppShell(appName)}
	}

	entrypoint := ""
	if selectedPod.Annotations["dokku.com/builder-type"] == "herokuish" {
		entrypoint = "/exec"
	}

	return enterPod(ctx, EnterPodInput{
		Clientset:   clientset,
		Command:     command,
		Entrypoint:  entrypoint,
		SelectedPod: selectedPod,
		WaitTimeout: 10,
	})
}

// TriggerSchedulerLogs displays logs for a given application
func TriggerSchedulerLogs(scheduler string, appName string, processType string, tail bool, quiet bool, numLines int64) error {
	if scheduler != "k3s" {
		return nil
	}

	clientset, err := NewKubernetesClient()
	if err != nil {
		return fmt.Errorf("Error creating kubernetes client: %w", err)
	}

	return clientset.StreamLogs(context.Background(), StreamLogsInput{
		Namespace:     getComputedNamespace(appName),
		ContainerName: processType,
		LabelSelector: []string{fmt.Sprintf("app.kubernetes.io/part-of=%s", appName)},
		TailLines:     numLines,
		Follow:        tail,
		Quiet:         quiet,
	})
}

// TriggerSchedulerProxyConfig displays nginx config for a given application
func TriggerSchedulerProxyConfig(scheduler string, appName string, proxyType string) error {
	if scheduler != "k3s" || proxyType != "k3s" {
		return nil
	}

	clientset, err := NewKubernetesClient()
	if err != nil {
		return fmt.Errorf("Error creating kubernetes client: %w", err)
	}

	pods, err := clientset.ListPods(context.Background(), ListPodsInput{
		Namespace:     "ingress-nginx",
		LabelSelector: "app.kubernetes.io/name=ingress-nginx",
	})
	if err != nil {
		return fmt.Errorf("Error listing pods: %w", err)
	}

	if len(pods) == 0 {
		return errors.New("No pods found for ingress-nginx")
	}

	var buf bytes.Buffer
	w := io.MultiWriter(&buf)

	command := []string{"cat", "/etc/nginx/nginx.conf"}
	err = clientset.ExecCommand(context.Background(), ExecCommandInput{
		Command:       command,
		ContainerName: "controller",
		Name:          pods[0].Name,
		Namespace:     "ingress-nginx",
		Stdout:        w,
	})
	if err != nil {
		return fmt.Errorf("Error reading nginx config: %w", err)
	}

	// split buffer string into lines
	lines := strings.Split(buf.String(), "\n")

	// get every domain block from the nginx config
	// domain blocks begin with "## start server DOMAIN_NAME" and end with "## end server DOMAIN_NAME"
	// each line also may have space characters at the beginning
	domainBlocks := map[string][]string{}
	for i, line := range lines {
		strippedLine := strings.TrimSpace(line)
		if strings.HasPrefix(strippedLine, "## start server ") {
			domainName := strings.TrimPrefix(strippedLine, "## start server ")
			domainBlocks[domainName] = []string{}
			for _, nextLine := range lines[i+1:] {
				nextStrippedLine := strings.TrimSpace(nextLine)
				if strings.HasPrefix(nextStrippedLine, "## end server "+domainName) {
					break
				}
				domainBlocks[domainName] = append(domainBlocks[domainName], nextLine)
			}
		}
	}

	// get all domains for this app
	domains := []string{}
	_, err = common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "domains-vhost-enabled",
		Args:        []string{appName},
		StreamStdio: true,
	})
	if err == nil {
		results, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
			Trigger: "domains-list",
			Args:    []string{appName},
		})
		if err != nil {
			return fmt.Errorf("Error getting domains for deployment: %w", err)
		}

		for _, domain := range strings.Split(results.StdoutContents(), "\n") {
			domain = strings.TrimSpace(domain)
			if domain != "" {
				domains = append(domains, domain)
			}
		}
	}

	for _, domain := range domains {
		if domainBlock, ok := domainBlocks[domain]; ok {
			fmt.Println(strings.Join(domainBlock, "\n"))
		}
	}

	return err
}

// TriggerSchedulerProxyLogs displays nginx logs for a given application
func TriggerSchedulerProxyLogs(scheduler string, appName string, proxyType string, logType string, tail bool, numLines int64) error {
	if scheduler != "k3s" || proxyType != "k3s" {
		return nil
	}

	clientset, err := NewKubernetesClient()
	if err != nil {
		return fmt.Errorf("Error creating kubernetes client: %w", err)
	}

	filename := ""
	if logType == "access" {
		filename = nginxvhosts.ComputedAccessLogPath(appName)
	} else if logType == "error" {
		filename = nginxvhosts.ComputedErrorLogPath(appName)
	} else {
		return errors.New("Invalid log type")
	}

	command := []string{"tail"}
	if tail {
		command = append(command, "-F")
	}
	if numLines > 0 {
		command = append(command, "-n")
		command = append(command, strconv.FormatInt(numLines, 10))
	}
	command = append(command, filename)

	pods, err := clientset.ListPods(context.Background(), ListPodsInput{
		Namespace:     "ingress-nginx",
		LabelSelector: "app.kubernetes.io/name=ingress-nginx",
	})
	if err != nil {
		return fmt.Errorf("Error listing pods: %w", err)
	}

	if len(pods) == 0 {
		return errors.New("No pods found for ingress-nginx")
	}

	return clientset.ExecCommand(context.Background(), ExecCommandInput{
		Command:       command,
		ContainerName: "controller",
		Name:          pods[0].Name,
		Namespace:     "ingress-nginx",
	})
}

// TriggerSchedulerRun runs a command in an ephemeral container
func TriggerSchedulerRun(scheduler string, appName string, envCount int, args []string) error {
	if scheduler != "k3s" {
		return nil
	}

	extraEnv := map[string]string{}
	if envCount > 0 {
		var envPairs []string
		envPairs, args = args[0:envCount], args[envCount:]
		for _, envPair := range envPairs {
			parts := strings.SplitN(envPair, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("Invalid environment variable pair: %s", envPair)
			}
			extraEnv[parts[0]] = parts[1]
		}
	}

	imageTag, err := common.GetRunningImageTag(appName, "")
	if err != nil {
		return fmt.Errorf("Error getting running image tag: %w", err)
	}
	image, err := common.GetDeployingAppImageName(appName, imageTag, "")
	if err != nil {
		return fmt.Errorf("Error getting deploying app image name: %w", err)
	}

	imageStage, err := common.DockerInspect(image, "{{ index .Config.Labels \"com.dokku.image-stage\" }}")
	if err != nil {
		return fmt.Errorf("Error getting image stage: %w", err)
	}
	if imageStage != "release" {
		common.LogWarn("Invalid image stage detected: expected 'release', got '$IMAGE_STAGE'")
		return fmt.Errorf("Successfully deploy your app to fix dokku run calls")
	}

	dokkuRmContainer := os.Getenv("DOKKU_RM_CONTAINER")
	if dokkuRmContainer == "" {
		resp, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
			Trigger: "config-get",
			Args:    []string{appName, "DOKKU_RM_CONTAINER"},
		})
		if err != nil {
			resp, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
				Trigger: "config-get-global",
				Args:    []string{"DOKKU_RM_CONTAINER"},
			})
			if err == nil {
				dokkuRmContainer = strings.TrimSpace(resp.Stdout)
			}
		} else {
			dokkuRmContainer = strings.TrimSpace(resp.Stdout)
		}
	}
	if dokkuRmContainer == "" {
		dokkuRmContainer = "true"
	}

	rmContainer, err := strconv.ParseBool(dokkuRmContainer)
	if err != nil {
		return fmt.Errorf("Error parsing DOKKU_RM_CONTAINER value as boolean: %w", err)
	}

	labels := map[string]string{
		"app.kubernetes.io/part-of": appName,
	}
	annotations := map[string]string{}

	if os.Getenv("DOKKU_TRACE") == "1" {
		extraEnv["TRACE"] = "true"
	}

	namespace := getComputedNamespace(appName)
	clientset, err := NewKubernetesClient()
	if err != nil {
		return fmt.Errorf("Error creating kubernetes client: %w", err)
	}

	if err := clientset.Ping(); err != nil {
		return fmt.Errorf("kubernetes api not available: %w", err)
	}

	processType := "run"
	if os.Getenv("DOKKU_CRON_ID") != "" {
		processType = "cron"
		cronID := os.Getenv("DOKKU_CRON_ID")
		cronHash := cronIDLabelValue(cronID)
		labels["dokku.com/cron-hash"] = cronHash
		annotations["dokku.com/cron-hash"] = cronHash
		annotations["dokku.com/cron-id"] = cronID
		concurrencyPolicy := strings.ToUpper(os.Getenv("DOKKU_CONCURRENCY_POLICY"))
		switch concurrencyPolicy {
		case "forbid":
			pods, err := clientset.ListPods(context.Background(), ListPodsInput{
				Namespace:     namespace,
				LabelSelector: fmt.Sprintf("dokku.com/cron-hash=%s", cronHash),
			})
			if err != nil {
				return fmt.Errorf("Error listing pods: %w", err)
			}
			if len(pods) > 0 {
				return fmt.Errorf("There is a running pod with the same dokku.com/cron-hash label")
			}
		case "replace":
			err := clientset.DeletePod(context.Background(), DeletePodInput{
				Namespace:     namespace,
				LabelSelector: fmt.Sprintf("dokku.com/cron-hash=%s", cronHash),
			})
			if err != nil {
				return fmt.Errorf("Error deleting pod: %w", err)
			}
		}
	}

	imageSourceType, err := common.DockerInspect(image, "{{ index .Config.Labels \"com.dokku.builder-type\" }}")
	if err != nil {
		return fmt.Errorf("Error getting image builder type: %w", err)
	}

	// todo: do something with docker args
	command := args
	commandShell := common.GetDokkuAppShell(appName)
	if len(args) == 0 {
		command = []string{commandShell}
	} else if len(args) == 1 {
		resp, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
			Trigger: "procfile-get-command",
			Args:    []string{appName, args[0], "5000"},
		})
		if err == nil && resp.StdoutContents() != "" {
			common.LogInfo1Quiet(fmt.Sprintf("Found '%s' in Procfile, running that command", args[0]))
			words, err := shellquote.Split(resp.StdoutContents())
			if err != nil {
				return fmt.Errorf("Error parsing Procfile command: %w", err)
			}
			command = words
		}
	}

	entrypoint := ""
	switch imageSourceType {
	case "herokuish":
		entrypoint = "/exec"
	case "pack":
		entrypoint = "launcher"
	}

	helmAgent, err := NewHelmAgent(namespace, DevNullPrinter)
	if err != nil {
		return fmt.Errorf("Error creating helm agent: %w", err)
	}

	values, err := helmAgent.GetValues(appName)
	if err != nil {
		return fmt.Errorf("Error getting helm values: %w", err)
	}

	globalValues, ok := values["global"].(map[string]interface{})
	if !ok {
		return errors.New("Global helm values not found")
	}

	deploymentIDValue, ok := globalValues["deployment_id"].(string)
	if !ok {
		return fmt.Errorf("Deployment ID is not a string")
	}
	if len(deploymentIDValue) == 0 {
		return fmt.Errorf("Deployment ID is empty")
	}
	deploymentID, err := strconv.ParseInt(deploymentIDValue, 10, 64)
	if err != nil {
		return fmt.Errorf("Error parsing deployment ID: %w", err)
	}

	attachToPod := os.Getenv("DOKKU_DETACH_CONTAINER") != "1"

	forceTTY := common.ToBool(os.Getenv("DOKKU_FORCE_TTY"))
	disableTTY := common.ToBool(os.Getenv("DOKKU_DISABLE_TTY"))
	ttyProbe := term.TTY{Out: os.Stdout}
	hasTTY := ttyProbe.MonitorSize(ttyProbe.GetSize()) != nil
	streamLogsOnly := attachToPod && !forceTTY && (disableTTY || !hasTTY)

	imagePullSecrets := getComputedImagePullSecrets(appName)
	if imagePullSecrets == "" {
		imagePullSecrets = GetImagePullSecretName(appName)
		_, err := clientset.GetSecret(context.Background(), GetSecretInput{
			Name:      imagePullSecrets,
			Namespace: namespace,
		})

		if err != nil {
			if _, ok := err.(*NotFoundError); !ok {
				return fmt.Errorf("Error getting image pull secret: %w", err)
			}
			imagePullSecrets = ""
		}
	}

	securityContext, err := getSecurityContext(appName, "run")
	if err != nil {
		return fmt.Errorf("Error getting security context: %w", err)
	}

	activeDeadlineSeconds := int64(86400) // 24 hours
	if os.Getenv("DOKKU_RUN_TTL_SECONDS") != "" {
		activeDeadlineSeconds, err = strconv.ParseInt(os.Getenv("DOKKU_RUN_TTL_SECONDS"), 10, 64)
		if err != nil {
			return fmt.Errorf("Error parsing DOKKU_RUN_TTL_SECONDS value as int64: %w", err)
		}
	}

	workingDir := common.GetWorkingDir(appName, image)
	job, err := templateKubernetesJob(Job{
		ActiveDeadlineSeconds: activeDeadlineSeconds,
		Annotations:           annotations,
		AppName:               appName,
		Command:               command,
		DeploymentID:          deploymentID,
		Entrypoint:            entrypoint,
		Env:                   extraEnv,
		Image:                 image,
		ImagePullSecrets:      imagePullSecrets,
		ImageSourceType:       imageSourceType,
		Interactive:           attachToPod || common.ToBool(os.Getenv("DOKKU_FORCE_TTY")),
		Labels:                labels,
		Namespace:             namespace,
		ProcessType:           processType,
		RemoveContainer:       rmContainer,
		SecurityContext:       securityContext,
		WorkingDir:            workingDir,
	})
	if err != nil {
		return fmt.Errorf("Error templating job: %w", err)
	}

	if os.Getenv("FORCE_TTY") == "1" {
		color.NoColor = false
	}

	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM)
	go func() {
		<-signals
		if attachToPod {
			clientset.DeleteJob(ctx, DeleteJobInput{ // nolint: errcheck
				Name:      job.Name,
				Namespace: namespace,
			})
		}
		cancel()
	}()

	createdJob, err := clientset.CreateJob(ctx, CreateJobInput{
		Job:       job,
		Namespace: namespace,
	})
	if err != nil {
		return fmt.Errorf("Error creating job: %w", err)
	}

	if attachToPod {
		defer func() {
			clientset.DeleteJob(ctx, DeleteJobInput{ // nolint: errcheck
				Name:      job.Name,
				Namespace: namespace,
			})
		}()
	}

	batchJobSelector := fmt.Sprintf("batch.kubernetes.io/job-name=%s", createdJob.Name)
	pods, err := waitForPodToExist(ctx, WaitForPodToExistInput{
		Clientset:     clientset,
		Namespace:     namespace,
		RetryCount:    3,
		LabelSelector: batchJobSelector,
	})
	if err != nil {
		return fmt.Errorf("Error waiting for pod to exist: %w", err)
	}
	if !attachToPod {
		fmt.Println(pods[0].Name)
		return nil
	}

	err = waitForPodBySelectorRunning(ctx, WaitForPodBySelectorRunningInput{
		Clientset:     clientset,
		Namespace:     namespace,
		LabelSelector: batchJobSelector,
		Timeout:       30,
		Waiter:        isPodReady,
	})
	if streamLogsOnly && (err == nil || errors.Is(err, conditions.ErrPodCompleted)) {
		if streamErr := clientset.StreamLogs(ctx, StreamLogsInput{
			Follow:        true,
			LabelSelector: []string{batchJobSelector},
			Namespace:     namespace,
			Quiet:         true,
		}); streamErr != nil {
			return fmt.Errorf("Error streaming logs: %w", streamErr)
		}

		selectedPod, waitErr := waitForPodBySelectorCompleted(ctx, WaitForPodBySelectorCompletedInput{
			Clientset:     clientset,
			Namespace:     namespace,
			LabelSelector: batchJobSelector,
			Timeout:       10,
		})
		if waitErr != nil {
			return fmt.Errorf("Pod did not reach terminal state after logs streamed: %w", waitErr)
		}
		switch selectedPod.Status.Phase {
		case v1.PodSucceeded:
			return nil
		case v1.PodFailed:
			for _, status := range selectedPod.Status.ContainerStatuses {
				if status.Name != fmt.Sprintf("%s-%s", appName, processType) {
					continue
				}
				if status.State.Terminated == nil {
					continue
				}
				return fmt.Errorf("Unable to attach as the pod has already exited with a failed exit code: %s", status.State.Terminated.Message)
			}
			return fmt.Errorf("Unable to attach as the pod has already exited with a failed exit code")
		}
		return fmt.Errorf("Unable to attach as the pod is in an unknown state: %s", selectedPod.Status.Phase)
	}
	if err != nil {
		if errors.Is(err, conditions.ErrPodCompleted) {
			pods, podErr := clientset.ListPods(ctx, ListPodsInput{
				Namespace:     namespace,
				LabelSelector: batchJobSelector,
			})
			if podErr != nil {
				return fmt.Errorf("Error completed pod: %w", err)
			}
			selectedPod := pods[0]

			err := clientset.StreamLogs(ctx, StreamLogsInput{
				ContainerName: processType,
				Follow:        false,
				LabelSelector: []string{batchJobSelector},
				Namespace:     namespace,
				Quiet:         true,
				SinceSeconds:  10,
			})
			if err != nil {
				return fmt.Errorf("Error streaming logs: %w", err)
			}
			if selectedPod.Status.Phase == v1.PodFailed {
				for _, status := range selectedPod.Status.ContainerStatuses {
					if status.Name != fmt.Sprintf("%s-%s", appName, processType) {
						continue
					}
					if status.State.Terminated == nil {
						continue
					}
					return fmt.Errorf("Unable to attach as the pod has already exited with a failed exit code: %s", status.State.Terminated.Message)
				}

				return fmt.Errorf("Unable to attach as the pod has already exited with a failed exit code")
			} else if selectedPod.Status.Phase == v1.PodSucceeded {
				return nil
			}
		}

		return fmt.Errorf("Error waiting for pod to be running: %w", err)
	}

	pods, err = clientset.ListPods(ctx, ListPodsInput{
		Namespace:     namespace,
		LabelSelector: batchJobSelector,
	})
	if err != nil {
		return fmt.Errorf("Error getting pod: %w", err)
	}
	selectedPod := pods[0]

	switch selectedPod.Status.Phase {
	case v1.PodFailed, v1.PodSucceeded:
		if selectedPod.Status.Phase == v1.PodFailed {
			for _, status := range selectedPod.Status.ContainerStatuses {
				if status.Name != fmt.Sprintf("%s-%s", appName, processType) {
					continue
				}
				if status.State.Terminated == nil {
					continue
				}
				return fmt.Errorf("Unable to attach as the pod has already exited with a failed exit code: %s", status.State.Terminated.Message)
			}

			return fmt.Errorf("Unable to attach as the pod has already exited with a failed exit code")
		} else if selectedPod.Status.Phase == v1.PodSucceeded {
			return errors.New("Unable to attach as the pod has already exited with a successful exit code")
		}
	case v1.PodRunning:
		return enterPod(ctx, EnterPodInput{
			AllowCompletion: true,
			Clientset:       clientset,
			Command:         command,
			Entrypoint:      entrypoint,
			SelectedPod:     selectedPod,
			WaitTimeout:     10,
		})
	default:
		return fmt.Errorf("Unable to attach as the pod is in an unknown state: %s", selectedPod.Status.Phase)
	}
	// todo: support scheduler-post-run

	return nil
}

// TriggerSchedulerRunList lists one-off run pods for a given application
func TriggerSchedulerRunList(scheduler string, appName string, format string) error {
	if scheduler != "k3s" {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM)
	go func() {
		<-signals
		cancel()
	}()

	clientset, err := NewKubernetesClient()
	if err != nil {
		return fmt.Errorf("Error creating kubernetes client: %w", err)
	}

	if err := clientset.Ping(); err != nil {
		return fmt.Errorf("kubernetes api not available: %w", err)
	}

	namespace := getComputedNamespace(appName)
	cronJobs, err := clientset.ListCronJobs(ctx, ListCronJobsInput{
		LabelSelector: fmt.Sprintf("app.kubernetes.io/part-of=%s", appName),
		Namespace:     namespace,
	})
	if err != nil {
		return fmt.Errorf("Error getting cron jobs: %w", err)
	}

	type CronJobTask struct {
		ID       string `json:"id"`
		AppName  string `json:"app"`
		Command  string `json:"command"`
		Schedule string `json:"schedule"`
	}

	data := []CronJobTask{}
	lines := []string{"ID | Schedule | Command"}
	for _, cronJob := range cronJobs {
		command := ""
		for _, container := range cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers {
			if container.Name == fmt.Sprintf("%s-cron", appName) {
				command = strings.Join(container.Args, " ")
			}
		}

		cronID, ok := cronJob.Annotations["dokku.com/cron-id"]
		if !ok {
			common.LogWarn(fmt.Sprintf("Cron job %s does not have a dokku.com/cron-id annotation", cronJob.Name))
			continue
		}

		lines = append(lines, fmt.Sprintf("%s | %s | %s", cronID, cronJob.Spec.Schedule, command))
		data = append(data, CronJobTask{
			ID:       cronID,
			AppName:  appName,
			Command:  command,
			Schedule: cronJob.Spec.Schedule,
		})
	}

	if format == "stdout" {
		result := columnize.SimpleFormat(lines)
		fmt.Println(result)
	} else {
		b, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("Error marshalling cron jobs: %w", err)
		}
		fmt.Println(string(b))
	}

	return nil
}

// TriggerSchedulerPostDelete destroys the scheduler-k3s data for a given app container
func TriggerSchedulerPostDelete(scheduler string, appName string) error {
	if scheduler != "k3s" {
		return nil
	}

	dataErr := common.RemoveAppDataDirectory("logs", appName)
	propertyErr := common.PropertyDestroy("logs", appName)

	if dataErr != nil {
		return dataErr
	}

	if propertyErr != nil {
		return propertyErr
	}

	if isK3sKubernetes() {
		if err := isK3sInstalled(); err != nil {
			common.LogWarn("k3s is not installed, skipping")
			return nil
		}
	}

	if err := isKubernetesAvailable(); err != nil {
		return fmt.Errorf("kubernetes api not available: %w", err)
	}

	namespace := getComputedNamespace(appName)
	helmAgent, err := NewHelmAgent(namespace, DeployLogPrinter)
	if err != nil {
		return fmt.Errorf("Error creating helm agent: %w", err)
	}

	err = helmAgent.UninstallChart(appName)
	if err != nil {
		return fmt.Errorf("Error uninstalling chart: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if _, err := DeleteTLSSecret(ctx, appName); err != nil {
		common.LogWarn(fmt.Sprintf("Error deleting TLS secret for %s: %v", appName, err))
	}

	if err := DeleteConfigSecret(ctx, appName); err != nil {
		common.LogWarn(fmt.Sprintf("Error deleting config secret for %s: %v", appName, err))
	}

	if err := DeleteImagePullSecret(ctx, appName); err != nil {
		common.LogWarn(fmt.Sprintf("Error deleting image pull secret for %s: %v", appName, err))
	}

	return nil
}

// TriggerSchedulerStop stops an application
func TriggerSchedulerStop(scheduler string, appName string) error {
	if scheduler != "k3s" {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM)
	go func() {
		<-signals
		cancel()
	}()

	clientset, err := NewKubernetesClient()
	if err != nil {
		if isK3sKubernetes() {
			if err := isK3sInstalled(); err != nil {
				common.LogWarn("k3s is not installed, skipping")
				return nil
			}
		}
		return fmt.Errorf("Error creating kubernetes client: %w", err)
	}

	if err := clientset.Ping(); err != nil {
		return fmt.Errorf("kubernetes api not available: %w", err)
	}

	namespace := getComputedNamespace(appName)
	deployments, err := clientset.ListDeployments(ctx, ListDeploymentsInput{
		Namespace:     namespace,
		LabelSelector: fmt.Sprintf("app.kubernetes.io/part-of=%s", appName),
	})
	if err != nil {
		return fmt.Errorf("Error listing deployments: %w", err)
	}

	for _, deployment := range deployments {
		processType, ok := deployment.Labels["app.kubernetes.io/name"]
		if !ok {
			return fmt.Errorf("Deployment %s does not have a process type label", deployment.Name)
		}
		common.LogVerboseQuiet(fmt.Sprintf("Stopping %s process", processType))
		err := clientset.ScaleDeployment(ctx, ScaleDeploymentInput{
			Name:      deployment.Name,
			Namespace: namespace,
			Replicas:  0,
		})
		if err != nil {
			return fmt.Errorf("Error updating deployment scale: %w", err)
		}
	}

	// get all cronjobs for the app
	err = clientset.SuspendCronJobs(ctx, SuspendCronJobsInput{
		Namespace:     namespace,
		LabelSelector: fmt.Sprintf("app.kubernetes.io/part-of=%s", appName),
	})
	if err != nil {
		return fmt.Errorf("Error suspending cron jobs: %w", err)
	}

	return nil
}
