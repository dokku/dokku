package network

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
	"github.com/dokku/dokku/plugins/proxy"
)

// TriggerInstall runs the install step for the network plugin
func TriggerInstall() {
	if err := common.PropertySetup("network"); err != nil {
		common.LogFail(fmt.Sprintf("Unable to install the network plugin: %s", err.Error()))
	}

	apps, err := common.DokkuApps()
	if err != nil {
		return
	}
	for _, appName := range apps {
		if common.PropertyExists("network", appName, "bind-all-interfaces") {
			continue
		}
		if proxy.IsAppProxyEnabled(appName) {
			common.LogVerboseQuiet("Setting network property 'bind-all-interfaces' to false")
			if err := common.PropertyWrite("network", appName, "bind-all-interfaces", "false"); err != nil {
				common.LogWarn(err.Error())
			}
		} else {
			common.LogVerboseQuiet("Setting network property 'bind-all-interfaces' to true")
			if err := common.PropertyWrite("network", appName, "bind-all-interfaces", "true"); err != nil {
				common.LogWarn(err.Error())
			}
		}
	}
}

// TriggerNetworkComputePorts computes the ports for a given app container
func TriggerNetworkComputePorts(appName string, processType string, isHerokuishContainer bool) {
	if processType != "web" {
		return
	}

	var dockerfilePorts []string
	if !isHerokuishContainer {
		dokkuDockerfilePorts := strings.Trim(config.GetWithDefault(appName, "DOKKU_DOCKERFILE_PORTS", ""), " ")
		if utf8.RuneCountInString(dokkuDockerfilePorts) > 0 {
			dockerfilePorts = strings.Split(dokkuDockerfilePorts, " ")
		}
	}

	var ports []string
	if len(dockerfilePorts) == 0 {
		ports = append(ports, "5000")
	} else {
		for _, port := range dockerfilePorts {
			port = strings.TrimSuffix(strings.TrimSpace(port), "/tcp")
			if port == "" || strings.HasSuffix(port, "/udp") {
				continue
			}
			ports = append(ports, port)
		}
	}
	fmt.Println(strings.Join(ports, " "))
}

// TriggerNetworkConfigExists writes true or false to stdout whether a given app has network config
func TriggerNetworkConfigExists(appName string) {
	if HasNetworkConfig(appName) {
		fmt.Println("true")
		return
	}

	fmt.Println("false")
}

// TriggerNetworkGetIppaddr writes the ipaddress to stdout for a given app container
func TriggerNetworkGetIppaddr(appName string, processType string, containerID string) {
	ipAddress := GetContainerIpaddress(appName, processType, containerID)
	fmt.Println(ipAddress)
}

// TriggerNetworkGetListeners returns the listeners (host:port combinations) for a given app container
func TriggerNetworkGetListeners(appName string) {
	listeners := GetListeners(appName)
	fmt.Println(strings.Join(listeners, " "))
}

// TriggerNetworkGetPort writes the port to stdout for a given app container
func TriggerNetworkGetPort(appName string, processType string, isHerokuishContainer bool, containerID string) {
	port := GetContainerPort(appName, processType, isHerokuishContainer, containerID)
	fmt.Println(port)
}

// TriggerNetworkGetProperty writes the network property to stdout for a given app container
func TriggerNetworkGetProperty(appName string, property string) {
	defaultValue := GetDefaultValue(property)
	value := common.PropertyGetDefault("network", appName, property, defaultValue)
	fmt.Println(value)
}

// TriggerNetworkWriteIpaddr writes the ip to disk
func TriggerNetworkWriteIpaddr(appName string, processType string, containerIndex string, ip string) {
	if appName == "" {
		common.LogFail("Please specify an app to run the command on")
	}
	err := common.VerifyAppName(appName)
	if err != nil {
		common.LogFail(err.Error())
	}

	appRoot := strings.Join([]string{common.MustGetEnv("DOKKU_ROOT"), appName}, "/")
	filename := fmt.Sprintf("%v/IP.%v.%v", appRoot, processType, containerIndex)
	f, err := os.Create(filename)
	if err != nil {
		common.LogFail(err.Error())
	}
	defer f.Close()

	ipBytes := []byte(ip)
	_, err = f.Write(ipBytes)
	if err != nil {
		common.LogFail(err.Error())
	}
}

// TriggerNetworkWritePort writes the port to disk
func TriggerNetworkWritePort(appName string, processType string, containerIndex string, port string) {
	if appName == "" {
		common.LogFail("Please specify an app to run the command on")
	}
	err := common.VerifyAppName(appName)
	if err != nil {
		common.LogFail(err.Error())
	}

	appRoot := strings.Join([]string{common.MustGetEnv("DOKKU_ROOT"), appName}, "/")
	filename := fmt.Sprintf("%v/PORT.%v.%v", appRoot, processType, containerIndex)
	f, err := os.Create(filename)
	if err != nil {
		common.LogFail(err.Error())
	}
	defer f.Close()

	portBytes := []byte(port)
	_, err = f.Write(portBytes)
	if err != nil {
		common.LogFail(err.Error())
	}
}

// TriggerPostAppCloneSetup creates new network files
func TriggerPostAppCloneSetup(oldAppName string, newAppName string) error {
	success := ClearNetworkConfig(newAppName)
	if !success {
		os.Exit(1)
	}

	err := common.PropertyClone("network", oldAppName, newAppName)
	if err != nil {
		return err
	}

	return nil
}

// TriggerPostAppRenameSetup renames network files
func TriggerPostAppRenameSetup(oldAppName string, newAppName string) error {
	success := ClearNetworkConfig(newAppName)
	if !success {
		os.Exit(1)
	}

	if err := common.PropertyClone("network", oldAppName, newAppName); err != nil {
		return err
	}

	if err := common.PropertyDestroy("network", oldAppName); err != nil {
		return err
	}

	return nil
}

// TriggerPostContainerCreate associates the container with a specified network
func TriggerPostContainerCreate(containerType string, containerID string, appName string, phase string, processType string) {
	if containerType != "app" {
		return
	}

	property := "attach-post-create"
	defaultValue := GetDefaultValue(property)
	networkName := common.PropertyGetDefault("network", appName, property, defaultValue)
	if networkName == "" {
		return
	}

	exists, err := networkExists(networkName)
	if err != nil {
		common.LogFail(err.Error())
	}

	if !exists {
		common.LogFail(fmt.Sprintf("Network %v does not exist", networkName))
	}

	attachAppToNetwork(containerID, networkName, appName, phase, processType)
}

// TriggerPostCreate sets bind-all-interfaces to false by default
func TriggerPostCreate(appName string) {
	err := common.PropertyWrite("network", appName, "bind-all-interfaces", "false")
	if err != nil {
		common.LogWarn(err.Error())
	}
}

// TriggerPostDelete destroys the network property for a given app container
func TriggerPostDelete(appName string) {
	err := common.PropertyDestroy("network", appName)
	if err != nil {
		common.LogFail(err.Error())
	}
}

// TriggerCorePostDeploy associates the container with a specified network
func TriggerCorePostDeploy(appName string) {
	property := "attach-post-deploy"
	defaultValue := GetDefaultValue(property)
	networkName := common.PropertyGetDefault("network", appName, property, defaultValue)
	if networkName == "" {
		return
	}

	common.LogInfo1Quiet(fmt.Sprintf("Associating app with network %s", networkName))
	containerIDs, err := common.GetAppRunningContainerIDs(appName, "")
	if err != nil {
		common.LogFail(err.Error())
	}

	exists, err := networkExists(networkName)
	if err != nil {
		common.LogFail(err.Error())
	}

	if !exists {
		common.LogFail(fmt.Sprintf("Network %v does not exist", networkName))
	}

	for _, containerID := range containerIDs {
		processType, err := common.DockerInspect(containerID, "{{ index .Config.Labels \"com.dokku.process-type\"}}")
		if err != nil {
			common.LogFail(err.Error())
		}
		attachAppToNetwork(containerID, networkName, appName, "deploy", processType)
	}
}
