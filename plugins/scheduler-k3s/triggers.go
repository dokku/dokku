package scheduler_k3s

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
	"github.com/rancher/wharfie/pkg/registries"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
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
