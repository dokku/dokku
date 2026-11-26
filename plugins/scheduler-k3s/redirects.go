package scheduler_k3s

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// GetAppRedirects loads the proxy:redirect:set redirect set for an app via
// the proxy-redirects-list plugin trigger. The trigger emits one
// "source|destination|code" line per redirect, sorted by source. Returns an
// empty slice when no redirects are configured.
func GetAppRedirects(appName string) ([]ProcessRedirect, error) {
	result, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "proxy-redirects-list",
		Args:    []string{appName},
	})
	if err != nil {
		return nil, fmt.Errorf("error reading proxy redirects for %s: %w", appName, err)
	}

	return parseRedirects(appName, result.StdoutContents())
}

// parseRedirects parses the proxy-redirects-list trigger output
func parseRedirects(appName string, stdout string) ([]ProcessRedirect, error) {
	redirects := []ProcessRedirect{}
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, "|")
		if len(fields) != 3 || fields[0] == "" || fields[1] == "" {
			return nil, fmt.Errorf("invalid redirect line %q for %s", line, appName)
		}
		code, err := strconv.Atoi(fields[2])
		if err != nil {
			return nil, fmt.Errorf("invalid code %q in redirect line %q for %s: %w", fields[2], line, appName, err)
		}
		redirects = append(redirects, ProcessRedirect{
			Source:      fields[0],
			Destination: fields[1],
			Code:        code,
			Permanent:   code == 301 || code == 308,
			Slug:        domainSlug(fields[0]),
		})
	}
	return redirects, nil
}

// filterRedirectSources removes every redirect source from a list of domains
func filterRedirectSources(domains []string, redirects []ProcessRedirect) []string {
	sources := map[string]bool{}
	for _, redirect := range redirects {
		sources[redirect.Source] = true
	}

	filtered := []string{}
	for _, domain := range domains {
		if !sources[domain] {
			filtered = append(filtered, domain)
		}
	}
	return filtered
}
