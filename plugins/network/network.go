package network

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
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
	isHerokuishContainer := common.IsImageHerokuishBased(image, appName)
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
		processType := procParts[0]
		procCount, err := strconv.Atoi(procParts[1])
		if err != nil {
			continue
		}

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
			port := GetContainerPort(appName, processType, isHerokuishContainer, containerID)

			if ipAddress != "" {
				_, err := sh.Command("plugn", "trigger", "network-write-ipaddr", appName, processType, containerIndexString, ipAddress).Output()
				if err != nil {
					common.LogWarn(err.Error())
				}
			}

			if port != "" {
				_, err := sh.Command("plugn", "trigger", "network-write-port", appName, processType, containerIndexString, port).Output()
				if err != nil {
					common.LogWarn(err.Error())
				}
			}
		}
	}
}

// GetContainerIpaddress returns the ipaddr for a given app container
func GetContainerIpaddress(appName, processType, containerID string) (ipAddr string) {
	if processType != "web" {
		return
	}

	b, err := common.DockerInspect(containerID, "'{{.NetworkSettings.Networks.bridge.IPAddress}}'")
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
func GetContainerPort(appName, processType string, isHerokuishContainer bool, containerID string) (port string) {
	if processType != "web" {
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
		cmd := sh.Command(common.DockerBin(), "container", "port", containerID, port)
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

// PostAppCloneSetup removes old IP and PORT files for a newly cloned app
func PostAppCloneSetup(appName string) bool {
	dokkuRoot := common.MustGetEnv("DOKKU_ROOT")
	appRoot := strings.Join([]string{dokkuRoot, appName}, "/")
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

// ReportSingleApp is an internal function that displays the app report for one or more apps
func ReportSingleApp(appName, infoFlag string) {
	if err := common.VerifyAppName(appName); err != nil {
		common.LogFail(err.Error())
	}

	infoFlags := map[string]string{
		"--network-bind-all-interfaces": common.PropertyGet("network", appName, "bind-all-interfaces"),
		"--network-listeners":           strings.Join(GetListeners(appName), " "),
	}

	if len(infoFlag) == 0 {
		common.LogInfo2Quiet(fmt.Sprintf("%s network information", appName))
		for k, v := range infoFlags {
			key := common.UcFirst(strings.Replace(strings.TrimPrefix(k, "--"), "-", " ", -1))
			common.LogVerbose(fmt.Sprintf("%s%s", Right(fmt.Sprintf("%s:", key), 31, " "), v))
		}
		return
	}

	for k, v := range infoFlags {
		if infoFlag == k {
			fmt.Fprintln(os.Stdout, v)
			return
		}
	}

	keys := reflect.ValueOf(infoFlags).MapKeys()
	strkeys := make([]string, len(keys))
	for i := 0; i < len(keys); i++ {
		strkeys[i] = keys[i].String()
	}
	common.LogFail(fmt.Sprintf("Invalid flag passed, valid flags: %s", strings.Join(strkeys, ", ")))
}

func times(str string, n int) (out string) {
	for i := 0; i < n; i++ {
		out += str
	}
	return
}

// Right right-pads the string with pad up to len runes
func Right(str string, length int, pad string) string {
	return str + times(pad, length-len(str))
}
