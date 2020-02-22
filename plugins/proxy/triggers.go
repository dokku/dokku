package proxy

import (
	"fmt"

	"github.com/dokku/dokku/plugins/config"
)

func TriggerProxyIsEnabled(appName string) error {
	if IsAppProxyEnabled(appName) {
		fmt.Println("true")
	} else {
		fmt.Println("false")
	}

	return nil
}

func TriggerProxyType(appName string) error {
	proxyType := getAppProxyType(appName)
	fmt.Println(proxyType)

	return nil
}

func TriggerPostCertsRemove(appName string) error {
	keys := []string{"DOKKU_PROXY_SSL_PORT"}
	if err := config.UnsetMany(appName, keys, false); err != nil {
		return err
	}

	return removeProxyPorts(appName, filterAppProxyPorts(appName, "https", 443))
}

// TriggerPostCertsUpdate unsets port config vars after SSL cert is added
func TriggerPostCertsUpdate(appName string) error {
	port := config.GetWithDefault(appName, "DOKKU_PROXY_PORT", "")
	sslPort := config.GetWithDefault(appName, "DOKKU_PROXY_SSL_PORT", "")
	proxyPortMap := getProxyPortMap(appName)

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
	for _, portMap := range proxyPortMap {
		if portMap.Scheme == "http" && portMap.HostPort == 80 {
			http80Ports = append(http80Ports, portMap)
		}
	}

	if len(http80Ports) > 0 {
		var https443Ports []PortMap
		for _, portMap := range proxyPortMap {
			if portMap.Scheme == "https" && portMap.HostPort == 443 {
				https443Ports = append(https443Ports, portMap)
			}
		}

		if err := removeProxyPorts(appName, https443Ports); err != nil {
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

		if err := addProxyPorts(appName, toAdd); err != nil {
			return err
		}
	}

	return nil
}
