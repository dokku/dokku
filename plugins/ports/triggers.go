package ports

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
)

// TriggerRawTCPPorts extracts raw tcp port numbers from DOCKERFILE_PORTS config variable
func TriggerPortsDockerfileRawTCPPorts(appName string) error {
	ports := getDockerfileRawTCPPorts(appName)
	for _, port := range ports {
		common.Log(port)
	}

	return nil
}

// TriggerPortsConfigure ensures we have a port mapping
func TriggerPortsConfigure(appName string) error {
	rawTCPPorts := getDockerfileRawTCPPorts(appName)

	b, _ := common.PlugnTriggerOutput("config-get", []string{appName, "DOKKU_PROXY_PORT"}...)
	dokkuProxyPort := strings.TrimSpace(string(b[:]))

	b, _ = common.PlugnTriggerOutput("config-get", []string{appName, "DOKKU_PROXY_SSL_PORT"}...)
	dokkuProxySSLPort := strings.TrimSpace(string(b[:]))

	b, _ = common.PlugnTriggerOutput("config-get", []string{appName, "DOKKU_PROXY_PORT_MAP"}...)
	portMapString := strings.TrimSpace(string(b[:]))

	isAppVhostEnabled := true
	upstreamPort := "5000"

	if err := common.PlugnTrigger("domains-vhost-enabled", []string{appName}...); err != nil {
		isAppVhostEnabled = false
	}

	if dokkuProxyPort == "" && len(rawTCPPorts) == 0 {
		proxyPort := "80"
		if !isAppVhostEnabled {
			common.LogInfo1("No port set, setting to random open high port")
			b, _ = common.PlugnTriggerOutput("ports-get-available", []string{}...)
			proxyPort = strings.TrimSpace(string(b[:]))
		} else {
			b, _ = common.PlugnTriggerOutput("config-get-global", []string{"DOKKU_PROXY_PORT"}...)
			proxyPort = strings.TrimSpace(string(b[:]))
		}

		if proxyPort == "" {
			proxyPort = "80"
		}

		dokkuProxyPort = proxyPort
		err := common.EnvWrap(func() error {
			entries := map[string]string{
				"DOKKU_PROXY_PORT": proxyPort,
			}
			return config.SetMany(appName, entries, false)
		}, map[string]string{"DOKKU_QUIET_OUTPUT": "1"})
		if err != nil {
			return err
		}
	}

	if dokkuProxySSLPort == "" {
		b, _ = common.PlugnTriggerOutput("certs-exists", []string{appName}...)
		certsExists := strings.TrimSpace(string(b[:]))
		if certsExists == "true" {
			b, _ = common.PlugnTriggerOutput("config-get-global", []string{"PROXY_SSL_PORT"}...)
			proxySSLPort := strings.TrimSpace(string(b[:]))
			if proxySSLPort == "" {
				proxySSLPort = "443"
			}

			if len(rawTCPPorts) == 0 && !isAppVhostEnabled {
				common.LogInfo1("No ssl port set, setting to random open high port")
				b, _ = common.PlugnTriggerOutput("ports-get-available", []string{}...)
				proxySSLPort = strings.TrimSpace(string(b[:]))
			}

			dokkuProxySSLPort = proxySSLPort
			err := common.EnvWrap(func() error {
				entries := map[string]string{
					"DOKKU_PROXY_SSL_PORT": proxySSLPort,
				}
				return config.SetMany(appName, entries, false)
			}, map[string]string{"DOKKU_QUIET_OUTPUT": "1"})
			if err != nil {
				return err
			}
		}
	}

	if len(portMapString) == 0 {
		if len(rawTCPPorts) > 0 {
			for _, rawTcpPort := range rawTCPPorts {
				if rawTcpPort == "" {
					continue
				}

				portMapString = fmt.Sprintf("%s http:%s:%s", portMapString, rawTcpPort, rawTcpPort)
			}
		} else {
			portFile := filepath.Join(common.AppRoot(appName), "PORT.web.1")
			if common.FileExists(portFile) {
				upstreamPort = common.ReadFirstLine(portFile)
			}

			if dokkuProxyPort != "" {
				portMapString = fmt.Sprintf("%s http:%s:%s", portMapString, dokkuProxyPort, upstreamPort)
			}
			if dokkuProxySSLPort != "" {
				portMapString = fmt.Sprintf("%s https:%s:%s", portMapString, dokkuProxySSLPort, upstreamPort)
			}
		}

		portMapString = strings.TrimSpace(portMapString)
		if len(portMapString) > 0 {
			portMaps, err := parsePortMapString(portMapString)
			if err != nil {
				return err
			}

			err = common.EnvWrap(func() error {
				return setPortMaps(appName, portMaps)
			}, map[string]string{"DOKKU_QUIET_OUTPUT": "1"})
			if err != nil {
				return err
			}
		}

		return nil
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
