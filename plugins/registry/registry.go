package registry

var (
	// DefaultProperties is a map of all valid network properties with corresponding default property values
	DefaultProperties = map[string]string{
		"create-repository": "false",
		"image-repo":        "",
		"push-on-release":   "false",
		"server":            "",
	}

	// GlobalProperties is a map of all valid global network properties
	GlobalProperties = map[string]bool{
		"create-repository": true,
		"push-on-release":   true,
		"server":            true,
	}
)
