package scheduler

var (
	// DefaultProperties is a map of all valid scheduler properties with corresponding default property values
	DefaultProperties = map[string]string{
		"selected": "docker-local",
	}

	// GlobalProperties is a map of all valid global scheduler properties
	GlobalProperties = map[string]bool{
		"selected": true,
	}
)
