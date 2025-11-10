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

// addPortMaps adds port mappings to an app
func addPortMaps(appName string, portMaps []PortMap) error {
	allPortMaps := getPortMaps(appName)
	allPortMaps = append(allPortMaps, portMaps...)

	return setPortMaps(appName, allPortMaps)
}

// clearPorts clears all port mappings for an app
func clearPorts(appName string) error {
	if err := common.PropertyDelete("ports", appName, "map"); err != nil {
		return err
	}

	return common.PropertyDelete("ports", appName, "map-detected")
}

// doesCertExist checks if a cert exists for an app
func doesCertExist(appName string) bool {
	results, _ := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "certs-exists",
		Args:    []string{appName},
	})
	if results.StdoutContents() == "true" {
		return true
	}

	results, _ = common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "certs-force",
		Args:    []string{appName},
	})
	return results.StdoutContents() == "true"
}

// filterAppPortMaps filters the port mappings for an app
func filterAppPortMaps(appName string, scheme string, hostPort int) []PortMap {
	var filteredPortMaps []PortMap
	for _, portMap := range getPortMaps(appName) {
		if portMap.Scheme == scheme && portMap.HostPort == hostPort {
			filteredPortMaps = append(filteredPortMaps, portMap)
		}
	}

	return filteredPortMaps
}

// getAvailablePort gets an available port
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

// getComputedProxyPort gets the computed proxy port for an app
func getComputedProxyPort(appName string) int {
	port := getProxyPort(appName)
	if port == 0 {
		port = getGlobalProxyPort()
	}

	return port
}

// getComputedProxySSLPort gets the computed proxy ssl port for an app
func getComputedProxySSLPort(appName string) int {
	port := getProxySSLPort(appName)
	if port == 0 {
		port = getGlobalProxySSLPort()
	}

	return port
}

// getDetectedPortMaps gets the detected port mappings for an app
func getDetectedPortMaps(appName string) []PortMap {
	basePort := getComputedProxyPort(appName)
	if basePort == 0 {
		basePort = 80
	}
	defaultMapping := []PortMap{
		{
			ContainerPort: 5000,
			HostPort:      basePort,
			Scheme:        "http",
		},
	}

	portMaps := []PortMap{}
	value, err := common.PropertyListGet("ports", appName, "map-detected")
	if err == nil {
		portMaps, _ = parsePortMapString(strings.Join(value, " "))
	}

	if len(portMaps) == 0 {
		portMaps = defaultMapping
	}

	if doesCertExist(appName) {
		setSSLPort := false
		baseSSLPort := getComputedProxySSLPort(appName)
		if baseSSLPort == 0 {
			baseSSLPort = 443
		}

		for _, portMap := range portMaps {
			if portMap.Scheme != "http" || portMap.HostPort != 80 {
				continue
			}

			setSSLPort = true
			portMaps = append(portMaps, PortMap{
				ContainerPort: portMap.ContainerPort,
				HostPort:      baseSSLPort,
				Scheme:        "https",
			})
		}

		if !setSSLPort {
			for i, portMap := range portMaps {
				if portMap.Scheme != "http" {
					continue
				}

				portMaps[i].Scheme = "https"
			}
		}
	}

	return portMaps
}

// getGlobalProxyPort gets the global proxy port
func getGlobalProxyPort() int {
	port := 0
	results, _ := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "config-get-global",
		Args:    []string{"DOKKU_PROXY_PORT"},
	})
	if intVar, err := strconv.Atoi(results.StdoutContents()); err == nil {
		port = intVar
	}

	return port
}

// getGlobalProxySSLPort gets the global proxy ssl port
func getGlobalProxySSLPort() int {
	port := 0
	results, _ := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "config-get-global",
		Args:    []string{"DOKKU_PROXY_SSL_PORT"},
	})
	if intVar, err := strconv.Atoi(results.StdoutContents()); err == nil {
		port = intVar
	}

	return port
}

// getPortMaps gets the port mappings for an app
func getPortMaps(appName string) []PortMap {
	value, err := common.PropertyListGet("ports", appName, "map")
	if err != nil {
		return []PortMap{}
	}

	portMaps, _ := parsePortMapString(strings.Join(value, " "))
	return portMaps
}

// getProxyPort gets the proxy port for an app
func getProxyPort(appName string) int {
	port := 0
	results, _ := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "config-get",
		Args:    []string{appName, "DOKKU_PROXY_PORT"},
	})
	if intVar, err := strconv.Atoi(results.StdoutContents()); err == nil {
		port = intVar
	}

	return port
}

// getProxySSLPort gets the proxy ssl port for an app
func getProxySSLPort(appName string) int {
	port := 0
	results, _ := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "config-get",
		Args:    []string{appName, "DOKKU_PROXY_SSL_PORT"},
	})
	if intVar, err := strconv.Atoi(results.StdoutContents()); err == nil {
		port = intVar
	}

	return port
}

// initializeProxyPort initializes the proxy port for an app
func initializeProxyPort(appName string) error {
	port := getProxyPort(appName)
	if port != 0 {
		return nil
	}

	if isAppVhostEnabled(appName) {
		port = getGlobalProxyPort()
	} else {
		common.LogInfo1("No port set, setting to random open high port")
		port = getAvailablePort()
		common.LogInfo1(fmt.Sprintf("Random port %d", port))
	}

	if port == 0 {
		port = 80
	}

	if err := setProxyPort(appName, port); err != nil {
		return err
	}
	return nil
}

// initializeProxySSLPort initializes the proxy ssl port for an app
func initializeProxySSLPort(appName string) error {
	port := getProxySSLPort(appName)
	if port != 0 {
		return nil
	}

	if !doesCertExist(appName) {
		return nil
	}

	port = getGlobalProxySSLPort()
	if port == 0 {
		port = 443
	}

	if !isAppVhostEnabled(appName) {
		common.LogInfo1("No ssl port set, setting to random open high port")
		port = getAvailablePort()
	}

	if err := setProxySSLPort(appName, port); err != nil {
		return err
	}

	return nil
}

// inRange checks if a value is within a range
func inRange(value int, min int, max int) bool {
	return min < value && value < max
}

// isAppVhostEnabled checks if the app vhost is enabled
func isAppVhostEnabled(appName string) bool {
	_, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "domains-vhost-enabled",
		Args:        []string{appName},
		StreamStdio: true,
	})
	return err == nil
}

// listAppPortMaps lists the port mappings for an app
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

// parsePortMapString parses a port map string into a slice of PortMap structs
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

	return portMaps, nil
}

// removePortMaps removes specific port mappings from an app
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
		return common.PropertyDelete("ports", appName, "map")
	}

	return setPortMaps(appName, toSet)
}

// reusesSchemeHostPort returns true if the port maps reuse the same scheme:host-port
func reusesSchemeHostPort(portMaps []PortMap) error {
	found := map[string]bool{}

	for _, portMap := range portMaps {
		key := fmt.Sprintf("%s:%d", portMap.Scheme, portMap.HostPort)
		if found[key] {
			return fmt.Errorf("The same scheme:host-port is being reused: %s", key)
		}
		found[key] = true
	}

	return nil
}

// setPortMaps sets the port maps for an app
func setPortMaps(appName string, portMaps []PortMap) error {
	if err := reusesSchemeHostPort(portMaps); err != nil {
		return fmt.Errorf("Error validating port mappings: %s", err)
	}

	var value []string
	for _, portMap := range portMaps {
		if portMap.AllowsPersistence() {
			continue
		}

		value = append(value, portMap.String())
	}

	sort.Strings(value)
	return common.PropertyListWrite("ports", appName, "map", value)
}

// setProxyPort sets the proxy port for an app
func setProxyPort(appName string, port int) error {
	return common.EnvWrap(func() error {
		entries := map[string]string{
			"DOKKU_PROXY_PORT": fmt.Sprint(port),
		}
		return config.SetMany(appName, entries, false, false)
	}, map[string]string{"DOKKU_QUIET_OUTPUT": "1"})
}

// setProxySSLPort sets the proxy ssl port for an app
func setProxySSLPort(appName string, port int) error {
	return common.EnvWrap(func() error {
		entries := map[string]string{
			"DOKKU_PROXY_SSL_PORT": fmt.Sprint(port),
		}
		return config.SetMany(appName, entries, false, false)
	}, map[string]string{"DOKKU_QUIET_OUTPUT": "1"})
}

// uniquePortMaps returns a unique set of port maps
func uniquePortMaps(portMaps []PortMap) []PortMap {
	uniquePortMaps := []PortMap{}
	found := map[string]bool{}

	for _, portMap := range portMaps {
		key := fmt.Sprintf("%s:%d", portMap.Scheme, portMap.HostPort)
		if found[key] {
			continue
		}

		found[key] = true
		uniquePortMaps = append(uniquePortMaps, portMap)
	}

	return uniquePortMaps
}
