package ports

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
)

// TriggerPortsClear removes all ports for the specified app
func TriggerPortsClear(appName string) error {
	return clearPorts(appName)
}

// TriggerPortsConfigure ensures we have a port mapping
func TriggerPortsConfigure(appName string) error {
	rawTCPPorts := getDockerfileRawTCPPorts(appName)

	dokkuProxyPort := getProxyPort(appName)
	dokkuProxySSLPort := getProxySSLPort(appName)
	portMaps := getPortMaps(appName)

	vhostEnabled := isAppVhostEnabled(appName)

	if dokkuProxyPort == 0 && len(rawTCPPorts) == 0 {
		proxyPort := 80
		if !vhostEnabled {
			common.LogInfo1("No port set, setting to random open high port")
			proxyPort = getAvailablePort()
		} else {
			proxyPort = getGlobalProxyPort()
		}

		if proxyPort == 0 {
			proxyPort = 80
		}

		dokkuProxyPort = proxyPort
		err := common.EnvWrap(func() error {
			entries := map[string]string{
				"DOKKU_PROXY_PORT": fmt.Sprint(proxyPort),
			}
			return config.SetMany(appName, entries, false)
		}, map[string]string{"DOKKU_QUIET_OUTPUT": "1"})
		if err != nil {
			return err
		}
	}

	if dokkuProxySSLPort == 0 {
		if doesCertExist(appName) {
			proxySSLPort := getGlobalProxySSLPort()
			if proxySSLPort == 0 {
				proxySSLPort = 443
			}

			if len(rawTCPPorts) == 0 && !vhostEnabled {
				common.LogInfo1("No ssl port set, setting to random open high port")
				proxySSLPort = getAvailablePort()
			}

			dokkuProxySSLPort = proxySSLPort
			err := common.EnvWrap(func() error {
				entries := map[string]string{
					"DOKKU_PROXY_SSL_PORT": fmt.Sprint(proxySSLPort),
				}
				return config.SetMany(appName, entries, false)
			}, map[string]string{"DOKKU_QUIET_OUTPUT": "1"})
			if err != nil {
				return err
			}
		}
	}

	if len(portMaps) == 0 {
		if len(rawTCPPorts) > 0 {
			for _, rawTcpPort := range rawTCPPorts {
				portMaps = append(portMaps, PortMap{
					ContainerPort: rawTcpPort,
					HostPort:      rawTcpPort,
					Scheme:        "http",
				})
			}
		} else {
			upstreamPort := 5000
			portFile := filepath.Join(common.AppRoot(appName), "PORT.web.1")
			if common.FileExists(portFile) {
				if port, err := strconv.Atoi(common.ReadFirstLine(portFile)); err == nil {
					upstreamPort = port
				}
			}

			if dokkuProxyPort != 0 {
				portMaps = append(portMaps, PortMap{
					ContainerPort: upstreamPort,
					HostPort:      dokkuProxyPort,
					Scheme:        "http",
				})
			}
			if dokkuProxySSLPort != 0 {
				portMaps = append(portMaps, PortMap{
					ContainerPort: upstreamPort,
					HostPort:      dokkuProxySSLPort,
					Scheme:        "https",
				})
			}
		}

		if len(portMaps) > 0 {
			return setPortMaps(appName, portMaps)
		}

		return nil
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
