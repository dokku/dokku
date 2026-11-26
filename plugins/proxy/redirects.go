package proxy

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// DefaultRedirectCode is the http status code used when --code is omitted on
// proxy:redirect:set.
const DefaultRedirectCode = 301

// validRedirectCodes are the http status codes accepted for a redirect
var validRedirectCodes = map[int]bool{
	301: true,
	302: true,
	303: true,
	307: true,
	308: true,
}

// Redirect describes a single domain redirect rule.
type Redirect struct {
	Source      string
	Destination string
	Code        int
}

// Encode returns the colon-delimited storage representation:
// "<destination>:<code>".
func (r Redirect) Encode() string {
	return fmt.Sprintf("%s:%d", r.Destination, r.Code)
}

// IsPermanent returns true when the redirect code is a permanent redirect
func (r Redirect) IsPermanent() bool {
	return r.Code == 301 || r.Code == 308
}

// DecodeRedirect parses the storage representation into a Redirect. The code
// field is optional and defaults to DefaultRedirectCode when absent.
func DecodeRedirect(source string, value string) (Redirect, error) {
	destination, codeValue, hasCode := strings.Cut(value, ":")
	if destination == "" {
		return Redirect{}, fmt.Errorf("invalid redirect value %q for source %s", value, source)
	}

	code := DefaultRedirectCode
	if hasCode {
		parsed, err := strconv.Atoi(codeValue)
		if err != nil {
			return Redirect{}, fmt.Errorf("invalid code %q for source %s: %v", codeValue, source, err)
		}
		code = parsed
	}

	return Redirect{
		Source:      source,
		Destination: destination,
		Code:        code,
	}, nil
}

// NormalizeRedirectHostname lowercases and trims a redirect hostname
func NormalizeRedirectHostname(hostname string) string {
	return strings.ToLower(strings.TrimSpace(hostname))
}

// ValidateRedirectHostname ensures a redirect source or destination is a plain
// hostname. Wildcards and the nginx catch-all "_" are rejected as not every
// proxy backend can express them as a redirect source.
func ValidateRedirectHostname(kind string, hostname string) error {
	if hostname == "" {
		return fmt.Errorf("%s is required", kind)
	}
	if hostname == "_" {
		return fmt.Errorf("%s %q cannot be the catch-all hostname", kind, hostname)
	}
	if strings.Contains(hostname, "*") {
		return fmt.Errorf("%s %q cannot be a wildcard hostname", kind, hostname)
	}
	if strings.HasPrefix(hostname, ".") || strings.HasPrefix(hostname, "-") || strings.Contains(hostname, "..") {
		return fmt.Errorf("%s %q is not a valid hostname", kind, hostname)
	}
	if !common.IsValidHostname(hostname) {
		return fmt.Errorf("%s %q is not a valid hostname", kind, hostname)
	}
	return nil
}

// ValidateRedirectCode ensures the code is a supported redirect status code
func ValidateRedirectCode(code int) error {
	if !validRedirectCodes[code] {
		return fmt.Errorf("invalid redirect code %d, must be one of 301, 302, 303, 307, 308", code)
	}
	return nil
}

// ValidateRedirect validates every field of a redirect
func ValidateRedirect(redirect Redirect) error {
	if err := ValidateRedirectHostname("source", redirect.Source); err != nil {
		return err
	}
	if err := ValidateRedirectHostname("destination", redirect.Destination); err != nil {
		return err
	}
	if redirect.Source == redirect.Destination {
		return fmt.Errorf("source and destination cannot be the same hostname")
	}
	return ValidateRedirectCode(redirect.Code)
}

// SortRedirects sorts redirects alphabetically by source
func SortRedirects(redirects []Redirect) {
	sort.SliceStable(redirects, func(i, j int) bool {
		return redirects[i].Source < redirects[j].Source
	})
}

// GetRedirects returns the redirect set for an app, sorted by source.
func GetRedirects(appName string) ([]Redirect, error) {
	m, err := common.PropertyMapGet("proxy", appName, "redirects")
	if err != nil {
		return nil, err
	}
	redirects := make([]Redirect, 0, len(m))
	for source, value := range m {
		redirect, err := DecodeRedirect(source, value)
		if err != nil {
			return nil, err
		}
		redirects = append(redirects, redirect)
	}
	SortRedirects(redirects)
	return redirects, nil
}

// AddRedirect upserts a redirect. Returns whether the stored state changed.
func AddRedirect(appName string, redirect Redirect) (bool, error) {
	if err := ValidateRedirect(redirect); err != nil {
		return false, err
	}

	existing, err := common.PropertyMapGet("proxy", appName, "redirects")
	if err != nil {
		return false, err
	}
	next := redirect.Encode()
	if existing[redirect.Source] == next {
		return false, nil
	}
	return true, common.PropertyMapSet("proxy", appName, "redirects", redirect.Source, next)
}

// RemoveRedirect removes a redirect by source. Returns whether the stored
// state changed (false when the source was absent - removal is idempotent).
func RemoveRedirect(appName string, source string) (bool, error) {
	existing, err := common.PropertyMapGet("proxy", appName, "redirects")
	if err != nil {
		return false, err
	}
	if _, ok := existing[source]; !ok {
		return false, nil
	}
	return true, common.PropertyMapDelete("proxy", appName, "redirects", source)
}

// ClearRedirects removes every redirect for an app. Returns whether anything
// was removed (false when the redirect set was already empty).
func ClearRedirects(appName string) (bool, error) {
	existing, err := common.PropertyMapGet("proxy", appName, "redirects")
	if err != nil {
		return false, err
	}
	if len(existing) == 0 {
		return false, nil
	}
	return true, common.PropertyDelete("proxy", appName, "redirects")
}
