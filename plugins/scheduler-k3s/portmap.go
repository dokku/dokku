package scheduler_k3s

import (
	"encoding/json"
	"fmt"

	"github.com/dokku/dokku/plugins/common"
)

type PortMap struct {
	ContainerPort int32  `json:"container_port"`
	HostPort      int32  `json:"host_port"`
	Scheme        string `json:"scheme"`
}

func (p PortMap) String() string {
	return fmt.Sprintf("%s-%d-%d", p.Scheme, p.HostPort, p.ContainerPort)
}

func (p PortMap) IsAllowedHttp() bool {
	return p.Scheme == "http" || p.ContainerPort == 80
}

func (p PortMap) IsAllowedHttps() bool {
	return p.Scheme == "https" || p.ContainerPort == 443
}

func getPortMaps(appName string) (map[string]PortMap, error) {
	portMaps := []PortMap{}

	allowedMappings := map[string]PortMap{}
	results, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "ports-get",
		Args:    []string{appName, "json"},
	})
	if err != nil {
		return allowedMappings, err
	}

	err = json.Unmarshal([]byte(results.StdoutContents()), &portMaps)
	if err != nil {
		return allowedMappings, err
	}

	for _, portMap := range portMaps {
		if !portMap.IsAllowedHttp() && !portMap.IsAllowedHttps() {
			// todo: log warning
			continue
		}

		allowedMappings[portMap.String()] = portMap
	}

	return allowedMappings, nil
}
