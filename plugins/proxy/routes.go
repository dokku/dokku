package proxy

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// DefaultRoutePort is the upstream container port assumed when --port is
// omitted on proxy:route:set. It matches the existing default container port
// for the web process.
const DefaultRoutePort = 5000

// Route describes a single path-prefix routing rule.
type Route struct {
	Path        string
	Process     string
	Port        int
	StripPrefix bool
}

// Encode returns the colon-delimited storage representation:
// "<process>:<port>:<strip>" where <strip> is 0 or 1.
func (r Route) Encode() string {
	strip := "0"
	if r.StripPrefix {
		strip = "1"
	}
	return fmt.Sprintf("%s:%d:%s", r.Process, r.Port, strip)
}

// Slug returns a deterministic kebab-case identifier derived from the route's
// path. Used as a stable name fragment for generated labels and CRDs.
// Example: "/api/v0" -> "api-v0".
func (r Route) Slug() string {
	s := strings.Trim(r.Path, "/")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ToLower(s)
	if s == "" {
		return "root"
	}
	return s
}

// DecodeRoute parses the storage representation into a Route. The third
// (strip) field is optional and defaults to 0 (no strip) when absent.
func DecodeRoute(path string, value string) (Route, error) {
	parts := strings.SplitN(value, ":", 3)
	if len(parts) < 2 {
		return Route{}, fmt.Errorf("invalid route value %q for path %s", value, path)
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return Route{}, fmt.Errorf("invalid port %q for path %s: %v", parts[1], path, err)
	}
	strip := false
	if len(parts) == 3 && parts[2] == "1" {
		strip = true
	}
	return Route{
		Path:        path,
		Process:     parts[0],
		Port:        port,
		StripPrefix: strip,
	}, nil
}

var processNameRE = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// ValidatePath enforces the path constraints documented in the proposal:
// - must start with "/"
// - cannot be "/" (reserved for web)
// - cannot end with "/" (other than root, already rejected above)
func ValidatePath(path string) error {
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path must start with /")
	}
	if path == "/" {
		return fmt.Errorf("path / is reserved for the web process")
	}
	if strings.HasSuffix(path, "/") {
		return fmt.Errorf("path must not end with /")
	}
	return nil
}

// ValidateProcess enforces process-name constraints:
// - cannot be empty
// - cannot be "web" (which is the implicit catch-all)
// - must match Procfile name syntax
func ValidateProcess(process string) error {
	if process == "" {
		return fmt.Errorf("process name is required")
	}
	if process == "web" {
		return fmt.Errorf("web cannot be a route target; web is the implicit catch-all")
	}
	if !processNameRE.MatchString(process) {
		return fmt.Errorf("process name %q contains invalid characters", process)
	}
	return nil
}

// ValidatePort ensures the port is in the legal TCP range.
func ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("port %d out of range (1-65535)", port)
	}
	return nil
}

// GetRoutes returns the route set for an app, sorted by descending path
// length (longest-prefix-first), then alphabetically as a tiebreaker.
func GetRoutes(appName string) ([]Route, error) {
	m, err := common.PropertyMapGet("proxy", appName, "routes")
	if err != nil {
		return nil, err
	}
	routes := make([]Route, 0, len(m))
	for path, value := range m {
		route, err := DecodeRoute(path, value)
		if err != nil {
			return nil, err
		}
		routes = append(routes, route)
	}
	sort.SliceStable(routes, func(i, j int) bool {
		if len(routes[i].Path) != len(routes[j].Path) {
			return len(routes[i].Path) > len(routes[j].Path)
		}
		return routes[i].Path < routes[j].Path
	})
	return routes, nil
}

// GetRoutesByProcess returns the subset of routes targeting a specific
// process, preserving the longest-prefix-first ordering.
func GetRoutesByProcess(appName string, processName string) ([]Route, error) {
	all, err := GetRoutes(appName)
	if err != nil {
		return nil, err
	}
	filtered := make([]Route, 0, len(all))
	for _, route := range all {
		if route.Process == processName {
			filtered = append(filtered, route)
		}
	}
	return filtered, nil
}

// RoutedProcessTypes returns the unique process types referenced by any route.
func RoutedProcessTypes(appName string) ([]string, error) {
	routes, err := GetRoutes(appName)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	out := []string{}
	for _, route := range routes {
		if !seen[route.Process] {
			seen[route.Process] = true
			out = append(out, route.Process)
		}
	}
	sort.Strings(out)
	return out, nil
}

// SupportsRoutes returns true if the proxy backend selected for the app
// supports path-based routing, as declared by the proxy-supports-routes
// trigger. Backends that do not implement the trigger are treated as not
// supporting routes (conservative default).
func SupportsRoutes(appName string) (bool, error) {
	result, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "proxy-supports-routes",
		Args:    []string{appName},
	})
	if err != nil {
		return false, err
	}
	return result.StdoutContents() == "true", nil
}

// AddRoute upserts a route. Returns whether the stored state changed.
func AddRoute(appName string, route Route) (bool, error) {
	if err := ValidatePath(route.Path); err != nil {
		return false, err
	}
	if err := ValidateProcess(route.Process); err != nil {
		return false, err
	}
	if err := ValidatePort(route.Port); err != nil {
		return false, err
	}

	existing, err := common.PropertyMapGet("proxy", appName, "routes")
	if err != nil {
		return false, err
	}
	next := route.Encode()
	if existing[route.Path] == next {
		return false, nil
	}
	return true, common.PropertyMapSet("proxy", appName, "routes", route.Path, next)
}

// RemoveRoute removes a route by path. Returns whether the stored state
// changed (false when the path was absent - removal is idempotent).
func RemoveRoute(appName string, path string) (bool, error) {
	existing, err := common.PropertyMapGet("proxy", appName, "routes")
	if err != nil {
		return false, err
	}
	if _, ok := existing[path]; !ok {
		return false, nil
	}
	return true, common.PropertyMapDelete("proxy", appName, "routes", path)
}

// ClearRoutes removes every route for an app. Returns whether anything was
// removed (false when the route set was already empty).
func ClearRoutes(appName string) (bool, error) {
	existing, err := common.PropertyMapGet("proxy", appName, "routes")
	if err != nil {
		return false, err
	}
	if len(existing) == 0 {
		return false, nil
	}
	return true, common.PropertyDelete("proxy", appName, "routes")
}

// getProcessScale returns the configured replica count for a process, or 0
// when no scale entry exists. The ps plugin stores scale as a property list
// of "<proc>=<count>" entries.
func getProcessScale(appName string, processName string) int {
	lines, err := common.PropertyListGet("ps", appName, "scale")
	if err != nil {
		return 0
	}
	prefix := processName + "="
	for _, line := range lines {
		if strings.HasPrefix(line, prefix) {
			n, err := strconv.Atoi(strings.TrimPrefix(line, prefix))
			if err != nil {
				return 0
			}
			return n
		}
	}
	return 0
}
