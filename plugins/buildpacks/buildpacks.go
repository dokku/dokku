package buildpacks

var (
	// DefaultProperties is a map of all valid buildpacks properties with corresponding default property values
	DefaultProperties = map[string]string{
		"stack": "",
	}

	// GlobalProperties is a map of all valid global buildpacks properties
	GlobalProperties = map[string]bool{
		"stack": true,
	}
)
