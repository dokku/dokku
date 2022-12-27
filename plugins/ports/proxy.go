package ports

import (
	"fmt"
)

// RunInSerial is the default value for whether to run a command in parallel or not
// and defaults to -1 (false)
const RunInSerial = 0

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
