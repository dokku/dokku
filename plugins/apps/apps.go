package apps

var (
	// DefaultProperties is a map of all valid apps properties with corresponding default property values
	DefaultProperties = map[string]string{
		"deploy-source":          "",
		"deploy-source-metadata": "",
	}

	// GlobalProperties is a map of all valid global apps properties
	GlobalProperties = map[string]bool{
		"deploy-source":          true,
		"deploy-source-metadata": true,
	}
)
