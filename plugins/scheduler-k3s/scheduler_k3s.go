package scheduler_k3s

var (
	// DefaultProperties is a map of all valid network properties with corresponding default property values
	DefaultProperties = map[string]string{}

	// GlobalProperties is a map of all valid global network properties
	GlobalProperties = map[string]bool{
		"network-interface": true,
		"token":             true,
	}
)
