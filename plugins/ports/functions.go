package ports

import (
	"errors"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
	"github.com/ryanuber/columnize"
)

func addPortMaps(appName string, portMaps []PortMap) error {
	allPortMaps := getPortMaps(appName)
	allPortMaps = append(allPortMaps, portMaps...)

	return setPortMaps(appName, allPortMaps)
}

func clearPorts(appName string) error {
	return common.EnvWrap(func() error {
		keys := []string{"DOKKU_PROXY_PORT_MAP"}
		return config.UnsetMany(appName, keys, false)
	}, map[string]string{"DOKKU_QUIET_OUTPUT": "1"})
}

func doesCertExist(appName string) bool {
	b, _ := common.PlugnTriggerOutput("certs-exists", []string{appName}...)
	certsExists := strings.TrimSpace(string(b[:]))
	return certsExists == "true"
}

func filterAppPortMaps(appName string, scheme string, hostPort int) []PortMap {
	var filteredPortMaps []PortMap
	for _, portMap := range getPortMaps(appName) {
		if portMap.Scheme == scheme && portMap.HostPort == hostPort {
			filteredPortMaps = append(filteredPortMaps, portMap)
		}
	}

	return filteredPortMaps
}

func getAvailablePort() int {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0
	}

	for {
		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			return 0
		}
		defer l.Close()

		port := l.Addr().(*net.TCPAddr).Port
		if port >= 1025 && port <= 65535 {
			return port
		}
	}
}

func getDockerfileRawTCPPorts(appName string) []int {
	b, _ := common.PlugnTriggerOutput("config-get", []string{appName, "DOKKU_DOCKERFILE_PORTS"}...)
	dockerfilePorts := strings.TrimSpace(string(b[:]))

	ports := []int{}
	for _, dockerfilePort := range strings.Split(dockerfilePorts, " ") {
		dockerfilePort = strings.TrimSpace(dockerfilePort)
		if strings.HasSuffix(dockerfilePort, "/udp") {
			continue
		}

		dockerfilePort = strings.TrimSuffix(dockerfilePort, "/tcp")
		if dockerfilePort == "" {
			continue
		}

		port, err := strconv.Atoi(dockerfilePort)
		if err != nil {
			continue
		}

		if port == 0 {
			continue
		}

		ports = append(ports, port)
	}

	return ports
}

func getGlobalProxyPort() int {
	port := 0
	b, _ := common.PlugnTriggerOutput("config-get-global", []string{"DOKKU_PROXY_PORT"}...)
	if intVar, err := strconv.Atoi(strings.TrimSpace(string(b[:]))); err == nil {
		port = intVar
	}

	return port
}

func getGlobalProxySSLPort() int {
	port := 0
	b, _ := common.PlugnTriggerOutput("config-get-global", []string{"DOKKU_PROXY_SSL_PORT"}...)
	if intVar, err := strconv.Atoi(strings.TrimSpace(string(b[:]))); err == nil {
		port = intVar
	}

	return port
}

func getPortMaps(appName string) []PortMap {
	value := config.GetWithDefault(appName, "DOKKU_PROXY_PORT_MAP", "")
	portMaps, _ := parsePortMapString(value)
	return portMaps
}

func getProxyPort(appName string) int {
	port := 0
	b, _ := common.PlugnTriggerOutput("config-get", []string{appName, "DOKKU_PROXY_PORT"}...)
	if intVar, err := strconv.Atoi(strings.TrimSpace(string(b[:]))); err == nil {
		port = intVar
	}

	return port
}

func getProxySSLPort(appName string) int {
	port := 0
	b, _ := common.PlugnTriggerOutput("config-get", []string{appName, "DOKKU_PROXY_SSL_PORT"}...)
	if intVar, err := strconv.Atoi(strings.TrimSpace(string(b[:]))); err == nil {
		port = intVar
	}

	return port
}

func isAppVhostEnabled(appName string) bool {
	if err := common.PlugnTrigger("domains-vhost-enabled", []string{appName}...); err != nil {
		return false
	}
	return true
}

func inRange(value int, min int, max int) bool {
	return min < value && value < max
}

func listAppPortMaps(appName string) error {
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
	var portMaps []PortMap

	for _, v := range strings.Split(strings.TrimSpace(stringPortMap), " ") {
		parts := strings.SplitN(v, ":", 3)
		if len(parts) == 1 {
			hostPort, err := strconv.Atoi(v)
			if err != nil {
				return portMaps, fmt.Errorf("Invalid port map %s [err=%s]", v, err.Error())
			}

			if !inRange(hostPort, 0, 65536) {
				return portMaps, fmt.Errorf("Invalid port map %s [hostPort=%d]", v, hostPort)
			}

			portMaps = append(portMaps, PortMap{
				HostPort: hostPort,
				Scheme:   "__internal__",
			})
			continue
		}

		if len(parts) != 3 {
			return portMaps, fmt.Errorf("Invalid port map %s [len=%d]", v, len(parts))
		}

		hostPort, err := strconv.Atoi(parts[1])
		if err != nil {
			return portMaps, fmt.Errorf("Invalid port map %s [err=%s]", v, err.Error())
		}

		containerPort, err := strconv.Atoi(parts[2])
		if err != nil {
			return portMaps, fmt.Errorf("Invalid port map %s [err=%s]", v, err.Error())
		}

		if !inRange(hostPort, 0, 65536) {
			return portMaps, fmt.Errorf("Invalid port map %s [hostPort=%d]", v, hostPort)
		}

		if !inRange(containerPort, 0, 65536) {
			return portMaps, fmt.Errorf("Invalid port map %s [containerPort=%d]", v, containerPort)
		}

		portMaps = append(portMaps, PortMap{
			ContainerPort: containerPort,
			HostPort:      hostPort,
			Scheme:        parts[0],
		})
	}

	return uniquePortMaps(portMaps), nil
}

func removePortMaps(appName string, portMaps []PortMap) error {
	toRemove := map[string]bool{}
	toRemoveByPort := map[int]bool{}

	for _, portMap := range portMaps {
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
		keys := []string{"DOKKU_PROXY_PORT_MAP"}
		return config.UnsetMany(appName, keys, false)
	}

	return setPortMaps(appName, toSet)
}

func setPortMaps(appName string, portMaps []PortMap) error {
	var value []string
	for _, portMap := range uniquePortMaps(portMaps) {
		if portMap.AllowsPersistence() {
			continue
		}

		value = append(value, portMap.String())
	}

	return common.EnvWrap(func() error {
		sort.Strings(value)
		entries := map[string]string{
			"DOKKU_PROXY_PORT_MAP": strings.Join(value, " "),
		}
		return config.SetMany(appName, entries, false)
	}, map[string]string{"DOKKU_QUIET_OUTPUT": "1"})
}

func uniquePortMaps(portMaps []PortMap) []PortMap {
	var unique []PortMap
	existingPortMaps := map[string]bool{}

	for _, portMap := range portMaps {
		if existingPortMaps[portMap.String()] {
			continue
		}

		existingPortMaps[portMap.String()] = true
		unique = append(unique, portMap)
	}

	return unique
}
