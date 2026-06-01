package scheduler_k3s

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// GetAppRoutes loads the proxy:route:set route set for an app via the
// proxy-routes-list plugin trigger. The trigger emits one
// "process|port|path|strip" line per route, sorted by descending path length.
// Returns an empty slice when no routes are configured.
func GetAppRoutes(appName string) ([]GlobalRoute, error) {
	result, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "proxy-routes-list",
		Args:    []string{appName},
	})
	if err != nil {
		return nil, fmt.Errorf("error reading proxy routes for %s: %w", appName, err)
	}

	stdout := result.StdoutContents()
	if stdout == "" {
		return nil, nil
	}

	routes := []GlobalRoute{}
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.SplitN(line, "|", 4)
		if len(fields) < 3 {
			return nil, fmt.Errorf("invalid route line %q for %s", line, appName)
		}
		port, err := strconv.Atoi(fields[1])
		if err != nil {
			return nil, fmt.Errorf("invalid port %q in route line %q for %s: %w", fields[1], line, appName, err)
		}
		strip := false
		if len(fields) == 4 && fields[3] == "1" {
			strip = true
		}
		routes = append(routes, GlobalRoute{
			Process:     fields[0],
			Path:        fields[2],
			Port:        int32(port),
			StripPrefix: strip,
			Slug:        routeSlug(fields[2]),
		})
	}
	return routes, nil
}

// routeSlug returns a deterministic kebab-case identifier derived from a
// route path. Example: "/api/v0" -> "api-v0".
func routeSlug(path string) string {
	s := strings.Trim(path, "/")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ToLower(s)
	if s == "" {
		return "root"
	}
	return s
}

// RoutedProcessPorts returns, per process type, the set of upstream container
// ports referenced by any route targeting that process. The chart-values
// builder uses this to synthesize port_maps on routed non-web processes so
// their Kubernetes Services expose the right port.
func RoutedProcessPorts(routes []GlobalRoute) map[string]map[int32]bool {
	out := map[string]map[int32]bool{}
	for _, route := range routes {
		if _, ok := out[route.Process]; !ok {
			out[route.Process] = map[int32]bool{}
		}
		out[route.Process][route.Port] = true
	}
	return out
}

// SortedRoutedPorts returns the ports for a process in deterministic order
// so the resulting Service has a stable port list across renders.
func SortedRoutedPorts(routedPorts map[string]map[int32]bool, processType string) []int32 {
	set := routedPorts[processType]
	ports := make([]int32, 0, len(set))
	for port := range set {
		ports = append(ports, port)
	}
	sort.Slice(ports, func(i, j int) bool { return ports[i] < ports[j] })
	return ports
}
