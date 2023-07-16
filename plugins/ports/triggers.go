package ports

import (
	"fmt"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
)

// TriggerInstall migrates the ports config to properties
func TriggerInstall() error {
	if err := common.PropertySetup("ports"); err != nil {
		return fmt.Errorf("Unable to install the ports plugin: %s", err.Error())
	}

	apps, err := common.UnfilteredDokkuApps()
	if err != nil {
		return nil
	}

	for _, appName := range apps {
		if common.PropertyExists("ports", appName, "map") {
			continue
		}

		common.LogInfo1(fmt.Sprintf("Migrating DOKKU_PROXY_PORT_MAP to ports map property for %s", appName))
		portMapString := config.GetWithDefault(appName, "DOKKU_PROXY_PORT_MAP", "")
		portMaps, _ := parsePortMapString(portMapString)

		propertyValue := []string{}
		for _, portMap := range portMaps {
			propertyValue = append(propertyValue, portMap.String())
		}

		if err := common.PropertyListWrite("ports", appName, "map", propertyValue); err != nil {
			return err
		}

		if err := config.UnsetMany(appName, []string{"DOKKU_PROXY_PORT_MAP"}, false); err != nil {
			return err
		}
	}

	return nil
}

// TriggerPortsClear removes all ports for the specified app
func TriggerPortsClear(appName string) error {
	return clearPorts(appName)
}

// TriggerPortsConfigure ensures we have a port mapping
func TriggerPortsConfigure(appName string) error {
	rawTCPPorts := getDockerfileRawTCPPorts(appName)

	if err := initializeProxyPort(appName, rawTCPPorts); err != nil {
		return err
	}

	if err := initializeProxySSLPort(appName, rawTCPPorts); err != nil {
		return err
	}

	if err := initializePortMap(appName, rawTCPPorts); err != nil {
		return err
	}

	return nil
}

// TriggerRawTCPPorts extracts raw tcp port numbers from DOCKERFILE_PORTS config variable
func TriggerPortsDockerfileRawTCPPorts(appName string) error {
	ports := getDockerfileRawTCPPorts(appName)
	for _, port := range ports {
		common.Log(fmt.Sprint(port))
	}

	return nil
}

// TriggerPortsGet prints out the port mapping for a given app
func TriggerPortsGet(appName string) error {
	for _, portMap := range getPortMaps(appName) {
		if portMap.AllowsPersistence() {
			continue
		}

		fmt.Println(portMap)
	}

	return nil
}

// TriggerPortsGetAvailable prints out an available port greater than 1024
func TriggerPortsGetAvailable() error {
	port := getAvailablePort()
	if port > 0 {
		common.Log(fmt.Sprint(port))
	}

	return nil
}

// TriggerPostAppCloneSetup creates new ports files
func TriggerPostAppCloneSetup(oldAppName string, newAppName string) error {
	err := common.PropertyClone("ports", oldAppName, newAppName)
	if err != nil {
		return err
	}

	return nil
}

// TriggerPostAppRenameSetup renames ports files
func TriggerPostAppRenameSetup(oldAppName string, newAppName string) error {
	if err := common.PropertyClone("ports", oldAppName, newAppName); err != nil {
		return err
	}

	if err := common.PropertyDestroy("ports", oldAppName); err != nil {
		return err
	}

	return nil
}

// TriggerPostCertsRemove unsets port config vars after SSL cert is added
func TriggerPostCertsRemove(appName string) error {
	keys := []string{"DOKKU_PROXY_SSL_PORT"}
	if err := config.UnsetMany(appName, keys, false); err != nil {
		return err
	}

	return removePortMaps(appName, filterAppPortMaps(appName, "https", 443))
}

// TriggerPostCertsUpdate sets port config vars after SSL cert is added
func TriggerPostCertsUpdate(appName string) error {
	port := config.GetWithDefault(appName, "DOKKU_PROXY_PORT", "")
	sslPort := config.GetWithDefault(appName, "DOKKU_PROXY_SSL_PORT", "")
	portMaps := getPortMaps(appName)

	toUnset := []string{}
	if port == "80" {
		toUnset = append(toUnset, "DOKKU_PROXY_PORT")
	}
	if sslPort == "443" {
		toUnset = append(toUnset, "DOKKU_PROXY_SSL_PORT")
	}

	if len(toUnset) > 0 {
		if err := config.UnsetMany(appName, toUnset, false); err != nil {
			return err
		}
	}

	var http80Ports []PortMap
	for _, portMap := range portMaps {
		if portMap.Scheme == "http" && portMap.HostPort == 80 {
			http80Ports = append(http80Ports, portMap)
		}
	}

	if len(http80Ports) > 0 {
		var https443Ports []PortMap
		for _, portMap := range portMaps {
			if portMap.Scheme == "https" && portMap.HostPort == 443 {
				https443Ports = append(https443Ports, portMap)
			}
		}

		if err := removePortMaps(appName, https443Ports); err != nil {
			return err
		}

		var toAdd []PortMap
		for _, portMap := range http80Ports {
			toAdd = append(toAdd, PortMap{
				Scheme:        "https",
				HostPort:      443,
				ContainerPort: portMap.ContainerPort,
			})
		}

		if err := addPortMaps(appName, toAdd); err != nil {
			return err
		}
	}

	return nil
}

// TriggerPostDelete is the ports post-delete plugin trigger
func TriggerPostDelete(appName string) error {
	if err := common.PropertyDestroy("ports", appName); err != nil {
		common.LogWarn(err.Error())
	}

	return nil
}
