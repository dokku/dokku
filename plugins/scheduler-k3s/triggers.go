package scheduler_k3s

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
	"github.com/fatih/color"
	"github.com/kballard/go-shellquote"
	"github.com/rancher/wharfie/pkg/registries"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/util/term"
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
	namespace := common.PropertyGetDefault("scheduler-k3s", appName, "namespace", "default")
	helmAgent, err := NewHelmAgent(namespace, DeployLogPrinter)
	if err != nil {
		return fmt.Errorf("Error creating helm agent: %w", err)
	}

	err = helmAgent.UninstallChart(fmt.Sprintf("dokku-%s", appName))
	if err != nil {
		return fmt.Errorf("Error uninstalling chart: %w", err)
	}

	dataErr := common.RemoveAppDataDirectory("scheduler-k3s", appName)
	propertyErr := common.PropertyDestroy("scheduler-k3s", appName)

	if dataErr != nil {
		return dataErr
	}

	return propertyErr
}

// TriggerPostRegistryLogin updates the `/etc/rancher/k3s/registries.yaml` to include
// auth information for the registry. Note that if the file does not exist, it won't be updated.
func TriggerPostRegistryLogin(server string, username string, password string) error {
	if !common.FileExists("/usr/local/bin/k3s") {
		return nil
	}

	registry := registries.Registry{}
	registryFile := "/etc/rancher/k3s/registries.yaml"
	yamlFile, err := os.ReadFile(registryFile)
	if err != nil {
		return fmt.Errorf("Unable to read existing registries.yaml: %w", err)
	}

	err = yaml.Unmarshal(yamlFile, registry)
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

	if err := os.WriteFile(registryFile, data, os.FileMode(0644)); err != nil {
		return fmt.Errorf("Unable to write registry configuration to file: %w", err)
	}

	// todo: auth against all nodes in cluster
	return nil
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

	ctx := context.Background()
	namespace := common.PropertyGetDefault("scheduler-k3s", appName, "namespace", "default")
	if err := createKubernetesNamespace(ctx, namespace); err != nil {
		return fmt.Errorf("Error creating kubernetes namespace for deployment: %w", err)
	}

	image, err := common.GetDeployingAppImageName(appName, imageTag, "")
	if err != nil {
		return fmt.Errorf("Error getting deploying app image name: %w", err)
	}

	deployTimeout := common.PropertyGetDefault("scheduler-k3s", appName, "deploy-timeout", "300s")
	if _, err := strconv.Atoi(deployTimeout); err == nil {
		deployTimeout = fmt.Sprintf("%ss", deployTimeout)
	}

	deployRollback := common.PropertyGetDefault("scheduler-k3s", appName, "rollback-on-failure", "false")
	allowRollbacks, err := strconv.ParseBool(deployRollback)
	if err != nil {
		return fmt.Errorf("Error parsing rollback-on-failure value as boolean: %w", err)
	}

	imagePullSecrets := common.PropertyGetDefault("scheduler-k3s", appName, "image-pull-secrets", "")

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

	common.LogDebug(fmt.Sprintf("Using chart directory: %s", chartDir))
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

		// todo: implement healthchecks
		// todo: implement deployment annotations
		// todo: implement pod annotations
		// todo: implement volumes
		deployment, err := templateKubernetesDeployment(Deployment{
			AppName:          appName,
			Command:          startCommand.Command,
			Image:            image,
			ImagePullSecrets: imagePullSecrets,
			ImageSourceType:  imageSourceType,
			Namespace:        namespace,
			PrimaryPort:      primaryPort,
			PortMaps:         portMaps,
			ProcessType:      processType,
			Replicas:         replicaCountPlaceholder,
			WorkingDir:       common.GetWorkingDir(appName, image),
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
		replacements.Delete(fmt.Sprintf("replicas: %d", replicaCountPlaceholder))

		if err != nil {
			return fmt.Errorf("Error printing deployment: %w", err)
		}
	}

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

		err = createIngressRoutesFiles(CreateIngressRoutesInput{
			AppName:    appName,
			ChartDir:   chartDir,
			Deployment: deployment,
			Namespace:  namespace,
			PortMaps:   portMaps,
			Service:    service,
		})
		if err != nil {
			return fmt.Errorf("Error creating ingress routes: %w", err)
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
		Processes:    map[string]ValuesProcess{},
	}
	for processType, processCount := range processes {
		values.Processes[processType] = ValuesProcess{
			Replicas: int32(processCount),
		}
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

	err = helmAgent.InstallOrUpgradeChart(ChartInput{
		ChartPath:         chartPath,
		Namespace:         namespace,
		ReleaseName:       fmt.Sprintf("dokku-%s", appName),
		RollbackOnFailure: allowRollbacks,
		Timeout:           timeoutDuration,
	})
	if err != nil {
		return err
	}

	return nil
}

// TriggerSchedulerEnter enters a container for a given application
func TriggerSchedulerEnter(scheduler string, appName string, processType string, podName string, args []string) error {
	if scheduler != "k3s" {
		return nil
	}

	common.LogDebug(fmt.Sprintf("%s %s %s %s", scheduler, appName, processType, podName))
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
		cancel()
	}()

	namespace := common.PropertyGetDefault("scheduler-k3s", appName, "namespace", "default")

	labelSelector := []string{fmt.Sprintf("dokku.com/app-name=%s", appName)}
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
		labelSelector = append(labelSelector, fmt.Sprintf("dokku.com/process-type=%s", processType))
	}

	podList, err := clientset.Client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: strings.Join(labelSelector, ","),
	})
	if err != nil {
		return fmt.Errorf("Error listing pods: %w", err)
	}

	if podList.Items == nil {
		return fmt.Errorf("No pods found for app %s", appName)
	}
	if len(podList.Items) == 0 {
		return fmt.Errorf("No pods found for app %s", appName)
	}

	pods := []string{}
	for _, pod := range podList.Items {
		pods = append(pods, pod.Name)
	}
	slices.Sort(pods)

	processIndex--
	if processIndex > len(podList.Items) {
		return fmt.Errorf("Process index %d out of range for app %s", processIndex, appName)
	}

	var selectedPod corev1.Pod
	if podName != "" {
		for _, pod := range podList.Items {
			if pod.Name == podName {
				selectedPod = pod
				break
			}
		}
	} else {
		selectedPod = podList.Items[processIndex]
	}

	processType, ok := selectedPod.Labels["dokku.com/process-type"]
	if !ok {
		return fmt.Errorf("Pod %s does not have a process type label", selectedPod.Name)
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
		AppName:     appName,
		Clientset:   clientset,
		Command:     command,
		Entrypoint:  entrypoint,
		ProcessType: processType,
		SelectedPod: selectedPod,
	})
}

type EnterPodInput struct {
	AppName     string
	Clientset   KubernetesClient
	Command     []string
	Entrypoint  string
	ProcessType string
	SelectedPod v1.Pod
}

func enterPod(ctx context.Context, input EnterPodInput) error {
	coreclient, err := corev1client.NewForConfig(&input.Clientset.RestConfig)
	if err != nil {
		return fmt.Errorf("Error creating corev1 client: %w", err)
	}

	req := coreclient.RESTClient().Post().
		Resource("pods").
		Namespace(input.SelectedPod.Namespace).
		Name(input.SelectedPod.Name).
		SubResource("exec")

	req.Param("container", fmt.Sprintf("%s-%s", input.AppName, input.ProcessType))
	req.Param("stdin", "true")
	req.Param("stdout", "true")
	req.Param("stderr", "true")
	req.Param("tty", "true")

	if input.Entrypoint != "" {
		req.Param("command", input.Entrypoint)
	}
	for _, cmd := range input.Command {
		req.Param("command", cmd)
	}

	t := term.TTY{
		In:  os.Stdin,
		Out: os.Stdout,
		Raw: true,
	}
	size := t.GetSize()
	sizeQueue := t.MonitorSize(size)

	return t.Safe(func() error {
		exec, err := remotecommand.NewSPDYExecutor(&input.Clientset.RestConfig, "POST", req.URL())
		if err != nil {
			return fmt.Errorf("Error creating executor: %w", err)
		}

		return exec.StreamWithContext(ctx, remotecommand.StreamOptions{
			Stdin:             os.Stdin,
			Stdout:            os.Stdout,
			Stderr:            os.Stderr,
			Tty:               true,
			TerminalSizeQueue: sizeQueue,
		})
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

	labelSelector := []string{fmt.Sprintf("dokku.com/app-name=%s", appName)}
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
		labelSelector = append(labelSelector, fmt.Sprintf("dokku.com/process-type=%s", processType))
	}

	namespace := common.PropertyGetDefault("scheduler-k3s", appName, "namespace", "default")
	podList, err := clientset.Client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: strings.Join(labelSelector, ","),
	})
	if err != nil {
		return fmt.Errorf("Error listing pods: %w", err)
	}
	if len(podList.Items) == 0 {
		return fmt.Errorf("No pods found for app %s", appName)
	}

	ch := make(chan bool)
	podItems := podList.Items

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
	for i := 0; i < len(podItems); i++ {
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
		podName := podItems[i].Name
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

// TriggerSchedulerStop stops an application
func TriggerSchedulerStop(scheduler string, appName string) error {
	if scheduler != "k3s" {
		return nil
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
		cancel()
	}()

	namespace := common.PropertyGetDefault("scheduler-k3s", appName, "namespace", "default")
	deploymentList, err := clientset.Client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("dokku.com/app-name=%s", appName),
	})
	if err != nil {
		return fmt.Errorf("Error listing deployments: %w", err)
	}

	for _, deployment := range deploymentList.Items {
		processType, ok := deployment.Annotations["dokku.com/process-type"]
		if !ok {
			return fmt.Errorf("Deployment %s does not have a process type annotation", deployment.Name)
		}
		common.LogVerboseQuiet(fmt.Sprintf("Stopping %s process", processType))
		_, err := clientset.Client.AppsV1().Deployments(namespace).UpdateScale(ctx, deployment.Name, &autoscalingv1.Scale{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deployment.Name,
				Namespace: namespace,
			},
			Spec: autoscalingv1.ScaleSpec{
				Replicas: 0,
			},
		}, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("Error updating deployment scale: %w", err)
		}
	}

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
	imageRepo := common.GetAppImageRepo(appName)
	image, err := common.GetDeployingAppImageName(appName, imageTag, imageRepo)
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
		appRmContainer, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
			Trigger:       "config-get",
			Args:          []string{appName, "DOKKU_RM_CONTAINER"},
			CaptureOutput: true,
			StreamStdio:   false,
		})
		if err != nil {
			globalRmContainer, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
				Trigger:       "config-get-global",
				Args:          []string{"DOKKU_RM_CONTAINER"},
				CaptureOutput: true,
				StreamStdio:   false,
			})
			if err == nil {
				dokkuRmContainer = globalRmContainer.Stdout
			}
		} else {
			dokkuRmContainer = appRmContainer.Stdout
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
		"dokku.com/app-name": appName,
	}

	if os.Getenv("DOKKU_TRACE") == "1" {
		extraEnv["TRACE"] = "true"
	}

	processType := "run"
	if os.Getenv("DOKKU_CRON_ID") != "" {
		processType = "cron"
		labels["dokku.com/cron-id"] = os.Getenv("DOKKU_CRON_ID")
	}

	dockerArgs := []string{}
	if b, err := common.PlugnTriggerSetup("docker-args-run", []string{appName, imageTag}...).SetInput("").Output(); err == nil {
		words, err := shellquote.Split(strings.TrimSpace(string(b[:])))
		if err != nil {
			return err
		}

		dockerArgs = append(dockerArgs, words...)
	}

	imageSourceType, err := common.DockerInspect(image, "{{ index .Config.Labels \"com.dokku.builder-type\" }}")
	if err != nil {
		return fmt.Errorf("Error getting image builder type: %w", err)
	}

	if b, err := common.PlugnTriggerSetup("docker-args-process-run", []string{appName, imageSourceType, imageTag}...).SetInput("").Output(); err == nil {
		words, err := shellquote.Split(strings.TrimSpace(string(b[:])))
		if err != nil {
			return err
		}

		dockerArgs = append(dockerArgs, words...)
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

	namespace := common.PropertyGetDefault("scheduler-k3s", appName, "namespace", "default")
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
	imagePullSecrets := common.PropertyGetDefault("scheduler-k3s", appName, "image-pull-secrets", "")
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

	err = writeResourceToFile(WriteResourceInput{
		Object: &job,
		Path:   "/tmp/job.yaml",
	})
	if err != nil {
		return fmt.Errorf("Error writing job: %w", err)
	}
	common.CatFile("/tmp/job.yaml")

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

	createdJob, err := createKubernetesJob(ctx, job)
	if err != nil {
		return fmt.Errorf("Error creating job: %w", err)
	}

	clientset, err := NewKubernetesClient()
	if err != nil {
		return fmt.Errorf("Error creating kubernetes client: %w", err)
	}

	batchJobSelector := fmt.Sprintf("batch.kubernetes.io/job-name=%s", createdJob.Name)
	podList, err := WaitForPodToExist(ctx, WaitForPodToExistInput{
		Clientset:  clientset,
		Namespace:  namespace,
		RetryCount: 3,
		Selector:   batchJobSelector,
	})
	if err != nil {
		return fmt.Errorf("Error waiting for pod to exist: %w", err)
	}

	for _, pod := range podList.Items {
		common.LogQuiet(pod.Name)
	}

	if !attachToPod {
		return nil
	}

	err = WaitForPodBySelectorRunning(ctx, WaitForPodBySelectorRunningInput{
		Clientset: clientset,
		Namespace: namespace,
		Selector:  batchJobSelector,
		Timeout:   300,
		Waiter:    isPodReady,
	})
	if err != nil {
		if errors.Is(err, conditions.ErrPodCompleted) {
			podList, podErr := GetPod(ctx, GetPodInput{
				Clientset: clientset,
				Namespace: namespace,
				Selector:  batchJobSelector,
			})
			if podErr != nil {
				return fmt.Errorf("Error completed pod: %w", err)
			}
			selectedPod := podList.Items[0]
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

	podList, err = GetPod(ctx, GetPodInput{
		Clientset: clientset,
		Namespace: namespace,
		Selector:  batchJobSelector,
	})
	if err != nil {
		return fmt.Errorf("Error getting pod: %w", err)
	}
	selectedPod := podList.Items[0]

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
		err := enterPod(ctx, EnterPodInput{
			AppName:     appName,
			Clientset:   clientset,
			Command:     command,
			Entrypoint:  entrypoint,
			ProcessType: processType,
			SelectedPod: selectedPod,
		})
		if err != nil {
			return err
		}

		err = clientset.Client.BatchV1().Jobs(namespace).Delete(ctx, createdJob.Name, metav1.DeleteOptions{
			PropagationPolicy: ptr.To(metav1.DeletePropagationForeground),
		})
		if err != nil {
			return fmt.Errorf("Error deleting pod: %w", err)
		}

		return nil
	default:
		return fmt.Errorf("Unable to attach as the pod is in an unknown state: %s", selectedPod.Status.Phase)
	}
	// todo: support scheduler-post-run

	return nil
}

func isPodReady(ctx context.Context, clientset KubernetesClient, podName, namespace string) wait.ConditionWithContextFunc {
	return func(ctx context.Context) (bool, error) {
		fmt.Printf(".") // progress bar!

		pod, err := clientset.Client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		switch pod.Status.Phase {
		case v1.PodRunning:
			return true, nil
		case v1.PodFailed, v1.PodSucceeded:
			return false, conditions.ErrPodCompleted
		}
		return false, nil
	}
}

func isPodComplete(ctx context.Context, clientset KubernetesClient, podName, namespace string) wait.ConditionWithContextFunc {
	return func(ctx context.Context) (bool, error) {
		fmt.Printf(".") // progress bar!

		pod, err := clientset.Client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		switch pod.Status.Phase {
		case v1.PodFailed, v1.PodSucceeded:
			return true, nil
		}
		return false, nil
	}
}

type WaitForPodBySelectorRunningInput struct {
	Clientset KubernetesClient
	Namespace string
	Selector  string
	Timeout   int
	Waiter    func(ctx context.Context, clientset KubernetesClient, podName, namespace string) wait.ConditionWithContextFunc
}

type WaitForPodToExistInput struct {
	Clientset  KubernetesClient
	Namespace  string
	RetryCount int
	Selector   string
}

func WaitForPodToExist(ctx context.Context, input WaitForPodToExistInput) (v1.PodList, error) {
	var podList v1.PodList
	var err error
	for i := 0; i < input.RetryCount; i++ {
		podList, err = GetPod(ctx, GetPodInput{
			Clientset: input.Clientset,
			Namespace: input.Namespace,
			Selector:  input.Selector,
		})
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		return podList, fmt.Errorf("Error listing pods: %w", err)
	}
	return podList, nil
}

type GetPodInput struct {
	Clientset KubernetesClient
	Namespace string
	Selector  string
}

func GetPod(ctx context.Context, input GetPodInput) (v1.PodList, error) {
	listOptions := metav1.ListOptions{LabelSelector: input.Selector}
	podList, err := input.Clientset.Client.CoreV1().Pods(input.Namespace).List(ctx, listOptions)
	return *podList, err
}

func WaitForPodBySelectorRunning(ctx context.Context, input WaitForPodBySelectorRunningInput) error {
	podList, err := WaitForPodToExist(ctx, WaitForPodToExistInput{
		Clientset:  input.Clientset,
		Namespace:  input.Namespace,
		RetryCount: 3,
		Selector:   input.Selector,
	})
	if err != nil {
		return fmt.Errorf("Error waiting for pod to exist: %w", err)
	}

	if len(podList.Items) == 0 {
		return fmt.Errorf("no pods in %s with selector %s", input.Namespace, input.Selector)
	}

	timeout := time.Duration(input.Timeout) * time.Second
	for _, pod := range podList.Items {
		if err := wait.PollUntilContextTimeout(ctx, time.Second, timeout, false, input.Waiter(ctx, input.Clientset, pod.Name, pod.Namespace)); err != nil {
			print("\n")
			return err
		}
	}
	print("\n")
	return nil
}
