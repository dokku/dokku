package buildpacks

import "strings"

var (
	// DefaultProperties is a map of all valid buildpacks properties with corresponding default property values
	DefaultProperties = map[string]string{}

	// GlobalProperties is a map of all valid global buildpacks properties
	GlobalProperties = map[string]bool{}
)

// isHerokuishStack reports whether a stack value targets the herokuish builder
func isHerokuishStack(value string) bool {
	return strings.Contains(value, "herokuish")
}

// builderForStack returns the builder plugin name a stack value should be written to
func builderForStack(value string) string {
	if isHerokuishStack(value) {
		return "builder-herokuish"
	}
	return "builder-pack"
}
