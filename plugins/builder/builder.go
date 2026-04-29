package builder

var (
	// DefaultProperties is a map of all valid builder properties with corresponding default property values
	DefaultProperties = map[string]string{
		"build-dir":    "",
		"detected":     "",
		"selected":     "",
		"skip-cleanup": "",
	}

	// GlobalProperties is a map of all valid global builder properties
	GlobalProperties = map[string]bool{
		"build-dir":    true,
		"selected":     true,
		"skip-cleanup": true,
	}
)
