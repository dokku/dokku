package ports

import (
	"fmt"
)

// PortMap is a struct that contains a scheme:host-port:container-port mapping
type PortMap struct {
	ContainerPort int    `json:"container_port"`
	HostPort      int    `json:"host_port"`
	Scheme        string `json:"scheme"`
}

func (p PortMap) String() string {
	return fmt.Sprintf("%s:%d:%d", p.Scheme, p.HostPort, p.ContainerPort)
}

// AllowsPersistence returns true if the port map is not to be persisted
func (p PortMap) AllowsPersistence() bool {
	return p.Scheme == "__internal__"
}
