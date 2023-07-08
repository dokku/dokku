package ports

import (
	"fmt"

	"github.com/dokku/dokku/plugins/config"
)

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
