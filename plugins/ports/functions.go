package ports

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

func addPorts(appName string, portMap []PortMap) error {
	allPortMaps := getPortMaps(appName)
	allPortMaps = append(allPortMaps, portMap...)

	return setPorts(appName, allPortMaps)
}

func clearPorts(appName string) error {
	keys := []string{"DOKKU_PROXY_PORT_MAP"}
	return config.UnsetMany(appName, keys, false)
}

func filterAppPorts(appName string, scheme string, hostPort int) []PortMap {
	var filteredMaps []PortMap
	for _, portMap := range getPortMaps(appName) {
		if portMap.Scheme == scheme && portMap.HostPort == hostPort {
			filteredMaps = append(filteredMaps, portMap)
		}
	}

	return filteredMaps
}

func getPortMaps(appName string) []PortMap {
	value := config.GetWithDefault(appName, "DOKKU_PROXY_PORT_MAP", "")
	portMaps, _ := parsePortMapString(value)
	return portMaps
}

func inRange(value int, min int, max int) bool {
	return min < value && value < max
}

func listAppPorts(appName string) error {
	portMaps := getPortMaps(appName)

	if len(portMaps) == 0 {
		return errors.New("No port mappings configured for app")
	}

	var lines []string
	if os.Getenv("DOKKU_QUIET_OUTPUT") == "" {
		lines = append(lines, "-----> scheme:host port:container port")
	}

	for _, portMap := range portMaps {
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

func parsePortMapString(stringPortMap string) ([]PortMap, error) {
	var portMap []PortMap

	for _, v := range strings.Split(strings.TrimSpace(stringPortMap), " ") {
		parts := strings.SplitN(v, ":", 3)
		if len(parts) == 1 {
			hostPort, err := strconv.Atoi(v)
			if err != nil {
				return portMap, fmt.Errorf("Invalid port map %s [err=%s]", v, err.Error())
			}

			if !inRange(hostPort, 0, 65536) {
				return portMap, fmt.Errorf("Invalid port map %s [hostPort=%d]", v, hostPort)
			}

			portMap = append(portMap, PortMap{
				HostPort: hostPort,
				Scheme:   "__internal__",
			})
			continue
		}

		if len(parts) != 3 {
			return portMap, fmt.Errorf("Invalid port map %s [len=%d]", v, len(parts))
		}

		hostPort, err := strconv.Atoi(parts[1])
		if err != nil {
			return portMap, fmt.Errorf("Invalid port map %s [err=%s]", v, err.Error())
		}

		containerPort, err := strconv.Atoi(parts[2])
		if err != nil {
			return portMap, fmt.Errorf("Invalid port map %s [err=%s]", v, err.Error())
		}

		if !inRange(hostPort, 0, 65536) {
			return portMap, fmt.Errorf("Invalid port map %s [hostPort=%d]", v, hostPort)
		}

		if !inRange(containerPort, 0, 65536) {
			return portMap, fmt.Errorf("Invalid port map %s [containerPort=%d]", v, containerPort)
		}

		portMap = append(portMap, PortMap{
			ContainerPort: containerPort,
			HostPort:      hostPort,
			Scheme:        parts[0],
		})
	}

	return uniquePortMap(portMap), nil
}

func removePorts(appName string, portMap []PortMap) error {
	toRemove := map[string]bool{}
	toRemoveByPort := map[int]bool{}

	for _, portMap := range portMap {
		if portMap.AllowsPersistence() {
			toRemoveByPort[portMap.HostPort] = true
			continue
		}
		toRemove[portMap.String()] = true
	}

	var toSet []PortMap
	for _, portMap := range getPortMaps(appName) {
		if toRemove[portMap.String()] {
			continue
		}

		if toRemoveByPort[portMap.HostPort] {
			continue
		}

		toSet = append(toSet, portMap)
	}

	if len(toSet) == 0 {
		return clearPorts(appName)
	}

	return setPorts(appName, toSet)
}

func setPorts(appName string, portMap []PortMap) error {
	var value []string
	for _, portMap := range uniquePortMap(portMap) {
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

func uniquePortMap(portMap []PortMap) []PortMap {
	var unique []PortMap
	existingPortMaps := map[string]bool{}

	for _, portMap := range portMap {
		if existingPortMaps[portMap.String()] {
			continue
		}

		existingPortMaps[portMap.String()] = true
		unique = append(unique, portMap)
	}

	return unique
}
