package proxy

import (
	"fmt"

	"github.com/dokku/dokku/plugins/config"
)

// PortMap is a struct that contains a scheme:host-port:container-port mapping
type PortMap struct {
	ContainerPort int
	HostPort      int
	Scheme        string
}

func (p PortMap) String() string {
	return fmt.Sprintf("%s:%d:%d", p.Scheme, p.HostPort, p.ContainerPort)
}

// AllowsPersistence returns true if the port map is not to be persisted
func (p PortMap) AllowsPersistence() bool {
	return p.Scheme == "__internal__"
}

// IsAppProxyEnabled returns true if proxy is enabled; otherwise return false
func IsAppProxyEnabled(appName string) bool {
	proxyEnabled := true
	disableProxy := config.GetWithDefault(appName, "DOKKU_DISABLE_PROXY", "")
	if disableProxy != "" {
		proxyEnabled = false
	}
	return proxyEnabled
}
