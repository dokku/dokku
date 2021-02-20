package proxy

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
	"github.com/ryanuber/columnize"
)

func addProxyPorts(appName string, proxyPortMap []PortMap) error {
	allPortMaps := getProxyPortMap(appName)
	allPortMaps = append(allPortMaps, proxyPortMap...)

	return setProxyPorts(appName, allPortMaps)
}

func filterAppProxyPorts(appName string, scheme string, hostPort int) []PortMap {
	var filteredProxyMaps []PortMap
	for _, portMap := range getProxyPortMap(appName) {
		if portMap.Scheme == scheme && portMap.HostPort == hostPort {
			filteredProxyMaps = append(filteredProxyMaps, portMap)
		}
	}

	return filteredProxyMaps
}

func getAppProxyType(appName string) string {
	return config.GetWithDefault(appName, "DOKKU_APP_PROXY_TYPE", "nginx")
}

func getProxyPortMap(appName string) []PortMap {
	value := config.GetWithDefault(appName, "DOKKU_PROXY_PORT_MAP", "")
	portMaps, _ := parseProxyPortMapString(value)
	return portMaps
}

func inRange(value int, min int, max int) bool {
	return min < value && value < max
}

func listAppProxyPorts(appName string) error {
	proxyPortMap := getProxyPortMap(appName)

	if len(proxyPortMap) == 0 {
		return errors.New("No port mappings configured for app")
	}

	var lines []string
	if os.Getenv("DOKKU_QUIET_OUTPUT") == "" {
		lines = append(lines, "-----> scheme:host port:container port")
	}

	for _, portMap := range proxyPortMap {
		lines = append(lines, portMap.String())
	}

	sort.Strings(lines)
	common.LogInfo1Quiet(fmt.Sprintf("Port mappings for %s", appName))
	config := columnize.DefaultConfig()
	config.Delim = ":"
	config.Prefix = "    "
	config.Empty = ""
	fmt.Println(columnize.Format(lines, config))
	return nil
}

func parseProxyPortMapString(stringPortMap string) ([]PortMap, error) {
	var proxyPortMap []PortMap

	for _, v := range strings.Split(strings.TrimSpace(stringPortMap), " ") {
		parts := strings.SplitN(v, ":", 3)
		if len(parts) == 1 {
			hostPort, err := strconv.Atoi(v)
			if err != nil {
				return proxyPortMap, fmt.Errorf("Invalid port map %s [err=%s]", v, err.Error())
			}

			if !inRange(hostPort, 0, 65536) {
				return proxyPortMap, fmt.Errorf("Invalid port map %s [hostPort=%d]", v, hostPort)
			}

			proxyPortMap = append(proxyPortMap, PortMap{
				HostPort: hostPort,
				Scheme:   "__internal__",
			})
			continue
		}

		if len(parts) != 3 {
			return proxyPortMap, fmt.Errorf("Invalid port map %s [len=%d]", v, len(parts))
		}

		hostPort, err := strconv.Atoi(parts[1])
		if err != nil {
			return proxyPortMap, fmt.Errorf("Invalid port map %s [err=%s]", v, err.Error())
		}

		containerPort, err := strconv.Atoi(parts[2])
		if err != nil {
			return proxyPortMap, fmt.Errorf("Invalid port map %s [err=%s]", v, err.Error())
		}

		if !inRange(hostPort, 0, 65536) {
			return proxyPortMap, fmt.Errorf("Invalid port map %s [hostPort=%d]", v, hostPort)
		}

		if !inRange(containerPort, 0, 65536) {
			return proxyPortMap, fmt.Errorf("Invalid port map %s [containerPort=%d]", v, containerPort)
		}

		proxyPortMap = append(proxyPortMap, PortMap{
			ContainerPort: containerPort,
			HostPort:      hostPort,
			Scheme:        parts[0],
		})
	}

	return uniqueProxyPortMap(proxyPortMap), nil
}

func removeProxyPorts(appName string, proxyPortMap []PortMap) error {
	toRemove := map[string]bool{}
	toRemoveByPort := map[int]bool{}

	for _, portMap := range proxyPortMap {
		if portMap.AllowsPersistence() {
			toRemoveByPort[portMap.HostPort] = true
			continue
		}
		toRemove[portMap.String()] = true
	}

	var toSet []PortMap
	for _, portMap := range getProxyPortMap(appName) {
		if toRemove[portMap.String()] {
			continue
		}

		if toRemoveByPort[portMap.HostPort] {
			continue
		}

		toSet = append(toSet, portMap)
	}

	if len(toSet) == 0 {
		keys := []string{"DOKKU_PROXY_PORT_MAP"}
		return config.UnsetMany(appName, keys, false)
	}

	return setProxyPorts(appName, toSet)
}

func setProxyPorts(appName string, proxyPortMap []PortMap) error {
	var value []string
	for _, portMap := range uniqueProxyPortMap(proxyPortMap) {
		if portMap.AllowsPersistence() {
			continue
		}

		value = append(value, portMap.String())
	}

	sort.Strings(value)
	entries := map[string]string{
		"DOKKU_PROXY_PORT_MAP": strings.Join(value, " "),
	}
	return config.SetMany(appName, entries, false)
}

func uniqueProxyPortMap(proxyPortMap []PortMap) []PortMap {
	var unique []PortMap
	existingPortMaps := map[string]bool{}

	for _, portMap := range proxyPortMap {
		if existingPortMaps[portMap.String()] {
			continue
		}

		existingPortMaps[portMap.String()] = true
		unique = append(unique, portMap)
	}

	return unique
}
