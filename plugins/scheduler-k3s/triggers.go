package scheduler_k3s

import (
	"context"
	"encoding/base64"
	"fmt"
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
	"github.com/rancher/wharfie/pkg/registries"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/util/term"
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
	err = printResource(PrintInput{
		Object:       &secret,
		Path:         filepath.Join(chartDir, "templates/secret.yaml"),
		Name:         secret.Name,
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
		})
		if err != nil {
			return fmt.Errorf("Error templating deployment: %w", err)
		}

		replacements.Set(fmt.Sprintf("replicas: %d", replicaCountPlaceholder), fmt.Sprintf("replicas: {{.Values.processes.%s.replicas}}", processType))
		deployments[processType] = deployment
		err = printResource(PrintInput{
			Object:       &deployment,
			Path:         filepath.Join(chartDir, fmt.Sprintf("templates/deployment-%s.yaml", deployment.Name)),
			Name:         deployment.Name,
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
			Port:      deployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort,
		})

		err := printResource(PrintInput{
			Object:       &service,
			Path:         filepath.Join(chartDir, "templates/service-web.yaml"),
			Name:         service.Name,
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

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	cSignal := make(chan os.Signal, 2)
	signal.Notify(cSignal, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-cSignal
		common.LogWarn("Exiting pod")
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

	coreclient, err := corev1client.NewForConfig(&clientset.RestConfig)
	if err != nil {
		return fmt.Errorf("Error creating corev1 client: %w", err)
	}

	req := coreclient.RESTClient().Post().
		Resource("pods").
		Namespace(selectedPod.Namespace).
		Name(selectedPod.Name).
		SubResource("exec")

	req.Param("container", fmt.Sprintf("%s-%s", appName, processType))
	req.Param("stdin", "true")
	req.Param("stdout", "true")
	req.Param("stderr", "true")
	req.Param("tty", "true")

	if selectedPod.Annotations["dokku.com/builder-type"] == "herokuish" {
		req.Param("command", "/exec")
	}
	for _, cmd := range command {
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
		exec, err := remotecommand.NewSPDYExecutor(&clientset.RestConfig, "POST", req.URL())
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
