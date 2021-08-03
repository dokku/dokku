package registry

var (
	// DefaultProperties is a map of all valid network properties with corresponding default property values
	DefaultProperties = map[string]string{
		"image-repo":      "",
		"push-on-release": "false",
		"server":          "",
	}

	// GlobalProperties is a map of all valid global network properties
	GlobalProperties = map[string]bool{
		"push-on-release": true,
		"server":          true,
	}
)
