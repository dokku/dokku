package scheduler_k3s

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	appjson "github.com/dokku/dokku/plugins/app-json"
	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
	"github.com/dokku/dokku/plugins/cron"
	"github.com/dokku/dokku/plugins/registry"
	"github.com/gosimple/slug"
	"github.com/kballard/go-shellquote"
)

// BuildResult is the materialized output of BuildAppChart. It points at a
// temporary chart directory on disk; the caller is responsible for removing
// it once Helm is done with it.
type BuildResult struct {
	ChartDir          string
	ChartPath         string
	ReleaseName       string
	Namespace         string
	Timeout           time.Duration
	RollbackOnFailure bool
	KustomizeRootPath string
}

// BuildAppChart constructs a Helm chart on disk that reflects dokku's configured
// state for the given app and image tag. It reads from the Kubernetes cluster
// where required (for KEDA values and existing cron-job suffixes) but performs
// no writes to the cluster - no namespace, secrets, ingresses, or helm releases
// are created or modified.
//
// The returned BuildResult holds the absolute chart path and the parameters a
// subsequent Helm Install/Upgrade (dry-run or real) needs. The caller must
// os.RemoveAll(result.ChartDir) when finished.
func BuildAppChart(ctx context.Context, appName, imageTag string) (BuildResult, error) {
	var result BuildResult

	processesRes, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "ps-current-scale",
		Args:    []string{appName},
	})
	if err != nil {
		return result, err
	}

	processes, err := common.ParseScaleOutput(processesRes.StdoutBytes())
	if err != nil {
		return result, err
	}

	namespace := getComputedNamespace(appName)

	image, err := common.GetDeployingAppImageName(appName, imageTag, "")
	if err != nil {
		return result, fmt.Errorf("Error getting deploying app image name: %w", err)
	}

	deployTimeout := getComputedDeployTimeout(appName)
	if _, err := strconv.Atoi(deployTimeout); err == nil {
		deployTimeout = fmt.Sprintf("%ss", deployTimeout)
	}
	timeoutDuration, err := time.ParseDuration(deployTimeout)
	if err != nil {
		return result, fmt.Errorf("Error parsing deploy timeout duration: %w", err)
	}

	deployRollback := getComputedRollbackOnFailure(appName)
	allowRollbacks, err := strconv.ParseBool(deployRollback)
	if err != nil {
		return result, fmt.Errorf("Error parsing rollback-on-failure value as boolean: %w", err)
	}

	imageSourceType := "dockerfile"
	if common.IsImageCnbBased(image) {
		imageSourceType = "pack"
	} else if common.IsImageHerokuishBased(image, appName) {
		imageSourceType = "herokuish"
	}

	env, err := config.LoadMergedAppEnv(appName)
	if err != nil {
		return result, fmt.Errorf("Error loading environment for deployment: %w", err)
	}

	importedCertExists := false
	if HasImportedTLSCert(appName) {
		exists, err := TLSSecretExists(ctx, appName)
		if err == nil && exists {
			importedCertExists = true
		}
	}

	tlsEnabled := false
	issuerName := ""
	useImportedCert := false

	if importedCertExists {
		tlsEnabled = true
		useImportedCert = true
	} else {
		server := getComputedLetsencryptServer(appName)
		letsencryptEmailStag := getGlobalLetsencryptEmailStag()
		letsencryptEmailProd := getGlobalLetsencryptEmailProd()

		switch server {
		case "prod", "production":
			issuerName = "letsencrypt-prod"
			tlsEnabled = letsencryptEmailProd != ""
		case "stag", "staging":
			issuerName = "letsencrypt-stag"
			tlsEnabled = letsencryptEmailStag != ""
		case "false":
			issuerName = ""
			tlsEnabled = false
		default:
			return result, fmt.Errorf("Invalid letsencrypt server config: %s", server)
		}
	}

	chartDir, err := os.MkdirTemp("", "dokku-chart-")
	if err != nil {
		return result, fmt.Errorf("Error creating chart directory: %w", err)
	}

	// From this point on, any error returns require os.RemoveAll(chartDir).
	cleanup := func(retErr error) (BuildResult, error) {
		os.RemoveAll(chartDir)
		return result, retErr
	}

	if err := os.MkdirAll(filepath.Join(chartDir, "templates"), os.FileMode(0755)); err != nil {
		return cleanup(fmt.Errorf("Error creating chart templates directory: %w", err))
	}

	deploymentId := time.Now().Unix()
	imagePullSecrets := getComputedImagePullSecrets(appName)
	if imagePullSecrets == "" {
		dockerConfigPath := filepath.Join(registry.GetComputedAppRegistryConfigDir(appName), "config.json")
		if fi, err := os.Stat(dockerConfigPath); err == nil && !fi.IsDir() {
			imagePullSecrets = GetImagePullSecretName(appName)
		}
	}

	globalTemplateFiles := []string{"service-account"}
	for _, templateName := range globalTemplateFiles {
		b, err := templates.ReadFile(fmt.Sprintf("templates/chart/%s.yaml", templateName))
		if err != nil {
			return cleanup(fmt.Errorf("Error reading %s template: %w", templateName, err))
		}

		filename := filepath.Join(chartDir, "templates", fmt.Sprintf("%s.yaml", templateName))
		if err := os.WriteFile(filename, b, os.FileMode(0644)); err != nil {
			return cleanup(fmt.Errorf("Error writing %s template: %w", templateName, err))
		}

		if os.Getenv("DOKKU_TRACE") == "1" {
			common.CatFile(filename)
		}
	}

	portMaps, err := getPortMaps(appName)
	if err != nil {
		return cleanup(fmt.Errorf("Error getting port mappings for deployment: %w", err))
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
		return cleanup(fmt.Errorf("Error getting app.json for deployment: %w", err))
	}

	workingDir := common.GetWorkingDir(appName, image)

	cronTasks, err := cron.FetchCronTasks(cron.FetchCronTasksInput{AppName: appName})
	if err != nil {
		return cleanup(fmt.Errorf("Error fetching cron tasks: %w", err))
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
				return cleanup(fmt.Errorf("Error getting domains for deployment: %w", err))
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

	if err := writeYaml(WriteYamlInput{
		Object: chart,
		Path:   filepath.Join(chartDir, "Chart.yaml"),
	}); err != nil {
		return cleanup(fmt.Errorf("Error writing chart: %w", err))
	}

	globalAnnotations, err := getGlobalAnnotations(appName)
	if err != nil {
		return cleanup(fmt.Errorf("Error getting global annotations: %w", err))
	}

	globalLabels, err := getGlobalLabel(appName)
	if err != nil {
		return cleanup(fmt.Errorf("Error getting global labels: %w", err))
	}

	clientset, err := NewKubernetesClient()
	if err != nil {
		return cleanup(fmt.Errorf("Error creating kubernetes client: %w", err))
	}

	if err := clientset.Ping(); err != nil {
		return cleanup(fmt.Errorf("kubernetes api not available: %w", err))
	}

	kedaValues, err := getKedaValues(ctx, clientset, appName)
	if err != nil {
		return cleanup(fmt.Errorf("Error getting keda values: %w", err))
	}

	securityContext, err := getSecurityContext(appName, "deploy")
	if err != nil {
		return cleanup(fmt.Errorf("Error getting security context: %w", err))
	}

	values := &AppValues{
		Global: GlobalValues{
			Annotations:  globalAnnotations,
			AppName:      appName,
			DeploymentID: fmt.Sprint(deploymentId),
			Keda:         kedaValues,
			Image: GlobalImage{
				ImagePullSecrets: imagePullSecrets,
				Name:             image,
				Type:             imageSourceType,
				WorkingDir:       workingDir,
			},
			Labels:    globalLabels,
			Namespace: namespace,
			Network: GlobalNetwork{
				IngressClass:       getComputedIngressClass(),
				PrimaryPort:        primaryPort,
				PrimaryServicePort: primaryServicePort,
			},
			SecurityContext: securityContext,
		},
		Processes: map[string]ProcessValues{},
	}

	if len(kedaValues.Authentications) > 0 {
		templateFiles := []string{"keda-secret", "keda-trigger-authentication"}
		for _, templateName := range templateFiles {
			b, err := templates.ReadFile(fmt.Sprintf("templates/chart/%s.yaml", templateName))
			if err != nil {
				return cleanup(fmt.Errorf("Error reading %s template: %w", templateName, err))
			}

			filename := filepath.Join(chartDir, "templates", fmt.Sprintf("%s.yaml", templateName))
			if err := os.WriteFile(filename, b, os.FileMode(0644)); err != nil {
				return cleanup(fmt.Errorf("Error writing %s template: %w", templateName, err))
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

	deployMountPairs, err := LoadAppMounts(appName, "deploy")
	if err != nil {
		return cleanup(fmt.Errorf("error loading storage app mounts: %w", err))
	}
	deployVolumes, err := ToProcessVolumes(deployMountPairs)
	if err != nil {
		return cleanup(err)
	}
	processVolumes = append(processVolumes, deployVolumes...)

	for processType, processCount := range processes {
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
			return cleanup(fmt.Errorf("Error getting start command for deployment: %w", err))
		}
		args := startCommand.Command

		processResources, err := getProcessResources(appName, processType)
		if err != nil {
			return cleanup(fmt.Errorf("Error getting process resources: %w", err))
		}

		annotations, err := getAnnotations(appName, processType)
		if err != nil {
			return cleanup(fmt.Errorf("Error getting process annotations: %w", err))
		}

		labels, err := getLabels(appName, processType)
		if err != nil {
			return cleanup(fmt.Errorf("Error getting process labels: %w", err))
		}

		autoscaling, err := getAutoscaling(GetAutoscalingInput{
			AppName:     appName,
			ProcessType: processType,
			Replicas:    int(processCount),
			KedaValues:  kedaValues,
		})
		if err != nil {
			return cleanup(fmt.Errorf("Error getting autoscaling: %w", err))
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
					Enabled:         tlsEnabled,
					IssuerName:      issuerName,
					UseImportedCert: useImportedCert,
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
		} else if appJSON.Formation[processType].Service != nil && appJSON.Formation[processType].Service.Exposed {
			processValues.Web = ProcessWeb{
				Domains:  []ProcessDomains{},
				PortMaps: []ProcessPortMap{},
				TLS: ProcessTls{
					Enabled: false,
				},
			}

			processValues.Web.PortMaps = append(processValues.Web.PortMaps, ProcessPortMap{
				ContainerPort: 5000,
				HostPort:      5000,
				Name:          "http-5000-5000",
				Protocol:      PortmapProtocol_TCP,
				Scheme:        "http",
			})
		}

		values.Processes[processType] = processValues

		templateFiles := []string{"deployment", "keda-scaled-object"}
		if processType == "web" {
			templateFiles = append(templateFiles, "service", "certificate", "ingress", "ingress-route", "compression-middleware", "https-redirect-middleware", "keda-http-scaled-object", "keda-interceptor-proxy-service")
		}
		for _, templateName := range templateFiles {
			b, err := templates.ReadFile(fmt.Sprintf("templates/chart/%s.yaml", templateName))
			if err != nil {
				return cleanup(fmt.Errorf("Error reading %s template: %w", templateName, err))
			}

			filename := filepath.Join(chartDir, "templates", fmt.Sprintf("%s.yaml", templateName))
			if err := os.WriteFile(filename, b, os.FileMode(0644)); err != nil {
				return cleanup(fmt.Errorf("Error writing %s template: %w", templateName, err))
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
		return cleanup(fmt.Errorf("Error listing cron jobs: %w", err))
	}
	for _, cronTask := range cronTasks {
		suffix := ""
		for _, cronJob := range cronJobs {
			if cronJob.Annotations["dokku.com/cron-id"] == cronTask.ID {
				if value, ok := cronJob.Annotations["dokku.com/job-suffix"]; ok {
					suffix = value
				}
			}
		}
		if suffix == "" {
			n := 5
			b := make([]byte, n)
			if _, err := rand.Read(b); err != nil {
				return cleanup(fmt.Errorf("Error generating cron suffix: %w", err))
			}
			suffix = strings.ToLower(fmt.Sprintf("%X", b))
		}

		words, err := shellquote.Split(cronTask.Command)
		if err != nil {
			return cleanup(fmt.Errorf("Error parsing cron task command: %w", err))
		}

		processResources, err := getProcessResources(appName, cronTask.ID)
		if err != nil {
			return cleanup(fmt.Errorf("Error getting process resources: %w", err))
		}

		annotations, err := getAnnotations(appName, cronTask.ID)
		if err != nil {
			return cleanup(fmt.Errorf("Error getting process annotations: %w", err))
		}

		labels, err := getLabels(appName, cronTask.ID)
		if err != nil {
			return cleanup(fmt.Errorf("Error getting process labels: %w", err))
		}

		concurrencyPolicy := strings.ToUpper(cronTask.ConcurrencyPolicy)
		switch concurrencyPolicy {
		case "ALLOW":
			concurrencyPolicy = "Allow"
		case "FORBID":
			concurrencyPolicy = "Forbid"
		case "REPLACE":
			concurrencyPolicy = "Replace"
		default:
			return cleanup(fmt.Errorf("Invalid concurrency_policy specified: %v", concurrencyPolicy))
		}
		processValues := ProcessValues{
			Args:        words,
			Annotations: annotations,
			Cron: ProcessCron{
				ID:                cronTask.ID,
				Hash:              cronIDLabelValue(cronTask.ID),
				Schedule:          cronTask.Schedule,
				Suffix:            suffix,
				Suspend:           cronTask.Maintenance,
				ConcurrencyPolicy: ProcessCronConcurrencyPolicy(concurrencyPolicy),
			},
			Labels:      labels,
			ProcessType: ProcessType_Cron,
			Replicas:    1,
			Resources:   processResources,
			Volumes:     processVolumes,
		}
		values.Processes[cronTask.ID] = processValues
	}

	if len(cronTasks) > 0 {
		b, err := templates.ReadFile("templates/chart/cron-job.yaml")
		if err != nil {
			return cleanup(fmt.Errorf("Error reading cron job template: %w", err))
		}

		cronFile := filepath.Join(chartDir, "templates", "cron-job.yaml")
		if err := os.WriteFile(cronFile, b, os.FileMode(0644)); err != nil {
			return cleanup(fmt.Errorf("Error writing cron job template: %w", err))
		}

		if os.Getenv("DOKKU_TRACE") == "1" {
			common.CatFile(cronFile)
		}
	}

	helpersBytes, err := templates.ReadFile("templates/chart/_helpers.tpl")
	if err != nil {
		return cleanup(fmt.Errorf("Error reading _helpers template: %w", err))
	}

	helpersFile := filepath.Join(chartDir, "templates", "_helpers.tpl")
	if err := os.WriteFile(helpersFile, helpersBytes, os.FileMode(0644)); err != nil {
		return cleanup(fmt.Errorf("Error writing _helpers template: %w", err))
	}

	if os.Getenv("DOKKU_TRACE") == "1" {
		common.CatFile(helpersFile)
	}

	if err := writeYaml(WriteYamlInput{
		Object: values,
		Path:   filepath.Join(chartDir, "values.yaml"),
	}); err != nil {
		return cleanup(fmt.Errorf("Error writing chart: %w", err))
	}

	chartPath, err := filepath.Abs(chartDir)
	if err != nil {
		return cleanup(fmt.Errorf("Error getting chart path: %w", err))
	}

	kustomizeRootPath := ""
	if hasKustomizeDirectory(appName) {
		kustomizeRootPath = getProcessSpecificKustomizeRootPath(appName)
	}

	result = BuildResult{
		ChartDir:          chartDir,
		ChartPath:         chartPath,
		ReleaseName:       appName,
		Namespace:         namespace,
		Timeout:           timeoutDuration,
		RollbackOnFailure: allowRollbacks,
		KustomizeRootPath: kustomizeRootPath,
	}
	return result, nil
}
