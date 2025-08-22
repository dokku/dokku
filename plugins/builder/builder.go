package builder

var (
	// DefaultProperties is a map of all valid builder properties with corresponding default property values
	DefaultProperties = map[string]string{
		"selected":  "",
		"detected":  "",
		"build-dir": "",
	}

	// GlobalProperties is a map of all valid global builder properties
	GlobalProperties = map[string]bool{
		"selected":  true,
		"build-dir": true,
	}
)
