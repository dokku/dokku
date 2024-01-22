package scheduler_k3s

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
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
	"github.com/fatih/color"
	"github.com/hashicorp/go-multierror"
	"github.com/kballard/go-shellquote"
	"github.com/rancher/wharfie/pkg/registries"
	"github.com/ryanuber/columnize"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/client/conditions"
	"k8s.io/utils/ptr"
)

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

// TriggerPostRegistryLogin updates the `/etc/rancher/k3s/registries.yaml` to include
// auth information for the registry. Note that if the file does not exist, it won't be updated.
func TriggerPostRegistryLogin(server string, username string) error {
	if !common.FileExists("/usr/local/bin/k3s") {
		return nil
	}

	password := os.Getenv("DOCKER_REGISTRY_PASS")

	yamlFile, err := os.ReadFile(RegistryConfigPath)
	if err != nil {
		return fmt.Errorf("Unable to read existing registries.yaml: %w", err)
	}

	var registry registries.Registry
	err = yaml.Unmarshal(yamlFile, &registry)
	if err != nil {
		return fmt.Errorf("Unable to unmarshal registry configuration from yaml: %w", err)
	}

	common.LogInfo1("Updating k3s configuration")
	if registry.Auths == nil {
		registry.Auths = map[string]registries.AuthConfig{}
	}

	if server == "docker.io" {
		server = "registry-1.docker.io"
	}

	registry.Auths[server] = registries.AuthConfig{
		Username: username,
		Password: password,
	}

	data, err := yaml.Marshal(&registry)
	if err != nil {
		return fmt.Errorf("Unable to marshal registry configuration to yaml: %w", err)
	}

	if err := os.WriteFile(RegistryConfigPath, data, os.FileMode(0644)); err != nil {
		return fmt.Errorf("Unable to write registry configuration to file: %w", err)
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

	nodes, err := clientset.ListNodes(ctx, ListNodesInput{})
	if err != nil {
		return fmt.Errorf("Error listing nodes: %w", err)
	}

	var result error
	for _, node := range nodes {
		remoteHost, ok := node.Annotations["dokku.com/remote-host"]
		if !ok {
			continue
		}

		err := copyRegistryToNode(ctx, remoteHost)
		if err != nil {
			wrappedErr := fmt.Errorf("Error copying registry to node: %w", err)
			result = multierror.Append(result, wrappedErr)
			common.LogWarn(wrappedErr.Error())
		}
	}

	return result
}

// TriggerSchedulerDeploy deploys an image tag for a given application
func TriggerSchedulerDeploy(scheduler string, appName string, imageTag string) error {
	if scheduler != "k3s" {
		return nil
	}
	s, err := common.PlugnTriggerOutput("ps-current-scale", []string{appName}...)
	if err != nil {
		return err
	}

	processes, err := common.ParseScaleOutput(s)
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

	imagePullSecrets := getComputedImagePullSecrets(appName)

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

	chartDir, err := os.MkdirTemp("", "dokku-chart-")
	if err != nil {
		return fmt.Errorf("Error creating chart directory: %w", err)
	}
	defer os.RemoveAll(chartDir)

	if err := os.MkdirAll(filepath.Join(chartDir, "templates"), os.FileMode(0755)); err != nil {
		return fmt.Errorf("Error creating chart templates directory: %w", err)
	}

	deploymentId := time.Now().Unix()
	replacements := orderedmap.New[string, string]()
	replacements.Set("DEPLOYMENT_ID_QUOTED", "{{.Values.deploment_id | quote}}")
	replacements.Set("DEPLOYMENT_ID", "{{.Values.deploment_id}}")

	secret := templateKubernetesSecret(Secret{
		AppName:   appName,
		Env:       env.Map(),
		Namespace: namespace,
	})
	err = writeResourceToFile(WriteResourceInput{
		Object:       &secret,
		Path:         filepath.Join(chartDir, "templates/secret.yaml"),
		Replacements: replacements,
		AppendContents: `{{- with .Values.secrets }}
data:
  {{- toYaml . | nindent 2 }}
{{- end }}
`,
	})
	if err != nil {
		return fmt.Errorf("Error printing deployment: %w", err)
	}

	portMaps, err := getPortMaps(appName)
	if err != nil {
		return fmt.Errorf("Error getting port mappings for deployment: %w", err)
	}

	primaryPort := int32(5000)
	for _, portMap := range portMaps {
		primaryPort = portMap.ContainerPort
		if primaryPort != 0 {
			break
		}
	}

	appJSON, err := appjson.GetAppJSON(appName)
	if err != nil {
		return fmt.Errorf("Error getting app.json for deployment: %w", err)
	}

	workingDir := common.GetWorkingDir(appName, image)
	deployments := map[string]appsv1.Deployment{}
	i := 0
	for processType := range processes {
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

		i++
		replicaCountPlaceholder := int32(i * 1000)

		healthchecks, ok := appJSON.Healthchecks[processType]
		if !ok {
			healthchecks = []appjson.Healthcheck{}
		}

		// todo: implement deployment annotations
		// todo: implement pod annotations
		// todo: implement volumes
		deployment, err := templateKubernetesDeployment(Deployment{
			AppName:          appName,
			Command:          startCommand.Command,
			Image:            image,
			ImagePullSecrets: imagePullSecrets,
			ImageSourceType:  imageSourceType,
			Healthchecks:     healthchecks,
			Namespace:        namespace,
			PrimaryPort:      primaryPort,
			PortMaps:         portMaps,
			ProcessType:      processType,
			Replicas:         replicaCountPlaceholder,
			WorkingDir:       workingDir,
		})
		if err != nil {
			return fmt.Errorf("Error templating deployment: %w", err)
		}

		replacements.Set(fmt.Sprintf("replicas: %d", replicaCountPlaceholder), fmt.Sprintf("replicas: {{.Values.processes.%s.replicas}}", processType))
		deployments[processType] = deployment
		err = writeResourceToFile(WriteResourceInput{
			Object:       &deployment,
			Path:         filepath.Join(chartDir, fmt.Sprintf("templates/deployment-%s.yaml", deployment.Name)),
			Replacements: replacements,
		})
		if err != nil {
			return fmt.Errorf("Error printing deployment: %w", err)
		}

		replacements.Delete(fmt.Sprintf("replicas: %d", replicaCountPlaceholder))
	}

	cronEntries, err := cron.FetchCronEntries(appName)
	if err != nil {
		return fmt.Errorf("Error fetching cron entries: %w", err)
	}

	clientset, err := NewKubernetesClient()
	if err != nil {
		return fmt.Errorf("Error creating kubernetes client: %w", err)
	}

	cronJobs, err := clientset.ListCronJobs(ctx, ListCronJobsInput{
		LabelSelector: fmt.Sprintf("app.kubernetes.io/part-of=%s", appName),
		Namespace:     namespace,
	})
	if err != nil {
		return fmt.Errorf("Error listing cron jobs: %w", err)
	}

	for _, cronEntry := range cronEntries {
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

		words, err := shellquote.Split(cronEntry.Command)
		if err != nil {
			return fmt.Errorf("Error parsing cron command: %w", err)
		}
		cronJob, err := templateKubernetesCronJob(Job{
			AppName:          appName,
			Command:          words,
			Env:              map[string]string{},
			ID:               cronEntry.ID,
			Image:            image,
			ImagePullSecrets: imagePullSecrets,
			ImageSourceType:  imageSourceType,
			Namespace:        namespace,
			ProcessType:      "cron",
			Schedule:         cronEntry.Schedule,
			Suffix:           suffix,
			WorkingDir:       workingDir,
		})
		if err != nil {
			return fmt.Errorf("Error templating cron job: %w", err)
		}

		err = writeResourceToFile(WriteResourceInput{
			Object:       &cronJob,
			Path:         filepath.Join(chartDir, fmt.Sprintf("templates/cron-job-%s.yaml", cronEntry.ID)),
			Replacements: replacements,
		})
		if err != nil {
			return fmt.Errorf("Error printing cron job: %w", err)
		}
	}

	// todo: make this configurable
	tls := false

	domains := []string{}
	if deployment, ok := deployments["web"]; ok {
		service := templateKubernetesService(Service{
			AppName:   appName,
			Namespace: namespace,
			PortMaps:  portMaps,
		})

		err := writeResourceToFile(WriteResourceInput{
			Object:       &service,
			Path:         filepath.Join(chartDir, "templates/service-web.yaml"),
			Replacements: replacements,
		})
		if err != nil {
			return fmt.Errorf("Error printing service: %w", err)
		}

		err = common.PlugnTrigger("domains-vhost-enabled", []string{appName}...)
		if err == nil {
			b, err := common.PlugnTriggerOutput("domains-list", []string{appName}...)
			if err != nil {
				return fmt.Errorf("Error getting domains for deployment: %w", err)
			}

			for _, domain := range strings.Split(string(b), "\n") {
				domain = strings.TrimSpace(domain)
				if domain != "" {
					domains = append(domains, domain)
				}
			}
		}

		err = createIngressRoutesFiles(CreateIngressRoutesInput{
			AppName:     appName,
			ChartDir:    chartDir,
			Deployment:  deployment,
			Namespace:   namespace,
			PortMaps:    portMaps,
			ProcessType: "web",
			Service:     service,
		})
		if err != nil {
			return fmt.Errorf("Error creating ingress routes: %w", err)
		}

		err = createCertificateFile(CreateCertificateFileInput{
			ChartDir: chartDir,
			Certificate: Certificate{
				AppName:   appName,
				Name:      fmt.Sprintf("%s-%s", appName, "web"),
				Namespace: namespace,
				TLS:       tls,
			},
			IssuerName:  "letsencrypt-prod",
			ProcessType: "web",
		})
		if err != nil {
			return fmt.Errorf("Error creating certificate files: %w", err)
		}
	}

	chart := &Chart{
		ApiVersion: "v2",
		AppVersion: "1.0.0",
		Name:       appName,
		Version:    fmt.Sprintf("0.0.%d", deploymentId),
	}

	err = writeYaml(WriteYamlInput{
		Object: chart,
		Path:   filepath.Join(chartDir, "Chart.yaml"),
	})
	if err != nil {
		return fmt.Errorf("Error writing chart: %w", err)
	}

	values := &Values{
		DeploymentID: fmt.Sprint(deploymentId),
		Secrets:      map[string]string{},
		Processes:    map[string]ProcessValues{},
	}
	for processType, processCount := range processes {
		processValues := ProcessValues{
			Replicas: int32(processCount),
		}
		if processType == "web" {
			sort.Strings(domains)
			processValues.Domains = domains
			processValues.TLS = tls
		}

		values.Processes[processType] = processValues
	}

	for key, value := range env.Map() {
		values.Secrets[key] = base64.StdEncoding.EncodeToString([]byte(value))
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

	err = helmAgent.InstallOrUpgradeChart(ctx, ChartInput{
		ChartPath:         chartPath,
		Namespace:         namespace,
		ReleaseName:       fmt.Sprintf("dokku-%s", appName),
		RollbackOnFailure: allowRollbacks,
		Timeout:           timeoutDuration,
	})
	if err != nil {
		return err
	}

	common.LogInfo1("Running post-deploy")
	_, err = common.CallPlugnTrigger(common.PlugnTriggerInput{
		Args:          []string{appName, "", "", imageTag},
		CaptureOutput: false,
		StreamStdio:   true,
		Trigger:       "core-post-deploy",
	})
	if err != nil {
		return fmt.Errorf("Error running core-post-deploy: %w", err)

	}
	_, err = common.CallPlugnTrigger(common.PlugnTriggerInput{
		Args:          []string{appName, "", "", imageTag},
		CaptureOutput: false,
		StreamStdio:   true,
		Trigger:       "post-deploy",
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
		command = []string{"/bin/bash"}
		if globalShell, err := common.PlugnTriggerOutputAsString("config-get-global", []string{"DOKKU_APP_SHELL"}...); err == nil && globalShell != "" {
			command = []string{globalShell}
		}
		if appShell, err := common.PlugnTriggerOutputAsString("config-get", []string{appName, "DOKKU_APP_SHELL"}...); err == nil && appShell != "" {
			command = []string{appShell}
		}
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

	labelSelector := []string{fmt.Sprintf("app.kubernetes.io/part-of=%s", appName)}
	processIndex := 0
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

	namespace := getComputedNamespace(appName)
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

	ch := make(chan bool)

	if os.Getenv("FORCE_TTY") == "1" {
		color.NoColor = false
	}

	colors := []color.Attribute{
		color.FgRed,
		color.FgYellow,
		color.FgGreen,
		color.FgCyan,
		color.FgBlue,
		color.FgMagenta,
	}
	// colorIndex := 0
	for i := 0; i < len(pods); i++ {
		if processIndex > 0 && i != (processIndex-1) {
			continue
		}

		logOptions := v1.PodLogOptions{
			Follow: tail,
		}
		if numLines > 0 {
			logOptions.TailLines = ptr.To(numLines)
		}

		podColor := colors[i%len(colors)]
		dynoText := color.New(podColor).SprintFunc()
		podName := pods[i].Name
		podLogs, err := clientset.Client.CoreV1().Pods(namespace).GetLogs(podName, &logOptions).Stream(ctx)
		if err != nil {
			return err
		}
		buffer := bufio.NewReader(podLogs)
		go func(ctx context.Context, buffer *bufio.Reader, prettyText func(a ...interface{}) string, ch chan bool) {
			defer func() {
				ch <- true
			}()
			for {
				select {
				case <-ctx.Done(): // if cancel() execute
					ch <- true
					return
				default:
					str, readErr := buffer.ReadString('\n')
					if readErr == io.EOF {
						break
					}

					if str == "" {
						continue
					}

					if !quiet {
						str = fmt.Sprintf("%s %s", dynoText(fmt.Sprintf("app[%s]:", podName)), str)
					}

					_, err := fmt.Print(str)
					if err != nil {
						return
					}
				}
			}
		}(ctx, buffer, dynoText, ch)
	}
	<-ch

	return nil
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
			Trigger:       "config-get",
			Args:          []string{appName, "DOKKU_RM_CONTAINER"},
			CaptureOutput: true,
			StreamStdio:   false,
		})
		if err != nil {
			resp, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
				Trigger:       "config-get-global",
				Args:          []string{"DOKKU_RM_CONTAINER"},
				CaptureOutput: true,
				StreamStdio:   false,
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
	if len(args) == 0 {
		command = []string{"/bin/bash"}
		if globalShell, err := common.PlugnTriggerOutputAsString("config-get-global", []string{"DOKKU_APP_SHELL"}...); err == nil && globalShell != "" {
			command = []string{globalShell}
		}
		if appShell, err := common.PlugnTriggerOutputAsString("config-get", []string{appName, "DOKKU_APP_SHELL"}...); err == nil && appShell != "" {
			command = []string{appShell}
		}
	} else if len(args) == 1 {
		resp, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
			Trigger:       "procfile-get-command",
			Args:          []string{appName, args[0], "5000"},
			CaptureOutput: true,
			StreamStdio:   false,
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

	values, err := helmAgent.GetValues(fmt.Sprintf("dokku-%s", appName))
	if err != nil {
		return fmt.Errorf("Error getting helm values: %w", err)
	}

	deploymentIDValue, ok := values["deploment_id"].(string)
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
	imagePullSecrets := getComputedImagePullSecrets(appName)
	workingDir := common.GetWorkingDir(appName, image)
	job, err := templateKubernetesJob(Job{
		AppName:          appName,
		Command:          command,
		DeploymentID:     deploymentID,
		Entrypoint:       entrypoint,
		Env:              extraEnv,
		Image:            image,
		ImagePullSecrets: imagePullSecrets,
		ImageSourceType:  imageSourceType,
		Interactive:      attachToPod,
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

	clientset, err := NewKubernetesClient()
	if err != nil {
		return fmt.Errorf("Error creating kubernetes client: %w", err)
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

	for _, pod := range pods {
		common.LogQuiet(pod.Name)
	}

	if !attachToPod {
		return nil
	}

	err = waitForPodBySelectorRunning(ctx, WaitForPodBySelectorRunningInput{
		Clientset:     clientset,
		Namespace:     namespace,
		LabelSelector: batchJobSelector,
		Timeout:       300,
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
			Clientset:   clientset,
			Command:     command,
			Entrypoint:  entrypoint,
			SelectedPod: selectedPod,
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
	namespace := getComputedNamespace(appName)
	helmAgent, err := NewHelmAgent(namespace, DeployLogPrinter)
	if err != nil {
		return fmt.Errorf("Error creating helm agent: %w", err)
	}

	err = helmAgent.UninstallChart(fmt.Sprintf("dokku-%s", appName))
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
		return fmt.Errorf("Error creating kubernetes client: %w", err)
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
		processType, ok := deployment.Annotations["app.kubernetes.io/name"]
		if !ok {
			return fmt.Errorf("Deployment %s does not have a process type annotation", deployment.Name)
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

	return nil
}
