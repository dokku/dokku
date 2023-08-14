package network

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

var (
	// DefaultProperties is a map of all valid network properties with corresponding default property values
	DefaultProperties = map[string]string{
		"attach-post-create":  "",
		"attach-post-deploy":  "",
		"bind-all-interfaces": "",
		"initial-network":     "",
		"static-web-listener": "",
		"tld":                 "",
	}

	// GlobalProperties is a map of all valid global network properties
	GlobalProperties = map[string]bool{
		"attach-post-create":  true,
		"attach-post-deploy":  true,
		"bind-all-interfaces": true,
		"initial-network":     true,
		"tld":                 true,
	}
)

// BuildConfig builds network config files
func BuildConfig(appName string) error {
	if !common.IsDeployed(appName) {
		return nil
	}

	if staticWebListener := reportStaticWebListener(appName); staticWebListener != "" {
		return nil
	}

	appRoot := common.AppRoot(appName)
	s, err := common.PlugnTriggerOutput("ps-current-scale", []string{appName}...)
	if err != nil {
		return err
	}

	scale, err := common.ParseScaleOutput(s)
	if err != nil {
		return err
	}

	if len(scale) == 0 {
		return nil
	}

	if common.GetAppScheduler(appName) != "docker-local" {
		return nil
	}

	common.LogInfo1(fmt.Sprintf("Ensuring network configuration is in sync for %s", appName))

	for processType, procCount := range scale {
		containerIndex := 0
		for containerIndex < procCount {
			containerIndex++
			containerIndexString := strconv.Itoa(containerIndex)
			containerIDFile := fmt.Sprintf("%v/CONTAINER.%v.%v", appRoot, processType, containerIndex)

			containerID := common.ReadFirstLine(containerIDFile)
			if containerID == "" || !common.ContainerIsRunning(containerID) {
				continue
			}

			ipAddress := GetContainerIpaddress(appName, processType, containerID)
			if ipAddress != "" {
				args := []string{appName, processType, containerIndexString, ipAddress}
				_, err := common.PlugnTriggerOutput("network-write-ipaddr", args...)
				if err != nil {
					common.LogWarn(err.Error())
				}
			}
		}
	}

	return nil
}

// GetContainerIpaddress returns the ipaddr for a given app container
func GetContainerIpaddress(appName, processType, containerID string) (ipAddr string) {
	if processType == "web" {
		if staticWebListener := reportStaticWebListener(appName); staticWebListener != "" {
			ip, _, err := net.SplitHostPort(staticWebListener)
			if err == nil {
				return ip
			}

			ip2 := net.ParseIP(staticWebListener)
			if ip2 != nil {
				return ip2.String()
			}

			return "127.0.0.1"
		}
	}

	if b, err := common.DockerInspect(containerID, "{{ .HostConfig.NetworkMode }}"); err == nil {
		if string(b[:]) == "host" {
			return "127.0.0.1"
		}
	}

	initialNetwork := reportComputedInitialNetwork(appName)
	if initialNetwork == "" {
		initialNetwork = "bridge"
	}

	b, err := common.DockerInspect(containerID, fmt.Sprintf("{{ $network := index .NetworkSettings.Networks \"%s\" }}{{ $network.IPAddress}}", initialNetwork))
	if err != nil || len(b) == 0 {
		// Deprecated: docker < 1.9 compatibility
		b, err = common.DockerInspect(containerID, "{{ .NetworkSettings.IPAddress }}")
	}

	if err == nil {
		return string(b[:])
	}

	return
}

// GetListeners returns a string array of app listeners
func GetListeners(appName string, processType string) []string {
	if processType == "web" {
		if staticWebListener := reportStaticWebListener(appName); staticWebListener != "" {
			return []string{staticWebListener}
		}
	}

	appRoot := common.AppRoot(appName)

	ipPrefix := fmt.Sprintf("/IP.%s.", processType)
	portPrefix := fmt.Sprintf("/PORT.%s.", processType)

	files, _ := filepath.Glob(appRoot + ipPrefix + "*")

	var listeners []string
	for _, ipfile := range files {
		portfile := strings.Replace(ipfile, ipPrefix, portPrefix, 1)
		ipAddress := common.ReadFirstLine(ipfile)
		port := common.ReadFirstLine(portfile)
		if port == "" {
			port = "5000"
		}
		listeners = append(listeners, fmt.Sprintf("%s:%s", ipAddress, port))
	}
	return listeners
}

// HasNetworkConfig returns whether the network configuration for a given app exists
func HasNetworkConfig(appName string) bool {
	appRoot := common.AppRoot(appName)
	ipfile := fmt.Sprintf("%v/IP.web.1", appRoot)
	portfile := fmt.Sprintf("%v/PORT.web.1", appRoot)

	if common.FileExists(ipfile) && common.FileExists(portfile) {
		return true
	}

	return reportStaticWebListener(appName) != ""
}

// ClearNetworkConfig removes old IP and PORT files for a newly cloned app
func ClearNetworkConfig(appName string) bool {
	appRoot := common.AppRoot(appName)
	success := true

	ipFiles, _ := filepath.Glob(appRoot + "/IP.*")
	for _, file := range ipFiles {
		if err := os.Remove(file); err != nil {
			common.LogWarn(fmt.Sprintf("Unable to remove file %s", file))
			success = false
		}
	}
	portFiles, _ := filepath.Glob(appRoot + "/PORT.*")
	for _, file := range portFiles {
		if err := os.Remove(file); err != nil {
			common.LogWarn(fmt.Sprintf("Unable to remove file %s", file))
			success = false
		}
	}
	return success
}
