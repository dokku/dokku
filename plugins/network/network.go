package network

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"

	sh "github.com/codeskyblue/go-sh"
)

var (
	// DefaultProperties is a map of all valid network properties with corresponding default property values
	DefaultProperties = map[string]string{
		"bind-all-interfaces": "false",
	}
)

// BuildConfig builds network config files
func BuildConfig(appName string) {
	if err := common.VerifyAppName(appName); err != nil {
		common.LogFail(err.Error())
	}
	if !common.IsDeployed(appName) {
		return
	}
	appRoot := strings.Join([]string{common.MustGetEnv("DOKKU_ROOT"), appName}, "/")
	scaleFile := strings.Join([]string{appRoot, "DOKKU_SCALE"}, "/")
	if !common.FileExists(scaleFile) {
		return
	}

	image := common.GetAppImageName(appName, "", "")
	isHerokuishContainer := common.IsImageHerokuishBased(image)
	common.LogInfo1(fmt.Sprintf("Ensuring network configuration is in sync for %s", appName))
	lines, err := common.FileToSlice(scaleFile)
	if err != nil {
		return
	}
	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		procParts := strings.SplitN(line, "=", 2)
		if len(procParts) != 2 {
			continue
		}
		procType := procParts[0]
		procCount, err := strconv.Atoi(procParts[1])
		if err != nil {
			continue
		}

		containerIndex := 0
		for containerIndex < procCount {
			containerIndex++
			containerIndexString := strconv.Itoa(containerIndex)
			containerIDFile := fmt.Sprintf("%v/CONTAINER.%v.%v", appRoot, procType, containerIndex)

			containerID := common.ReadFirstLine(containerIDFile)
			if containerID == "" || !common.ContainerIsRunning(containerID) {
				continue
			}

			ipAddress := GetContainerIpaddress(appName, procType, containerID)
			port := GetContainerPort(appName, procType, isHerokuishContainer, containerID)

			if ipAddress != "" {
				_, err := sh.Command("plugn", "trigger", "network-write-ipaddr", appName, procType, containerIndexString, ipAddress).Output()
				if err != nil {
					common.LogWarn(err.Error())
				}
			}

			if port != "" {
				_, err := sh.Command("plugn", "trigger", "network-write-port", appName, procType, containerIndexString, port).Output()
				if err != nil {
					common.LogWarn(err.Error())
				}
			}
		}
	}
}

// GetContainerIpaddress returns the ipaddr for a given app container
func GetContainerIpaddress(appName, procType, containerID string) (ipAddr string) {
	if procType != "web" {
		return
	}

	b, err := common.DockerInspect(containerID, "'{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}'")
	if err != nil || len(b) == 0 {
		// docker < 1.9 compatibility
		b, err = common.DockerInspect(containerID, "'{{ .NetworkSettings.IPAddress }}'")
	}

	if err == nil {
		return string(b[:])
	}

	return
}

// GetContainerPort returns the port for a given app container
func GetContainerPort(appName, procType string, isHerokuishContainer bool, containerID string) (port string) {
	if procType != "web" {
		return
	}

	dockerfilePorts := make([]string, 0)
	if !isHerokuishContainer {
		configValue := config.GetWithDefault(appName, "DOKKU_DOCKERFILE_PORTS", "")
		if configValue != "" {
			dockerfilePorts = strings.Split(configValue, " ")
		}
	}

	if len(dockerfilePorts) > 0 {
		for _, p := range dockerfilePorts {
			if strings.HasSuffix(p, "/udp") {
				continue
			}
			port = strings.TrimSuffix(p, "/tcp")
			if port != "" {
				break
			}
		}
		cmd := sh.Command("docker", "port", containerID, port)
		cmd.Stderr = ioutil.Discard
		b, err := cmd.Output()
		if err == nil {
			port = strings.Split(string(b[:]), ":")[1]
		}
	} else {
		port = "5000"
	}

	return
}

// GetDefaultValue returns the default value for a given property
func GetDefaultValue(property string) (value string) {
	value, ok := DefaultProperties[property]
	if ok {
		return
	}
	return
}

// GetListeners returns a string array of app listeners
func GetListeners(appName string) []string {
	dokkuRoot := common.MustGetEnv("DOKKU_ROOT")
	appRoot := strings.Join([]string{dokkuRoot, appName}, "/")

	files, _ := filepath.Glob(appRoot + "/IP.web.*")

	var listeners []string
	for _, ipfile := range files {
		portfile := strings.Replace(ipfile, "/IP.web.", "/PORT.web.", 1)
		ipAddress := common.ReadFirstLine(ipfile)
		port := common.ReadFirstLine(portfile)
		listeners = append(listeners, fmt.Sprintf("%s:%s", ipAddress, port))
	}
	return listeners
}

// HasNetworkConfig returns whether the network configuration for a given app exists
func HasNetworkConfig(appName string) bool {
	appRoot := strings.Join([]string{common.MustGetEnv("DOKKU_ROOT"), appName}, "/")
	ipfile := fmt.Sprintf("%v/IP.web.1", appRoot)
	portfile := fmt.Sprintf("%v/PORT.web.1", appRoot)

	return common.FileExists(ipfile) && common.FileExists(portfile)
}
