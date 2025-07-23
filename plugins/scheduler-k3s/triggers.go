package scheduler_k3s

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	appjson "github.com/dokku/dokku/plugins/app-json"
	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
	"github.com/dokku/dokku/plugins/cron"
	nginxvhosts "github.com/dokku/dokku/plugins/nginx-vhosts"
	"github.com/fatih/color"
	"github.com/gosimple/slug"
	"github.com/kballard/go-shellquote"
	"github.com/ryanuber/columnize"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/client/conditions"
)

// TriggerCorePostDeploy moves a configured kustomize root path to be in the app root dir
func TriggerCorePostDeploy(appName string) error {
	existingKustomizeRootPath := getComputedKustomizeRootPath(appName)
	processSpecificKustomizeRootPath := fmt.Sprintf("%s.%s", existingKustomizeRootPath, os.Getenv("DOKKU_PID"))
	if common.DirectoryExists(processSpecificKustomizeRootPath) {
		if err := os.Rename(processSpecificKustomizeRootPath, existingKustomizeRootPath); err != nil {
			return err
		}
	} else if common.FileExists(fmt.Sprintf("%s.missing", processSpecificKustomizeRootPath)) {
		if err := os.RemoveAll(fmt.Sprintf("%s.missing", processSpecificKustomizeRootPath)); err != nil {
			return err
		}

		if common.DirectoryExists(existingKustomizeRootPath) {
			if err := os.RemoveAll(existingKustomizeRootPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// TriggerCorePostExtract moves a configured kustomize root path to be in the app root dir
func TriggerCorePostExtract(appName string, sourceWorkDir string) error {
	kustomizeRootPath := getComputedKustomizeRootPath(appName)
	if kustomizeRootPath == "" {
		return nil
	}

	directory := filepath.Join(common.MustGetEnv("DOKKU_LIB_ROOT"), "data", "scheduler-k3s", appName)
	existingKustomizeDirectory := filepath.Join(directory, "kustomization")
	files, err := filepath.Glob(fmt.Sprintf("%s.*", existingKustomizeDirectory))
	if err != nil {
		return err
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			return err
		}
	}

	processSpecificKustomizeRootPath := fmt.Sprintf("%s.%s", kustomizeRootPath, os.Getenv("DOKKU_PID"))
	results, _ := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "git-get-property",
		Args:    []string{appName, "source-image"},
	})
	appSourceImage := results.StdoutContents()

	if appSourceImage == "" {
		repoDefaultKustomizeRootPath := path.Join(sourceWorkDir, "kustomization")
		repoKustomizeRootPath := path.Join(sourceWorkDir, kustomizeRootPath)
		if !common.DirectoryExists(repoKustomizeRootPath) {
			if kustomizeRootPath != "kustomization" && common.DirectoryExists(repoDefaultKustomizeRootPath) {
				if err := os.RemoveAll(repoDefaultKustomizeRootPath); err != nil {
					return fmt.Errorf("Unable to remove existing kustomize directory: %s", err.Error())
				}
			}
			return common.TouchDir(fmt.Sprintf("%s.missing", processSpecificKustomizeRootPath))
		}

		if err := common.Copy(repoKustomizeRootPath, processSpecificKustomizeRootPath); err != nil {
			return fmt.Errorf("Unable to extract kustomize root path: %s", err.Error())
		}

		if kustomizeRootPath != "kustomization" {
			if err := common.Copy(repoKustomizeRootPath, repoDefaultKustomizeRootPath); err != nil {
				return fmt.Errorf("Unable to move kustomize root path into place: %s", err.Error())
			}
		}
	} else {
		if err := common.CopyDirFromImage(appName, appSourceImage, kustomizeRootPath, processSpecificKustomizeRootPath); err != nil {
			return common.TouchDir(fmt.Sprintf("%s.missing", processSpecificKustomizeRootPath))
		}
	}

	return nil
}

// TriggerInstall runs the install step for the scheduler-k3s plugin
func TriggerInstall() error {
	if err := common.PropertySetup("scheduler-k3s"); err != nil {
		return fmt.Errorf("Unable to install the scheduler-k3s plugin: %s", err.Error())
	}
	return nil
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

	return nil
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

// TriggerSchedulerDeploy deploys an image tag for a given application
func TriggerSchedulerDeploy(scheduler string, appName string, imageTag string) error {
	if scheduler != "k3s" {
		return nil
	}
	results, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "ps-current-scale",
		Args:    []string{appName},
	})
	if err != nil {
		return err
	}

	processes, err := common.ParseScaleOutput(results.StdoutBytes())
	if err != nil {
		return err
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

	image, err := common.GetDeployingAppImageName(appName, imageTag, "")
	if err != nil {
		return fmt.Errorf("Error getting deploying app image name: %w", err)
	}

	deployTimeout := getComputedDeployTimeout(appName)
	if _, err := strconv.Atoi(deployTimeout); err == nil {
		deployTimeout = fmt.Sprintf("%ss", deployTimeout)
	}

	deployRollback := getComputedRollbackOnFailure(appName)
	allowRollbacks, err := strconv.ParseBool(deployRollback)
	if err != nil {
		return fmt.Errorf("Error parsing rollback-on-failure value as boolean: %w", err)
	}

	imageSourceType := "dockerfile"
	if common.IsImageCnbBased(image) {
		imageSourceType = "pack"
	} else if common.IsImageHerokuishBased(image, appName) {
		imageSourceType = "herokuish"
	}

	env, err := config.LoadMergedAppEnv(appName)
	if err != nil {
		return fmt.Errorf("Error loading environment for deployment: %w", err)
	}

	issuerName := "letsencrypt-stag"
	server := getComputedLetsencryptServer(appName)
	if server == "prod" || server == "production" {
		issuerName = "letsencrypt-prod"
	} else if server != "stag" && server != "staging" {
		return fmt.Errorf("Invalid letsencrypt server config: %s", server)
	}

	tlsEnabled := false
	letsencryptEmailStag := getGlobalLetsencryptEmailStag()
	letsencryptEmailProd := getGlobalLetsencryptEmailProd()
	if issuerName == "letsencrypt-stag" {
		tlsEnabled = letsencryptEmailStag != ""
	}
	if issuerName == "letsencrypt-prod" {
		tlsEnabled = letsencryptEmailProd != ""
	}

	chartDir, err := os.MkdirTemp("", "dokku-chart-")
	if err != nil {
		return fmt.Errorf("Error creating chart directory: %w", err)
	}
	defer os.RemoveAll(chartDir)

	if err := os.MkdirAll(filepath.Join(chartDir, "templates"), os.FileMode(0755)); err != nil {
		return fmt.Errorf("Error creating chart templates directory: %w", err)
	}

	deploymentId := time.Now().Unix()
	pullSecretBase64 := base64.StdEncoding.EncodeToString([]byte(""))
	imagePullSecrets := getComputedImagePullSecrets(appName)
	if imagePullSecrets == "" {
		dockerConfigPath := filepath.Join(os.Getenv("DOKKU_ROOT"), ".docker/config.json")
		if fi, err := os.Stat(dockerConfigPath); err == nil && !fi.IsDir() {
			b, err := os.ReadFile(dockerConfigPath)
			if err != nil {
				return fmt.Errorf("Error reading docker config: %w", err)
			}

			imagePullSecrets = fmt.Sprintf("ims-%s.%d", appName, deploymentId)
			pullSecretBase64 = base64.StdEncoding.EncodeToString(b)
		}
	}

	globalTemplateFiles := []string{"service-account", "secret", "image-pull-secret"}
	for _, templateName := range globalTemplateFiles {
		b, err := templates.ReadFile(fmt.Sprintf("templates/chart/%s.yaml", templateName))
		if err != nil {
			return fmt.Errorf("Error reading %s template: %w", templateName, err)
		}

		filename := filepath.Join(chartDir, "templates", fmt.Sprintf("%s.yaml", templateName))
		err = os.WriteFile(filename, b, os.FileMode(0644))
		if err != nil {
			return fmt.Errorf("Error writing %s template: %w", templateName, err)
		}

		if os.Getenv("DOKKU_TRACE") == "1" {
			common.CatFile(filename)
		}
	}

	portMaps, err := getPortMaps(appName)
	if err != nil {
		return fmt.Errorf("Error getting port mappings for deployment: %w", err)
	}

	primaryPort := int32(5000)
	primaryServicePort := int32(80)
	for _, portMap := range portMaps {
		primaryPort = portMap.ContainerPort
		primaryServicePort = portMap.HostPort
		if primaryPort != 0 {
			break
		}
	}

	appJSON, err := appjson.GetAppJSON(appName)
	if err != nil {
		return fmt.Errorf("Error getting app.json for deployment: %w", err)
	}

	workingDir := common.GetWorkingDir(appName, image)

	allCronEntries, err := cron.FetchCronEntries(cron.FetchCronEntriesInput{AppName: appName})
	if err != nil {
		return fmt.Errorf("Error fetching cron entries: %w", err)
	}
	// remove maintenance cron entries
	cronEntries := []cron.TemplateCommand{}
	for _, cronEntry := range allCronEntries {
		if !cronEntry.Maintenance {
			cronEntries = append(cronEntries, cronEntry)
		}
	}

	domains := []string{}
	if _, ok := processes["web"]; ok {
		_, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
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
	}

	chart := &Chart{
		ApiVersion: "v2",
		AppVersion: "1.0.0",
		Name:       appName,
		Icon:       "https://dokku.com/assets/dokku-logo.svg",
		Version:    fmt.Sprintf("0.0.%d", deploymentId),
	}

	err = writeYaml(WriteYamlInput{
		Object: chart,
		Path:   filepath.Join(chartDir, "Chart.yaml"),
	})
	if err != nil {
		return fmt.Errorf("Error writing chart: %w", err)
	}

	globalAnnotations, err := getGlobalAnnotations(appName)
	if err != nil {
		return fmt.Errorf("Error getting global annotations: %w", err)
	}

	globalLabels, err := getGlobalLabel(appName)
	if err != nil {
		return fmt.Errorf("Error getting global labels: %w", err)
	}

	clientset, err := NewKubernetesClient()
	if err != nil {
		return fmt.Errorf("Error creating kubernetes client: %w", err)
	}

	if err := clientset.Ping(); err != nil {
		return fmt.Errorf("kubernetes api not available: %w", err)
	}

	kedaValues, err := getKedaValues(ctx, clientset, appName)
	if err != nil {
		return fmt.Errorf("Error getting keda values: %w", err)
	}

	values := &AppValues{
		Global: GlobalValues{
			Annotations:  globalAnnotations,
			AppName:      appName,
			DeploymentID: fmt.Sprint(deploymentId),
			Keda:         kedaValues,
			Image: GlobalImage{
				ImagePullSecrets: imagePullSecrets,
				PullSecretBase64: pullSecretBase64,
				Name:             image,
				Type:             imageSourceType,
				WorkingDir:       workingDir,
			},
			Labels:    globalLabels,
			Namespace: namespace,
			Network: GlobalNetwork{
				IngressClass:       getGlobalIngressClass(),
				PrimaryPort:        primaryPort,
				PrimaryServicePort: primaryServicePort,
			},
			Secrets: map[string]string{},
		},
		Processes: map[string]ProcessValues{},
	}

	if len(kedaValues.Authentications) > 0 {
		templateFiles := []string{"keda-secret", "keda-trigger-authentication"}
		for _, templateName := range templateFiles {
			b, err := templates.ReadFile(fmt.Sprintf("templates/chart/%s.yaml", templateName))
			if err != nil {
				return fmt.Errorf("Error reading %s template: %w", templateName, err)
			}

			filename := filepath.Join(chartDir, "templates", fmt.Sprintf("%s.yaml", templateName))
			err = os.WriteFile(filename, b, os.FileMode(0644))
			if err != nil {
				return fmt.Errorf("Error writing %s template: %w", templateName, err)
			}

			if os.Getenv("DOKKU_TRACE") == "1" {
				common.CatFile(filename)
			}
		}
	}

	processVolumes := []ProcessVolume{}
	if shmSize := getComputedShmSize(appName); shmSize != "" {
		processVolumes = append(processVolumes, ProcessVolume{
			Name:      "shmem",
			MountPath: "/dev/shm",
			EmptyDir: &ProcessVolumeEmptyDir{
				Medium:    "Memory",
				SizeLimit: shmSize,
			},
		})
	}

	for processType, processCount := range processes {
		// todo: implement deployment annotations
		// todo: implement pod annotations
		// todo: implement volumes

		healthchecks, ok := appJSON.Healthchecks[processType]
		if !ok {
			healthchecks = []appjson.Healthcheck{}
		}
		processHealthchecks := getProcessHealtchecks(healthchecks, primaryPort)

		startCommand, err := getStartCommand(StartCommandInput{
			AppName:         appName,
			ProcessType:     processType,
			ImageSourceType: imageSourceType,
			Port:            primaryPort,
			Env:             env.Map(),
		})
		if err != nil {
			return fmt.Errorf("Error getting start command for deployment: %w", err)
		}
		args := startCommand.Command

		processResources, err := getProcessResources(appName, processType)
		if err != nil {
			return fmt.Errorf("Error getting process resources: %w", err)
		}

		annotations, err := getAnnotations(appName, processType)
		if err != nil {
			return fmt.Errorf("Error getting process annotations: %w", err)
		}

		labels, err := getLabels(appName, processType)
		if err != nil {
			return fmt.Errorf("Error getting process labels: %w", err)
		}

		autoscaling, err := getAutoscaling(GetAutoscalingInput{
			AppName:     appName,
			ProcessType: processType,
			Replicas:    int(processCount),
			KedaValues:  kedaValues,
		})
		if err != nil {
			return fmt.Errorf("Error getting autoscaling: %w", err)
		}

		processValues := ProcessValues{
			Annotations:  annotations,
			Autoscaling:  autoscaling,
			Args:         args,
			Healthchecks: processHealthchecks,
			Labels:       labels,
			ProcessType:  ProcessType_Worker,
			Replicas:     int32(processCount),
			Resources:    processResources,
			Volumes:      processVolumes,
		}

		if processType == "web" {
			sort.Strings(domains)
			domainValues := []ProcessDomains{}
			for _, domain := range domains {
				domainValues = append(domainValues, ProcessDomains{
					Name: domain,
					Slug: slug.Make(domain),
				})
			}

			processValues.Web = ProcessWeb{
				Domains:  domainValues,
				PortMaps: []ProcessPortMap{},
				TLS: ProcessTls{
					Enabled:    tlsEnabled,
					IssuerName: issuerName,
				},
			}

			processValues.ProcessType = ProcessType_Web
			for _, portMap := range portMaps {
				protocol := PortmapProtocol_TCP
				if portMap.Scheme == "udp" {
					protocol = PortmapProtocol_UDP
				}

				processValues.Web.PortMaps = append(processValues.Web.PortMaps, ProcessPortMap{
					ContainerPort: portMap.ContainerPort,
					HostPort:      portMap.HostPort,
					Name:          portMap.String(),
					Protocol:      protocol,
					Scheme:        portMap.Scheme,
				})
			}

			for _, portMap := range processValues.Web.PortMaps {
				_, httpOk := portMaps[fmt.Sprintf("http-80-%d", portMap.ContainerPort)]
				_, httpsOk := portMaps[fmt.Sprintf("https-443-%d", portMap.ContainerPort)]
				if portMap.Scheme == "http" && !httpsOk && tlsEnabled {
					processValues.Web.PortMaps = append(processValues.Web.PortMaps, ProcessPortMap{
						ContainerPort: portMap.ContainerPort,
						HostPort:      443,
						Name:          fmt.Sprintf("https-443-%d", portMap.ContainerPort),
						Protocol:      PortmapProtocol_TCP,
						Scheme:        "https",
					})
				}

				if portMap.Scheme == "https" && !httpOk {
					processValues.Web.PortMaps = append(processValues.Web.PortMaps, ProcessPortMap{
						ContainerPort: portMap.ContainerPort,
						HostPort:      80,
						Name:          fmt.Sprintf("http-80-%d", portMap.ContainerPort),
						Protocol:      PortmapProtocol_TCP,
						Scheme:        "http",
					})
				}
			}

			sort.Sort(NameSorter(processValues.Web.PortMaps))
		}

		values.Processes[processType] = processValues

		templateFiles := []string{"deployment", "keda-scaled-object"}
		if processType == "web" {
			templateFiles = append(templateFiles, "service", "certificate", "ingress", "ingress-route", "https-redirect-middleware", "keda-http-scaled-object", "keda-interceptor-proxy-service")
		}
		for _, templateName := range templateFiles {
			b, err := templates.ReadFile(fmt.Sprintf("templates/chart/%s.yaml", templateName))
			if err != nil {
				return fmt.Errorf("Error reading %s template: %w", templateName, err)
			}

			filename := filepath.Join(chartDir, "templates", fmt.Sprintf("%s.yaml", templateName))
			err = os.WriteFile(filename, b, os.FileMode(0644))
			if err != nil {
				return fmt.Errorf("Error writing %s template: %w", templateName, err)
			}

			if os.Getenv("DOKKU_TRACE") == "1" {
				common.CatFile(filename)
			}
		}
	}

	cronJobs, err := clientset.ListCronJobs(ctx, ListCronJobsInput{
		LabelSelector: fmt.Sprintf("app.kubernetes.io/part-of=%s", appName),
		Namespace:     namespace,
	})
	if err != nil {
		return fmt.Errorf("Error listing cron jobs: %w", err)
	}
	for _, cronEntry := range cronEntries {
		// todo: implement deployment annotations
		// todo: implement pod annotations
		// todo: implement volumes
		suffix := ""
		for _, cronJob := range cronJobs {
			if cronJob.Labels["dokku.com/cron-id"] == cronEntry.ID {
				var ok bool
				suffix, ok = cronJob.Annotations["dokku.com/job-suffix"]
				if !ok {
					suffix = ""
				}
			}
		}
		if suffix == "" {
			n := 5
			b := make([]byte, n)
			if _, err := rand.Read(b); err != nil {
				panic(err)
			}
			suffix = strings.ToLower(fmt.Sprintf("%X", b))
		}

		words, err := shellquote.Split(cronEntry.Command)
		if err != nil {
			return fmt.Errorf("Error parsing cron command: %w", err)
		}

		processResources, err := getProcessResources(appName, cronEntry.ID)
		if err != nil {
			return fmt.Errorf("Error getting process resources: %w", err)
		}

		annotations, err := getAnnotations(appName, cronEntry.ID)
		if err != nil {
			return fmt.Errorf("Error getting process annotations: %w", err)
		}

		labels, err := getLabels(appName, cronEntry.ID)
		if err != nil {
			return fmt.Errorf("Error getting process labels: %w", err)
		}

		processValues := ProcessValues{
			Args:        words,
			Annotations: annotations,
			Cron: ProcessCron{
				ID:       cronEntry.ID,
				Schedule: cronEntry.Schedule,
				Suffix:   suffix,
			},
			Labels:      labels,
			ProcessType: ProcessType_Cron,
			Replicas:    1,
			Resources:   processResources,
			Volumes:     processVolumes,
		}
		values.Processes[cronEntry.ID] = processValues
	}

	if len(cronEntries) > 0 {
		b, err := templates.ReadFile("templates/chart/cron-job.yaml")
		if err != nil {
			return fmt.Errorf("Error reading cron job template: %w", err)
		}

		cronFile := filepath.Join(chartDir, "templates", "cron-job.yaml")
		err = os.WriteFile(cronFile, b, os.FileMode(0644))
		if err != nil {
			return fmt.Errorf("Error writing cron job template: %w", err)
		}

		if os.Getenv("DOKKU_TRACE") == "1" {
			common.CatFile(cronFile)
		}
	}

	for key, value := range env.Map() {
		values.Global.Secrets[key] = base64.StdEncoding.EncodeToString([]byte(value))
	}

	b, err := templates.ReadFile("templates/chart/_helpers.tpl")
	if err != nil {
		return fmt.Errorf("Error reading _helpers template: %w", err)
	}

	helpersFile := filepath.Join(chartDir, "templates", "_helpers.tpl")
	err = os.WriteFile(helpersFile, b, os.FileMode(0644))
	if err != nil {
		return fmt.Errorf("Error writing _helpers template: %w", err)
	}

	if os.Getenv("DOKKU_TRACE") == "1" {
		common.CatFile(helpersFile)
	}

	err = writeYaml(WriteYamlInput{
		Object: values,
		Path:   filepath.Join(chartDir, "values.yaml"),
	})
	if err != nil {
		return fmt.Errorf("Error writing chart: %w", err)
	}

	helmAgent, err := NewHelmAgent(namespace, DeployLogPrinter)
	if err != nil {
		return fmt.Errorf("Error creating helm agent: %w", err)
	}

	chartPath, err := filepath.Abs(chartDir)
	if err != nil {
		return fmt.Errorf("Error getting chart path: %w", err)
	}

	timeoutDuration, err := time.ParseDuration(deployTimeout)
	if err != nil {
		return fmt.Errorf("Error parsing deploy timeout duration: %w", err)
	}

	ingresses, err := clientset.ListIngresses(ctx, ListIngressesInput{
		Namespace:     namespace,
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
			Namespace: namespace,
		})

		if err != nil {
			return fmt.Errorf("Error deleting ingress: %w", err)
		}
	}

	kustomizeRootPath := ""
	if hasKustomizeRootPath(appName) {
		kustomizeRootPath = getProcessSpecificKustomizeRootPath(appName)
	}

	common.LogInfo2(fmt.Sprintf("Installing %s", appName))
	err = helmAgent.InstallOrUpgradeChart(ctx, ChartInput{
		ChartPath:         chartPath,
		KustomizeRootPath: kustomizeRootPath,
		Namespace:         namespace,
		ReleaseName:       appName,
		RollbackOnFailure: allowRollbacks,
		Timeout:           timeoutDuration,
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

	if os.Getenv("DOKKU_TRACE") == "1" {
		extraEnv["TRACE"] = "true"
	}

	processType := "run"
	if os.Getenv("DOKKU_CRON_ID") != "" {
		processType = "cron"
		labels["dokku.com/cron-id"] = os.Getenv("DOKKU_CRON_ID")
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
		if err == nil && resp.Stdout != "" {
			common.LogInfo1Quiet(fmt.Sprintf("Found '%s' in Procfile, running that command", args[0]))
			return err
		}
		// todo: run command in procfile
	}

	entrypoint := ""
	if imageSourceType == "herokuish" {
		entrypoint = "/exec"
	}

	namespace := getComputedNamespace(appName)
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

	clientset, err := NewKubernetesClient()
	if err != nil {
		return fmt.Errorf("Error creating kubernetes client: %w", err)
	}

	if err := clientset.Ping(); err != nil {
		return fmt.Errorf("kubernetes api not available: %w", err)
	}

	imagePullSecrets := getComputedImagePullSecrets(appName)
	if imagePullSecrets == "" {
		imagePullSecrets = fmt.Sprintf("ims-%s.%d", appName, deploymentID)
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

	workingDir := common.GetWorkingDir(appName, image)
	job, err := templateKubernetesJob(Job{
		AppName:          appName,
		Command:          []string{commandShell},
		DeploymentID:     deploymentID,
		Entrypoint:       entrypoint,
		Env:              extraEnv,
		Image:            image,
		ImagePullSecrets: imagePullSecrets,
		ImageSourceType:  imageSourceType,
		Interactive:      attachToPod || common.ToBool(os.Getenv("DOKKU_FORCE_TTY")),
		Labels:           labels,
		Namespace:        namespace,
		ProcessType:      processType,
		RemoveContainer:  rmContainer,
		WorkingDir:       workingDir,
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
		Timeout:       10,
		Waiter:        isPodReady,
	})
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

	type CronJobEntry struct {
		ID       string `json:"id"`
		AppName  string `json:"app"`
		Command  string `json:"command"`
		Schedule string `json:"schedule"`
	}

	data := []CronJobEntry{}
	lines := []string{"ID | Schedule | Command"}
	for _, cronJob := range cronJobs {
		command := ""
		for _, container := range cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers {
			if container.Name == fmt.Sprintf("%s-cron", appName) {
				command = strings.Join(container.Args, " ")
			}
		}

		cronID, ok := cronJob.Labels["dokku.com/cron-id"]
		if !ok {
			common.LogWarn(fmt.Sprintf("Cron job %s does not have a cron ID label", cronJob.Name))
			continue
		}

		lines = append(lines, fmt.Sprintf("%s | %s | %s", cronID, cronJob.Spec.Schedule, command))
		data = append(data, CronJobEntry{
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
