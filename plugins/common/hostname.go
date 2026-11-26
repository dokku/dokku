package common

import (
	"regexp"
	"strings"
)

var hostnameRegex = regexp.MustCompile(`^[a-z0-9.*-]*[a-z0-9*-]$`)

// IsValidHostname returns true if the hostname is a valid dokku vhost. The
// nginx catch-all "_" and wildcard hostnames are both considered valid.
func IsValidHostname(hostname string) bool {
	hostname = strings.ToLower(hostname)
	if hostname == "_" {
		return true
	}

	return hostnameRegex.MatchString(hostname)
}
